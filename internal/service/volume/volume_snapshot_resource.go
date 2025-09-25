// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package volume

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-kakaocloud/internal/common"
	"terraform-provider-kakaocloud/internal/utils"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kakaoenterprise/kc-sdk-go/services/volume"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.ResourceWithConfigure   = &volumeSnapshotResource{}
	_ resource.ResourceWithImportState = &volumeSnapshotResource{}
)

// NewVolumeSnapshotResource is a helper function to simplify the provider implementation.
func NewVolumeSnapshotResource() resource.Resource {
	return &volumeSnapshotResource{}
}

type volumeSnapshotResource struct {
	kc *common.KakaoCloudClient
}

// Metadata returns the resource type name.
func (r *volumeSnapshotResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_volume_snapshot"
}

// Schema defines the schema for the resource.
func (r *volumeSnapshotResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Represents a volume snapshot resource.",
		Attributes: utils.MergeResourceSchemaAttributes(
			volumeSnapshotResourceSchemaAttributes,
			map[string]schema.Attribute{
				"timeouts": timeouts.AttributesAll(ctx),
			},
		),
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *volumeSnapshotResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan volumeSnapshotResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	timeout, diags := plan.Timeouts.Create(ctx, common.DefaultCreateTimeout)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	createReq := volume.CreateVolumeSnapshotModel{
		Name:          plan.Name.ValueString(),
		IsIncremental: plan.IsIncremental.ValueBool(),
	}

	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		createReq.SetDescription(plan.Description.ValueString())
	}

	body := volume.BodyCreateSnapshot{
		Snapshot: createReq,
	}

	ok := r.checkVolumeStatus(ctx, plan.VolumeId.ValueString(), resp)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*volume.ResponseCreateVolumeSnapshotModel, *http.Response, error) {
			return r.kc.ApiClient.VolumeAPI.CreateSnapshot(ctx, plan.VolumeId.ValueString()).
				XAuthToken(r.kc.XAuthToken).BodyCreateSnapshot(body).Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "CreateSnapshot", err, &resp.Diagnostics)
		return
	}

	plan.Id = types.StringValue(respModel.Snapshot.Id)

	result, ok := common.PollUntilResult(
		ctx,
		r,
		2*time.Second,
		[]string{common.VolumeSnapshotStatusAvailable, common.VolumeSnapshotStatusError},
		&resp.Diagnostics,
		func(ctx context.Context) (*volume.BcsVolumeV1ApiGetSnapshotModelVolumeSnapshotModel, *http.Response, error) {
			respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
				func() (*volume.BcsVolumeV1ApiGetSnapshotModelResponseVolumeSnapshotModel, *http.Response, error) {
					return r.kc.ApiClient.VolumeSnapshotAPI.
						GetSnapshot(ctx, plan.Id.ValueString()).
						XAuthToken(r.kc.XAuthToken).
						Execute()
				},
			)
			if err != nil {
				return nil, httpResp, err
			}
			return &respModel.Snapshot, httpResp, nil
		},
		func(v *volume.BcsVolumeV1ApiGetSnapshotModelVolumeSnapshotModel) string {
			return *v.Status.Get()
		},
	)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	common.CheckResourceAvailableStatus(ctx, r, result.Status.Get(), []string{common.VolumeSnapshotStatusAvailable}, &resp.Diagnostics)

	ok = mapVolumeSnapshotBaseModel(&plan.volumeSnapshotBaseModel, result, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

// Read refreshes the Terraform state with the latest data.
func (r *volumeSnapshotResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state volumeSnapshotResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	timeout, diags := state.Timeouts.Read(ctx, common.DefaultReadTimeout)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*volume.BcsVolumeV1ApiGetSnapshotModelResponseVolumeSnapshotModel, *http.Response, error) {
			return r.kc.ApiClient.VolumeSnapshotAPI.GetSnapshot(ctx, state.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).Execute()
		},
	)

	if httpResp != nil && httpResp.StatusCode == 404 {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "GetSnapshot", err, &resp.Diagnostics)
		return
	}

	volumeSnapshotResult := respModel.Snapshot
	ok := mapVolumeSnapshotBaseModel(&state.volumeSnapshotBaseModel, &volumeSnapshotResult, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *volumeSnapshotResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state volumeSnapshotResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	timeout, diags := plan.Timeouts.Update(ctx, common.DefaultUpdateTimeout)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if !plan.Name.Equal(state.Name) || !plan.Description.Equal(state.Description) {
		editReq := volume.EditVolumeSnapshotModel{}
		if !plan.Name.Equal(state.Name) {
			editReq.SetName(plan.Name.ValueString())
		}
		if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
			editReq.SetDescription(plan.Description.ValueString())
		}

		body := *volume.NewBodyUpdateSnapshot(editReq)

		respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
			func() (*volume.BcsVolumeV1ApiUpdateSnapshotModelResponseVolumeSnapshotModel, *http.Response, error) {
				return r.kc.ApiClient.VolumeSnapshotAPI.UpdateSnapshot(ctx, plan.Id.ValueString()).
					XAuthToken(r.kc.XAuthToken).
					BodyUpdateSnapshot(body).
					Execute()
			},
		)
		if err != nil {
			common.AddApiActionError(ctx, r, httpResp, "UpdateSnapshot", err, &resp.Diagnostics)
			return
		}

		state.Name = types.StringValue(respModel.Snapshot.Name)
		state.Description = plan.Description
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *volumeSnapshotResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state volumeSnapshotResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	timeout, diags := state.Timeouts.Delete(ctx, common.DefaultDeleteTimeout)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (interface{}, *http.Response, error) {
			return r.kc.ApiClient.VolumeSnapshotAPI.DeleteSnapshot(ctx, state.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				Execute()
		},
	)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			return
		}
		common.AddApiActionError(ctx, r, httpResp, "DeleteSnapshot", err, &resp.Diagnostics)
		return
	}

	common.PollUntilDeletion(ctx, r, 2*time.Second, &resp.Diagnostics, func(ctx context.Context) (bool, *http.Response, error) {
		_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
			func() (*volume.BcsVolumeV1ApiGetSnapshotModelResponseVolumeSnapshotModel, *http.Response, error) {
				_, httpResp, err := r.kc.ApiClient.VolumeSnapshotAPI.
					GetSnapshot(ctx, state.Id.ValueString()).
					XAuthToken(r.kc.XAuthToken).
					Execute()
				return nil, httpResp, err
			},
		)
		return false, httpResp, err
	})
}

func (r *volumeSnapshotResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Add a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*common.KakaoCloudClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *kakaocloud.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.kc = client
}

func (r *volumeSnapshotResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *volumeSnapshotResource) checkVolumeStatus(
	ctx context.Context,
	volumeId string,
	resp *resource.CreateResponse,
) bool {
	timeout := common.DefaultPollingTimeout
	interval := 1 * time.Second
	_, ok := common.PollUntilResultWithTimeout(
		ctx,
		r,
		interval,
		&timeout,
		[]string{common.VolumeStatusAvailable, common.VolumeStatusInUse},
		&resp.Diagnostics,
		func(ctx context.Context) (*volume.BcsVolumeV1ApiGetVolumeModelVolumeModel, *http.Response, error) {
			respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
				func() (*volume.BcsVolumeV1ApiGetVolumeModelResponseVolumeModel, *http.Response, error) {
					return r.kc.ApiClient.VolumeAPI.
						GetVolume(ctx, volumeId).
						XAuthToken(r.kc.XAuthToken).
						Execute()
				},
			)
			if err != nil {
				return nil, httpResp, err
			}
			return &respModel.Volume, httpResp, nil
		},
		func(v *volume.BcsVolumeV1ApiGetVolumeModelVolumeModel) string {
			return *v.Status.Get()
		},
	)
	if !ok {
		common.AddGeneralError(ctx, r, &resp.Diagnostics,
			fmt.Sprintf("Volume did not reach one of the following states: '%v'.", []string{common.VolumeStatusAvailable, common.VolumeStatusInUse}))
		return false
	}
	if resp.Diagnostics.HasError() {
		return false
	}
	return true
}

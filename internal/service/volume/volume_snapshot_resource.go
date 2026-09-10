// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package volume

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"terraform-provider-kakaocloud/internal/common"
	"terraform-provider-kakaocloud/internal/utils"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kakaoenterprise/kc-sdk-go/v2/services/volume"
)

var (
	_ resource.ResourceWithConfigure   = &volumeSnapshotResource{}
	_ resource.ResourceWithImportState = &volumeSnapshotResource{}
)

func NewVolumeSnapshotResource() resource.Resource {
	return &volumeSnapshotResource{}
}

type volumeSnapshotResource struct {
	kc *common.KakaoCloudClient
}

func (r *volumeSnapshotResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_volume_snapshot"
}

func (r *volumeSnapshotResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: utils.MergeResourceSchemaAttributes(
			volumeSnapshotResourceSchemaAttributes,
			map[string]schema.Attribute{
				"timeouts": timeouts.AttributesAll(ctx),
			},
		),
	}
}

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

	createReq := volume.CreateSnapshot{
		Name:          plan.Name.ValueString(),
		IsIncremental: plan.IsIncremental.ValueBool(),
	}

	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		createReq.SetDescription(plan.Description.ValueString())
	}

	body := *volume.NewCreateSnapshotRequest(createReq)
	_, ok := CheckVolumeStatus(ctx, r.kc, r, plan.VolumeId.ValueString(), StatusesReadyGetOrUpdateForSize, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*volume.CreateSnapshotResponse, *http.Response, error) {
			return r.kc.ApiClient.VolumeSnapshotAPI.CreateSnapshot(ctx, plan.VolumeId.ValueString()).
				XAuthToken(r.kc.XAuthToken).CreateSnapshotRequest(body).Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "CreateSnapshot", err, &resp.Diagnostics)
		return
	}

	plan.Id = types.StringValue(respModel.Snapshot.Id)

	result, ok := CheckVolumeSnapshotStatus(ctx, r.kc, r, plan.Id.ValueString(), false, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Volume snapshot creation status could not be confirmed",
			fmt.Sprintf(
				"Volume snapshot %q was created, but it was not saved to Terraform state because readiness polling did not complete. Check the snapshot in KakaoCloud. If it should be managed by this configuration, run `terraform import <resource-address> %s`. Otherwise, delete it before applying again.",
				plan.Id.ValueString(),
				plan.Id.ValueString(),
			),
		)
		return
	}

	ok = mapVolumeSnapshotModel(&plan.volumeSnapshotBaseModel, result, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

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
		func() (*volume.GetSnapshotResponse, *http.Response, error) {
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
	ok := mapVolumeSnapshotModel(&state.volumeSnapshotBaseModel, &volumeSnapshotResult, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

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

	_, ok := CheckVolumeSnapshotStatus(ctx, r.kc, r, plan.Id.ValueString(), true, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	if !plan.Name.Equal(state.Name) ||
		(!plan.Description.IsNull() && !plan.Description.IsUnknown() && !plan.Description.Equal(state.Description)) {
		editReq := volume.UpdateSnapshot{}
		if !plan.Name.Equal(state.Name) {
			editReq.SetName(plan.Name.ValueString())
			state.Name = plan.Name
		}
		if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
			editReq.SetDescription(plan.Description.ValueString())
			state.Description = plan.Description
		}

		body := *volume.NewUpdateSnapshotRequest(editReq)

		_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
			func() (*volume.UpdateSnapshotResponse, *http.Response, error) {
				return r.kc.ApiClient.VolumeSnapshotAPI.UpdateSnapshot(ctx, plan.Id.ValueString()).
					XAuthToken(r.kc.XAuthToken).
					UpdateSnapshotRequest(body).
					Execute()
			},
		)
		if err != nil {
			common.AddApiActionError(ctx, r, httpResp, "UpdateSnapshot", err, &resp.Diagnostics)
			return
		}
	}

	state.Timeouts = plan.Timeouts
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

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

	snapshotResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*volume.GetSnapshotResponse, *http.Response, error) {
			return r.kc.ApiClient.VolumeSnapshotAPI.GetSnapshot(ctx, state.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				Execute()
		},
	)
	if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
		return
	}
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "GetSnapshot", err, &resp.Diagnostics)
		return
	}

	pollUntilDeleted := func() {
		common.PollUntilDeletion(ctx, r, 2*time.Second, &resp.Diagnostics, func(ctx context.Context) (bool, *http.Response, error) {
			_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
				func() (*volume.GetSnapshotResponse, *http.Response, error) {
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

	status := snapshotResp.Snapshot.Status.Get()
	if status != nil && *status == common.VolumeSnapshotStatusDeleting {
		pollUntilDeleted()
		return
	}

	if status == nil || (*status != common.VolumeSnapshotStatusAvailable && *status != common.VolumeSnapshotStatusError) {
		_, ok := CheckVolumeSnapshotStatus(ctx, r.kc, r, state.Id.ValueString(), true, &resp.Diagnostics)
		if !ok || resp.Diagnostics.HasError() {
			return
		}
	}

	_, httpResp, err = common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
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

	pollUntilDeleted()
}

func (r *volumeSnapshotResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {

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

	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *volumeSnapshotResource) checkVolumeStatus(
	ctx context.Context,
	volumeId string,
	resp *resource.CreateResponse,
) bool {
	interval := 1 * time.Second
	result, ok := common.PollUntilResult(
		ctx,
		r,
		interval,
		"volume",
		volumeId,
		[]string{common.VolumeStatusAvailable, common.VolumeStatusInUse, common.VolumeStatusError, common.VolumeStatusErrorRestore, common.VolumeStatusDeleting},
		&resp.Diagnostics,
		func(ctx context.Context) (*volume.Volume, *http.Response, error) {
			respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
				func() (*volume.GetVolumeResponse, *http.Response, error) {
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
		func(v *volume.Volume) string {
			return *v.Status.Get()
		},
	)
	if !ok {
		for _, d := range resp.Diagnostics.Errors() {
			if strings.Contains(d.Detail(), "context deadline exceeded") {
				common.AddGeneralError(ctx, r, &resp.Diagnostics,
					fmt.Sprintf("Volume did not reach one of the following states: '%v'.", []string{common.VolumeStatusAvailable, common.VolumeStatusInUse}))
				return false
			}
		}
	}
	common.CheckResourceAvailableStatus(ctx, r, result.Status.Get(), []string{common.VolumeStatusAvailable, common.VolumeStatusInUse}, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return false
	}
	return true
}

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package volume

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-kakaocloud/internal/common"
	. "terraform-provider-kakaocloud/internal/utils"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kakaoenterprise/kc-sdk-go/services/volume"
)

var (
	_ resource.ResourceWithConfigure      = &volumeResource{}
	_ resource.ResourceWithImportState    = &volumeResource{}
	_ resource.ResourceWithValidateConfig = &volumeResource{}
)

func NewVolumeResource() resource.Resource {
	return &volumeResource{}
}

type volumeResource struct {
	kc *common.KakaoCloudClient
}

func (r *volumeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_volume"
}

func (r *volumeResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: MergeResourceSchemaAttributes(
			volumeResourceSchemaAttributes,
			map[string]schema.Attribute{
				"timeouts": timeouts.AttributesAll(ctx),
			},
		),
	}
}

func (r *volumeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan volumeResourceModel
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

	if !plan.VolumeSnapshotId.IsNull() && !plan.VolumeSnapshotId.IsUnknown() {
		r.restoreVolumeFromSnapshot(ctx, &plan, resp)
	} else {
		r.createVolumeFromBody(ctx, &plan, resp)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	result, ok := CheckVolumeStatus(ctx, r.kc, r, plan.Id.ValueString(), StatusesReadyGetOrUpdateForSize, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	if !plan.VolumeSnapshotId.IsNull() && !plan.VolumeSnapshotId.IsUnknown() {
		currentSize := result.Size.Get()
		if !plan.Size.IsNull() && !plan.Size.IsUnknown() && plan.Size.ValueInt32() > *currentSize {
			if ok := r.UpdateVolumeSize(ctx, r.kc, plan.Id.ValueString(), plan.Size.ValueInt32(), &resp.Diagnostics); !ok {
				return
			}
			newSize := plan.Size.ValueInt32()
			result.Size.Set(&newSize)
		}
	}

	ok = mapVolumeBaseModel(ctx, &plan.volumeBaseModel, result, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

func (r *volumeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state volumeResourceModel
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
		func() (*volume.BcsVolumeV1ApiGetVolumeModelResponseVolumeModel, *http.Response, error) {
			return r.kc.ApiClient.VolumeAPI.GetVolume(ctx, state.Id.ValueString()).XAuthToken(r.kc.XAuthToken).Execute()
		},
	)

	if httpResp != nil && httpResp.StatusCode == 404 {
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "GetVolume", err, &resp.Diagnostics)
		return
	}

	volumeResult := respModel.Volume
	ok := mapVolumeBaseModel(ctx, &state.volumeBaseModel, &volumeResult, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *volumeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state volumeResourceModel
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

	sizeChanged := !plan.Size.IsNull() &&
		!plan.Size.IsUnknown() &&
		!plan.Size.Equal(state.Size)

	var requiredStatuses []string
	if sizeChanged {
		requiredStatuses = StatusesReadyGetOrUpdateForSize
	} else {
		requiredStatuses = StatusesReadyForNameDesc
	}

	_, ok := CheckVolumeStatus(ctx, r.kc, r, plan.Id.ValueString(), requiredStatuses, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	if plan.Name != state.Name || (!plan.Description.IsUnknown() && plan.Description != state.Description) {
		editReq := volume.EditVolumeModel{
			Name: plan.Name.ValueString(),
		}
		if !plan.Description.IsNull() {
			editReq.SetDescription(plan.Description.ValueString())
		} else {
			editReq.SetDescriptionNil()
		}

		body := *volume.NewBodyUpdateVolume(editReq)

		volumeResult, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
			func() (*volume.BcsVolumeV1ApiUpdateVolumeModelResponseVolumeModel, *http.Response, error) {
				return r.kc.ApiClient.VolumeAPI.UpdateVolume(ctx, state.Id.ValueString()).
					XAuthToken(r.kc.XAuthToken).
					BodyUpdateVolume(body).
					Execute()
			},
		)

		if err != nil {
			common.AddApiActionError(ctx, r, httpResp, "UpdateVolume", err, &resp.Diagnostics)
			return
		}

		state.Name = types.StringValue(volumeResult.Volume.Name)
		state.Description = ConvertNullableString(volumeResult.Volume.Description)
	}

	if !plan.Size.IsNull() && !plan.Size.IsUnknown() && !plan.Size.Equal(state.Size) {
		if ok := r.UpdateVolumeSize(ctx, r.kc, state.Id.ValueString(), plan.Size.ValueInt32(), &resp.Diagnostics); !ok {
			return
		}

		state.Size = plan.Size
	}

	state.Timeouts = plan.Timeouts
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *volumeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state volumeResourceModel
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

	_, ok := CheckVolumeStatus(ctx, r.kc, r, state.Id.ValueString(), StatusesReadyToDelete, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (interface{}, *http.Response, error) {
			return r.kc.ApiClient.VolumeAPI.DeleteVolume(ctx, state.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				Execute()
		},
	)

	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			return
		}
		common.AddApiActionError(ctx, r, httpResp, "DeleteVolume", err, &resp.Diagnostics)
		return
	}

	common.PollUntilDeletion(ctx, r, 2*time.Second, &resp.Diagnostics, func(ctx context.Context) (bool, *http.Response, error) {
		_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
			func() (*volume.BcsVolumeV1ApiGetVolumeModelResponseVolumeModel, *http.Response, error) {
				_, httpResp, err := r.kc.ApiClient.VolumeAPI.
					GetVolume(ctx, state.Id.ValueString()).
					XAuthToken(r.kc.XAuthToken).
					Execute()
				return nil, httpResp, err
			},
		)
		return false, httpResp, err
	})
}

func (r *volumeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {

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

func (r *volumeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {

	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *volumeResource) createVolumeFromBody(ctx context.Context, plan *volumeResourceModel, resp *resource.CreateResponse) {
	createReq := volume.CreateVolumeModel{
		Name:             plan.Name.ValueString(),
		Size:             plan.Size.ValueInt32(),
		AvailabilityZone: volume.AvailabilityZone(plan.AvailabilityZone.ValueString()),
	}

	if !plan.VolumeTypeId.IsNull() {
		createReq.SetVolumeTypeId(plan.VolumeTypeId.ValueString())
	}

	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		createReq.SetDescription(plan.Description.ValueString())
	}

	if !plan.EncryptionSecretId.IsNull() {
		createReq.SetEncryptionSecretId(plan.EncryptionSecretId.ValueString())
	}

	if !plan.ImageId.IsNull() {
		createReq.SetImageId(plan.ImageId.ValueString())
	}

	if !plan.SourceVolumeId.IsNull() {
		createReq.SetSourceVolumeId(plan.SourceVolumeId.ValueString())
	}

	body := volume.BodyCreateVolume{
		Volume: createReq,
	}

	respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*volume.BcsVolumeV1ApiCreateVolumeModelResponseVolumeModel, *http.Response, error) {
			return r.kc.ApiClient.VolumeAPI.CreateVolume(ctx).XAuthToken(r.kc.XAuthToken).BodyCreateVolume(body).Execute()
		},
	)

	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "CreateVolume", err, &resp.Diagnostics)
		return
	}

	plan.Id = types.StringValue(respModel.Volume.Id)
}

func (r *volumeResource) restoreVolumeFromSnapshot(ctx context.Context, plan *volumeResourceModel, resp *resource.CreateResponse) {
	snapshotResult, ok := CheckVolumeSnapshotStatus(ctx, r.kc, r, plan.VolumeSnapshotId.ValueString(), false, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	if !plan.Size.IsNull() && !plan.Size.IsUnknown() {
		snapshotSize := snapshotResult.Size.Get()
		if snapshotSize != nil && int32(*snapshotSize) > plan.Size.ValueInt32() {
			common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
				fmt.Sprintf("Invalid size: cannot be smaller than the snapshot volume size: %d", *snapshotSize))
			return
		}
	}

	restoreReq := volume.RequestRestoreVolumeSnapshotModel{
		Name:             plan.Name.ValueString(),
		AvailabilityZone: volume.AvailabilityZone(plan.AvailabilityZone.ValueString()),
	}

	if !plan.VolumeTypeId.IsNull() {
		restoreReq.SetVolumeTypeId(plan.VolumeTypeId.ValueString())
	}

	body := volume.BodyRestoreSnapshot{
		Restore: restoreReq,
	}

	respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*volume.ResponseRestoreVolumeSnapshotModel, *http.Response, error) {
			return r.kc.ApiClient.VolumeSnapshotAPI.RestoreSnapshot(ctx, plan.VolumeSnapshotId.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				BodyRestoreSnapshot(body).
				Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "RestoreSnapshot", err, &resp.Diagnostics)
		return
	}

	plan.Id = types.StringValue(respModel.Restore.VolumeId)
}

func (r *volumeResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config volumeResourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.validateAvailabilityZoneConfig(config, resp)
	r.validateVolumeConfig(ctx, config, resp)
}

func (r *volumeResource) validateAvailabilityZoneConfig(config volumeResourceModel, resp *resource.ValidateConfigResponse) {
	common.ValidateAvailabilityZone(
		path.Root("availability_zone"),
		config.AvailabilityZone,
		r.kc,
		&resp.Diagnostics,
	)
}

func (r *volumeResource) validateVolumeConfig(ctx context.Context, config volumeResourceModel, resp *resource.ValidateConfigResponse) {

	sourceCount := 0
	if !config.VolumeSnapshotId.IsNull() {
		sourceCount++
	}
	if !config.ImageId.IsNull() {
		sourceCount++
	}
	if !config.SourceVolumeId.IsNull() {
		sourceCount++
	}
	if sourceCount > 1 {
		common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
			"Only one of 'volume_snapshot_id', 'image_id', 'source_volume_id' can be set.")
		return
	}
	if config.VolumeSnapshotId.IsNull() && config.Size.IsNull() {
		common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
			"'size' must be set when 'volume_snapshot_id' is not provided.")
	}
	if !config.VolumeSnapshotId.IsNull() && !(config.Description.IsNull() || config.Description.IsUnknown()) {
		common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
			"'description' is determined by the snapshot, when 'volume_snapshot_id' is provided.")
	}
}

func (r *volumeResource) UpdateVolumeSize(ctx context.Context, kc *common.KakaoCloudClient, volumeId string, newSize int32, diags *diag.Diagnostics) bool {
	body := volume.BodyExtendVolume{
		Volume: volume.ExtendVolumeModel{
			NewSize: newSize,
		},
	}

	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, diags,
		func() (interface{}, *http.Response, error) {
			return kc.ApiClient.VolumeAPI.ExtendVolume(ctx, volumeId).
				XAuthToken(kc.XAuthToken).
				BodyExtendVolume(body).
				Execute()
		},
	)

	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "ExtendVolume", err, diags)
		return false
	}

	return true
}

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package bcs

import (
	"fmt"
	"net/http"
	"sort"
	"terraform-provider-kakaocloud/internal/common"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/kakaoenterprise/kc-sdk-go/v2/services/bcs"
	"github.com/kakaoenterprise/kc-sdk-go/v2/services/volume"
	"golang.org/x/net/context"
)

func (r *instanceResource) setRequestVolumeId(ctx context.Context, plan *instanceResourceModel, respDiags *diag.Diagnostics) {
	if plan.Volumes.IsNull() || plan.Volumes.Elements() == nil {
		return
	}
	needsUpdate := false

	planList, planDiags := r.convertListToInstanceVolumeModel(ctx, plan.Volumes)
	respDiags.Append(planDiags...)

	for _, reqVolume := range planList {
		if reqVolume.Id.IsNull() || reqVolume.Id.IsUnknown() {
			needsUpdate = true
			break
		}
	}
	if !needsUpdate {
		return
	}
	attachedElems := plan.AttachedVolumes.Elements()
	sort.Slice(attachedElems, func(i, j int) bool {
		left, _ := attachedElems[i].(types.Object).Attributes()["mount_point"]
		right, _ := attachedElems[j].(types.Object).Attributes()["mount_point"]
		return left.(types.String).ValueString() < right.(types.String).ValueString()
	})

	for i := range planList {
		if i >= len(attachedElems) {
			break
		}
		attached := attachedElems[i].(types.Object)
		attachedId := attached.Attributes()["id"].(types.String)
		planList[i].Id = attachedId
	}
	convertedList, diags := types.ListValueFrom(ctx, plan.Volumes.ElementType(ctx), planList)
	respDiags.Append(diags...)
	plan.Volumes = convertedList
}

func (r *instanceResource) updateAttachedVolumes(
	ctx context.Context,
	instanceId string,
	plans *[]instanceVolumeModel,
	states *[]instanceAttachedVolumeModel,
	resp *diag.Diagnostics,
) bool {
	stateMap := make(map[string]instanceAttachedVolumeModel)
	rootCount := 0
	for _, s := range *states {
		if s.Id.IsNull() || s.Id.IsUnknown() {
			common.AddGeneralError(ctx, r, resp,
				fmt.Sprintf("Attached volume with unknown ID returned for instance %q", instanceId))
			return false
		}

		stateMap[s.Id.ValueString()] = s
		if !s.IsRoot.IsNull() && !s.IsRoot.IsUnknown() && s.IsRoot.ValueBool() {
			rootCount++
		}
	}
	if rootCount != 1 {
		common.AddGeneralError(ctx, r, resp,
			fmt.Sprintf("VM instance %q must have exactly one root volume, got %d", instanceId, rootCount))
		return false
	}

	planMap := make(map[string]instanceVolumeModel)
	for _, s := range *plans {
		if !s.Id.IsNull() && !s.Id.IsUnknown() {
			planMap[s.Id.ValueString()] = s
		} else {
			common.AddGeneralError(ctx, r, resp, fmt.Sprintf("Unknown volume ID for instance: %v", instanceId))
			return false
		}
	}

	for _, s := range *states {
		volumeID := s.Id.ValueString()
		plan, exists := planMap[volumeID]
		if !exists {
			if !s.IsRoot.IsNull() && !s.IsRoot.IsUnknown() && s.IsRoot.ValueBool() {
				common.AddGeneralError(ctx, r, resp,
					fmt.Sprintf("Root volume %q cannot be detached from instance %q", volumeID, instanceId))
				return false
			}

			ok := r.detachVolume(ctx, instanceId, volumeID, resp)
			if !ok {
				return false
			}
			continue
		}

		if !plan.IsDeleteOnTermination.IsNull() && !plan.IsDeleteOnTermination.IsUnknown() &&
			!plan.IsDeleteOnTermination.Equal(s.IsDeleteOnTermination) {
			if !r.updateAttachedVolumeData(ctx, instanceId, &plan, resp) {
				return false
			}
		}

		if !plan.Size.IsNull() && !plan.Size.IsUnknown() && !plan.Size.Equal(s.Size) {
			if !r.UpdateVolumeSize(ctx, r.kc, volumeID, plan.Size.ValueInt32(), resp) {
				return false
			}
		}
	}

	for _, plan := range *plans {
		_, exists := stateMap[plan.Id.ValueString()]
		if !exists {
			ok := r.attacheVolume(ctx, instanceId, &plan, resp)
			if !ok {
				return false
			}
		}
	}
	return true
}

func (r *instanceResource) updateAttachedVolumeData(
	ctx context.Context,
	instanceId string,
	plan *instanceVolumeModel,
	resp *diag.Diagnostics,
) bool {
	editReq := bcs.UpdateInstanceVolume{}
	editReq.SetIsDeleteOnTermination(plan.IsDeleteOnTermination.ValueBool())

	body := *bcs.NewUpdateInstanceVolumeRequest(editReq)

	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, resp,
		func() (interface{}, *http.Response, error) {
			return r.kc.ApiClient.InstanceAPI.UpdateInstanceVolume(ctx, instanceId, plan.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				UpdateInstanceVolumeRequest(body).
				Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "UpdateInstanceVolume", err, resp)
		return false
	}

	return true
}

func (r *instanceResource) attacheVolume(
	ctx context.Context,
	instanceId string,
	plan *instanceVolumeModel,
	resp *diag.Diagnostics,
) bool {
	editReq := bcs.AttachVolume{}
	if !plan.IsDeleteOnTermination.IsNull() && !plan.IsDeleteOnTermination.IsUnknown() {
		editReq.SetIsDeleteOnTermination(plan.IsDeleteOnTermination.ValueBool())
	} else {
		editReq.SetIsDeleteOnTermination(false)
	}

	body := *bcs.NewAttachVolumeRequest(editReq)

	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, resp,
		func() (*bcs.AttachVolumeResponse, *http.Response, error) {
			return r.kc.ApiClient.InstanceAPI.AttachVolume(ctx, instanceId, plan.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				AttachVolumeRequest(body).
				Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "AttachVolume", err, resp)
		return false
	}

	return r.pollInstanceUntilAllVolumesOk(ctx, instanceId, plan.Id.ValueString(), volumeActionAttach, resp)
}

func (r *instanceResource) detachVolume(
	ctx context.Context,
	instanceId string,
	volumeId string,
	resp *diag.Diagnostics,
) bool {
	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, resp,
		func() (interface{}, *http.Response, error) {
			return r.kc.ApiClient.InstanceAPI.DetachVolume(ctx, instanceId, volumeId).
				XAuthToken(r.kc.XAuthToken).
				Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "DetachVolume", err, resp)
		return false
	}

	return r.pollInstanceUntilAllVolumesOk(ctx, instanceId, volumeId, volumeActionDetach, resp)
}

func (r *instanceResource) UpdateVolumeSize(ctx context.Context, kc *common.KakaoCloudClient, volumeId string, newSize int32, diags *diag.Diagnostics) bool {
	body := *volume.NewExtendVolumeRequest(*volume.NewExtendVolume(newSize))
	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, diags,
		func() (interface{}, *http.Response, error) {
			return kc.ApiClient.VolumeAPI.ExtendVolume(ctx, volumeId).
				XAuthToken(kc.XAuthToken).
				ExtendVolumeRequest(body).
				Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "ExtendVolume", err, diags)
		return false
	}
	return true
}

func (r *instanceResource) convertListToInstanceVolumeModel(
	ctx context.Context,
	list types.List,
) ([]instanceVolumeModel, diag.Diagnostics) {
	var result []instanceVolumeModel
	var diags diag.Diagnostics

	for _, elem := range list.Elements() {
		if obj, ok := elem.(types.Object); ok {
			var model instanceVolumeModel
			elemDiags := obj.As(ctx, &model, basetypes.ObjectAsOptions{})
			diags.Append(elemDiags...)
			result = append(result, model)
		}
	}

	return result, diags
}

const (
	volumeActionAttach = "attach"
	volumeActionDetach = "detach"
)

func (r *instanceResource) pollInstanceUntilAllVolumesOk(
	ctx context.Context,
	instanceID string,
	volumeID string,
	action string,
	diags *diag.Diagnostics,
) bool {
	if action != volumeActionAttach && action != volumeActionDetach {
		diags.AddError(
			"Invalid volume action",
			fmt.Sprintf("Unsupported volume action: %q", action),
		)
		return false
	}

	for {
		respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, diags,
			func() (*bcs.GetInstanceResponse, *http.Response, error) {
				return r.kc.ApiClient.InstanceAPI.
					GetInstance(ctx, instanceID).
					XAuthToken(r.kc.XAuthToken).
					Execute()
			},
		)
		if err != nil {
			common.AddApiActionError(ctx, r, httpResp, "GetInstance", err, diags)
			return false
		}

		found := false
		inUse := false

		for _, attachedVolume := range respModel.Instance.AttachedVolumes {
			attachedVolumeID := attachedVolume.Id.Get()
			if attachedVolumeID == nil || *attachedVolumeID != volumeID {
				continue
			}

			found = true
			status := attachedVolume.Status.Get()
			inUse = status != nil && *status == common.VolumeStatusInUse
			break
		}

		switch action {
		case volumeActionAttach:
			if found && inUse {
				return true
			}
		case volumeActionDetach:
			if !found {
				return true
			}
		}

		select {
		case <-time.After(2 * time.Second):
		case <-ctx.Done():
			diags.AddError(
				"Volume polling canceled",
				fmt.Sprintf(
					"Waiting for volume %q to %s failed: %s",
					volumeID,
					action,
					ctx.Err(),
				),
			)
			return false
		}
	}
}

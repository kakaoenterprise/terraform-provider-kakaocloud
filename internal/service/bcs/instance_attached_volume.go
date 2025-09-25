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
	"github.com/kakaoenterprise/kc-sdk-go/services/bcs"
	"github.com/kakaoenterprise/kc-sdk-go/services/volume"
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
	states *[]instanceVolumeModel,
	resp *diag.Diagnostics,
) bool {
	stateMap := make(map[string]instanceVolumeModel)
	for _, s := range *states {
		if !s.Id.IsNull() && !s.Id.IsUnknown() {
			stateMap[s.Id.ValueString()] = s
		}
	}

	planMap := make(map[string]instanceVolumeModel)
	for _, s := range *plans {
		if !s.Id.IsNull() && !s.Id.IsUnknown() {
			planMap[s.Id.ValueString()] = s
		} else {
			common.AddGeneralError(ctx, r, resp, fmt.Sprintf("Unknown volume Id for instance : %v", instanceId))
			return false
		}
	}

	// Detach or Replace
	for _, s := range *states {
		if _, exists := planMap[s.Id.ValueString()]; !exists {
			ok := r.detachVolume(ctx, instanceId, s.Id.ValueString(), resp)
			if !ok {
				return false
			}

		} else {
			plan := planMap[s.Id.ValueString()]
			// case: IsDeleteOnTermination
			if !plan.IsDeleteOnTermination.IsNull() && !plan.IsDeleteOnTermination.IsUnknown() &&
				!plan.IsDeleteOnTermination.Equal(s.IsDeleteOnTermination) {
				ok := r.updateAttachedVolumeData(ctx, instanceId, &plan, resp)
				if !ok {
					return false
				}
			}
			// case: size
			if !plan.Size.IsNull() && !plan.Size.IsUnknown() && !plan.Size.Equal(s.Size) {
				ok := r.UpdateVolumeSize(ctx, r.kc, plan.Id.ValueString(), plan.Size.ValueInt32(), resp)
				if !ok {
					return false
				}
			}
			// case: TypeId, EncryptionSecretId -> Detach and Attach
			if !plan.TypeId.IsNull() && !plan.TypeId.IsUnknown() && !plan.TypeId.Equal(s.TypeId) ||
				!plan.EncryptionSecretId.IsNull() && !plan.EncryptionSecretId.IsUnknown() && !plan.EncryptionSecretId.Equal(s.EncryptionSecretId) {
				ok := r.detachVolume(ctx, instanceId, s.Id.ValueString(), resp)
				if !ok {
					return false
				}
				ok = r.attacheVolume(ctx, instanceId, &plan, resp)
				if !ok {
					return false
				}
			}
		}
	}

	for _, plan := range *plans {
		_, exists := stateMap[plan.Id.ValueString()]
		// Attach
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
	editReq := bcs.EditVolumeModel{}
	editReq.SetIsDeleteOnTermination(plan.IsDeleteOnTermination.ValueBool())

	body := *bcs.NewBodyUpdateInstanceVolume(editReq)

	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, resp,
		func() (interface{}, *http.Response, error) {
			return r.kc.ApiClient.InstanceAttachedVolumeAPI.UpdateInstanceVolume(ctx, instanceId, plan.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				BodyUpdateInstanceVolume(body).
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
	editReq := bcs.CreateVolumeModel{}
	if !plan.IsDeleteOnTermination.IsNull() && !plan.IsDeleteOnTermination.IsUnknown() {
		editReq.SetIsDeleteOnTermination(plan.IsDeleteOnTermination.ValueBool())
	} else {
		editReq.SetIsDeleteOnTermination(false)
	}

	body := *bcs.NewBodyAttachVolume(editReq)

	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, resp,
		func() (*bcs.InstanceAttachedVolumeModelResponse, *http.Response, error) {
			return r.kc.ApiClient.InstanceAttachedVolumeAPI.AttachVolume(ctx, instanceId, plan.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				BodyAttachVolume(body).
				Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "AttachVolume", err, resp)
		return false
	}

	return r.pollInstanceUntilAllVolumesOk(ctx, instanceId, plan.Id.ValueString(), "attach", resp)
}

func (r *instanceResource) detachVolume(
	ctx context.Context,
	instanceId string,
	volumeId string,
	resp *diag.Diagnostics,
) bool {
	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, resp,
		func() (interface{}, *http.Response, error) {
			return r.kc.ApiClient.InstanceAttachedVolumeAPI.DetachVolume(ctx, instanceId, volumeId).
				XAuthToken(r.kc.XAuthToken).
				Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "DetachVolume", err, resp)
		return false
	}

	return r.pollInstanceUntilAllVolumesOk(ctx, instanceId, volumeId, "detach", resp)
}

func (r *instanceResource) UpdateVolumeSize(ctx context.Context, kc *common.KakaoCloudClient, volumeId string, newSize int32, diags *diag.Diagnostics) bool {
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

func (r *instanceResource) pollInstanceUntilAllVolumesOk(
	ctx context.Context,
	instanceId string,
	volumeId string,
	action string,
	diag *diag.Diagnostics,
) bool {
	for {
		isOk := false
		respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, diag,
			func() (*bcs.ResponseInstanceModel, *http.Response, error) {
				return r.kc.ApiClient.InstanceAPI.
					GetInstance(ctx, instanceId).
					XAuthToken(r.kc.XAuthToken).
					Execute()
			},
		)
		if err != nil {
			common.AddApiActionError(ctx, r, httpResp, "GetInstance", err, diag)
			return false
		}

		for _, attachedVolume := range respModel.Instance.AttachedVolumes {
			if action == "attach" {
				if attachedVolume.Id == volumeId && string(*attachedVolume.Status.Get()) == common.VolumeStatusInUse {
					isOk = true
					break
				}
			} else if action == "detach" {
				isOk = true
				if attachedVolume.Id == volumeId {
					isOk = false
					break
				}
			}
		}

		if isOk {
			return true
		}
		time.Sleep(2 * time.Second)
	}
}

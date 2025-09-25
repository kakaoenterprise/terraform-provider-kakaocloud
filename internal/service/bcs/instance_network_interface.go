// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bcs

import (
	"fmt"
	"net/http"
	"terraform-provider-kakaocloud/internal/common"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/kakaoenterprise/kc-sdk-go/services/bcs"
	"golang.org/x/net/context"
)

func (r *instanceResource) setOneNetworkInterfaceId(ctx context.Context, plan *instanceResourceModel, respDiags *diag.Diagnostics) {
	if plan.Subnets.IsNull() || plan.Subnets.Elements() == nil {
		return
	}

	planList, planDiags := r.convertListToInstanceSubnetModel(ctx, plan.Subnets)
	respDiags.Append(planDiags...)
	if respDiags.HasError() {
		return
	}

	if planList[0].NetworkInterfaceId.IsNull() || planList[0].NetworkInterfaceId.ValueString() == "" {
		address := plan.Addresses.Elements()[0].(types.Object)
		planList[0].NetworkInterfaceId = address.Attributes()["network_interface_id"].(types.String)
	}

	convertedList, diags := types.ListValueFrom(ctx, plan.Subnets.ElementType(ctx), planList)
	respDiags.Append(diags...)
	plan.Subnets = convertedList
}

func (r *instanceResource) updateNetworkInterface(
	ctx context.Context,
	instanceId string,
	plans *[]instanceSubnetModel,
	states *[]instanceSubnetModel,
	resp *diag.Diagnostics,
) bool {
	stateMap := make(map[string]instanceSubnetModel)
	for _, s := range *states {
		if !s.NetworkInterfaceId.IsNull() && !s.NetworkInterfaceId.IsUnknown() {
			stateMap[s.NetworkInterfaceId.ValueString()] = s
		}
	}

	planMap := make(map[string]instanceSubnetModel)
	for _, p := range *plans {
		if !p.NetworkInterfaceId.IsNull() && !p.NetworkInterfaceId.IsUnknown() {
			planMap[p.NetworkInterfaceId.ValueString()] = p
		} else {
			common.AddGeneralError(ctx, r, resp, fmt.Sprintf("Unknown network interface Id for instance : %v", instanceId))
			return false
		}
	}

	// Detach
	for _, s := range *states {
		if _, exists := planMap[s.NetworkInterfaceId.ValueString()]; !exists {
			ok := r.detachNetworkInterface(ctx, instanceId, s.NetworkInterfaceId.ValueString(), resp)
			if !ok {
				return false
			}
		}
	}

	// Attach
	for _, s := range *plans {
		if _, exists := stateMap[s.NetworkInterfaceId.ValueString()]; !exists {
			ok := r.attacheNetworkInterface(ctx, instanceId, s.NetworkInterfaceId.ValueString(), resp)
			if !ok {
				return false
			}
		}
	}
	return true
}

func (r *instanceResource) attacheNetworkInterface(
	ctx context.Context,
	instanceId string,
	networkInterfaceId string,
	resp *diag.Diagnostics,
) bool {
	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, resp,
		func() (*bcs.BcsInstanceV1ApiAttachNetworkInterfaceModelResponseInstanceNetworkInterfaceModel, *http.Response, error) {
			return r.kc.ApiClient.InstanceNetworkInterfaceAPI.AttachNetworkInterface(ctx, instanceId, networkInterfaceId).
				XAuthToken(r.kc.XAuthToken).
				Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "AttachNetworkInterface", err, resp)
		return false
	}
	return r.pollInstanceUntilNetworkInterfaceOk(ctx, instanceId, networkInterfaceId, "attach", resp)
}

func (r *instanceResource) detachNetworkInterface(
	ctx context.Context,
	instanceId string,
	networkInterfaceId string,
	resp *diag.Diagnostics,
) bool {
	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, resp,
		func() (interface{}, *http.Response, error) {
			return r.kc.ApiClient.InstanceNetworkInterfaceAPI.DetachNetworkInterface(ctx, instanceId, networkInterfaceId).
				XAuthToken(r.kc.XAuthToken).
				Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "DetachNetworkInterface", err, resp)
		return false
	}
	return r.pollInstanceUntilNetworkInterfaceOk(ctx, instanceId, networkInterfaceId, "detach", resp)
}

func (r *instanceResource) convertListToInstanceSubnetModel(
	ctx context.Context,
	list types.List,
) ([]instanceSubnetModel, diag.Diagnostics) {
	var result []instanceSubnetModel
	var diags diag.Diagnostics

	for _, elem := range list.Elements() {
		if obj, ok := elem.(types.Object); ok {
			var model instanceSubnetModel
			elemDiags := obj.As(ctx, &model, basetypes.ObjectAsOptions{})
			diags.Append(elemDiags...)
			result = append(result, model)
		}
	}
	return result, diags
}

func (r *instanceResource) pollInstanceUntilNetworkInterfaceOk(
	ctx context.Context,
	instanceId string,
	networkInterfaceId string,
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

		for _, address := range respModel.Instance.Addresses {
			if action == "attach" {
				if address.NetworkInterfaceId.Get() != nil && *address.NetworkInterfaceId.Get() == networkInterfaceId {
					isOk = true
					break
				}
			} else if action == "detach" {
				isOk = true
				if address.NetworkInterfaceId.Get() != nil && *address.NetworkInterfaceId.Get() == networkInterfaceId {
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

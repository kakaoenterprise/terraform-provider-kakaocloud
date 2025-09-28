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
	"github.com/kakaoenterprise/kc-sdk-go/services/vpc"
	"golang.org/x/net/context"
)

func (r *instanceResource) setNetworkInterfaceId(ctx context.Context, plan *instanceResourceModel, respDiags *diag.Diagnostics) {
	if plan.Subnets.IsNull() || plan.Subnets.Elements() == nil {
		return
	}

	subnetList, subnetDiags := r.convertListToInstanceSubnetModel(ctx, plan.Subnets)
	respDiags.Append(subnetDiags...)
	if respDiags.HasError() {
		return
	}

	needsSet := false
	for _, subnet := range subnetList {
		if subnet.NetworkInterfaceId.IsNull() || subnet.NetworkInterfaceId.ValueString() == "" {
			needsSet = true
			break
		}
	}

	if !needsSet {
		return
	}

	usedNicIds := make(map[string]bool)
	for _, subnet := range subnetList {
		if !subnet.NetworkInterfaceId.IsNull() {
			usedNicIds[subnet.NetworkInterfaceId.ValueString()] = true
		}
	}

	addressList, addressDiags := r.convertListToInstanceAddressModel(ctx, plan.Addresses)
	respDiags.Append(addressDiags...)
	if respDiags.HasError() {
		return
	}

	subnetMap := make(map[string][]string)
	for _, address := range addressList {
		nicId := address.NetworkInterfaceId.ValueString()
		if nicId == "" || usedNicIds[nicId] {
			continue
		}
		nicResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, respDiags,
			func() (*vpc.BnsVpcV1ApiGetNetworkInterfaceModelResponseNetworkInterfaceModel, *http.Response, error) {
				return r.kc.ApiClient.NetworkInterfaceAPI.GetNetworkInterface(ctx, nicId).XAuthToken(r.kc.XAuthToken).Execute()
			},
		)
		if err != nil {
			common.AddApiActionError(ctx, r, httpResp, "GetNetworkInterface", err, respDiags)
			return
		}
		if nicResp.NetworkInterface.SubnetId.Get() != nil {
			subnetId := nicResp.NetworkInterface.SubnetId.Get()
			subnetMap[*subnetId] = append(subnetMap[*subnetId], nicId)
		}
	}

	for i, subnet := range subnetList {
		if subnet.NetworkInterfaceId.IsNull() || subnet.NetworkInterfaceId.ValueString() == "" {
			nicIds, exists := subnetMap[subnet.Id.ValueString()]
			if !exists {
				common.AddGeneralError(ctx, r, respDiags,
					fmt.Sprintf("No network interface found for subnet_id: %s", subnet.Id.ValueString()))
				return
			}

			assigned := false
			for _, nicId := range nicIds {
				if !usedNicIds[nicId] {
					subnetList[i].NetworkInterfaceId = types.StringValue(nicId)
					usedNicIds[nicId] = true
					assigned = true
					break
				}
			}
			if !assigned {
				common.AddGeneralError(ctx, r, respDiags,
					fmt.Sprintf("All network interfaces for subnet_id %s have already been used", subnet.Id.ValueString()))
				return
			}
		}
	}

	convertedList, diags := types.ListValueFrom(ctx, plan.Subnets.ElementType(ctx), subnetList)
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

	for _, s := range *states {
		if _, exists := planMap[s.NetworkInterfaceId.ValueString()]; !exists {
			ok := r.detachNetworkInterface(ctx, instanceId, s.NetworkInterfaceId.ValueString(), resp)
			if !ok {
				return false
			}
		}
	}

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

func (r *instanceResource) convertListToInstanceAddressModel(
	ctx context.Context,
	list types.List,
) ([]instanceAddressModel, diag.Diagnostics) {
	var result []instanceAddressModel
	var diags diag.Diagnostics

	for _, elem := range list.Elements() {
		if obj, ok := elem.(types.Object); ok {
			var model instanceAddressModel
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

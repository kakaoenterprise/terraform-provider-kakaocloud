// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package mysql

import (
	"context"

	"terraform-provider-kakaocloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	mysqlsdk "github.com/kakaoenterprise/kc-sdk-go/services/mysql"
)

func toInstanceModel(
	ctx context.Context,
	item mysqlsdk.InstanceResponseModel,
	respDiags *diag.Diagnostics,
) (instanceModel, bool) {
	specContent, ok := instanceSpecContentValue(ctx, item.SpecContent, respDiags)
	if !ok {
		return instanceModel{}, false
	}

	statusContent, ok := instanceStatusContentValue(ctx, item.StatusContent, respDiags)
	if !ok {
		return instanceModel{}, false
	}

	return instanceModel{
		Id:                 types.StringValue(item.Id),
		ProjectId:          types.StringValue(item.ProjectId),
		InstanceGroupId:    types.StringValue(item.InstanceGroupId),
		InstanceGroupName:  types.StringValue(item.InstanceGroupName),
		Name:               types.StringValue(item.Name),
		Status:             types.StringValue(string(item.Status)),
		AvailabilityStatus: utils.ConvertNullableString(item.AvailabilityStatus),
		StatusContent:      statusContent,
		Role:               types.StringValue(string(item.Role)),
		DataDiskUsage:      types.Int32Value(item.DataDiskUsage),
		LogDiskUsage:       types.Int32Value(item.LogDiskUsage),
		SpecContent:        specContent,
		CreatedAt:          types.StringValue(item.CreatedAt),
		UpdatedAt:          types.StringValue(item.UpdatedAt),
		StartTime:          utils.ConvertNullableStringWithEmptyToNull(item.StartTime),
	}, true
}

func instanceSpecContentValue(
	ctx context.Context,
	src mysqlsdk.MysqlV1ApiListMysqlInstancesModelSpecContentResponseModel,
	respDiags *diag.Diagnostics,
) (types.Object, bool) {
	networkPorts, ok := instanceNetworkPortsValue(ctx, src.NetworkPorts, respDiags)
	if !ok {
		return types.ObjectNull(instanceSpecContentAttrTypes), false
	}

	specContent, diags := types.ObjectValueFrom(ctx, instanceSpecContentAttrTypes, instanceSpecContentModel{
		AvailabilityZone: types.StringValue(src.AvailabilityZone),
		FlavorId:         types.StringValue(src.FlavorId),
		DataDiskSize:     types.Int32Value(src.DataDiskSize),
		LogDiskSize:      types.Int32Value(src.LogDiskSize),
		EngineVersion:    types.StringValue(src.EngineVersion),
		NetworkPorts:     networkPorts,
	})
	respDiags.Append(diags...)
	if respDiags.HasError() {
		return types.ObjectNull(instanceSpecContentAttrTypes), false
	}

	return specContent, true
}

func instanceNetworkPortsValue(ctx context.Context, src []mysqlsdk.NetworkPortResponseModel, respDiags *diag.Diagnostics) (types.List, bool) {
	networkPorts, diags := utils.ConvertListFromModel(ctx, src, networkPortAttrTypes, func(port mysqlsdk.NetworkPortResponseModel) any {
		securityGroupIds, portDiags := utils.SetFromStrings(ctx, port.SecurityGroupIds)
		respDiags.Append(portDiags...)

		return networkPortModel{
			SubnetId:         utils.ConvertNullableString(port.SubnetId),
			SecurityGroupIds: securityGroupIds,
		}
	})
	respDiags.Append(diags...)
	if respDiags.HasError() {
		return types.ListNull(types.ObjectType{AttrTypes: networkPortAttrTypes}), false
	}

	return networkPorts, true
}

func instanceStatusContentValue(ctx context.Context, src mysqlsdk.StatusContentResponseModel, respDiags *diag.Diagnostics) (types.Object, bool) {
	needsRestartReason, diags := utils.ListFromStrings(ctx, src.NeedsRestartReason)
	respDiags.Append(diags...)
	if respDiags.HasError() {
		return types.ObjectNull(instanceStatusContentAttrTypes), false
	}

	statusContent, diags := types.ObjectValueFrom(ctx, instanceStatusContentAttrTypes, instanceStatusContentModel{
		NeedsRestart:       types.BoolValue(src.NeedsRestart),
		NeedsRestartReason: needsRestartReason,
	})
	respDiags.Append(diags...)
	if respDiags.HasError() {
		return types.ObjectNull(instanceStatusContentAttrTypes), false
	}

	return statusContent, true
}

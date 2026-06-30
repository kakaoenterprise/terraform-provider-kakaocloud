// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package mysql

import (
	"context"

	"terraform-provider-kakaocloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	mysqlsdk "github.com/kakaoenterprise/kc-sdk-go/services/mysql"
)

func toInstanceGroupModel(ctx context.Context, src mysqlsdk.MysqlV1ApiGetMysqlInstanceGroupModelInstanceGroupResponseModel, respDiags *diag.Diagnostics) (instanceGroupModel, bool) {
	endpoint, ok := instanceGroupEndpointValue(ctx, src.Endpoint, respDiags)
	if !ok {
		return instanceGroupModel{}, false
	}

	networkInfo, ok := instanceGroupNetworkInfoValue(ctx, src.NetworkInfo, respDiags)
	if !ok {
		return instanceGroupModel{}, false
	}

	specContent, ok := instanceGroupSpecContentValue(ctx, src.SpecContent, respDiags)
	if !ok {
		return instanceGroupModel{}, false
	}

	backupSchedule, ok := instanceGroupBackupScheduleValue(ctx, src.BackupSchedule, respDiags)
	if !ok {
		return instanceGroupModel{}, false
	}

	parameterGroup, ok := instanceGroupParameterGroupValue(ctx, src.ParameterGroup, respDiags)
	if !ok {
		return instanceGroupModel{}, false
	}

	extraInfo, ok := instanceGroupExtraInfoValue(ctx, src.ExtraInfo, respDiags)
	if !ok {
		return instanceGroupModel{}, false
	}

	instances, ok := instanceGroupInstancesValue(ctx, src.Instances, respDiags)
	if !ok {
		return instanceGroupModel{}, false
	}

	return instanceGroupModel{
		Id:             types.StringValue(src.Id),
		CreatedAt:      types.StringValue(src.CreatedAt),
		UpdatedAt:      types.StringValue(src.UpdatedAt),
		License:        types.StringValue(src.License),
		Name:           types.StringValue(src.Name),
		ProjectId:      types.StringValue(src.ProjectId),
		Description:    utils.ConvertNullableString(src.Description),
		Creator:        utils.ConvertNullableString(src.Creator),
		SourceBackupId: utils.ConvertNullableString(src.SourceBackupId),
		IsMultiAz:      types.BoolValue(src.IsMultiAz),
		Endpoint:       endpoint,
		Status:         types.StringValue(string(src.Status)),
		NetworkInfo:    networkInfo,
		SpecContent:    specContent,
		BackupSchedule: backupSchedule,
		ParameterGroup: parameterGroup,
		ExtraInfo:      extraInfo,
		Instances:      instances,
	}, true
}

func toInstanceGroupModelFromList(
	ctx context.Context,
	src mysqlsdk.MysqlV1ApiListMysqlInstanceGroupsModelInstanceGroupResponseModel,
	respDiags *diag.Diagnostics,
) (instanceGroupListModel, bool) {
	endpoint, ok := instanceGroupEndpointValue(ctx, src.Endpoint, respDiags)
	if !ok {
		return instanceGroupListModel{}, false
	}

	networkInfo, ok := instanceGroupNetworkInfoValue(ctx, src.NetworkInfo, respDiags)
	if !ok {
		return instanceGroupListModel{}, false
	}

	specContent, ok := instanceGroupSpecContentFromListValue(ctx, src.SpecContent, respDiags)
	if !ok {
		return instanceGroupListModel{}, false
	}

	backupSchedule, ok := instanceGroupBackupScheduleValue(ctx, src.BackupSchedule, respDiags)
	if !ok {
		return instanceGroupListModel{}, false
	}

	parameterGroup, ok := instanceGroupParameterGroupValue(ctx, src.ParameterGroup, respDiags)
	if !ok {
		return instanceGroupListModel{}, false
	}

	extraInfo, ok := instanceGroupExtraInfoFromListValue(ctx, src.ExtraInfo, respDiags)
	if !ok {
		return instanceGroupListModel{}, false
	}

	instances, ok := instanceGroupInstancesFromListValue(ctx, src.Instances, respDiags)
	if !ok {
		return instanceGroupListModel{}, false
	}

	return instanceGroupListModel{
		Id:             types.StringValue(src.Id),
		CreatedAt:      types.StringValue(src.CreatedAt),
		UpdatedAt:      types.StringValue(src.UpdatedAt),
		License:        types.StringValue(src.License),
		Name:           types.StringValue(src.Name),
		ProjectId:      types.StringValue(src.ProjectId),
		Description:    utils.ConvertNullableString(src.Description),
		Creator:        utils.ConvertNullableString(src.Creator),
		SourceBackupId: utils.ConvertNullableString(src.SourceBackupId),
		IsMultiAz:      types.BoolValue(src.IsMultiAz),
		Endpoint:       endpoint,
		Status:         types.StringValue(string(src.Status)),
		NetworkInfo:    networkInfo,
		SpecContent:    specContent,
		BackupSchedule: backupSchedule,
		ParameterGroup: parameterGroup,
		ExtraInfo:      extraInfo,
		Instances:      instances,
	}, true
}

func instanceGroupEndpointValue(ctx context.Context, src []string, respDiags *diag.Diagnostics) (types.List, bool) {
	endpoint, diags := utils.ListFromStrings(ctx, src)
	respDiags.Append(diags...)
	if respDiags.HasError() {
		return types.ListNull(types.StringType), false
	}
	return endpoint, true
}

func instanceGroupNetworkInfoValue(ctx context.Context, src mysqlsdk.NullableNetworkInfoResponseModel, respDiags *diag.Diagnostics) (types.Object, bool) {
	if !src.IsSet() || src.Get() == nil {
		return types.ObjectNull(networkInfoAttrTypes), true
	}

	networkInfoModel := src.Get()
	securityGroupIds, diags := utils.SetFromStrings(ctx, networkInfoModel.SecurityGroupIds)
	respDiags.Append(diags...)
	if respDiags.HasError() {
		return types.ObjectNull(networkInfoAttrTypes), false
	}

	primarySubnetInfo := types.ObjectNull(subnetInfoAttrTypes)
	if primarySubnet, ok := networkInfoModel.GetPrimarySubnetInfoOk(); ok && primarySubnet != nil {
		var diags diag.Diagnostics
		primarySubnetInfo, diags = types.ObjectValueFrom(ctx, subnetInfoAttrTypes, subnetInfoModel{
			Replicas:         types.Int32Value(primarySubnet.Replicas),
			AvailabilityZone: types.StringValue(primarySubnet.AvailabilityZone),
			SubnetId:         types.StringValue(primarySubnet.SubnetId),
		})
		respDiags.Append(diags...)
		if respDiags.HasError() {
			return types.ObjectNull(networkInfoAttrTypes), false
		}
	}

	standbySubnetInfos, diags := utils.ConvertListFromModel(ctx, networkInfoModel.StandbySubnetInfo, subnetInfoAttrTypes, func(subnet mysqlsdk.SubnetInfoResponseModel) any {
		return subnetInfoModel{
			Replicas:         types.Int32Value(subnet.Replicas),
			AvailabilityZone: types.StringValue(subnet.AvailabilityZone),
			SubnetId:         types.StringValue(subnet.SubnetId),
		}
	})
	respDiags.Append(diags...)
	if respDiags.HasError() {
		return types.ObjectNull(networkInfoAttrTypes), false
	}

	networkInfo, diags := types.ObjectValue(networkInfoAttrTypes, map[string]attr.Value{
		"primary_subnet_info": primarySubnetInfo,
		"standby_subnet_info": standbySubnetInfos,
		"security_group_ids":  securityGroupIds,
	})
	respDiags.Append(diags...)
	if respDiags.HasError() {
		return types.ObjectNull(networkInfoAttrTypes), false
	}

	return networkInfo, true
}

func instanceGroupSpecContentValue(
	ctx context.Context,
	src mysqlsdk.MysqlV1ApiGetMysqlInstanceGroupModelSpecContentResponseModel,
	respDiags *diag.Diagnostics,
) (types.Object, bool) {
	specContent, diags := types.ObjectValueFrom(ctx, specContentAttrTypes, specContentModel{
		DatabaseUserName:  types.StringValue(src.DatabaseUserName),
		PrimaryPort:       types.Int32Value(src.PrimaryPort),
		StandbyPort:       types.Int32Value(src.StandbyPort),
		EngineVersion:     types.StringValue(src.EngineVersion),
		FlavorId:          types.StringValue(src.FlavorId),
		Vcpu:              types.Int32Value(src.Vcpu),
		Memory:            types.Int32Value(src.Memory),
		LogDiskSize:       types.Int32Value(src.LogDiskSize),
		DataDiskSize:      types.Int32Value(src.DataDiskSize),
		InstanceGroupType: types.StringValue(string(src.InstanceGroupType)),
		NodeSize:          types.Int32Value(src.NodeSize),
	})
	respDiags.Append(diags...)
	if respDiags.HasError() {
		return types.ObjectNull(specContentAttrTypes), false
	}
	return specContent, true
}

func instanceGroupSpecContentFromListValue(
	ctx context.Context,
	src mysqlsdk.MysqlV1ApiListMysqlInstanceGroupsModelSpecContentResponseModel,
	respDiags *diag.Diagnostics,
) (types.Object, bool) {
	specContent, diags := types.ObjectValueFrom(ctx, specContentAttrTypes, specContentModel{
		DatabaseUserName:  types.StringValue(src.DatabaseUserName),
		PrimaryPort:       types.Int32Value(src.PrimaryPort),
		StandbyPort:       types.Int32Value(src.StandbyPort),
		EngineVersion:     types.StringValue(src.EngineVersion),
		FlavorId:          types.StringValue(src.FlavorId),
		Vcpu:              types.Int32Value(src.Vcpu),
		Memory:            types.Int32Value(src.Memory),
		LogDiskSize:       types.Int32Value(src.LogDiskSize),
		DataDiskSize:      types.Int32Value(src.DataDiskSize),
		InstanceGroupType: types.StringValue(string(src.InstanceGroupType)),
		NodeSize:          types.Int32Value(src.NodeSize),
	})
	respDiags.Append(diags...)
	if respDiags.HasError() {
		return types.ObjectNull(specContentAttrTypes), false
	}
	return specContent, true
}

func instanceGroupBackupScheduleValue(ctx context.Context, src mysqlsdk.NullableBackupScheduleResponseModel, respDiags *diag.Diagnostics) (types.Object, bool) {
	backupSchedule, diags := utils.ConvertObjectFromModel(ctx, src, backupScheduleAttrTypes, func(schedule mysqlsdk.BackupScheduleResponseModel) any {
		return backupScheduleModel{
			Id:             types.StringValue(schedule.Id),
			Type:           utils.ConvertNullableString(schedule.Type),
			StartTime:      utils.ConvertNullableString(schedule.StartTime),
			ExpiryDuration: utils.ConvertNullableInt32(schedule.ExpiryDuration),
			Enabled:        types.BoolValue(schedule.Enabled),
		}
	})
	respDiags.Append(diags...)
	if respDiags.HasError() {
		return types.ObjectNull(backupScheduleAttrTypes), false
	}
	return backupSchedule, true
}

func instanceGroupParameterGroupValue(ctx context.Context, src mysqlsdk.ParameterGroupResponseModel, respDiags *diag.Diagnostics) (types.Object, bool) {
	parameterGroup, diags := types.ObjectValueFrom(ctx, parameterGroupAttrTypes, parameterGroupModel{
		Id:                      types.StringValue(src.Id),
		Type:                    types.StringValue(string(src.Type)),
		ApplyStatus:             utils.ConvertNullableString(src.ApplyStatus),
		EngineVersion:           types.StringValue(src.EngineVersion),
		IsEngineVersionMismatch: types.BoolValue(src.IsEngineVersionMismatch),
	})
	respDiags.Append(diags...)
	if respDiags.HasError() {
		return types.ObjectNull(parameterGroupAttrTypes), false
	}
	return parameterGroup, true
}

func instanceGroupExtraInfoValue(
	ctx context.Context,
	src mysqlsdk.NullableMysqlV1ApiGetMysqlInstanceGroupModelExtraInfoResponseModel,
	respDiags *diag.Diagnostics,
) (types.Object, bool) {
	extraInfo, diags := utils.ConvertObjectFromModel(ctx, src, extraInfoAttrTypes, func(info mysqlsdk.MysqlV1ApiGetMysqlInstanceGroupModelExtraInfoResponseModel) any {
		return extraInfoModel{
			UseCaseSensitiveTableNames: utils.ConvertNullableBool(info.UseCaseSensitiveTableNames),
		}
	})
	respDiags.Append(diags...)
	if respDiags.HasError() {
		return types.ObjectNull(extraInfoAttrTypes), false
	}
	return extraInfo, true
}

func instanceGroupExtraInfoFromListValue(
	ctx context.Context,
	src mysqlsdk.NullableMysqlV1ApiListMysqlInstanceGroupsModelExtraInfoResponseModel,
	respDiags *diag.Diagnostics,
) (types.Object, bool) {
	extraInfo, diags := utils.ConvertObjectFromModel(ctx, src, extraInfoAttrTypes, func(info mysqlsdk.MysqlV1ApiListMysqlInstanceGroupsModelExtraInfoResponseModel) any {
		return extraInfoModel{
			UseCaseSensitiveTableNames: utils.ConvertNullableBool(info.UseCaseSensitiveTableNames),
		}
	})
	respDiags.Append(diags...)
	if respDiags.HasError() {
		return types.ObjectNull(extraInfoAttrTypes), false
	}
	return extraInfo, true
}

func instanceGroupInstancesValue(
	ctx context.Context,
	src mysqlsdk.NullableMysqlV1ApiGetMysqlInstanceGroupModelTopologyResponseModel,
	respDiags *diag.Diagnostics,
) (types.Object, bool) {
	instances, diags := utils.ConvertObjectFromModel(ctx, src, instancesAttrTypes, func(instances mysqlsdk.MysqlV1ApiGetMysqlInstanceGroupModelTopologyResponseModel) any {
		primary := types.ObjectNull(instanceNodeAttrTypes)
		if primaryNode, ok := instances.GetPrimaryOk(); ok && primaryNode != nil {
			var primaryDiags diag.Diagnostics
			primary, primaryDiags = instanceGroupInstanceNodeValue(ctx, *primaryNode)
			respDiags.Append(primaryDiags...)
		}

		standby, standbyDiags := utils.ConvertListFromModel(ctx, instances.Standby, instanceNodeAttrTypes, func(node mysqlsdk.MysqlV1ApiGetMysqlInstanceGroupModelTopologyInfoResponseModel) any {
			return instanceNodeModel{
				InstanceId:       utils.ConvertNullableString(node.InstanceId),
				SubnetId:         utils.ConvertNullableString(node.SubnetId),
				AvailabilityZone: utils.ConvertNullableString(node.AvailabilityZone),
			}
		})
		respDiags.Append(standbyDiags...)

		return struct {
			Primary types.Object `tfsdk:"primary"`
			Standby types.List   `tfsdk:"standby"`
		}{
			Primary: primary,
			Standby: standby,
		}
	})
	respDiags.Append(diags...)
	if respDiags.HasError() {
		return types.ObjectNull(instancesAttrTypes), false
	}
	return instances, true
}

func instanceGroupInstancesFromListValue(
	ctx context.Context,
	src mysqlsdk.NullableMysqlV1ApiListMysqlInstanceGroupsModelTopologyResponseModel,
	respDiags *diag.Diagnostics,
) (types.Object, bool) {
	instances, diags := utils.ConvertObjectFromModel(ctx, src, instanceGroupListInstancesAttrTypes, func(instances mysqlsdk.MysqlV1ApiListMysqlInstanceGroupsModelTopologyResponseModel) any {
		primary := types.ObjectNull(instanceGroupListInstanceNodeAttrTypes)
		if primaryNode, ok := instances.GetPrimaryOk(); ok && primaryNode != nil {
			var primaryDiags diag.Diagnostics
			primary, primaryDiags = instanceGroupInstanceNodeFromListValue(ctx, *primaryNode)
			respDiags.Append(primaryDiags...)
		}

		standby, standbyDiags := utils.ConvertListFromModel(ctx, instances.Standby, instanceGroupListInstanceNodeAttrTypes, func(node mysqlsdk.MysqlV1ApiListMysqlInstanceGroupsModelTopologyInfoResponseModel) any {
			return instanceGroupListNodeModel{
				InstanceId: utils.ConvertNullableString(node.InstanceId),
			}
		})
		respDiags.Append(standbyDiags...)

		return struct {
			Primary types.Object `tfsdk:"primary"`
			Standby types.List   `tfsdk:"standby"`
		}{
			Primary: primary,
			Standby: standby,
		}
	})
	respDiags.Append(diags...)
	if respDiags.HasError() {
		return types.ObjectNull(instanceGroupListInstancesAttrTypes), false
	}
	return instances, true
}

func instanceGroupInstanceNodeValue(ctx context.Context, src mysqlsdk.MysqlV1ApiGetMysqlInstanceGroupModelTopologyInfoResponseModel) (types.Object, diag.Diagnostics) {
	return types.ObjectValueFrom(ctx, instanceNodeAttrTypes, instanceNodeModel{
		InstanceId:       utils.ConvertNullableString(src.InstanceId),
		SubnetId:         utils.ConvertNullableString(src.SubnetId),
		AvailabilityZone: utils.ConvertNullableString(src.AvailabilityZone),
	})
}

func instanceGroupInstanceNodeFromListValue(
	ctx context.Context,
	src mysqlsdk.MysqlV1ApiListMysqlInstanceGroupsModelTopologyInfoResponseModel,
) (types.Object, diag.Diagnostics) {
	return types.ObjectValueFrom(ctx, instanceGroupListInstanceNodeAttrTypes, instanceGroupListNodeModel{
		InstanceId: utils.ConvertNullableString(src.InstanceId),
	})
}

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package mysql

import (
	"context"

	"terraform-provider-kakaocloud/internal/utils"

	resourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	mysqlsdk "github.com/kakaoenterprise/kc-sdk-go/services/mysql"
)

func toInstanceGroupResourceModel(
	ctx context.Context,
	src mysqlsdk.MysqlV1ApiGetMysqlInstanceGroupModelInstanceGroupResponseModel,
	desiredNetworkInfo types.Object,
	source types.Object,
	timeouts resourceTimeouts.Value,
) (instanceGroupResourceModel, diag.Diagnostics, bool) {
	var diags diag.Diagnostics

	mapped, ok := toInstanceGroupModel(ctx, src, &diags)
	if !ok {
		return instanceGroupResourceModel{}, diags, false
	}

	desiredNetworkInfo, ok = instanceGroupResourceDesiredNetworkInfoForState(ctx, mapped.NetworkInfo, desiredNetworkInfo, &diags)
	if !ok {
		return instanceGroupResourceModel{}, diags, false
	}

	networkInfo, ok := instanceGroupResourceNetworkInfoValue(ctx, src.NetworkInfo, &diags)
	if !ok {
		return instanceGroupResourceModel{}, diags, false
	}

	var spec specContentModel
	diags.Append(mapped.SpecContent.As(ctx, &spec, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return instanceGroupResourceModel{}, diags, false
	}

	resourceSpec, ok := instanceGroupResourceSpecContentValue(ctx, spec, &diags)
	if !ok {
		return instanceGroupResourceModel{}, diags, false
	}

	return instanceGroupResourceModel{
		Id:                 mapped.Id,
		CreatedAt:          mapped.CreatedAt,
		UpdatedAt:          mapped.UpdatedAt,
		License:            mapped.License,
		Name:               mapped.Name,
		ProjectId:          mapped.ProjectId,
		Description:        mapped.Description,
		Creator:            mapped.Creator,
		SourceBackupId:     mapped.SourceBackupId,
		IsMultiAz:          mapped.IsMultiAz,
		Endpoint:           mapped.Endpoint,
		Status:             mapped.Status,
		NetworkInfo:        networkInfo,
		DesiredNetworkInfo: desiredNetworkInfo,
		SpecContent:        resourceSpec,
		Source:             source,
		BackupSchedule:     mapped.BackupSchedule,
		ParameterGroup:     mapped.ParameterGroup,
		ExtraInfo:          mapped.ExtraInfo,
		Instances:          mapped.Instances,
		Timeouts:           timeouts,
	}, diags, true
}

func instanceGroupResourceSpecContentValue(
	ctx context.Context,
	spec specContentModel,
	respDiags *diag.Diagnostics,
) (types.Object, bool) {
	specContent, diags := types.ObjectValueFrom(ctx, instanceGroupResourceSpecContentAttrTypes, instanceGroupSpecContentResourceModel{
		DatabaseUserName:  spec.DatabaseUserName,
		PrimaryPort:       spec.PrimaryPort,
		StandbyPort:       resourceStandbyPortValue(spec.StandbyPort),
		EngineVersion:     spec.EngineVersion,
		FlavorId:          spec.FlavorId,
		Vcpu:              spec.Vcpu,
		Memory:            spec.Memory,
		LogDiskSize:       spec.LogDiskSize,
		DataDiskSize:      spec.DataDiskSize,
		InstanceGroupType: spec.InstanceGroupType,
		NodeSize:          spec.NodeSize,
	})
	respDiags.Append(diags...)
	if respDiags.HasError() {
		return types.ObjectNull(instanceGroupResourceSpecContentAttrTypes), false
	}
	return specContent, true
}

func instanceGroupResourceDesiredNetworkInfoForState(
	ctx context.Context,
	remoteNetworkInfo types.Object,
	desiredNetworkInfo types.Object,
	respDiags *diag.Diagnostics,
) (types.Object, bool) {
	if !desiredNetworkInfo.IsNull() && !desiredNetworkInfo.IsUnknown() {
		return desiredNetworkInfo, true
	}
	return instanceGroupResourceDesiredNetworkInfoFromRemote(ctx, remoteNetworkInfo, respDiags)
}

func resourceStandbyPortValue(value types.Int32) types.Int32 {
	if value.IsNull() || value.IsUnknown() || value.ValueInt32() != 0 {
		return value
	}
	return types.Int32Null()
}

func instanceGroupResourceDesiredNetworkInfoFromRemote(
	ctx context.Context,
	value types.Object,
	respDiags *diag.Diagnostics,
) (types.Object, bool) {
	if value.IsNull() || value.IsUnknown() {
		return types.ObjectNull(instanceGroupResourceDesiredNetworkInfoAttrTypes), true
	}

	var networkInfo networkInfoModel
	respDiags.Append(value.As(ctx, &networkInfo, basetypes.ObjectAsOptions{})...)
	if respDiags.HasError() {
		return types.ObjectNull(instanceGroupResourceDesiredNetworkInfoAttrTypes), false
	}

	resourceNetworkInfo, ok := toInstanceGroupResourceDesiredNetworkInfoModel(ctx, networkInfo, respDiags)
	if !ok {
		return types.ObjectNull(instanceGroupResourceDesiredNetworkInfoAttrTypes), false
	}

	networkInfoValue, diags := types.ObjectValueFrom(ctx, instanceGroupResourceDesiredNetworkInfoAttrTypes, resourceNetworkInfo)
	respDiags.Append(diags...)
	if respDiags.HasError() {
		return types.ObjectNull(instanceGroupResourceDesiredNetworkInfoAttrTypes), false
	}
	return networkInfoValue, true
}

func toInstanceGroupResourceDesiredNetworkInfoModel(
	ctx context.Context,
	networkInfo networkInfoModel,
	respDiags *diag.Diagnostics,
) (instanceGroupResourceDesiredNetworkInfoModel, bool) {
	primarySubnetInfo, ok := toInstanceGroupResourceDesiredSubnetInfoObject(ctx, networkInfo.PrimarySubnetInfo, respDiags)
	if !ok {
		return instanceGroupResourceDesiredNetworkInfoModel{}, false
	}

	standbySubnetInfo, ok := toInstanceGroupResourceDesiredSubnetInfoSet(ctx, networkInfo.StandbySubnetInfo, respDiags)
	if !ok {
		return instanceGroupResourceDesiredNetworkInfoModel{}, false
	}

	return instanceGroupResourceDesiredNetworkInfoModel{
		PrimarySubnetInfo: primarySubnetInfo,
		StandbySubnetInfo: standbySubnetInfo,
		SecurityGroupIds:  networkInfo.SecurityGroupIds,
	}, true
}

func toInstanceGroupResourceDesiredSubnetInfoObject(
	ctx context.Context,
	value types.Object,
	respDiags *diag.Diagnostics,
) (types.Object, bool) {
	if value.IsNull() || value.IsUnknown() {
		return types.ObjectNull(instanceGroupResourceDesiredSubnetInfoAttrTypes), true
	}

	var subnet subnetInfoModel
	respDiags.Append(value.As(ctx, &subnet, basetypes.ObjectAsOptions{})...)
	if respDiags.HasError() {
		return types.ObjectNull(instanceGroupResourceDesiredSubnetInfoAttrTypes), false
	}

	subnetInfo, diags := types.ObjectValueFrom(ctx, instanceGroupResourceDesiredSubnetInfoAttrTypes, instanceGroupResourceDesiredSubnetInfoModel{
		Replicas: subnet.Replicas,
		SubnetId: subnet.SubnetId,
	})
	respDiags.Append(diags...)
	if respDiags.HasError() {
		return types.ObjectNull(instanceGroupResourceDesiredSubnetInfoAttrTypes), false
	}
	return subnetInfo, true
}

func toInstanceGroupResourceDesiredSubnetInfoSet(
	ctx context.Context,
	value types.List,
	respDiags *diag.Diagnostics,
) (types.Set, bool) {
	elemType := types.ObjectType{AttrTypes: instanceGroupResourceDesiredSubnetInfoAttrTypes}
	if value.IsNull() || value.IsUnknown() {
		return types.SetNull(elemType), true
	}
	if len(value.Elements()) == 0 {

		return types.SetNull(elemType), true
	}

	var subnets []subnetInfoModel
	respDiags.Append(value.ElementsAs(ctx, &subnets, false)...)
	if respDiags.HasError() {
		return types.SetNull(elemType), false
	}

	resourceSubnets := make([]instanceGroupResourceDesiredSubnetInfoModel, 0, len(subnets))
	for _, subnet := range subnets {
		resourceSubnets = append(resourceSubnets, instanceGroupResourceDesiredSubnetInfoModel{
			Replicas: subnet.Replicas,
			SubnetId: subnet.SubnetId,
		})
	}

	subnetInfos, diags := types.SetValueFrom(ctx, elemType, resourceSubnets)
	respDiags.Append(diags...)
	if respDiags.HasError() {
		return types.SetNull(elemType), false
	}
	return subnetInfos, true
}

func instanceGroupResourceNetworkInfoValue(
	ctx context.Context,
	src mysqlsdk.NullableNetworkInfoResponseModel,
	respDiags *diag.Diagnostics,
) (types.Object, bool) {
	if !src.IsSet() || src.Get() == nil {
		return types.ObjectNull(instanceGroupResourceNetworkInfoAttrTypes), true
	}

	networkInfoModel := src.Get()
	securityGroupIds, diags := utils.SetFromStrings(ctx, networkInfoModel.SecurityGroupIds)
	respDiags.Append(diags...)
	if respDiags.HasError() {
		return types.ObjectNull(instanceGroupResourceNetworkInfoAttrTypes), false
	}

	primarySubnetInfo := types.ObjectNull(instanceGroupResourceSubnetInfoAttrTypes)
	if primarySubnet, ok := networkInfoModel.GetPrimarySubnetInfoOk(); ok && primarySubnet != nil {
		var diags diag.Diagnostics
		primarySubnetInfo, diags = types.ObjectValueFrom(ctx, instanceGroupResourceSubnetInfoAttrTypes, instanceGroupResourceSubnetInfoModel{
			Replicas:         types.Int32Value(primarySubnet.Replicas),
			AvailabilityZone: types.StringValue(primarySubnet.AvailabilityZone),
			SubnetId:         types.StringValue(primarySubnet.SubnetId),
		})
		respDiags.Append(diags...)
		if respDiags.HasError() {
			return types.ObjectNull(instanceGroupResourceNetworkInfoAttrTypes), false
		}
	}

	standbySubnets := make([]instanceGroupResourceSubnetInfoModel, 0, len(networkInfoModel.StandbySubnetInfo))
	for _, subnet := range networkInfoModel.StandbySubnetInfo {
		standbySubnets = append(standbySubnets, instanceGroupResourceSubnetInfoModel{
			Replicas:         types.Int32Value(subnet.Replicas),
			AvailabilityZone: types.StringValue(subnet.AvailabilityZone),
			SubnetId:         types.StringValue(subnet.SubnetId),
		})
	}
	standbySubnetInfo, diags := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: instanceGroupResourceSubnetInfoAttrTypes}, standbySubnets)
	respDiags.Append(diags...)
	if respDiags.HasError() {
		return types.ObjectNull(instanceGroupResourceNetworkInfoAttrTypes), false
	}

	networkInfo, diags := types.ObjectValueFrom(ctx, instanceGroupResourceNetworkInfoAttrTypes, instanceGroupResourceNetworkInfoModel{
		PrimarySubnetInfo: primarySubnetInfo,
		StandbySubnetInfo: standbySubnetInfo,
		SecurityGroupIds:  securityGroupIds,
	})
	respDiags.Append(diags...)
	if respDiags.HasError() {
		return types.ObjectNull(instanceGroupResourceNetworkInfoAttrTypes), false
	}
	return networkInfo, true
}

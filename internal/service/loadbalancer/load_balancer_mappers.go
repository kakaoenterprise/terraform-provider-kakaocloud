// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package loadbalancer

import (
	"context"

	"terraform-provider-kakaocloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/kakaoenterprise/kc-sdk-go/services/loadbalancer"
)

var accessLogResourceAttrType = map[string]attr.Type{
	"bucket":     types.StringType,
	"access_key": types.StringType,
	"secret_key": types.StringType,
}

var accessLogDataSourceAttrType = map[string]attr.Type{
	"bucket": types.StringType,
}

func mapLoadBalancer(
	ctx context.Context,
	base *loadBalancerBaseModel,
	lb *loadbalancer.BnsLoadBalancerV1ApiGetLoadBalancerModelLoadBalancerModel,
	diags *diag.Diagnostics,
) bool {
	base.Id = types.StringValue(lb.Id)
	base.Name = utils.ConvertNullableString(lb.Name)
	base.Description = utils.ConvertNullableString(lb.Description)
	base.Type = utils.ConvertNullableString(lb.Type)
	base.ProjectId = utils.ConvertNullableString(lb.ProjectId)
	base.ProvisioningStatus = utils.ConvertNullableString(lb.ProvisioningStatus)
	base.OperatingStatus = utils.ConvertNullableString(lb.OperatingStatus)
	base.CreatedAt = utils.ConvertNullableTime(lb.CreatedAt)
	base.UpdatedAt = utils.ConvertNullableTime(lb.UpdatedAt)
	base.AvailabilityZone = utils.ConvertNullableString(lb.AvailabilityZone)
	base.BeyondLoadBalancerId = utils.ConvertNullableString(lb.BeyondLoadBalancerId)
	base.BeyondLoadBalancerName = utils.ConvertNullableString(lb.BeyondLoadBalancerName)
	base.BeyondLoadBalancerDnsName = utils.ConvertNullableString(lb.BeyondLoadBalancerDnsName)
	base.TargetGroupCount = utils.ConvertNullableInt64(lb.TargetGroupCount)
	base.ListenerCount = utils.ConvertNullableInt64(lb.ListenerCount)
	base.PrivateVip = utils.ConvertNullableString(lb.PrivateVip)
	base.PublicVip = utils.ConvertNullableString(lb.PublicVip)
	base.SubnetName = utils.ConvertNullableString(lb.SubnetName)
	base.SubnetCidrBlock = utils.ConvertNullableString(lb.SubnetCidrBlock)
	base.VpcId = utils.ConvertNullableString(lb.VpcId)
	base.VpcName = utils.ConvertNullableString(lb.VpcName)
	base.SubnetId = utils.ConvertNullableString(lb.SubnetId)

	listenerIds, d := types.ListValueFrom(ctx, types.StringType, lb.ListenerIds)
	diags.Append(d...)
	base.ListenerIds = listenerIds

	if lb.AccessLogs.Get() == nil {
		base.AccessLogs = types.ObjectNull(accessLogResourceAttrType)
	} else {
		var accessLogs accessLogModel
		var convertDiags diag.Diagnostics

		if !base.AccessLogs.IsNull() {
			convertDiags = base.AccessLogs.As(ctx, &accessLogs, basetypes.ObjectAsOptions{})
			diags.Append(convertDiags...)
			if diags.HasError() {
				return false
			}
		}

		accessLogs.Bucket = types.StringValue(lb.AccessLogs.Get().Bucket)
		base.AccessLogs, convertDiags = types.ObjectValueFrom(ctx, accessLogResourceAttrType, accessLogs)
		diags.Append(convertDiags...)
		if diags.HasError() {
			return false
		}
	}

	return !diags.HasError()
}

func mapLoadBalancerBaseForDataSource(
	ctx context.Context,
	base *loadBalancerBaseModel,
	lb *loadbalancer.BnsLoadBalancerV1ApiGetLoadBalancerModelLoadBalancerModel,
	diags *diag.Diagnostics,
) bool {
	base.Id = types.StringValue(lb.Id)
	base.Name = utils.ConvertNullableString(lb.Name)
	base.Description = utils.ConvertNullableString(lb.Description)
	base.Type = utils.ConvertNullableString(lb.Type)
	base.ProjectId = utils.ConvertNullableString(lb.ProjectId)
	base.ProvisioningStatus = utils.ConvertNullableString(lb.ProvisioningStatus)
	base.OperatingStatus = utils.ConvertNullableString(lb.OperatingStatus)
	base.CreatedAt = utils.ConvertNullableTime(lb.CreatedAt)
	base.UpdatedAt = utils.ConvertNullableTime(lb.UpdatedAt)
	base.AvailabilityZone = utils.ConvertNullableString(lb.AvailabilityZone)

	accessLogsObj, accessLogsDiags := utils.ConvertObjectFromModel(ctx, lb.AccessLogs, accessLogDataSourceAttrType, func(src loadbalancer.AccessLogsModel) any {
		return accessLogDataSourceModel{
			Bucket: types.StringValue(src.Bucket),
		}
	})
	diags.Append(accessLogsDiags...)
	base.AccessLogs = accessLogsObj

	base.BeyondLoadBalancerId = utils.ConvertNullableString(lb.BeyondLoadBalancerId)
	base.BeyondLoadBalancerName = utils.ConvertNullableString(lb.BeyondLoadBalancerName)
	base.BeyondLoadBalancerDnsName = utils.ConvertNullableString(lb.BeyondLoadBalancerDnsName)
	base.TargetGroupCount = utils.ConvertNullableInt64(lb.TargetGroupCount)
	base.ListenerCount = utils.ConvertNullableInt64(lb.ListenerCount)
	base.PrivateVip = utils.ConvertNullableString(lb.PrivateVip)
	base.PublicVip = utils.ConvertNullableString(lb.PublicVip)
	base.SubnetName = utils.ConvertNullableString(lb.SubnetName)
	base.SubnetCidrBlock = utils.ConvertNullableString(lb.SubnetCidrBlock)
	base.VpcId = utils.ConvertNullableString(lb.VpcId)
	base.VpcName = utils.ConvertNullableString(lb.VpcName)
	base.SubnetId = utils.ConvertNullableString(lb.SubnetId)

	listenerIds, d := types.ListValueFrom(ctx, types.StringType, lb.ListenerIds)
	diags.Append(d...)
	base.ListenerIds = listenerIds

	return !diags.HasError()
}

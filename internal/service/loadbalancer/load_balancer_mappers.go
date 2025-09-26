// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package loadbalancer

import (
	"context"

	"terraform-provider-kakaocloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kakaoenterprise/kc-sdk-go/services/loadbalancer"
)

func mapAccessLogsToString(apiAccessLogs string) types.String {
	return types.StringValue(apiAccessLogs)
}

var accessLogAttrType = map[string]attr.Type{
	"bucket":     types.StringType,
	"access_key": types.StringType,
	"secret_key": types.StringType,
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

	return !diags.HasError()
}

func mapLoadBalancerBaseForDataSource(
	ctx context.Context,
	base *loadBalancerDataSourceBaseModel,
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
	base.AccessLogs = mapAccessLogsToString(utils.ConvertNullableString(lb.AccessLogs).ValueString())
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

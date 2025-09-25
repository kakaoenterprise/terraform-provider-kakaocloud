// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package loadbalancer

import (
	"terraform-provider-kakaocloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kakaoenterprise/kc-sdk-go/services/loadbalancer"
	"golang.org/x/net/context"
)

func mapBeyondLoadBalancerBaseModel(
	ctx context.Context,
	base *beyondLoadBalancerBaseModel,
	src *loadbalancer.BnsLoadBalancerV1ApiGetHaGroupModelBeyondLoadBalancerModel,
	diags *diag.Diagnostics,
) bool {

	loadBalancers, lbDiags := utils.ConvertListFromModel(ctx, src.LoadBalancers, blbLoadBalancerAttrType, func(lb loadbalancer.BnsLoadBalancerV1ApiGetHaGroupModelLoadBalancerModel) any {
		return blbLoadBalancerModel{
			Id:                 types.StringValue(lb.Id),
			Name:               utils.ConvertNullableString(lb.Name),
			Description:        utils.ConvertNullableString(lb.Description),
			Type:               utils.ConvertNullableString(lb.Type),
			ProvisioningStatus: utils.ConvertNullableString(lb.ProvisioningStatus),
			OperatingStatus:    utils.ConvertNullableString(lb.OperatingStatus),
			AvailabilityZone:   utils.ConvertNullableString(lb.AvailabilityZone),
			TypeId:             utils.ConvertNullableString(lb.TypeId),
			SubnetId:           utils.ConvertNullableString(lb.SubnetId),
			SubnetName:         utils.ConvertNullableString(lb.SubnetName),
			SubnetCidrBlock:    utils.ConvertNullableString(lb.SubnetCidrBlock),
			CreatedAt:          utils.ConvertNullableTime(lb.CreatedAt),
			UpdatedAt:          utils.ConvertNullableTime(lb.UpdatedAt),
		}
	})
	diags.Append(lbDiags...)

	base.Id = types.StringValue(src.Id)
	base.Name = utils.ConvertNullableString(src.Name)
	base.Description = utils.ConvertNullableString(src.Description)
	base.ProviderName = utils.ConvertNullableString(src.Provider)
	base.Scheme = utils.ConvertNullableString(src.Scheme)
	base.ProjectId = utils.ConvertNullableString(src.ProjectId)
	base.DnsName = utils.ConvertNullableString(src.DnsName)
	base.TypeId = utils.ConvertNullableString(src.TypeId)
	base.CreatedAt = utils.ConvertNullableTime(src.CreatedAt)
	base.UpdatedAt = utils.ConvertNullableTime(src.UpdatedAt)
	base.ProvisioningStatus = utils.ConvertNullableString(src.ProvisioningStatus)
	base.OperatingStatus = utils.ConvertNullableString(src.OperatingStatus)
	base.VpcId = utils.ConvertNullableString(src.VpcId)
	base.Type = utils.ConvertNullableString(src.Type)
	base.VpcName = utils.ConvertNullableString(src.VpcName)
	base.VpcCidrBlock = utils.ConvertNullableString(src.VpcCidrBlock)
	base.AvailabilityZones = utils.ConvertNullableStringList(src.AvailabilityZones)
	base.LoadBalancers = loadBalancers

	return !diags.HasError()
}

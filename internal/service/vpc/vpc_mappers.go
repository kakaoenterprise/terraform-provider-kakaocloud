// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package vpc

import (
	. "terraform-provider-kakaocloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework-nettypes/cidrtypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/kakaoenterprise/kc-sdk-go/v2/services/vpc"
	"golang.org/x/net/context"
)

func mapVpcModel(
	ctx context.Context,
	base *vpcBaseModel,
	vpcResult *vpc.Vpc,
	respDiags *diag.Diagnostics,
) bool {
	igw, igwDiags := ConvertObjectFromModel(ctx, vpcResult.Igw, igwAttrType, func(src vpc.VpcIgw) any {
		return igwModel{
			Id:                 ConvertNullableString(src.Id),
			Name:               ConvertNullableString(src.Name),
			Description:        ConvertNullableString(src.Description),
			Region:             ConvertNullableString(src.Region),
			NatIp:              ConvertNullableString(src.NatIp),
			ProjectId:          ConvertNullableString(src.ProjectId),
			OperatingStatus:    ConvertNullableString(src.OperatingStatus),
			ProvisioningStatus: ConvertNullableString(src.ProvisioningStatus),
			CreatedAt:          ConvertNullableTime(src.CreatedAt),
			UpdatedAt:          ConvertNullableTime(src.UpdatedAt),
		}
	})
	respDiags.Append(igwDiags...)

	routeTable, rtDiags := ConvertObjectFromModel(ctx, vpcResult.DefaultRouteTable, defaultRouteTableAttrType, func(src vpc.DefaultRouteTable) any {
		return defaultRouteTableModel{
			Id:                 ConvertNullableString(src.Id),
			Name:               ConvertNullableString(src.Name),
			Description:        ConvertNullableString(src.Description),
			ProvisioningStatus: ConvertNullableString(src.ProvisioningStatus),
			CreatedAt:          ConvertNullableTime(src.CreatedAt),
			UpdatedAt:          ConvertNullableTime(src.UpdatedAt),
		}
	})
	respDiags.Append(rtDiags...)

	base.Id = ConvertNullableString(vpcResult.Id)
	base.Name = ConvertNullableString(vpcResult.Name)
	base.Description = ConvertNullableString(vpcResult.Description)
	base.Region = ConvertNullableString(vpcResult.Region)
	base.ProjectId = ConvertNullableString(vpcResult.ProjectId)
	base.ProjectName = ConvertNullableString(vpcResult.ProjectName)
	base.CidrBlock = cidrtypes.NewIPPrefixValue(*vpcResult.CidrBlock.Get())
	base.IsDefault = ConvertNullableBool(vpcResult.IsDefault)
	base.ProvisioningStatus = ConvertNullableString(vpcResult.ProvisioningStatus)
	base.IsEnableDnsSupport = ConvertNullableBool(vpcResult.IsEnableDnsSupport)
	base.CreatedAt = ConvertNullableTime(vpcResult.CreatedAt)
	base.UpdatedAt = ConvertNullableTime(vpcResult.UpdatedAt)
	base.Igw = igw
	base.DefaultRouteTable = routeTable

	if respDiags.HasError() {
		return false
	}

	return true
}

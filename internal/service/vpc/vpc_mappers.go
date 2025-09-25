// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpc

import (
	"github.com/hashicorp/terraform-plugin-framework-nettypes/cidrtypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kakaoenterprise/kc-sdk-go/services/vpc"
	"golang.org/x/net/context"
	. "terraform-provider-kakaocloud/internal/utils"
)

func mapVpcBaseModel(
	ctx context.Context,
	base *vpcBaseModel,
	vpcResult *vpc.BnsVpcV1ApiGetVpcModelVpcModel,
	respDiags *diag.Diagnostics,
) bool {
	igw, igwDiags := ConvertObjectFromModel(ctx, vpcResult.Igw, igwAttrType, func(src vpc.BnsVpcV1ApiGetVpcModelIgwModel) any {
		return igwModel{
			Id:                 types.StringValue(src.Id),
			Name:               ConvertNullableString(src.Name),
			Description:        ConvertNullableString(src.Description),
			Region:             ConvertNullableString(src.Region),
			ProjectId:          ConvertNullableString(src.ProjectId),
			OperatingStatus:    ConvertNullableString(src.OperatingStatus),
			ProvisioningStatus: ConvertNullableString(src.ProvisioningStatus),
			CreatedAt:          ConvertNullableTime(src.CreatedAt),
			UpdatedAt:          ConvertNullableTime(src.UpdatedAt),
		}
	})
	respDiags.Append(igwDiags...)

	routeTable, rtDiags := ConvertObjectFromModel(ctx, vpcResult.DefaultRouteTable, defaultRouteTableAttrType, func(src vpc.BnsVpcV1ApiGetVpcModelRouteTableModel) any {
		return defaultRouteTableModel{
			Id:                 types.StringValue(src.Id),
			Name:               ConvertNullableString(src.Name),
			Description:        ConvertNullableString(src.Description),
			ProvisioningStatus: ConvertNullableString(src.ProvisioningStatus),
			CreatedAt:          ConvertNullableTime(src.CreatedAt),
			UpdatedAt:          ConvertNullableTime(src.UpdatedAt),
		}
	})
	respDiags.Append(rtDiags...)

	base.Id = types.StringValue(vpcResult.Id)
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

func mapVpcListModel(
	ctx context.Context,
	base *vpcBaseModel,
	vpcResult *vpc.BnsVpcV1ApiListVpcsModelVpcModel,
	respDiags *diag.Diagnostics,
) bool {
	igw, igwDiags := ConvertObjectFromModel(ctx, vpcResult.Igw, igwAttrType, func(src vpc.BnsVpcV1ApiListVpcsModelIgwModel) any {
		return igwModel{
			Id:                 types.StringValue(src.Id),
			Name:               ConvertNullableString(src.Name),
			Description:        ConvertNullableString(src.Description),
			Region:             ConvertNullableString(src.Region),
			ProjectId:          ConvertNullableString(src.ProjectId),
			OperatingStatus:    ConvertNullableString(src.OperatingStatus),
			ProvisioningStatus: ConvertNullableString(src.ProvisioningStatus),
			CreatedAt:          ConvertNullableTime(src.CreatedAt),
			UpdatedAt:          ConvertNullableTime(src.UpdatedAt),
		}
	})
	respDiags.Append(igwDiags...)

	routeTable, rtDiags := ConvertObjectFromModel(ctx, vpcResult.DefaultRouteTable, defaultRouteTableAttrType, func(src vpc.BnsVpcV1ApiListVpcsModelRouteTableModel) any {
		return defaultRouteTableModel{
			Id:                 types.StringValue(src.Id),
			Name:               ConvertNullableString(src.Name),
			Description:        ConvertNullableString(src.Description),
			ProvisioningStatus: ConvertNullableString(src.ProvisioningStatus),
			CreatedAt:          ConvertNullableTime(src.CreatedAt),
			UpdatedAt:          ConvertNullableTime(src.UpdatedAt),
		}
	})
	respDiags.Append(rtDiags...)

	base.Id = types.StringValue(vpcResult.Id)
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

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpc

import (
	. "terraform-provider-kakaocloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kakaoenterprise/kc-sdk-go/services/vpc"
	"golang.org/x/net/context"
)

func mapRouteTableBaseModel(
	ctx context.Context,
	base *routeTableBaseModel,
	result *vpc.BnsVpcV1ApiGetRouteTableModelRouteTableModel,
	diags *diag.Diagnostics,

) bool {

	associations, assocDiags := ConvertListFromModel(ctx, result.Associations, routeTableAssociationAttrType, func(item vpc.BnsVpcV1ApiGetRouteTableModelAssociationModel) any {
		return routeTableAssociationModel{
			Id:                 types.StringValue(item.Id),
			ProvisioningStatus: ConvertNullableString(item.ProvisioningStatus),
			VpcId:              ConvertNullableString(item.VpcId),
			VpcName:            ConvertNullableString(item.VpcName),
			SubnetId:           ConvertNullableString(item.SubnetId),
			SubnetName:         ConvertNullableString(item.SubnetName),
			SubnetCidrBlock:    ConvertNullableString(item.SubnetCidrBlock),
			AvailabilityZone:   ConvertNullableString(item.AvailabilityZone),
		}
	})
	diags.Append(assocDiags...)

	routes, routeDiags := ConvertListFromModel(ctx, result.Routes, routeTableRouteAttrType, func(item vpc.BnsVpcV1ApiGetRouteTableModelRouteModel) any {
		return routeTableRouteModel{
			Id:                 types.StringValue(item.Id),
			Destination:        ConvertNullableString(item.Destination),
			ProvisioningStatus: ConvertNullableString(item.ProvisioningStatus),
			TargetType:         ConvertNullableString(item.TargetType),
			TargetName:         ConvertNullableString(item.TargetName),
			IsLocalRoute:       ConvertNullableBool(item.IsLocalRoute),
			TargetId:           ConvertNullableString(item.TargetId),
		}
	})
	diags.Append(routeDiags...)

	base.Id = types.StringValue(result.Id)
	base.Name = ConvertNullableString(result.Name)
	base.Associations = associations
	base.Routes = routes
	base.VpcId = ConvertNullableString(result.VpcId)
	base.ProvisioningStatus = ConvertNullableString(result.ProvisioningStatus)
	base.VpcName = ConvertNullableString(result.VpcName)
	base.VpcProvisioningStatus = ConvertNullableString(result.VpcProvisioningStatus)
	base.ProjectId = ConvertNullableString(result.ProjectId)
	base.ProjectName = ConvertNullableString(result.ProjectName)
	base.IsMain = ConvertNullableBool(result.IsMain)
	base.CreatedAt = ConvertNullableTime(result.CreatedAt)
	base.UpdatedAt = ConvertNullableTime(result.UpdatedAt)

	return !diags.HasError()
}

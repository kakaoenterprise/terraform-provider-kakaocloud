// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package vpc

import (
	. "terraform-provider-kakaocloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/kakaoenterprise/kc-sdk-go/v2/services/vpc"
	"golang.org/x/net/context"
)

func mapRouteTableModel(
	ctx context.Context,
	base *routeTableBaseModel,
	result *vpc.VpcRouteTable,
	diags *diag.Diagnostics,
) bool {

	associations, assocDiags := ConvertListFromModel(ctx, result.Associations, routeTableAssociationAttrType, func(item vpc.VpcAssociation) any {
		return routeTableAssociationModel{
			Id:                 ConvertNullableString(item.Id),
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

	routes, routeDiags := ConvertListFromModel(ctx, result.Routes, routeTableRouteAttrType, func(item vpc.VpcRoute) any {
		return routeTableRouteModel{
			Id:                 ConvertNullableString(item.Id),
			Destination:        ConvertNullableString(item.Destination),
			ProvisioningStatus: ConvertNullableString(item.ProvisioningStatus),
			TargetType:         ConvertNullableString(item.TargetType),
			TargetName:         ConvertNullableString(item.TargetName),
			IsLocalRoute:       ConvertNullableBool(item.IsLocalRoute),
			TargetId:           ConvertNullableString(item.TargetId),
		}
	})
	diags.Append(routeDiags...)

	base.Id = ConvertNullableString(result.Id)
	base.Name = ConvertNullableString(result.Name)
	base.Description = ConvertNullableString(result.Description)
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

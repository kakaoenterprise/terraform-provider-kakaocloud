// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package tgw

import (
	"context"

	. "terraform-provider-kakaocloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kakaoenterprise/kc-sdk-go/v2/services/tgw"
)

func mapTransitGatewayRouteTableBaseModel(
	base *transitGatewayRouteTableBaseModel,
	result *tgw.TgwRouteTable,
	diags *diag.Diagnostics,
) bool {
	base.Id = ConvertNullableString(result.Id)
	base.Name = ConvertNullableString(result.Name)
	base.TgwId = ConvertNullableString(result.TgwId)
	base.Region = ConvertNullableString(result.Region)
	base.ProjectId = ConvertNullableString(result.ProjectId)
	base.ProjectName = ConvertNullableString(result.ProjectName)
	base.TgwName = ConvertNullableString(result.TgwName)
	base.IsDefaultAssociationRouteTable = ConvertNullableBool(result.IsDefaultAssociationRouteTable)
	base.IsDefaultPropagationRouteTable = ConvertNullableBool(result.IsDefaultPropagationRouteTable)
	base.ProvisioningStatus = ConvertNullableString(result.ProvisioningStatus)
	base.CreatedAt = ConvertNullableTime(result.CreatedAt)
	base.UpdatedAt = ConvertNullableTime(result.UpdatedAt)

	return !diags.HasError()
}

func mapTransitGatewayRouteTableResourceModel(
	ctx context.Context,
	base *transitGatewayRouteTableResourceModel,
	result *tgw.TgwRouteTable,
	diags *diag.Diagnostics,
) bool {
	mapTransitGatewayRouteTableBaseModel(&base.transitGatewayRouteTableBaseModel, result, diags)

	var associations []tgwRouteTableRequestAssociationModel
	if result.Associations != nil && len(result.Associations) > 0 {
		for _, association := range result.Associations {
			associations = append(associations,
				tgwRouteTableRequestAssociationModel{
					Id:              ConvertNullableString(association.Id),
					TgwAttachmentId: ConvertNullableString(association.ResourceAttachmentId),
				})
		}
	}
	associationsElemType := types.ObjectType{AttrTypes: tgwRouteTableRequestAssociationAttrType}
	var mapDiags diag.Diagnostics
	base.Associations, mapDiags = types.ListValueFrom(ctx, associationsElemType, associations)
	diags.Append(mapDiags...)
	if diags.HasError() {
		return false
	}

	return !diags.HasError()
}

func mapTransitGatewayRouteTableDataSourceModel(
	ctx context.Context,
	base *transitGatewayRouteTableDataSourceBaseModel,
	result *tgw.TgwRouteTable,
	diags *diag.Diagnostics,
) bool {
	mapTransitGatewayRouteTableBaseModel(&base.transitGatewayRouteTableBaseModel, result, diags)

	routes, routeDiags := ConvertListFromModel(ctx, result.Routes, tgwRouteTableRouteAttrType,
		func(route tgw.Route) any {
			resourceObj, resourceDiags := mapTgwRouteTableResourceObject(ctx, route.Resource)
			diags.Append(resourceDiags...)

			return tgwRouteTableRouteNestedModel{
				Id:                   ConvertNullableString(route.Id),
				RouteType:            ConvertNullableString(route.RouteType),
				DestinationCidrBlock: ConvertNullableString(route.DestinationCidrBlock),
				ResourceAttachmentId: ConvertNullableString(route.ResourceAttachmentId),
				ResourceId:           ConvertNullableString(route.ResourceId),
				ResourceType:         ConvertNullableString(route.ResourceType),
				TgwRouteTableId:      ConvertNullableString(route.TgwRouteTableId),
				ProvisioningStatus:   ConvertNullableString(route.ProvisioningStatus),
				Resource:             resourceObj,
			}
		})
	diags.Append(routeDiags...)

	associations, assocDiags := ConvertListFromModel(ctx, result.Associations, tgwRouteTableAssociationAttrType,
		func(assoc tgw.Association) any {
			resourceObj, resourceDiags := mapTgwRouteTableResourceObject(ctx, assoc.Resource)
			diags.Append(resourceDiags...)

			return tgwRouteTableAssociationNestedModel{
				Id:                   ConvertNullableString(assoc.Id),
				ResourceAttachmentId: ConvertNullableString(assoc.ResourceAttachmentId),
				ResourceId:           ConvertNullableString(assoc.ResourceId),
				ResourceType:         ConvertNullableString(assoc.ResourceType),
				TgwRouteTableId:      ConvertNullableString(assoc.TgwRouteTableId),
				ProvisioningStatus:   ConvertNullableString(assoc.ProvisioningStatus),
				Resource:             resourceObj,
			}
		})
	diags.Append(assocDiags...)

	base.Routes = routes
	base.Associations = associations

	return !diags.HasError()
}

func mapTgwRouteTableResourceObject(ctx context.Context, resource tgw.NullableResource) (types.Object, diag.Diagnostics) {
	return ConvertObjectFromModel(ctx, resource, tgwRouteTableResourceAttrType, func(src tgw.Resource) any {
		return tgwRouteTableResourceNestedModel{
			Id:                 ConvertNullableString(src.Id),
			Name:               ConvertNullableString(src.Name),
			CidrBlock:          ConvertNullableString(src.CidrBlock),
			ProjectId:          ConvertNullableString(src.ProjectId),
			ProjectName:        ConvertNullableString(src.ProjectName),
			ProvisioningStatus: ConvertNullableString(src.ProvisioningStatus),
		}
	})
}

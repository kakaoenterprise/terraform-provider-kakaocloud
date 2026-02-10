// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package tgw

import (
	"context"
	. "terraform-provider-kakaocloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kakaoenterprise/kc-sdk-go/services/tgw"
)

func mapTransitGatewayBaseModel(
	ctx context.Context,
	base *transitGatewayBaseModel,
	tgwResult *tgw.BnsTgwV1ApiGetTransitGatewayModelTgwResponseModel,
	respDiags *diag.Diagnostics,
) bool {
	options, optionsDiags := ConvertObjectFromModel(ctx, tgwResult.Options, tgwOptionsAttrType, func(src tgw.OptionResponseModel) any {
		return tgwOptionsModel{
			IsAutoAcceptSharedAttachments:  ConvertNullableBool(src.IsAutoAcceptSharedAttachments),
			IsDefaultRouteTableAssociation: ConvertNullableBool(src.IsDefaultRouteTableAssociation),
			AssociationDefaultRouteTableId: ConvertNullableString(src.AssociationDefaultRouteTableId),
		}
	})
	respDiags.Append(optionsDiags...)

	base.Id = types.StringValue(*tgwResult.Id.Get())
	base.Name = ConvertNullableString(tgwResult.Name)
	base.Region = ConvertNullableString(tgwResult.Region)
	base.IsShared = ConvertNullableBool(tgwResult.IsShared)

	attachmentsList, attachmentsDiags := ConvertListFromModel(ctx, tgwResult.Attachments, tgwAttachmentAttrType, func(att tgw.BnsTgwV1ApiGetTransitGatewayModelAttachmentResponseModel) any {
		return tgwAttachmentNestedModel{
			Id:                 ConvertNullableString(att.Id),
			ResourceType:       ConvertNullableString(att.ResourceType),
			ResourceId:         ConvertNullableString(att.ResourceId),
			ResourceName:       ConvertNullableString(att.ResourceName),
			TgwId:              ConvertNullableString(att.TgwId),
			ProvisioningStatus: ConvertNullableString(att.ProvisioningStatus),
			CreatedAt:          ConvertNullableTime(att.CreatedAt),
			UpdatedAt:          ConvertNullableTime(att.UpdatedAt),
		}
	})
	respDiags.Append(attachmentsDiags...)
	base.Attachments = attachmentsList

	base.Options = options
	base.ProvisioningStatus = ConvertNullableString(tgwResult.ProvisioningStatus)
	base.ProjectId = ConvertNullableString(tgwResult.ProjectId)
	base.ProjectName = ConvertNullableString(tgwResult.ProjectName)
	base.OwnerProjectId = ConvertNullableString(tgwResult.OwnerProjectId)
	base.OwnerProjectName = ConvertNullableString(tgwResult.OwnerProjectName)

	routeTablesList, routeTablesDiags := ConvertListFromModel(ctx, tgwResult.RouteTables, tgwRouteTableAttrType, func(rt tgw.BnsTgwV1ApiGetTransitGatewayModelRouteTableResponseModel) any {
		return tgwRouteTableNestedModel{
			Id:                             ConvertNullableString(rt.Id),
			Name:                           ConvertNullableString(rt.Name),
			Region:                         ConvertNullableString(rt.Region),
			ProjectId:                      ConvertNullableString(rt.ProjectId),
			TgwId:                          ConvertNullableString(rt.TgwId),
			IsDefaultAssociationRouteTable: ConvertNullableBool(rt.IsDefaultAssociationRouteTable),
			IsDefaultPropagationRouteTable: ConvertNullableBool(rt.IsDefaultPropagationRouteTable),
			ProvisioningStatus:             ConvertNullableString(rt.ProvisioningStatus),
			CreatedAt:                      ConvertNullableTime(rt.CreatedAt),
			UpdatedAt:                      ConvertNullableTime(rt.UpdatedAt),
		}
	})
	respDiags.Append(routeTablesDiags...)
	base.RouteTables = routeTablesList

	base.CreatedAt = ConvertNullableTime(tgwResult.CreatedAt)
	base.UpdatedAt = ConvertNullableTime(tgwResult.UpdatedAt)

	return !respDiags.HasError()
}

func mapTransitGatewayListModel(
	ctx context.Context,
	base *transitGatewayBaseModel,
	tgwResult *tgw.BnsTgwV1ApiListTransitGatewaysModelTgwResponseModel,
	respDiags *diag.Diagnostics,
) bool {
	options, optionsDiags := ConvertObjectFromModel(ctx, tgwResult.Options, tgwOptionsAttrType, func(src tgw.OptionResponseModel) any {
		return tgwOptionsModel{
			IsAutoAcceptSharedAttachments:  ConvertNullableBool(src.IsAutoAcceptSharedAttachments),
			IsDefaultRouteTableAssociation: ConvertNullableBool(src.IsDefaultRouteTableAssociation),
			AssociationDefaultRouteTableId: ConvertNullableString(src.AssociationDefaultRouteTableId),
		}
	})
	respDiags.Append(optionsDiags...)

	base.Id = types.StringValue(*tgwResult.Id.Get())
	base.Name = ConvertNullableString(tgwResult.Name)
	base.Region = ConvertNullableString(tgwResult.Region)
	base.IsShared = ConvertNullableBool(tgwResult.IsShared)

	attachmentsList, attachmentsDiags := ConvertListFromModel(ctx, tgwResult.Attachments, tgwAttachmentAttrType, func(att tgw.BnsTgwV1ApiListTransitGatewaysModelAttachmentResponseModel) any {
		return tgwAttachmentNestedModel{
			Id:                 ConvertNullableString(att.Id),
			ResourceType:       ConvertNullableString(att.ResourceType),
			ResourceId:         ConvertNullableString(att.ResourceId),
			ResourceName:       ConvertNullableString(att.ResourceName),
			TgwId:              ConvertNullableString(att.TgwId),
			ProvisioningStatus: ConvertNullableString(att.ProvisioningStatus),
			CreatedAt:          ConvertNullableTime(att.CreatedAt),
			UpdatedAt:          ConvertNullableTime(att.UpdatedAt),
		}
	})
	respDiags.Append(attachmentsDiags...)
	base.Attachments = attachmentsList

	base.Options = options
	base.ProvisioningStatus = ConvertNullableString(tgwResult.ProvisioningStatus)
	base.ProjectId = ConvertNullableString(tgwResult.ProjectId)
	base.ProjectName = ConvertNullableString(tgwResult.ProjectName)
	base.OwnerProjectId = ConvertNullableString(tgwResult.OwnerProjectId)
	base.OwnerProjectName = ConvertNullableString(tgwResult.OwnerProjectName)

	routeTablesList, routeTablesDiags := ConvertListFromModel(ctx, tgwResult.RouteTables, tgwRouteTableAttrType, func(rt tgw.BnsTgwV1ApiListTransitGatewaysModelRouteTableResponseModel) any {
		return tgwRouteTableNestedModel{
			Id:                             ConvertNullableString(rt.Id),
			Name:                           ConvertNullableString(rt.Name),
			Region:                         ConvertNullableString(rt.Region),
			ProjectId:                      ConvertNullableString(rt.ProjectId),
			TgwId:                          ConvertNullableString(rt.TgwId),
			IsDefaultAssociationRouteTable: ConvertNullableBool(rt.IsDefaultAssociationRouteTable),
			IsDefaultPropagationRouteTable: ConvertNullableBool(rt.IsDefaultPropagationRouteTable),
			ProvisioningStatus:             ConvertNullableString(rt.ProvisioningStatus),
			CreatedAt:                      ConvertNullableTime(rt.CreatedAt),
			UpdatedAt:                      ConvertNullableTime(rt.UpdatedAt),
		}
	})
	respDiags.Append(routeTablesDiags...)
	base.RouteTables = routeTablesList

	base.CreatedAt = ConvertNullableTime(tgwResult.CreatedAt)
	base.UpdatedAt = ConvertNullableTime(tgwResult.UpdatedAt)

	return !respDiags.HasError()
}

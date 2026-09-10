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

func mapTransitGatewayResourceBaseModel(
	ctx context.Context,
	base *transitGatewayResourceBaseModel,
	tgwResult *tgw.Tgw,
	respDiags *diag.Diagnostics,
) bool {
	options, optionsDiags := ConvertObjectFromModel(ctx, tgwResult.Options, tgwOptionsAttrType, func(src tgw.Option) any {
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

	base.Options = options
	base.ProvisioningStatus = ConvertNullableString(tgwResult.ProvisioningStatus)
	base.ProjectId = ConvertNullableString(tgwResult.ProjectId)
	base.ProjectName = ConvertNullableString(tgwResult.ProjectName)
	base.OwnerProjectId = ConvertNullableString(tgwResult.OwnerProjectId)
	base.OwnerProjectName = ConvertNullableString(tgwResult.OwnerProjectName)

	base.CreatedAt = ConvertNullableTime(tgwResult.CreatedAt)
	base.UpdatedAt = ConvertNullableTime(tgwResult.UpdatedAt)

	return !respDiags.HasError()
}

func mapTransitGatewayModel(
	ctx context.Context,
	base *transitGatewayBaseModel,
	tgwResult *tgw.Tgw,
	respDiags *diag.Diagnostics,
) bool {
	mapTransitGatewayResourceBaseModel(ctx, &base.transitGatewayResourceBaseModel, tgwResult, respDiags)

	attachmentsList, attachmentsDiags := ConvertListFromModel(ctx, tgwResult.Attachments, tgwAttachmentAttrType, func(att tgw.Attachment) any {
		return tgwAttachmentNestedModel{
			Id:                 ConvertNullableString(att.Id),
			Name:               ConvertNullableString(att.Name),
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

	routeTablesList, routeTablesDiags := ConvertListFromModel(ctx, tgwResult.RouteTables, tgwRouteTableAttrType, func(rt tgw.RouteTable) any {
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

	return !respDiags.HasError()
}

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package tgw

import (
	"context"
	"terraform-provider-kakaocloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kakaoenterprise/kc-sdk-go/services/tgw"
)

func mapTransitGatewayAttachmentModelFromGet(
	ctx context.Context,
	base *transitGatewayAttachmentBaseModel,
	attachResult *tgw.BnsTgwV1ApiGetTgwAttachmentModelTgwAttachmentResponseModel,
	respDiags *diag.Diagnostics,
) bool {
	base.Id = types.StringValue(attachResult.Id)
	base.ProvisioningStatus = utils.ConvertNullableString(attachResult.ProvisioningStatus)
	base.ProjectId = utils.ConvertNullableString(attachResult.ProjectId)
	base.ProjectName = utils.ConvertNullableString(attachResult.ProjectName)
	base.ResourceType = utils.ConvertNullableString(attachResult.ResourceType)
	base.ResourceId = utils.ConvertNullableString(attachResult.ResourceId)
	base.ResourceName = utils.ConvertNullableString(attachResult.ResourceName)
	base.ResourceCidrBlock = utils.ConvertNullableString(attachResult.ResourceCidrBlock)

	tgwModel := tgwAttachmentTgwNestedModel{
		Id:          utils.ConvertNullableString(attachResult.Tgw.Id),
		Name:        utils.ConvertNullableString(attachResult.Tgw.Name),
		ProjectId:   utils.ConvertNullableString(attachResult.Tgw.ProjectId),
		ProjectName: utils.ConvertNullableString(attachResult.Tgw.ProjectName),
	}

	tgwObj, tgwDiags := types.ObjectValueFrom(ctx, tgwAttachmentTgwAttrType, tgwModel)
	respDiags.Append(tgwDiags...)
	base.Tgw = tgwObj

	resourcesList, resourcesDiags := utils.ConvertListFromModel(ctx, attachResult.Resources, tgwAttachmentResourceAttrType, func(res tgw.BnsTgwV1ApiGetTgwAttachmentModelResourceResponseModel) any {
		return tgwAttachmentResourceNestedModel{
			Id:                 utils.ConvertNullableString(res.Id),
			Name:               utils.ConvertNullableString(res.Name),
			Description:        utils.ConvertNullableString(res.Description),
			AvailabilityZone:   utils.ConvertNullableString(res.AvailabilityZone),
			CidrBlock:          utils.ConvertNullableString(res.CidrBlock),
			OperatingStatus:    utils.ConvertNullableString(res.OperatingStatus),
			ProvisioningStatus: utils.ConvertNullableString(res.ProvisioningStatus),
			VpcId:              utils.ConvertNullableString(res.VpcId),
			CreatedAt:          utils.ConvertNullableTime(res.CreatedAt),
			UpdatedAt:          utils.ConvertNullableTime(res.UpdatedAt),
		}
	})
	respDiags.Append(resourcesDiags...)
	base.Resources = resourcesList

	if attachResult.RouteTable.IsSet() && attachResult.RouteTable.Get() != nil {
		rt := attachResult.RouteTable.Get()
		rtModel := tgwAttachmentRouteTableNestedModel{
			Id:                             utils.ConvertNullableString(rt.Id),
			Name:                           utils.ConvertNullableString(rt.Name),
			ProjectId:                      utils.ConvertNullableString(rt.ProjectId),
			Region:                         utils.ConvertNullableString(rt.Region),
			TgwId:                          utils.ConvertNullableString(rt.TgwId),
			IsDefaultAssociationRouteTable: utils.ConvertNullableBool(rt.IsDefaultAssociationRouteTable),
			IsDefaultPropagationRouteTable: utils.ConvertNullableBool(rt.IsDefaultPropagationRouteTable),
			ProvisioningStatus:             utils.ConvertNullableString(rt.ProvisioningStatus),
			CreatedAt:                      utils.ConvertNullableTime(rt.CreatedAt),
			UpdatedAt:                      utils.ConvertNullableTime(rt.UpdatedAt),
		}
		rtObj, rtDiags := types.ObjectValueFrom(ctx, tgwAttachmentRouteTableAttrType, rtModel)
		respDiags.Append(rtDiags...)
		base.RouteTable = rtObj
	} else {
		base.RouteTable = types.ObjectNull(tgwAttachmentRouteTableAttrType)
	}

	base.CreatedAt = utils.ConvertNullableTime(attachResult.CreatedAt)
	base.UpdatedAt = utils.ConvertNullableTime(attachResult.UpdatedAt)

	return !respDiags.HasError()
}

func mapTransitGatewayAttachmentModelFromList(
	ctx context.Context,
	base *transitGatewayAttachmentBaseModel,
	attachResult *tgw.BnsTgwV1ApiListTgwAttachmentsModelTgwAttachmentResponseModel,
	respDiags *diag.Diagnostics,
) bool {
	base.Id = types.StringValue(attachResult.Id)
	base.ProvisioningStatus = utils.ConvertNullableString(attachResult.ProvisioningStatus)
	base.ProjectId = utils.ConvertNullableString(attachResult.ProjectId)
	base.ProjectName = utils.ConvertNullableString(attachResult.ProjectName)
	base.ResourceType = utils.ConvertNullableString(attachResult.ResourceType)
	base.ResourceId = utils.ConvertNullableString(attachResult.ResourceId)
	base.ResourceName = utils.ConvertNullableString(attachResult.ResourceName)
	base.ResourceCidrBlock = utils.ConvertNullableString(attachResult.ResourceCidrBlock)

	tgwModel := tgwAttachmentTgwNestedModel{
		Id:          utils.ConvertNullableString(attachResult.Tgw.Id),
		Name:        utils.ConvertNullableString(attachResult.Tgw.Name),
		ProjectId:   utils.ConvertNullableString(attachResult.Tgw.ProjectId),
		ProjectName: utils.ConvertNullableString(attachResult.Tgw.ProjectName),
	}

	tgwObj, tgwDiags := types.ObjectValueFrom(ctx, tgwAttachmentTgwAttrType, tgwModel)
	respDiags.Append(tgwDiags...)
	base.Tgw = tgwObj

	resourcesList, resourcesDiags := utils.ConvertListFromModel(ctx, attachResult.Resources, tgwAttachmentResourceAttrType, func(res tgw.BnsTgwV1ApiListTgwAttachmentsModelResourceResponseModel) any {
		return tgwAttachmentResourceNestedModel{
			Id:                 utils.ConvertNullableString(res.Id),
			Name:               utils.ConvertNullableString(res.Name),
			Description:        utils.ConvertNullableString(res.Description),
			AvailabilityZone:   utils.ConvertNullableString(res.AvailabilityZone),
			CidrBlock:          utils.ConvertNullableString(res.CidrBlock),
			OperatingStatus:    utils.ConvertNullableString(res.OperatingStatus),
			ProvisioningStatus: utils.ConvertNullableString(res.ProvisioningStatus),
			VpcId:              utils.ConvertNullableString(res.VpcId),
			CreatedAt:          utils.ConvertNullableTime(res.CreatedAt),
			UpdatedAt:          utils.ConvertNullableTime(res.UpdatedAt),
		}
	})
	respDiags.Append(resourcesDiags...)
	base.Resources = resourcesList

	if attachResult.RouteTable.IsSet() && attachResult.RouteTable.Get() != nil {
		rt := attachResult.RouteTable.Get()
		rtModel := tgwAttachmentRouteTableNestedModel{
			Id:                             utils.ConvertNullableString(rt.Id),
			Name:                           utils.ConvertNullableString(rt.Name),
			ProjectId:                      utils.ConvertNullableString(rt.ProjectId),
			Region:                         utils.ConvertNullableString(rt.Region),
			TgwId:                          utils.ConvertNullableString(rt.TgwId),
			IsDefaultAssociationRouteTable: utils.ConvertNullableBool(rt.IsDefaultAssociationRouteTable),
			IsDefaultPropagationRouteTable: utils.ConvertNullableBool(rt.IsDefaultPropagationRouteTable),
			ProvisioningStatus:             utils.ConvertNullableString(rt.ProvisioningStatus),
			CreatedAt:                      utils.ConvertNullableTime(rt.CreatedAt),
			UpdatedAt:                      utils.ConvertNullableTime(rt.UpdatedAt),
		}
		rtObj, rtDiags := types.ObjectValueFrom(ctx, tgwAttachmentRouteTableAttrType, rtModel)
		respDiags.Append(rtDiags...)
		base.RouteTable = rtObj
	} else {
		base.RouteTable = types.ObjectNull(tgwAttachmentRouteTableAttrType)
	}

	base.CreatedAt = utils.ConvertNullableTime(attachResult.CreatedAt)
	base.UpdatedAt = utils.ConvertNullableTime(attachResult.UpdatedAt)

	return !respDiags.HasError()
}

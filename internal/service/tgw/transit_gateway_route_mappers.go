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

func mapTransitGatewayRouteListModel(
	ctx context.Context,
	base *transitGatewayRouteBaseModel,
	routeResult *tgw.BnsTgwV1ApiListTgwRoutesModelRouteResponseModel,
	routeTableId string,
	respDiags *diag.Diagnostics,
) bool {
	base.Id = utils.ConvertNullableString(routeResult.Id)
	base.RouteTableId = types.StringValue(routeTableId)
	base.RouteType = utils.ConvertNullableString(routeResult.RouteType)
	base.DestinationCidrBlock = utils.ConvertNullableString(routeResult.DestinationCidrBlock)
	base.ResourceAttachmentId = utils.ConvertNullableString(routeResult.ResourceAttachmentId)
	base.ResourceId = utils.ConvertNullableString(routeResult.ResourceId)
	base.ResourceType = utils.ConvertNullableString(routeResult.ResourceType)
	base.TgwRouteTableId = utils.ConvertNullableString(routeResult.TgwRouteTableId)
	base.ProvisioningStatus = utils.ConvertNullableString(routeResult.ProvisioningStatus)

	resource := routeResult.Resource
	resourceModel := transitGatewayRouteResourceNestedModel{
		Id:                 utils.ConvertNullableString(resource.Id),
		Name:               utils.ConvertNullableString(resource.Name),
		CidrBlock:          utils.ConvertNullableString(resource.CidrBlock),
		ProjectId:          utils.ConvertNullableString(resource.ProjectId),
		ProjectName:        utils.ConvertNullableString(resource.ProjectName),
		ProvisioningStatus: utils.ConvertNullableString(resource.ProvisioningStatus),
	}

	resourceObj, diags := types.ObjectValueFrom(ctx, transitGatewayRouteResourceAttrType, resourceModel)
	respDiags.Append(diags...)
	base.Resource = resourceObj

	return !respDiags.HasError()
}

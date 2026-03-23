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

func mapTransitGatewayRouteTableAssociationBaseModel(
	ctx context.Context,
	base *transitGatewayRouteTableAssociationBaseModel,
	assocResult *tgw.BnsTgwV1ApiListTgwRouteTableAssociationsModelAssociationResponseModel,
	routeTableId string,
	respDiags *diag.Diagnostics,
) bool {
	base.Id = utils.ConvertNullableString(assocResult.Id)
	base.RouteTableId = types.StringValue(routeTableId)
	base.ResourceAttachmentId = utils.ConvertNullableString(assocResult.ResourceAttachmentId)
	base.ResourceId = utils.ConvertNullableString(assocResult.ResourceId)
	base.ResourceType = utils.ConvertNullableString(assocResult.ResourceType)
	base.TgwRouteTableId = utils.ConvertNullableString(assocResult.TgwRouteTableId)
	base.ProvisioningStatus = utils.ConvertNullableString(assocResult.ProvisioningStatus)

	resource := assocResult.Resource
	resourceModel := tgwRouteTableResourceNestedModel{
		Id:                 utils.ConvertNullableString(resource.Id),
		Name:               utils.ConvertNullableString(resource.Name),
		CidrBlock:          utils.ConvertNullableString(resource.CidrBlock),
		ProjectId:          utils.ConvertNullableString(resource.ProjectId),
		ProjectName:        utils.ConvertNullableString(resource.ProjectName),
		ProvisioningStatus: utils.ConvertNullableString(resource.ProvisioningStatus),
	}

	resourceObj, diags := types.ObjectValueFrom(ctx, tgwRouteTableResourceAttrType, resourceModel)
	respDiags.Append(diags...)
	base.Resource = resourceObj

	return !respDiags.HasError()
}

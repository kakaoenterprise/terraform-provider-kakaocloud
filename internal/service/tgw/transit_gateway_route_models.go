// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package tgw

import (
	"terraform-provider-kakaocloud/internal/common"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type transitGatewayRouteResourceNestedModel struct {
	Id                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	CidrBlock          types.String `tfsdk:"cidr_block"`
	ProjectId          types.String `tfsdk:"project_id"`
	ProjectName        types.String `tfsdk:"project_name"`
	ProvisioningStatus types.String `tfsdk:"provisioning_status"`
}

type transitGatewayRouteBaseModel struct {
	Id                   types.String `tfsdk:"id"`
	RouteTableId         types.String `tfsdk:"route_table_id"`
	RouteType            types.String `tfsdk:"route_type"`
	DestinationCidrBlock types.String `tfsdk:"destination_cidr_block"`
	ResourceAttachmentId types.String `tfsdk:"resource_attachment_id"`
	ResourceId           types.String `tfsdk:"resource_id"`
	ResourceType         types.String `tfsdk:"resource_type"`
	TgwAttachmentId      types.String `tfsdk:"tgw_attachment_id"`
	TgwRouteTableId      types.String `tfsdk:"tgw_route_table_id"`
	ProvisioningStatus   types.String `tfsdk:"provisioning_status"`
	Resource             types.Object `tfsdk:"resource"`
}

type transitGatewayRoutesDataSourceModel struct {
	RouteTableId         types.String                   `tfsdk:"route_table_id"`
	Filter               []common.FilterModel           `tfsdk:"filter"`
	TransitGatewayRoutes []transitGatewayRouteBaseModel `tfsdk:"transit_gateway_routes"`
	Timeouts             timeouts.Value                 `tfsdk:"timeouts"`
}

var transitGatewayRouteResourceAttrType = map[string]attr.Type{
	"id":                  types.StringType,
	"name":                types.StringType,
	"cidr_block":          types.StringType,
	"project_id":          types.StringType,
	"project_name":        types.StringType,
	"provisioning_status": types.StringType,
}

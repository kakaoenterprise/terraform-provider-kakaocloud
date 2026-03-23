// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package tgw

import (
	"terraform-provider-kakaocloud/internal/common"

	datasourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	resourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type transitGatewayRouteTableAssociationBaseModel struct {
	Id                   types.String `tfsdk:"id"`
	RouteTableId         types.String `tfsdk:"route_table_id"`
	ResourceAttachmentId types.String `tfsdk:"resource_attachment_id"`
	ResourceId           types.String `tfsdk:"resource_id"`
	ResourceType         types.String `tfsdk:"resource_type"`
	TgwRouteTableId      types.String `tfsdk:"tgw_route_table_id"`
	ProvisioningStatus   types.String `tfsdk:"provisioning_status"`
	Resource             types.Object `tfsdk:"resource"`
}

type transitGatewayRouteTableAssociationsDataSourceModel struct {
	RouteTableId types.String                                   `tfsdk:"route_table_id"`
	Filter       []common.FilterModel                           `tfsdk:"filter"`
	Associations []transitGatewayRouteTableAssociationBaseModel `tfsdk:"associations"`
	Timeouts     datasourceTimeouts.Value                       `tfsdk:"timeouts"`
}

type transitGatewayRouteTableAssociationResourceModel struct {
	Id                 types.String           `tfsdk:"id"`
	RouteTableId       types.String           `tfsdk:"route_table_id"`
	TgwAttachmentId    types.String           `tfsdk:"tgw_attachment_id"`
	ProvisioningStatus types.String           `tfsdk:"provisioning_status"`
	Timeouts           resourceTimeouts.Value `tfsdk:"timeouts"`
}

type routeTableAssociationFetchResultModel struct {
	AssociationId      types.String
	TgwAttachmentId    types.String
	ProvisioningStatus types.String
}

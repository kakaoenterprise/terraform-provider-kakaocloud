// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package tgw

import (
	"terraform-provider-kakaocloud/internal/common"

	datasourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	resourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type tgwAttachmentNestedModel struct {
	Id                 types.String `tfsdk:"id"`
	ResourceType       types.String `tfsdk:"resource_type"`
	ResourceId         types.String `tfsdk:"resource_id"`
	ResourceName       types.String `tfsdk:"resource_name"`
	TgwId              types.String `tfsdk:"tgw_id"`
	ProvisioningStatus types.String `tfsdk:"provisioning_status"`
	CreatedAt          types.String `tfsdk:"created_at"`
	UpdatedAt          types.String `tfsdk:"updated_at"`
}

type tgwRouteTableNestedModel struct {
	Id                             types.String `tfsdk:"id"`
	Name                           types.String `tfsdk:"name"`
	Region                         types.String `tfsdk:"region"`
	ProjectId                      types.String `tfsdk:"project_id"`
	TgwId                          types.String `tfsdk:"tgw_id"`
	IsDefaultAssociationRouteTable types.Bool   `tfsdk:"is_default_association_route_table"`
	IsDefaultPropagationRouteTable types.Bool   `tfsdk:"is_default_propagation_route_table"`
	ProvisioningStatus             types.String `tfsdk:"provisioning_status"`
	CreatedAt                      types.String `tfsdk:"created_at"`
	UpdatedAt                      types.String `tfsdk:"updated_at"`
}

type transitGatewayBaseModel struct {
	Id                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	Region             types.String `tfsdk:"region"`
	IsShared           types.Bool   `tfsdk:"is_shared"`
	Attachments        types.List   `tfsdk:"attachments"`
	Options            types.Object `tfsdk:"options"`
	ProvisioningStatus types.String `tfsdk:"provisioning_status"`
	ProjectId          types.String `tfsdk:"project_id"`
	ProjectName        types.String `tfsdk:"project_name"`
	OwnerProjectId     types.String `tfsdk:"owner_project_id"`
	OwnerProjectName   types.String `tfsdk:"owner_project_name"`
	RouteTables        types.List   `tfsdk:"route_tables"`
	CreatedAt          types.String `tfsdk:"created_at"`
	UpdatedAt          types.String `tfsdk:"updated_at"`
}

type transitGatewayResourceModel struct {
	transitGatewayBaseModel
	Timeouts resourceTimeouts.Value `tfsdk:"timeouts"`
}

type transitGatewayDataSourceModel struct {
	transitGatewayBaseModel
	Timeouts datasourceTimeouts.Value `tfsdk:"timeouts"`
}

type transitGatewaysDataSourceModel struct {
	Filter          []common.FilterModel      `tfsdk:"filter"`
	TransitGateways []transitGatewayBaseModel `tfsdk:"transit_gateways"`
	Timeouts        datasourceTimeouts.Value  `tfsdk:"timeouts"`
}

type tgwOptionsModel struct {
	IsAutoAcceptSharedAttachments  types.Bool   `tfsdk:"is_auto_accept_shared_attachments"`
	IsDefaultRouteTableAssociation types.Bool   `tfsdk:"is_default_route_table_association"`
	AssociationDefaultRouteTableId types.String `tfsdk:"association_default_route_table_id"`
}

var tgwAttachmentAttrType = map[string]attr.Type{
	"id":                  types.StringType,
	"resource_type":       types.StringType,
	"resource_id":         types.StringType,
	"resource_name":       types.StringType,
	"tgw_id":              types.StringType,
	"provisioning_status": types.StringType,
	"created_at":          types.StringType,
	"updated_at":          types.StringType,
}

var tgwRouteTableAttrType = map[string]attr.Type{
	"id":                                 types.StringType,
	"name":                               types.StringType,
	"region":                             types.StringType,
	"project_id":                         types.StringType,
	"tgw_id":                             types.StringType,
	"is_default_association_route_table": types.BoolType,
	"is_default_propagation_route_table": types.BoolType,
	"provisioning_status":                types.StringType,
	"created_at":                         types.StringType,
	"updated_at":                         types.StringType,
}

var tgwOptionsAttrType = map[string]attr.Type{
	"is_auto_accept_shared_attachments":  types.BoolType,
	"is_default_route_table_association": types.BoolType,
	"association_default_route_table_id": types.StringType,
}

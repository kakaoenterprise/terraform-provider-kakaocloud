// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package tgw

import (
	"terraform-provider-kakaocloud/internal/common"

	"github.com/hashicorp/terraform-plugin-framework-nettypes/cidrtypes"
	datasourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	resourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type tgwRouteTableRouteNestedModel struct {
	Id                   types.String `tfsdk:"id"`
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

type tgwRouteTableAssociationNestedModel struct {
	Id                   types.String `tfsdk:"id"`
	ResourceAttachmentId types.String `tfsdk:"resource_attachment_id"`
	ResourceId           types.String `tfsdk:"resource_id"`
	ResourceType         types.String `tfsdk:"resource_type"`
	TgwAttachmentId      types.String `tfsdk:"tgw_attachment_id"`
	TgwRouteTableId      types.String `tfsdk:"tgw_route_table_id"`
	ProvisioningStatus   types.String `tfsdk:"provisioning_status"`
	Resource             types.Object `tfsdk:"resource"`
}

type tgwRouteTableResourceNestedModel struct {
	Id                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	CidrBlock          types.String `tfsdk:"cidr_block"`
	ProjectId          types.String `tfsdk:"project_id"`
	ProjectName        types.String `tfsdk:"project_name"`
	ProvisioningStatus types.String `tfsdk:"provisioning_status"`
}

var tgwRouteTableResourceAttrType = map[string]attr.Type{
	"id":                  types.StringType,
	"name":                types.StringType,
	"cidr_block":          types.StringType,
	"project_id":          types.StringType,
	"project_name":        types.StringType,
	"provisioning_status": types.StringType,
}

var tgwRouteTableRouteAttrType = map[string]attr.Type{
	"id":                     types.StringType,
	"route_type":             types.StringType,
	"destination_cidr_block": types.StringType,
	"resource_attachment_id": types.StringType,
	"resource_id":            types.StringType,
	"resource_type":          types.StringType,
	"tgw_attachment_id":      types.StringType,
	"tgw_route_table_id":     types.StringType,
	"provisioning_status":    types.StringType,
	"resource":               types.ObjectType{AttrTypes: tgwRouteTableResourceAttrType},
}

var tgwRouteTableAssociationAttrType = map[string]attr.Type{
	"id":                     types.StringType,
	"resource_attachment_id": types.StringType,
	"resource_id":            types.StringType,
	"resource_type":          types.StringType,
	"tgw_attachment_id":      types.StringType,
	"tgw_route_table_id":     types.StringType,
	"provisioning_status":    types.StringType,
	"resource":               types.ObjectType{AttrTypes: tgwRouteTableResourceAttrType},
}

type transitGatewayRouteTableBaseModel struct {
	Id                             types.String `tfsdk:"id"`
	Name                           types.String `tfsdk:"name"`
	TgwId                          types.String `tfsdk:"tgw_id"`
	Region                         types.String `tfsdk:"region"`
	ProjectId                      types.String `tfsdk:"project_id"`
	ProjectName                    types.String `tfsdk:"project_name"`
	TgwName                        types.String `tfsdk:"tgw_name"`
	IsDefaultAssociationRouteTable types.Bool   `tfsdk:"is_default_association_route_table"`
	IsDefaultPropagationRouteTable types.Bool   `tfsdk:"is_default_propagation_route_table"`
	ProvisioningStatus             types.String `tfsdk:"provisioning_status"`
	Routes                         types.List   `tfsdk:"routes"`
	Associations                   types.List   `tfsdk:"associations"`
	CreatedAt                      types.String `tfsdk:"created_at"`
	UpdatedAt                      types.String `tfsdk:"updated_at"`
}

var tgwRouteTableRequestRouteAttrType = map[string]attr.Type{
	"id":                     types.StringType,
	"destination_cidr_block": cidrtypes.IPPrefixType{},
	"tgw_attachment_id":      types.StringType,
}

type tgwRouteTableRequestRouteModel struct {
	Id                   types.String       `tfsdk:"id"`
	DestinationCidrBlock cidrtypes.IPPrefix `tfsdk:"destination_cidr_block"`
	TgwAttachmentId      types.String       `tfsdk:"tgw_attachment_id"`
}

var tgwRouteTableRequestAssociationAttrType = map[string]attr.Type{
	"id":                types.StringType,
	"tgw_attachment_id": types.StringType,
}

type tgwRouteTableRequestAssociationModel struct {
	Id              types.String `tfsdk:"id"`
	TgwAttachmentId types.String `tfsdk:"tgw_attachment_id"`
}

type transitGatewayRouteTableResourceModel struct {
	transitGatewayRouteTableBaseModel
	RequestRoutes       types.Set              `tfsdk:"request_routes"`
	RequestAssociations types.Set              `tfsdk:"request_associations"`
	Timeouts            resourceTimeouts.Value `tfsdk:"timeouts"`
}

type transitGatewayRouteTableDataSourceModel struct {
	transitGatewayRouteTableBaseModel
	Timeouts datasourceTimeouts.Value `tfsdk:"timeouts"`
}

type transitGatewayRouteTablesDataSourceModel struct {
	Filter                    []common.FilterModel                `tfsdk:"filter"`
	TransitGatewayRouteTables []transitGatewayRouteTableBaseModel `tfsdk:"transit_gateway_route_tables"`
	Timeouts                  datasourceTimeouts.Value            `tfsdk:"timeouts"`
}

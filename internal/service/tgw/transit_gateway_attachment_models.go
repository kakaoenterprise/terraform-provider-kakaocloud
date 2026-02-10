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

type tgwAttachmentTgwNestedModel struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	ProjectId   types.String `tfsdk:"project_id"`
	ProjectName types.String `tfsdk:"project_name"`
}

type tgwAttachmentResourceNestedModel struct {
	Id                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	Description        types.String `tfsdk:"description"`
	AvailabilityZone   types.String `tfsdk:"availability_zone"`
	CidrBlock          types.String `tfsdk:"cidr_block"`
	OperatingStatus    types.String `tfsdk:"operating_status"`
	ProvisioningStatus types.String `tfsdk:"provisioning_status"`
	VpcId              types.String `tfsdk:"vpc_id"`
	CreatedAt          types.String `tfsdk:"created_at"`
	UpdatedAt          types.String `tfsdk:"updated_at"`
}

type tgwAttachmentRouteTableNestedModel struct {
	Id                             types.String `tfsdk:"id"`
	Name                           types.String `tfsdk:"name"`
	ProjectId                      types.String `tfsdk:"project_id"`
	Region                         types.String `tfsdk:"region"`
	TgwId                          types.String `tfsdk:"tgw_id"`
	IsDefaultAssociationRouteTable types.Bool   `tfsdk:"is_default_association_route_table"`
	IsDefaultPropagationRouteTable types.Bool   `tfsdk:"is_default_propagation_route_table"`
	ProvisioningStatus             types.String `tfsdk:"provisioning_status"`
	CreatedAt                      types.String `tfsdk:"created_at"`
	UpdatedAt                      types.String `tfsdk:"updated_at"`
}

type transitGatewayAttachmentBaseModel struct {
	Id                 types.String `tfsdk:"id"`
	ProvisioningStatus types.String `tfsdk:"provisioning_status"`
	ProjectId          types.String `tfsdk:"project_id"`
	ProjectName        types.String `tfsdk:"project_name"`
	ResourceType       types.String `tfsdk:"resource_type"`
	ResourceId         types.String `tfsdk:"resource_id"`
	ResourceName       types.String `tfsdk:"resource_name"`
	ResourceCidrBlock  types.String `tfsdk:"resource_cidr_block"`
	Tgw                types.Object `tfsdk:"tgw"`
	Resources          types.List   `tfsdk:"resources"`
	RouteTable         types.Object `tfsdk:"route_table"`
	CreatedAt          types.String `tfsdk:"created_at"`
	UpdatedAt          types.String `tfsdk:"updated_at"`
}

type transitGatewayAttachmentResourceModel struct {
	transitGatewayAttachmentBaseModel
	TgwId     types.String           `tfsdk:"tgw_id"`
	SubnetIds types.Set              `tfsdk:"subnet_ids"`
	Timeouts  resourceTimeouts.Value `tfsdk:"timeouts"`
}

type transitGatewayAttachmentDataSourceModel struct {
	transitGatewayAttachmentBaseModel
	Timeouts datasourceTimeouts.Value `tfsdk:"timeouts"`
}

type transitGatewayAttachmentsDataSourceModel struct {
	Filter                    []common.FilterModel                `tfsdk:"filter"`
	TransitGatewayAttachments []transitGatewayAttachmentBaseModel `tfsdk:"transit_gateway_attachments"`
	Timeouts                  datasourceTimeouts.Value            `tfsdk:"timeouts"`
}

var tgwAttachmentTgwAttrType = map[string]attr.Type{
	"id":           types.StringType,
	"name":         types.StringType,
	"project_id":   types.StringType,
	"project_name": types.StringType,
}

var tgwAttachmentResourceAttrType = map[string]attr.Type{
	"id":                  types.StringType,
	"name":                types.StringType,
	"description":         types.StringType,
	"availability_zone":   types.StringType,
	"cidr_block":          types.StringType,
	"operating_status":    types.StringType,
	"provisioning_status": types.StringType,
	"vpc_id":              types.StringType,
	"created_at":          types.StringType,
	"updated_at":          types.StringType,
}

var tgwAttachmentRouteTableAttrType = map[string]attr.Type{
	"id":                                 types.StringType,
	"name":                               types.StringType,
	"project_id":                         types.StringType,
	"region":                             types.StringType,
	"tgw_id":                             types.StringType,
	"is_default_association_route_table": types.BoolType,
	"is_default_propagation_route_table": types.BoolType,
	"provisioning_status":                types.StringType,
	"created_at":                         types.StringType,
	"updated_at":                         types.StringType,
}

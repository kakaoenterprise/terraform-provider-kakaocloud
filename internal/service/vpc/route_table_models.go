// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package vpc

import (
	"terraform-provider-kakaocloud/internal/common"

	"github.com/hashicorp/terraform-plugin-framework-nettypes/cidrtypes"
	datasourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	resourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type routeTableBaseModel struct {
	Id                    types.String `tfsdk:"id"`
	Name                  types.String `tfsdk:"name"`
	Associations          types.List   `tfsdk:"associations"`
	Routes                types.List   `tfsdk:"routes"`
	VpcId                 types.String `tfsdk:"vpc_id"`
	ProvisioningStatus    types.String `tfsdk:"provisioning_status"`
	VpcName               types.String `tfsdk:"vpc_name"`
	VpcProvisioningStatus types.String `tfsdk:"vpc_provisioning_status"`
	ProjectId             types.String `tfsdk:"project_id"`
	ProjectName           types.String `tfsdk:"project_name"`
	IsMain                types.Bool   `tfsdk:"is_main"`
	CreatedAt             types.String `tfsdk:"created_at"`
	UpdatedAt             types.String `tfsdk:"updated_at"`
}

type routeTableAssociationModel struct {
	Id                 types.String `tfsdk:"id"`
	ProvisioningStatus types.String `tfsdk:"provisioning_status"`
	VpcId              types.String `tfsdk:"vpc_id"`
	VpcName            types.String `tfsdk:"vpc_name"`
	SubnetId           types.String `tfsdk:"subnet_id"`
	SubnetName         types.String `tfsdk:"subnet_name"`
	SubnetCidrBlock    types.String `tfsdk:"subnet_cidr_block"`
	AvailabilityZone   types.String `tfsdk:"availability_zone"`
}

var routeTableAssociationAttrType = map[string]attr.Type{
	"id":                  types.StringType,
	"provisioning_status": types.StringType,
	"vpc_id":              types.StringType,
	"vpc_name":            types.StringType,
	"subnet_id":           types.StringType,
	"subnet_name":         types.StringType,
	"subnet_cidr_block":   types.StringType,
	"availability_zone":   types.StringType,
}

type routeTableRouteModel struct {
	Id                 types.String `tfsdk:"id"`
	Destination        types.String `tfsdk:"destination"`
	ProvisioningStatus types.String `tfsdk:"provisioning_status"`
	TargetType         types.String `tfsdk:"target_type"`
	TargetName         types.String `tfsdk:"target_name"`
	IsLocalRoute       types.Bool   `tfsdk:"is_local_route"`
	TargetId           types.String `tfsdk:"target_id"`
}

var routeTableRouteAttrType = map[string]attr.Type{
	"id":                  types.StringType,
	"destination":         types.StringType,
	"provisioning_status": types.StringType,
	"target_type":         types.StringType,
	"target_name":         types.StringType,
	"is_local_route":      types.BoolType,
	"target_id":           types.StringType,
}

type routeTableRequestRouteModel struct {
	Id          types.String       `tfsdk:"id"`
	Destination cidrtypes.IPPrefix `tfsdk:"destination"`
	TargetType  types.String       `tfsdk:"target_type"`
	TargetId    types.String       `tfsdk:"target_id"`
}

var routeTableRequestRouteAttrType = map[string]attr.Type{
	"id":          types.StringType,
	"destination": cidrtypes.IPPrefixType{},
	"target_type": types.StringType,
	"target_id":   types.StringType,
}

type routeTableResourceModel struct {
	routeTableBaseModel
	RequestRoutes types.List             `tfsdk:"request_routes"`
	Timeouts      resourceTimeouts.Value `tfsdk:"timeouts"`
}

type routeTableDataSourceModel struct {
	routeTableBaseModel
	Timeouts datasourceTimeouts.Value `tfsdk:"timeouts"`
}

type routeTablesDataSourceModel struct {
	Filter      []common.FilterModel     `tfsdk:"filter"`
	RouteTables []routeTableBaseModel    `tfsdk:"route_tables"`
	Timeouts    datasourceTimeouts.Value `tfsdk:"timeouts"`
}

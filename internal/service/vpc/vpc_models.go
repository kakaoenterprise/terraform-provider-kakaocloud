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

type vpcBaseModel struct {
	Id                 types.String       `tfsdk:"id"`
	Name               types.String       `tfsdk:"name"`
	Description        types.String       `tfsdk:"description"`
	Region             types.String       `tfsdk:"region"`
	Igw                types.Object       `tfsdk:"igw"`                 // Nested object
	DefaultRouteTable  types.Object       `tfsdk:"default_route_table"` // Nested object
	ProjectId          types.String       `tfsdk:"project_id"`
	ProjectName        types.String       `tfsdk:"project_name"`
	CidrBlock          cidrtypes.IPPrefix `tfsdk:"cidr_block"`
	IsDefault          types.Bool         `tfsdk:"is_default"`
	ProvisioningStatus types.String       `tfsdk:"provisioning_status"`
	IsEnableDnsSupport types.Bool         `tfsdk:"is_enable_dns_support"`
	CreatedAt          types.String       `tfsdk:"created_at"`
	UpdatedAt          types.String       `tfsdk:"updated_at"`
}

type vpcResourceModel struct {
	vpcBaseModel
	Subnet   types.Object           `tfsdk:"subnet"`
	Timeouts resourceTimeouts.Value `tfsdk:"timeouts"`
}

type vpcDataSourceModel struct {
	vpcBaseModel
	Timeouts datasourceTimeouts.Value `tfsdk:"timeouts"`
}

type vpcsDataSourceModel struct {
	Filter   []common.FilterModel     `tfsdk:"filter"`
	Vpcs     []vpcBaseModel           `tfsdk:"vpcs"`
	Timeouts datasourceTimeouts.Value `tfsdk:"timeouts"`
}

type igwModel struct {
	Id                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	Description        types.String `tfsdk:"description"`
	Region             types.String `tfsdk:"region"`
	ProjectId          types.String `tfsdk:"project_id"`
	OperatingStatus    types.String `tfsdk:"operating_status"`
	ProvisioningStatus types.String `tfsdk:"provisioning_status"`
	CreatedAt          types.String `tfsdk:"created_at"`
	UpdatedAt          types.String `tfsdk:"updated_at"`
}

var igwAttrType = map[string]attr.Type{
	"id":                  types.StringType,
	"name":                types.StringType,
	"description":         types.StringType,
	"region":              types.StringType,
	"project_id":          types.StringType,
	"operating_status":    types.StringType,
	"provisioning_status": types.StringType,
	"created_at":          types.StringType,
	"updated_at":          types.StringType,
}

type defaultRouteTableModel struct {
	Id                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	Description        types.String `tfsdk:"description"`
	ProvisioningStatus types.String `tfsdk:"provisioning_status"`
	CreatedAt          types.String `tfsdk:"created_at"`
	UpdatedAt          types.String `tfsdk:"updated_at"`
}

var defaultRouteTableAttrType = map[string]attr.Type{
	"id":                  types.StringType,
	"name":                types.StringType,
	"description":         types.StringType,
	"provisioning_status": types.StringType,
	"created_at":          types.StringType,
	"updated_at":          types.StringType,
}

type vpcSubnetModel struct {
	CidrBlock        cidrtypes.IPPrefix `tfsdk:"cidr_block"`
	AvailabilityZone types.String       `tfsdk:"availability_zone"`
}

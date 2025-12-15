// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package network

import (
	"terraform-provider-kakaocloud/internal/common"

	"github.com/hashicorp/terraform-plugin-framework-nettypes/cidrtypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"

	datasourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	resourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type securityGroupBaseModel struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	ProjectId   types.String `tfsdk:"project_id"`
	ProjectName types.String `tfsdk:"project_name"`
	IsStateful  types.Bool   `tfsdk:"is_stateful"`
	Rules       types.Set    `tfsdk:"rules"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

type securityGroupRuleModel struct {
	Id              types.String       `tfsdk:"id"`
	Direction       types.String       `tfsdk:"direction"`
	Protocol        types.String       `tfsdk:"protocol"`
	PortRangeMin    types.Int32        `tfsdk:"port_range_min"`
	PortRangeMax    types.Int32        `tfsdk:"port_range_max"`
	RemoteIpPrefix  cidrtypes.IPPrefix `tfsdk:"remote_ip_prefix"`
	RemoteGroupId   types.String       `tfsdk:"remote_group_id"`
	RemoteGroupName types.String       `tfsdk:"remote_group_name"`
	Description     types.String       `tfsdk:"description"`
	CreatedAt       types.String       `tfsdk:"created_at"`
	UpdatedAt       types.String       `tfsdk:"updated_at"`
}

var securityGroupRuleAttrType = map[string]attr.Type{
	"id":                types.StringType,
	"direction":         types.StringType,
	"protocol":          types.StringType,
	"port_range_min":    types.Int32Type,
	"port_range_max":    types.Int32Type,
	"remote_ip_prefix":  cidrtypes.IPPrefixType{},
	"remote_group_id":   types.StringType,
	"remote_group_name": types.StringType,
	"description":       types.StringType,
	"created_at":        types.StringType,
	"updated_at":        types.StringType,
}

type securityGroupResourceModel struct {
	securityGroupBaseModel
	Timeouts resourceTimeouts.Value `tfsdk:"timeouts"`
}

type securityGroupDataSourceModel struct {
	securityGroupBaseModel
	Timeouts datasourceTimeouts.Value `tfsdk:"timeouts"`
}

type securityGroupsDataSourceModel struct {
	Filter         []common.FilterModel     `tfsdk:"filter"`
	SecurityGroups []securityGroupBaseModel `tfsdk:"security_groups"`
	Timeouts       datasourceTimeouts.Value `tfsdk:"timeouts"`
}

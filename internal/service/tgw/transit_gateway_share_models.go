// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package tgw

import (
	datasourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	resourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type transitGatewayShareResourceModel struct {
	Id              types.String           `tfsdk:"id"`
	TgwId           types.String           `tfsdk:"tgw_id"`
	TargetProjectId types.String           `tfsdk:"target_project_id"`
	Timeouts        resourceTimeouts.Value `tfsdk:"timeouts"`
}

type sharedProjectModel struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Nickname    types.String `tfsdk:"nickname"`
	Description types.String `tfsdk:"description"`
	DomainId    types.String `tfsdk:"domain_id"`
	IsEnabled   types.Bool   `tfsdk:"is_enabled"`
	CreatedAt   types.String `tfsdk:"created_at"`
	DisabledAt  types.String `tfsdk:"disabled_at"`
}

type transitGatewaySharedProjectsDataSourceModel struct {
	TgwId          types.String             `tfsdk:"tgw_id"`
	SharedProjects []sharedProjectModel     `tfsdk:"shared_projects"`
	Timeouts       datasourceTimeouts.Value `tfsdk:"timeouts"`
}

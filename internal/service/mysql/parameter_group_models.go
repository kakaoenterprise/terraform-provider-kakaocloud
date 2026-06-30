// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package mysql

import (
	datasourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type mysqlParameterModel struct {
	Key                   types.String `tfsdk:"key"`
	Value                 types.String `tfsdk:"value"`
	DefaultParameterValue types.String `tfsdk:"default_parameter_value"`
	ParameterType         types.String `tfsdk:"parameter_type"`
	DataType              types.String `tfsdk:"data_type"`
	ValidationValueFormat types.String `tfsdk:"validation_value_format"`
	IsEditable            types.Bool   `tfsdk:"is_editable"`
	IsRequired            types.Bool   `tfsdk:"is_required"`
}

type customParameterGroupListModel struct {
	Id                      types.String `tfsdk:"id"`
	Name                    types.String `tfsdk:"name"`
	Description             types.String `tfsdk:"description"`
	EngineVersion           types.String `tfsdk:"engine_version"`
	DefaultParameterGroupId types.String `tfsdk:"default_parameter_group_id"`
	ExistErrorSync          types.Bool   `tfsdk:"exist_error_sync"`
	InstanceGroupCount      types.Int32  `tfsdk:"instance_group_count"`
	IsRollbackPossible      types.Bool   `tfsdk:"is_rollback_possible"`
}

type customParameterGroupSingleModel struct {
	Id                      types.String `tfsdk:"id"`
	Name                    types.String `tfsdk:"name"`
	Description             types.String `tfsdk:"description"`
	EngineVersion           types.String `tfsdk:"engine_version"`
	DefaultParameterGroupId types.String `tfsdk:"default_parameter_group_id"`
	ExistErrorSync          types.Bool   `tfsdk:"exist_error_sync"`
	InstanceGroupCount      types.Int32  `tfsdk:"instance_group_count"`
	IsRollbackPossible      types.Bool   `tfsdk:"is_rollback_possible"`
	Parameters              types.List   `tfsdk:"parameters"`
}

type defaultParameterGroupListModel struct {
	Id                         types.String `tfsdk:"id"`
	Name                       types.String `tfsdk:"name"`
	Description                types.String `tfsdk:"description"`
	EngineVersion              types.String `tfsdk:"engine_version"`
	ExistErrorSync             types.Bool   `tfsdk:"exist_error_sync"`
	ExistEngineVersionMismatch types.Bool   `tfsdk:"exist_engine_version_mismatch"`
	InstanceGroupCount         types.Int32  `tfsdk:"instance_group_count"`
}

type defaultParameterGroupSingleModel struct {
	Id                         types.String `tfsdk:"id"`
	Name                       types.String `tfsdk:"name"`
	Description                types.String `tfsdk:"description"`
	EngineVersion              types.String `tfsdk:"engine_version"`
	ExistErrorSync             types.Bool   `tfsdk:"exist_error_sync"`
	ExistEngineVersionMismatch types.Bool   `tfsdk:"exist_engine_version_mismatch"`
	InstanceGroupCount         types.Int32  `tfsdk:"instance_group_count"`
	Parameters                 types.List   `tfsdk:"parameters"`
}

type mysqlParameterGroupEventModel struct {
	CreatedAt   types.String `tfsdk:"created_at"`
	Description types.String `tfsdk:"description"`
	Name        types.String `tfsdk:"name"`
}

type mysqlParameterGroupInstanceGroupModel struct {
	Id                   types.String `tfsdk:"id"`
	Name                 types.String `tfsdk:"name"`
	Status               types.String `tfsdk:"status"`
	EngineVersion        types.String `tfsdk:"engine_version"`
	FlavorId             types.String `tfsdk:"flavor_id"`
	ParameterGroupStatus types.String `tfsdk:"parameter_group_status"`
	InstanceGroupType    types.String `tfsdk:"instance_group_type"`
	IsMultiAz            types.Bool   `tfsdk:"is_multi_az"`
}

type customParameterGroupDataSourceModel struct {
	customParameterGroupSingleModel
	Timeouts datasourceTimeouts.Value `tfsdk:"timeouts"`
}

type customParameterGroupsDataSourceModel struct {
	CustomParameterGroups types.List               `tfsdk:"custom_parameter_groups"`
	Timeouts              datasourceTimeouts.Value `tfsdk:"timeouts"`
}

type defaultParameterGroupDataSourceModel struct {
	defaultParameterGroupSingleModel
	Timeouts datasourceTimeouts.Value `tfsdk:"timeouts"`
}

type defaultParameterGroupsDataSourceModel struct {
	DefaultParameterGroups types.List               `tfsdk:"default_parameter_groups"`
	Timeouts               datasourceTimeouts.Value `tfsdk:"timeouts"`
}

type customParameterGroupEventsDataSourceModel struct {
	CustomParameterGroupId types.String             `tfsdk:"custom_parameter_group_id"`
	Events                 types.List               `tfsdk:"events"`
	Timeouts               datasourceTimeouts.Value `tfsdk:"timeouts"`
}

type defaultParameterGroupEventsDataSourceModel struct {
	DefaultParameterGroupId types.String             `tfsdk:"default_parameter_group_id"`
	Events                  types.List               `tfsdk:"events"`
	Timeouts                datasourceTimeouts.Value `tfsdk:"timeouts"`
}

type instanceGroupsUsingCustomParameterGroupDataSourceModel struct {
	CustomParameterGroupId types.String             `tfsdk:"custom_parameter_group_id"`
	InstanceGroups         types.List               `tfsdk:"instance_groups"`
	Timeouts               datasourceTimeouts.Value `tfsdk:"timeouts"`
}

type instanceGroupsUsingDefaultParameterGroupDataSourceModel struct {
	DefaultParameterGroupId types.String             `tfsdk:"default_parameter_group_id"`
	InstanceGroups          types.List               `tfsdk:"instance_groups"`
	Timeouts                datasourceTimeouts.Value `tfsdk:"timeouts"`
}

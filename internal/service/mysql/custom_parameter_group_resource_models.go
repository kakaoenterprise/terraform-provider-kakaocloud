// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package mysql

import (
	resourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type customParameterOverrideModel struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

type customParameterGroupResourceModel struct {
	Id                       types.String           `tfsdk:"id"`
	Name                     types.String           `tfsdk:"name"`
	SourceParameterGroupId   types.String           `tfsdk:"source_parameter_group_id"`
	SourceParameterGroupType types.String           `tfsdk:"source_parameter_group_type"`
	DefaultParameterGroupId  types.String           `tfsdk:"default_parameter_group_id"`
	Description              types.String           `tfsdk:"description"`
	ApplyMode                types.String           `tfsdk:"apply_mode"`
	ParameterOverrides       types.Set              `tfsdk:"parameter_overrides"`
	EngineVersion            types.String           `tfsdk:"engine_version"`
	ExistErrorSync           types.Bool             `tfsdk:"exist_error_sync"`
	InstanceGroupCount       types.Int32            `tfsdk:"instance_group_count"`
	IsRollbackPossible       types.Bool             `tfsdk:"is_rollback_possible"`
	Timeouts                 resourceTimeouts.Value `tfsdk:"timeouts"`
}

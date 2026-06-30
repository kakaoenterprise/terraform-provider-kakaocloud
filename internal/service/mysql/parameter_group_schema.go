// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package mysql

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var mysqlParameterSchemaAttributes = map[string]schema.Attribute{
	"key":                     schema.StringAttribute{Computed: true},
	"value":                   schema.StringAttribute{Computed: true},
	"default_parameter_value": schema.StringAttribute{Computed: true},
	"parameter_type":          schema.StringAttribute{Computed: true},
	"data_type":               schema.StringAttribute{Computed: true},
	"validation_value_format": schema.StringAttribute{Computed: true},
	"is_editable":             schema.BoolAttribute{Computed: true},
	"is_required":             schema.BoolAttribute{Computed: true},
}

var mysqlParameterAttrTypes = map[string]attr.Type{
	"key":                     types.StringType,
	"value":                   types.StringType,
	"default_parameter_value": types.StringType,
	"parameter_type":          types.StringType,
	"data_type":               types.StringType,
	"validation_value_format": types.StringType,
	"is_editable":             types.BoolType,
	"is_required":             types.BoolType,
}

var customParameterGroupListSchemaAttributes = map[string]schema.Attribute{
	"id":                         schema.StringAttribute{Computed: true},
	"name":                       schema.StringAttribute{Computed: true},
	"description":                schema.StringAttribute{Computed: true},
	"engine_version":             schema.StringAttribute{Computed: true},
	"default_parameter_group_id": schema.StringAttribute{Computed: true},
	"exist_error_sync":           schema.BoolAttribute{Computed: true},
	"instance_group_count":       schema.Int32Attribute{Computed: true},
	"is_rollback_possible":       schema.BoolAttribute{Computed: true},
}

var customParameterGroupSingleSchemaAttributes = map[string]schema.Attribute{
	"id":                         schema.StringAttribute{Computed: true},
	"name":                       schema.StringAttribute{Computed: true},
	"description":                schema.StringAttribute{Computed: true},
	"engine_version":             schema.StringAttribute{Computed: true},
	"default_parameter_group_id": schema.StringAttribute{Computed: true},
	"exist_error_sync":           schema.BoolAttribute{Computed: true},
	"instance_group_count":       schema.Int32Attribute{Computed: true},
	"is_rollback_possible":       schema.BoolAttribute{Computed: true},
	"parameters": schema.ListNestedAttribute{
		Computed: true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: mysqlParameterSchemaAttributes,
		},
	},
}

var customParameterGroupListAttrTypes = map[string]attr.Type{
	"id":                         types.StringType,
	"name":                       types.StringType,
	"description":                types.StringType,
	"engine_version":             types.StringType,
	"default_parameter_group_id": types.StringType,
	"exist_error_sync":           types.BoolType,
	"instance_group_count":       types.Int32Type,
	"is_rollback_possible":       types.BoolType,
}

var defaultParameterGroupListSchemaAttributes = map[string]schema.Attribute{
	"id":                            schema.StringAttribute{Computed: true},
	"name":                          schema.StringAttribute{Computed: true},
	"description":                   schema.StringAttribute{Computed: true},
	"engine_version":                schema.StringAttribute{Computed: true},
	"exist_error_sync":              schema.BoolAttribute{Computed: true},
	"exist_engine_version_mismatch": schema.BoolAttribute{Computed: true},
	"instance_group_count":          schema.Int32Attribute{Computed: true},
}

var defaultParameterGroupSingleSchemaAttributes = map[string]schema.Attribute{
	"id":                            schema.StringAttribute{Computed: true},
	"name":                          schema.StringAttribute{Computed: true},
	"description":                   schema.StringAttribute{Computed: true},
	"engine_version":                schema.StringAttribute{Computed: true},
	"exist_error_sync":              schema.BoolAttribute{Computed: true},
	"exist_engine_version_mismatch": schema.BoolAttribute{Computed: true},
	"instance_group_count":          schema.Int32Attribute{Computed: true},
	"parameters": schema.ListNestedAttribute{
		Computed: true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: mysqlParameterSchemaAttributes,
		},
	},
}

var defaultParameterGroupListAttrTypes = map[string]attr.Type{
	"id":                            types.StringType,
	"name":                          types.StringType,
	"description":                   types.StringType,
	"engine_version":                types.StringType,
	"exist_error_sync":              types.BoolType,
	"exist_engine_version_mismatch": types.BoolType,
	"instance_group_count":          types.Int32Type,
}

var mysqlParameterGroupEventSchemaAttributes = map[string]schema.Attribute{
	"created_at":  schema.StringAttribute{Computed: true},
	"description": schema.StringAttribute{Computed: true},
	"name":        schema.StringAttribute{Computed: true},
}

var mysqlParameterGroupEventAttrTypes = map[string]attr.Type{
	"created_at":  types.StringType,
	"description": types.StringType,
	"name":        types.StringType,
}

var mysqlParameterGroupInstanceGroupSchemaAttributes = map[string]schema.Attribute{
	"id":                     schema.StringAttribute{Computed: true},
	"name":                   schema.StringAttribute{Computed: true},
	"status":                 schema.StringAttribute{Computed: true},
	"engine_version":         schema.StringAttribute{Computed: true},
	"flavor_id":              schema.StringAttribute{Computed: true},
	"parameter_group_status": schema.StringAttribute{Computed: true},
	"instance_group_type":    schema.StringAttribute{Computed: true},
	"is_multi_az":            schema.BoolAttribute{Computed: true},
}

var mysqlParameterGroupInstanceGroupAttrTypes = map[string]attr.Type{
	"id":                     types.StringType,
	"name":                   types.StringType,
	"status":                 types.StringType,
	"engine_version":         types.StringType,
	"flavor_id":              types.StringType,
	"parameter_group_status": types.StringType,
	"instance_group_type":    types.StringType,
	"is_multi_az":            types.BoolType,
}

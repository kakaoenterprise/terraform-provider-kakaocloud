// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package mysql

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var backupExtraInfoSchemaAttributes = map[string]schema.Attribute{
	"use_case_sensitive_table_names": schema.BoolAttribute{Computed: true},
}

var backupExtraInfoAttrTypes = map[string]attr.Type{
	"use_case_sensitive_table_names": types.BoolType,
}

var backupSchemaAttributes = map[string]schema.Attribute{
	"id":                  schema.StringAttribute{Computed: true},
	"name":                schema.StringAttribute{Computed: true},
	"created_at":          schema.StringAttribute{Computed: true},
	"creator_name":        schema.StringAttribute{Computed: true},
	"description":         schema.StringAttribute{Computed: true},
	"disk_size":           schema.Int32Attribute{Computed: true},
	"expire_at":           schema.StringAttribute{Computed: true},
	"expiry_duration":     schema.Int32Attribute{Computed: true},
	"extra_info":          schema.SingleNestedAttribute{Computed: true, Attributes: backupExtraInfoSchemaAttributes},
	"instance_group_id":   schema.StringAttribute{Computed: true},
	"instance_group_name": schema.StringAttribute{Computed: true},
	"project_id":          schema.StringAttribute{Computed: true},
	"size":                schema.Int64Attribute{Computed: true},
	"status":              schema.StringAttribute{Computed: true},
	"type":                schema.StringAttribute{Computed: true},
	"started_at":          schema.StringAttribute{Computed: true},
	"updated_at":          schema.StringAttribute{Computed: true},
	"engine_version":      schema.StringAttribute{Computed: true},
}

var backupAttrTypes = map[string]attr.Type{
	"id":                  types.StringType,
	"name":                types.StringType,
	"created_at":          types.StringType,
	"creator_name":        types.StringType,
	"description":         types.StringType,
	"disk_size":           types.Int32Type,
	"expire_at":           types.StringType,
	"expiry_duration":     types.Int32Type,
	"extra_info":          types.ObjectType{AttrTypes: backupExtraInfoAttrTypes},
	"instance_group_id":   types.StringType,
	"instance_group_name": types.StringType,
	"project_id":          types.StringType,
	"size":                types.Int64Type,
	"status":              types.StringType,
	"type":                types.StringType,
	"started_at":          types.StringType,
	"updated_at":          types.StringType,
	"engine_version":      types.StringType,
}

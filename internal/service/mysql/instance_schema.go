// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package mysql

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var networkPortSchemaAttributes = map[string]schema.Attribute{
	"subnet_id":          schema.StringAttribute{Computed: true},
	"security_group_ids": schema.SetAttribute{Computed: true, ElementType: types.StringType},
}

var networkPortAttrTypes = map[string]attr.Type{
	"subnet_id":          types.StringType,
	"security_group_ids": types.SetType{ElemType: types.StringType},
}

var instanceStatusContentSchemaAttributes = map[string]schema.Attribute{
	"needs_restart":        schema.BoolAttribute{Computed: true},
	"needs_restart_reason": schema.ListAttribute{Computed: true, ElementType: types.StringType},
}

var instanceStatusContentAttrTypes = map[string]attr.Type{
	"needs_restart":        types.BoolType,
	"needs_restart_reason": types.ListType{ElemType: types.StringType},
}

var instanceSpecContentSchemaAttributes = map[string]schema.Attribute{
	"availability_zone": schema.StringAttribute{Computed: true},
	"flavor_id":         schema.StringAttribute{Computed: true},
	"data_disk_size":    schema.Int32Attribute{Computed: true},
	"log_disk_size":     schema.Int32Attribute{Computed: true},
	"engine_version":    schema.StringAttribute{Computed: true},
	"network_ports": schema.ListNestedAttribute{
		Computed: true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: networkPortSchemaAttributes,
		},
	},
}

var instanceSpecContentAttrTypes = map[string]attr.Type{
	"availability_zone": types.StringType,
	"flavor_id":         types.StringType,
	"data_disk_size":    types.Int32Type,
	"log_disk_size":     types.Int32Type,
	"engine_version":    types.StringType,
	"network_ports":     types.ListType{ElemType: types.ObjectType{AttrTypes: networkPortAttrTypes}},
}

var instanceSchemaAttributes = map[string]schema.Attribute{
	"id":                  schema.StringAttribute{Computed: true},
	"project_id":          schema.StringAttribute{Computed: true},
	"instance_group_id":   schema.StringAttribute{Computed: true},
	"instance_group_name": schema.StringAttribute{Computed: true},
	"name":                schema.StringAttribute{Computed: true},
	"status":              schema.StringAttribute{Computed: true},
	"availability_status": schema.StringAttribute{Computed: true},
	"status_content": schema.SingleNestedAttribute{
		Computed:   true,
		Attributes: instanceStatusContentSchemaAttributes,
	},
	"role":            schema.StringAttribute{Computed: true},
	"data_disk_usage": schema.Int32Attribute{Computed: true},
	"log_disk_usage":  schema.Int32Attribute{Computed: true},
	"spec_content": schema.SingleNestedAttribute{
		Computed:   true,
		Attributes: instanceSpecContentSchemaAttributes,
	},
	"created_at": schema.StringAttribute{Computed: true},
	"updated_at": schema.StringAttribute{Computed: true},
	"start_time": schema.StringAttribute{Computed: true},
}

var instanceAttrTypes = map[string]attr.Type{
	"id":                  types.StringType,
	"project_id":          types.StringType,
	"instance_group_id":   types.StringType,
	"instance_group_name": types.StringType,
	"name":                types.StringType,
	"status":              types.StringType,
	"availability_status": types.StringType,
	"status_content":      types.ObjectType{AttrTypes: instanceStatusContentAttrTypes},
	"role":                types.StringType,
	"data_disk_usage":     types.Int32Type,
	"log_disk_usage":      types.Int32Type,
	"spec_content":        types.ObjectType{AttrTypes: instanceSpecContentAttrTypes},
	"created_at":          types.StringType,
	"updated_at":          types.StringType,
	"start_time":          types.StringType,
}

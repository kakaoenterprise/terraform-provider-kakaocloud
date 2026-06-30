// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package mysql

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var subnetInfoSchemaAttributes = map[string]schema.Attribute{
	"replicas":          schema.Int32Attribute{Computed: true},
	"availability_zone": schema.StringAttribute{Computed: true},
	"subnet_id":         schema.StringAttribute{Computed: true},
}

var subnetInfoAttrTypes = map[string]attr.Type{
	"replicas":          types.Int32Type,
	"availability_zone": types.StringType,
	"subnet_id":         types.StringType,
}

var networkInfoSchemaAttributes = map[string]schema.Attribute{
	"primary_subnet_info": schema.SingleNestedAttribute{
		Computed:   true,
		Attributes: subnetInfoSchemaAttributes,
	},
	"standby_subnet_info": schema.ListNestedAttribute{
		Computed: true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: subnetInfoSchemaAttributes,
		},
	},
	"security_group_ids": schema.SetAttribute{
		Computed:    true,
		ElementType: types.StringType,
	},
}

var networkInfoAttrTypes = map[string]attr.Type{
	"primary_subnet_info": types.ObjectType{AttrTypes: subnetInfoAttrTypes},
	"standby_subnet_info": types.ListType{ElemType: types.ObjectType{AttrTypes: subnetInfoAttrTypes}},
	"security_group_ids":  types.SetType{ElemType: types.StringType},
}

var specContentSchemaAttributes = map[string]schema.Attribute{
	"database_user_name":  schema.StringAttribute{Computed: true},
	"primary_port":        schema.Int32Attribute{Computed: true},
	"standby_port":        schema.Int32Attribute{Computed: true},
	"engine_version":      schema.StringAttribute{Computed: true},
	"flavor_id":           schema.StringAttribute{Computed: true},
	"vcpu":                schema.Int32Attribute{Computed: true},
	"memory":              schema.Int32Attribute{Computed: true},
	"log_disk_size":       schema.Int32Attribute{Computed: true},
	"data_disk_size":      schema.Int32Attribute{Computed: true},
	"instance_group_type": schema.StringAttribute{Computed: true},
	"node_size":           schema.Int32Attribute{Computed: true},
}

var specContentAttrTypes = map[string]attr.Type{
	"database_user_name":  types.StringType,
	"primary_port":        types.Int32Type,
	"standby_port":        types.Int32Type,
	"engine_version":      types.StringType,
	"flavor_id":           types.StringType,
	"vcpu":                types.Int32Type,
	"memory":              types.Int32Type,
	"log_disk_size":       types.Int32Type,
	"data_disk_size":      types.Int32Type,
	"instance_group_type": types.StringType,
	"node_size":           types.Int32Type,
}

var backupScheduleSchemaAttributes = map[string]schema.Attribute{
	"id":              schema.StringAttribute{Computed: true},
	"type":            schema.StringAttribute{Computed: true},
	"start_time":      schema.StringAttribute{Computed: true},
	"expiry_duration": schema.Int32Attribute{Computed: true},
	"enabled":         schema.BoolAttribute{Computed: true},
}

var backupScheduleAttrTypes = map[string]attr.Type{
	"id":              types.StringType,
	"type":            types.StringType,
	"start_time":      types.StringType,
	"expiry_duration": types.Int32Type,
	"enabled":         types.BoolType,
}

var parameterGroupSchemaAttributes = map[string]schema.Attribute{
	"id":                         schema.StringAttribute{Computed: true},
	"type":                       schema.StringAttribute{Computed: true},
	"apply_status":               schema.StringAttribute{Computed: true},
	"engine_version":             schema.StringAttribute{Computed: true},
	"is_engine_version_mismatch": schema.BoolAttribute{Computed: true},
}

var parameterGroupAttrTypes = map[string]attr.Type{
	"id":                         types.StringType,
	"type":                       types.StringType,
	"apply_status":               types.StringType,
	"engine_version":             types.StringType,
	"is_engine_version_mismatch": types.BoolType,
}

var extraInfoSchemaAttributes = map[string]schema.Attribute{
	"use_case_sensitive_table_names": schema.BoolAttribute{Computed: true},
}

var extraInfoAttrTypes = map[string]attr.Type{
	"use_case_sensitive_table_names": types.BoolType,
}

var instanceNodeSchemaAttributes = map[string]schema.Attribute{
	"instance_id":       schema.StringAttribute{Computed: true},
	"subnet_id":         schema.StringAttribute{Computed: true},
	"availability_zone": schema.StringAttribute{Computed: true},
}

var instanceNodeAttrTypes = map[string]attr.Type{
	"instance_id":       types.StringType,
	"subnet_id":         types.StringType,
	"availability_zone": types.StringType,
}

var instancesSchemaAttributes = map[string]schema.Attribute{
	"primary": schema.SingleNestedAttribute{
		Computed:   true,
		Attributes: instanceNodeSchemaAttributes,
	},
	"standby": schema.ListNestedAttribute{
		Computed: true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: instanceNodeSchemaAttributes,
		},
	},
}

var instancesAttrTypes = map[string]attr.Type{
	"primary": types.ObjectType{AttrTypes: instanceNodeAttrTypes},
	"standby": types.ListType{ElemType: types.ObjectType{AttrTypes: instanceNodeAttrTypes}},
}

var instanceGroupListInstanceNodeSchemaAttributes = map[string]schema.Attribute{
	"instance_id": schema.StringAttribute{Computed: true},
}

var instanceGroupListInstanceNodeAttrTypes = map[string]attr.Type{
	"instance_id": types.StringType,
}

var instanceGroupListInstancesSchemaAttributes = map[string]schema.Attribute{
	"primary": schema.SingleNestedAttribute{
		Computed:   true,
		Attributes: instanceGroupListInstanceNodeSchemaAttributes,
	},
	"standby": schema.ListNestedAttribute{
		Computed: true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: instanceGroupListInstanceNodeSchemaAttributes,
		},
	},
}

var instanceGroupListInstancesAttrTypes = map[string]attr.Type{
	"primary": types.ObjectType{AttrTypes: instanceGroupListInstanceNodeAttrTypes},
	"standby": types.ListType{ElemType: types.ObjectType{AttrTypes: instanceGroupListInstanceNodeAttrTypes}},
}

var instanceGroupSchemaAttributes = map[string]schema.Attribute{
	"id":               schema.StringAttribute{Computed: true},
	"created_at":       schema.StringAttribute{Computed: true},
	"updated_at":       schema.StringAttribute{Computed: true},
	"license":          schema.StringAttribute{Computed: true},
	"name":             schema.StringAttribute{Computed: true},
	"project_id":       schema.StringAttribute{Computed: true},
	"description":      schema.StringAttribute{Computed: true},
	"creator":          schema.StringAttribute{Computed: true},
	"source_backup_id": schema.StringAttribute{Computed: true},
	"is_multi_az":      schema.BoolAttribute{Computed: true},
	"endpoint": schema.ListAttribute{
		Computed:    true,
		ElementType: types.StringType,
	},
	"status": schema.StringAttribute{Computed: true},
	"network_info": schema.SingleNestedAttribute{
		Computed:   true,
		Attributes: networkInfoSchemaAttributes,
	},
	"spec_content": schema.SingleNestedAttribute{
		Computed:   true,
		Attributes: specContentSchemaAttributes,
	},
	"backup_schedule": schema.SingleNestedAttribute{
		Computed:   true,
		Attributes: backupScheduleSchemaAttributes,
	},
	"parameter_group": schema.SingleNestedAttribute{
		Computed:   true,
		Attributes: parameterGroupSchemaAttributes,
	},
	"extra_info": schema.SingleNestedAttribute{
		Computed:   true,
		Attributes: extraInfoSchemaAttributes,
	},
	"instances": schema.SingleNestedAttribute{
		Computed:   true,
		Attributes: instancesSchemaAttributes,
	},
}

var instanceGroupListSchemaAttributes = map[string]schema.Attribute{
	"id":               instanceGroupSchemaAttributes["id"],
	"created_at":       instanceGroupSchemaAttributes["created_at"],
	"updated_at":       instanceGroupSchemaAttributes["updated_at"],
	"license":          instanceGroupSchemaAttributes["license"],
	"name":             instanceGroupSchemaAttributes["name"],
	"project_id":       instanceGroupSchemaAttributes["project_id"],
	"description":      instanceGroupSchemaAttributes["description"],
	"creator":          instanceGroupSchemaAttributes["creator"],
	"source_backup_id": instanceGroupSchemaAttributes["source_backup_id"],
	"is_multi_az":      instanceGroupSchemaAttributes["is_multi_az"],
	"endpoint":         instanceGroupSchemaAttributes["endpoint"],
	"status":           instanceGroupSchemaAttributes["status"],
	"network_info":     instanceGroupSchemaAttributes["network_info"],
	"spec_content":     instanceGroupSchemaAttributes["spec_content"],
	"backup_schedule":  instanceGroupSchemaAttributes["backup_schedule"],
	"parameter_group":  instanceGroupSchemaAttributes["parameter_group"],
	"extra_info":       instanceGroupSchemaAttributes["extra_info"],
	"instances": schema.SingleNestedAttribute{
		Computed:   true,
		Attributes: instanceGroupListInstancesSchemaAttributes,
	},
}

var instanceGroupAttrTypes = map[string]attr.Type{
	"id":               types.StringType,
	"created_at":       types.StringType,
	"updated_at":       types.StringType,
	"license":          types.StringType,
	"name":             types.StringType,
	"project_id":       types.StringType,
	"description":      types.StringType,
	"creator":          types.StringType,
	"source_backup_id": types.StringType,
	"is_multi_az":      types.BoolType,
	"endpoint":         types.ListType{ElemType: types.StringType},
	"status":           types.StringType,
	"network_info":     types.ObjectType{AttrTypes: networkInfoAttrTypes},
	"spec_content":     types.ObjectType{AttrTypes: specContentAttrTypes},
	"backup_schedule":  types.ObjectType{AttrTypes: backupScheduleAttrTypes},
	"parameter_group":  types.ObjectType{AttrTypes: parameterGroupAttrTypes},
	"extra_info":       types.ObjectType{AttrTypes: extraInfoAttrTypes},
	"instances":        types.ObjectType{AttrTypes: instancesAttrTypes},
}

var instanceGroupListAttrTypes = map[string]attr.Type{
	"id":               instanceGroupAttrTypes["id"],
	"created_at":       instanceGroupAttrTypes["created_at"],
	"updated_at":       instanceGroupAttrTypes["updated_at"],
	"license":          instanceGroupAttrTypes["license"],
	"name":             instanceGroupAttrTypes["name"],
	"project_id":       instanceGroupAttrTypes["project_id"],
	"description":      instanceGroupAttrTypes["description"],
	"creator":          instanceGroupAttrTypes["creator"],
	"source_backup_id": instanceGroupAttrTypes["source_backup_id"],
	"is_multi_az":      instanceGroupAttrTypes["is_multi_az"],
	"endpoint":         instanceGroupAttrTypes["endpoint"],
	"status":           instanceGroupAttrTypes["status"],
	"network_info":     instanceGroupAttrTypes["network_info"],
	"spec_content":     instanceGroupAttrTypes["spec_content"],
	"backup_schedule":  instanceGroupAttrTypes["backup_schedule"],
	"parameter_group":  instanceGroupAttrTypes["parameter_group"],
	"extra_info":       instanceGroupAttrTypes["extra_info"],
	"instances":        types.ObjectType{AttrTypes: instanceGroupListInstancesAttrTypes},
}

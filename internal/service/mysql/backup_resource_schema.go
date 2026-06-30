// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package mysql

import (
	"terraform-provider-kakaocloud/internal/common"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

var backupResourceSchemaAttributes = map[string]schema.Attribute{
	"id": schema.StringAttribute{
		Computed: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	},
	"name": schema.StringAttribute{
		Required:   true,
		Validators: mysqlBackupNameValidator(),
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	},
	"instance_group_id": schema.StringAttribute{
		Required:   true,
		Validators: common.UuidValidator(),
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	},
	"status": schema.StringAttribute{
		Computed: true,
	},
	"type": schema.StringAttribute{
		Computed: true,
	},
	"created_at": schema.StringAttribute{
		Computed: true,
	},
	"creator_name": schema.StringAttribute{
		Computed: true,
	},
	"description": schema.StringAttribute{
		Computed: true,
	},
	"disk_size": schema.Int32Attribute{
		Computed: true,
	},
	"expire_at": schema.StringAttribute{
		Computed: true,
	},
	"expiry_duration": schema.Int32Attribute{
		Computed: true,
	},
	"extra_info": schema.SingleNestedAttribute{
		Computed:   true,
		Attributes: backupResourceExtraInfoSchemaAttributes,
	},
	"instance_group_name": schema.StringAttribute{
		Computed: true,
	},
	"project_id": schema.StringAttribute{
		Computed: true,
	},
	"size": schema.Int64Attribute{
		Computed: true,
	},
	"started_at": schema.StringAttribute{
		Computed: true,
	},
	"updated_at": schema.StringAttribute{
		Computed: true,
	},
	"engine_version": schema.StringAttribute{
		Computed: true,
	},
}

var backupResourceExtraInfoSchemaAttributes = map[string]schema.Attribute{
	"use_case_sensitive_table_names": schema.BoolAttribute{Computed: true},
}

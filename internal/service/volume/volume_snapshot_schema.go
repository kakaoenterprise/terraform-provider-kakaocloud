// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package volume

import (
	"terraform-provider-kakaocloud/internal/common"

	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func getVolumeSnapshotDataSourceSchema() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"name": dschema.StringAttribute{
			Computed: true,
		},
		"description": dschema.StringAttribute{
			Computed: true,
		},
		"size": dschema.Int64Attribute{
			Computed: true,
		},
		"real_size": dschema.Int64Attribute{
			Computed: true,
		},
		"status": dschema.StringAttribute{
			Computed: true,
		},
		"volume_id": dschema.StringAttribute{
			Computed: true,
		},
		"project_id": dschema.StringAttribute{
			Computed: true,
		},
		"parent_id": dschema.StringAttribute{
			Computed: true,
		},
		"user_id": dschema.StringAttribute{
			Computed: true,
		},
		"is_incremental": dschema.BoolAttribute{
			Computed: true,
		},
		"is_dependent_snapshot": dschema.BoolAttribute{
			Computed: true,
		},
		"schedule_id": dschema.StringAttribute{
			Computed: true,
		},
		"created_at": dschema.StringAttribute{
			Computed: true,
		},
		"updated_at": dschema.StringAttribute{
			Computed: true,
		},
	}
}

func getVolumeSnapshotResourceSchema() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"name": rschema.StringAttribute{
			Required:   true,
			Validators: common.NameValidator(250),
		},
		"description": rschema.StringAttribute{
			Optional:   true,
			Computed:   true,
			Validators: common.DescriptionValidator(),
		},
		"size": rschema.Int64Attribute{
			Computed: true,
		},
		"real_size": rschema.Int64Attribute{
			Computed: true,
		},
		"status": rschema.StringAttribute{
			Computed: true,
		},
		"volume_id": rschema.StringAttribute{
			Required:   true,
			Validators: common.UuidValidator(),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"project_id": rschema.StringAttribute{
			Computed: true,
		},
		"parent_id": rschema.StringAttribute{
			Computed: true,
		},
		"user_id": rschema.StringAttribute{
			Computed: true,
		},
		"is_incremental": rschema.BoolAttribute{
			Required: true,
			PlanModifiers: []planmodifier.Bool{
				boolplanmodifier.RequiresReplace(),
			},
		},
		"is_dependent_snapshot": rschema.BoolAttribute{
			Computed: true,
		},
		"schedule_id": rschema.StringAttribute{
			Computed: true,
		},
		"created_at": rschema.StringAttribute{
			Computed: true,
		},
		"updated_at": rschema.StringAttribute{
			Computed: true,
		},
	}
}

var volumeSnapshotResourceSchemaAttributes = getVolumeSnapshotResourceSchema()
var volumeSnapshotDataSourceSchemaAttributes = getVolumeSnapshotDataSourceSchema()

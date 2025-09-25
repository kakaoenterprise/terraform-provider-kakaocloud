// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package volume

import (
	"terraform-provider-kakaocloud/internal/common"
	"terraform-provider-kakaocloud/internal/docs"

	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func getVolumeSnapshotDataSourceSchema() map[string]dschema.Attribute {
	desc := docs.Volume("bcs_volume__v1__api__get_snapshot__model__VolumeSnapshotModel")

	return map[string]dschema.Attribute{
		"name": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("name"),
		},
		"description": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("description"),
		},
		"size": dschema.Int64Attribute{
			Computed:    true,
			Description: desc.String("size"),
		},
		"real_size": dschema.Int64Attribute{
			Computed:    true,
			Description: desc.String("real_size"),
		},
		"status": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("status"),
		},
		"volume_id": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("volume_id"),
		},
		"project_id": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("project_id"),
		},
		"parent_id": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("parent_id"),
		},
		"user_id": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("user_id"),
		},
		"is_incremental": dschema.BoolAttribute{
			Computed:    true,
			Description: desc.String("is_incremental"),
		},
		"is_dependent_snapshot": dschema.BoolAttribute{
			Computed:    true,
			Description: desc.String("is_dependent_snapshot"),
		},
		"schedule_id": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("schedule_id"),
		},
		"created_at": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("created_at"),
		},
		"updated_at": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("updated_at"),
		},
	}
}

func getVolumeSnapshotResourceSchema() map[string]rschema.Attribute {
	desc := docs.Volume("bcs_volume__v1__api__get_snapshot__model__VolumeSnapshotModel")

	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("id"),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"name": rschema.StringAttribute{
			Required:    true,
			Description: desc.String("name"),
			Validators:  common.NameValidator(250),
		},
		"description": rschema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: desc.String("description"),
			Validators:  common.DescriptionValidator(),
		},
		"size": rschema.Int64Attribute{
			Computed:    true,
			Description: desc.String("size"),
		},
		"real_size": rschema.Int64Attribute{
			Computed:    true,
			Description: desc.String("real_size"),
		},
		"status": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("status"),
		},
		"volume_id": rschema.StringAttribute{
			Required:    true,
			Description: desc.String("volume_id"),
			Validators:  common.UuidValidator(),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"project_id": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("project_id"),
		},
		"parent_id": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("parent_id"),
		},
		"user_id": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("user_id"),
		},
		"is_incremental": rschema.BoolAttribute{
			Required:    true,
			Description: desc.String("is_incremental"),
			PlanModifiers: []planmodifier.Bool{
				boolplanmodifier.RequiresReplace(),
			},
		},
		"is_dependent_snapshot": rschema.BoolAttribute{
			Computed:    true,
			Description: desc.String("is_dependent_snapshot"),
		},
		"schedule_id": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("schedule_id"),
		},
		"created_at": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("created_at"),
		},
		"updated_at": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("updated_at"),
		},
	}
}

var volumeSnapshotResourceSchemaAttributes = getVolumeSnapshotResourceSchema()
var volumeSnapshotDataSourceSchemaAttributes = getVolumeSnapshotDataSourceSchema()

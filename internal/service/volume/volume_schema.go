// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package volume

import (
	"terraform-provider-kakaocloud/internal/common"

	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func getImageMetadataDataSourceSchema() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"container_format": dschema.StringAttribute{
			Computed: true,
		},
		"disk_format": dschema.StringAttribute{
			Computed: true,
		},
		"image_id": dschema.StringAttribute{
			Computed: true,
		},
		"image_name": dschema.StringAttribute{
			Computed: true,
		},
		"min_disk": dschema.StringAttribute{
			Computed: true,
		},
		"os_type": dschema.StringAttribute{
			Computed: true,
		},
		"min_ram": dschema.StringAttribute{
			Computed: true,
		},
		"size": dschema.StringAttribute{
			Computed: true,
		},
	}
}

func getVolumeDataSourceSchema() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"name": dschema.StringAttribute{
			Computed: true,
		},
		"description": dschema.StringAttribute{
			Computed: true,
		},
		"availability_zone": dschema.StringAttribute{
			Computed: true,
		},
		"status": dschema.StringAttribute{
			Computed: true,
		},
		"mount_point": dschema.StringAttribute{
			Computed: true,
		},
		"volume_type": dschema.StringAttribute{
			Computed: true,
		},
		"size": dschema.Int32Attribute{
			Computed: true,
		},
		"is_root": dschema.BoolAttribute{
			Computed: true,
		},
		"is_encrypted": dschema.BoolAttribute{
			Computed: true,
		},
		"is_bootable": dschema.BoolAttribute{
			Computed: true,
		},
		"type": dschema.StringAttribute{
			Computed: true,
		},
		"user_id": dschema.StringAttribute{
			Computed: true,
		},
		"project_id": dschema.StringAttribute{
			Computed: true,
		},
		"attach_status": dschema.StringAttribute{
			Computed: true,
		},
		"launched_at": dschema.StringAttribute{
			Computed: true,
		},
		"encryption_key_id": dschema.StringAttribute{
			Computed: true,
		},
		"previous_status": dschema.StringAttribute{
			Computed: true,
		},
		"created_at": dschema.StringAttribute{
			Computed: true,
		},
		"updated_at": dschema.StringAttribute{
			Computed: true,
		},
		"instance_id": dschema.StringAttribute{
			Computed: true,
		},
		"instance_name": dschema.StringAttribute{
			Computed: true,
		},
		"image_metadata": dschema.SingleNestedAttribute{
			Computed:   true,
			Attributes: imageMetadataDataSourceSchema,
		},
		"metadata": dschema.MapAttribute{
			ElementType: types.StringType,
			Computed:    true,
		},
	}
}

func getImageMetadataResourceSchema() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"container_format": rschema.StringAttribute{
			Computed: true,
		},
		"disk_format": rschema.StringAttribute{
			Computed: true,
		},
		"image_id": rschema.StringAttribute{
			Computed: true,
		},
		"image_name": rschema.StringAttribute{
			Computed: true,
		},
		"min_disk": rschema.StringAttribute{
			Computed: true,
		},
		"os_type": rschema.StringAttribute{
			Computed: true,
		},
		"min_ram": rschema.StringAttribute{
			Computed: true,
		},
		"size": rschema.StringAttribute{
			Computed: true,
		},
	}
}

func getVolumeResourceSchema() map[string]rschema.Attribute {
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
		"availability_zone": rschema.StringAttribute{
			Required: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"status": rschema.StringAttribute{
			Computed: true,
		},
		"mount_point": rschema.StringAttribute{
			Computed: true,
		},
		"volume_type_id": rschema.StringAttribute{
			Optional:   true,
			Validators: common.UuidValidator(),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"volume_type": rschema.StringAttribute{
			Computed: true,
		},
		"size": rschema.Int32Attribute{
			Optional:   true,
			Computed:   true,
			Validators: common.VolumeSizeValidator(),
			PlanModifiers: []planmodifier.Int32{
				common.PreventShrinkModifier[int32]{
					TypeName:        "Volume Size",
					DescriptionText: "Prevents reducing the volume size",
				},
			},
		},
		"is_root": rschema.BoolAttribute{
			Computed: true,
		},
		"is_encrypted": rschema.BoolAttribute{
			Computed: true,
		},
		"is_bootable": rschema.BoolAttribute{
			Computed: true,
		},
		"type": rschema.StringAttribute{
			Computed: true,
		},
		"user_id": rschema.StringAttribute{
			Computed: true,
		},
		"project_id": rschema.StringAttribute{
			Computed: true,
		},
		"attach_status": rschema.StringAttribute{
			Computed: true,
		},
		"launched_at": rschema.StringAttribute{
			Computed: true,
		},
		"encryption_key_id": rschema.StringAttribute{
			Computed: true,
		},
		"previous_status": rschema.StringAttribute{
			Computed: true,
		},
		"created_at": rschema.StringAttribute{
			Computed: true,
		},
		"updated_at": rschema.StringAttribute{
			Computed: true,
		},
		"instance_id": rschema.StringAttribute{
			Computed: true,
		},
		"instance_name": rschema.StringAttribute{
			Computed: true,
		},
		"image_id": rschema.StringAttribute{
			Optional:   true,
			Validators: common.UuidValidator(),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"volume_snapshot_id": rschema.StringAttribute{
			Optional:   true,
			Validators: common.UuidValidator(),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"source_volume_id": rschema.StringAttribute{
			Optional:   true,
			Validators: common.UuidValidator(),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"image_metadata": rschema.SingleNestedAttribute{
			Computed:   true,
			Attributes: imageMetadataResourceSchema,
		},
		"metadata": rschema.MapAttribute{
			ElementType: types.StringType,
			Computed:    true,
		},
		"encryption_secret_id": rschema.StringAttribute{
			Optional:   true,
			Validators: common.UuidValidator(),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
	}
}

var imageMetadataDataSourceSchema = getImageMetadataDataSourceSchema()
var volumeDataSourceSchemaAttributes = getVolumeDataSourceSchema()

var imageMetadataResourceSchema = getImageMetadataResourceSchema()
var volumeResourceSchemaAttributes = getVolumeResourceSchema()

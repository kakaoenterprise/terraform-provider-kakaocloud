// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package volume

import (
	"terraform-provider-kakaocloud/internal/common"
	"terraform-provider-kakaocloud/internal/docs"

	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func getImageMetadataDataSourceSchema() map[string]dschema.Attribute {
	desc := docs.Volume("bcs_volume__v1__api__get_volume__model__ImageMetaData")

	return map[string]dschema.Attribute{
		"container_format": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("container_format"),
		},
		"disk_format": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("disk_format"),
		},
		"image_id": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("image_id"),
		},
		"image_name": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("image_name"),
		},
		"min_disk": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("min_disk"),
		},
		"os_type": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("os_type"),
		},
		"min_ram": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("min_ram"),
		},
		"size": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("size"),
		},
	}
}

func getVolumeDataSourceSchema() map[string]dschema.Attribute {
	desc := docs.Volume("bcs_volume__v1__api__get_volume__model__VolumeModel")

	return map[string]dschema.Attribute{
		"name": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("name"),
		},
		"description": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("description"),
		},
		"availability_zone": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("availability_zone"),
		},
		"status": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("status"),
		},
		"mount_point": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("mount_point"),
		},
		"volume_type": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("volume_type"),
		},
		"size": dschema.Int32Attribute{
			Computed:    true,
			Description: desc.String("size"),
		},
		"is_root": dschema.BoolAttribute{
			Computed:    true,
			Description: desc.String("is_root"),
		},
		"is_encrypted": dschema.BoolAttribute{
			Computed:    true,
			Description: desc.String("is_encrypted"),
		},
		"is_bootable": dschema.BoolAttribute{
			Computed:    true,
			Description: desc.String("is_bootable"),
		},
		"type": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("type"),
		},
		"user_id": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("user_id"),
		},
		"project_id": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("project_id"),
		},
		"attach_status": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("attach_status"),
		},
		"launched_at": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("launched_at"),
		},
		"encryption_key_id": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("encryption_key_id"),
		},
		"previous_status": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("previous_status"),
		},
		"created_at": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("created_at"),
		},
		"updated_at": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("updated_at"),
		},
		"instance_id": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("instance_id"),
		},
		"instance_name": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("instance_name"),
		},
		"image_metadata": dschema.SingleNestedAttribute{
			Computed:    true,
			Description: desc.String("image_metadata"),
			Attributes:  imageMetadataDataSourceSchema,
		},
		"metadata": dschema.MapAttribute{
			ElementType: types.StringType,
			Computed:    true,
			Description: desc.String("metadata"),
		},
	}
}

func getImageMetadataResourceSchema() map[string]rschema.Attribute {
	desc := docs.Volume("bcs_volume__v1__api__get_volume__model__ImageMetaData")

	return map[string]rschema.Attribute{
		"container_format": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("container_format"),
		},
		"disk_format": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("disk_format"),
		},
		"image_id": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("image_id"),
		},
		"image_name": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("image_name"),
		},
		"min_disk": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("min_disk"),
		},
		"os_type": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("os_type"),
		},
		"min_ram": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("min_ram"),
		},
		"size": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("size"),
		},
	}
}

func getVolumeResourceSchema() map[string]rschema.Attribute {
	desc := docs.Volume("bcs_volume__v1__api__get_volume__model__VolumeModel")
	createDesc := docs.Volume("CreateVolumeModel")

	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			Description: desc.String("id"),
			Computed:    true,
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
		"availability_zone": rschema.StringAttribute{
			Required:    true,
			Description: desc.String("availability_zone"),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"status": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("status"),
		},
		"mount_point": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("mount_point"),
		},
		"volume_type_id": rschema.StringAttribute{
			Optional:    true,
			Description: createDesc.String("volume_type_id"),
			Validators:  common.UuidValidator(),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"volume_type": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("volume_type"),
		},
		"size": rschema.Int32Attribute{
			Optional:    true,
			Computed:    true,
			Description: createDesc.String("size"),
			Validators:  common.VolumeSizeValidator(),
			PlanModifiers: []planmodifier.Int32{
				common.PreventShrinkModifier[int32]{
					TypeName:        "Volume Size",
					DescriptionText: "Prevents reducing the volume size",
				},
			},
		},
		"is_root": rschema.BoolAttribute{
			Computed:    true,
			Description: desc.String("is_root"),
		},
		"is_encrypted": rschema.BoolAttribute{
			Computed:    true,
			Description: desc.String("is_encrypted"),
		},
		"is_bootable": rschema.BoolAttribute{
			Computed:    true,
			Description: desc.String("is_bootable"),
		},
		"type": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("type"),
		},
		"user_id": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("user_id"),
		},
		"project_id": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("project_id"),
		},
		"attach_status": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("attach_status"),
		},
		"launched_at": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("launched_at"),
		},
		"encryption_key_id": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("encryption_key_id"),
		},
		"previous_status": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("previous_status"),
		},
		"created_at": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("created_at"),
		},
		"updated_at": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("updated_at"),
		},
		"instance_id": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("instance_id"),
		},
		"instance_name": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("instance_name"),
		},
		"image_id": rschema.StringAttribute{
			Optional:    true,
			Description: createDesc.String("image_id"),
			Validators:  common.UuidValidator(),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"volume_snapshot_id": rschema.StringAttribute{
			Optional:    true,
			Description: createDesc.String("source_volume_id"),
			Validators:  common.UuidValidator(),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},

		"source_volume_id": rschema.StringAttribute{
			Optional:    true,
			Description: createDesc.String("source_volume_id"),
			Validators:  common.UuidValidator(),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"image_metadata": rschema.SingleNestedAttribute{
			Computed:    true,
			Description: desc.String("image_metadata"),
			Attributes:  imageMetadataResourceSchema,
		},
		"metadata": rschema.MapAttribute{
			ElementType: types.StringType,
			Computed:    true,
			Description: desc.String("metadata"),
		},
		"encryption_secret_id": rschema.StringAttribute{
			Optional:    true,
			Description: desc.String("encryption_secret_id"),
			Validators:  common.UuidValidator(),
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

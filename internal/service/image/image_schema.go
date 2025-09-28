// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package image

import (
	"terraform-provider-kakaocloud/internal/common"
	"terraform-provider-kakaocloud/internal/docs"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func getImageResourceSchema() map[string]rschema.Attribute {
	desc := docs.Image("bcs_image__v1__api__get_image__model__ImageModel")

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
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"volume_id": rschema.StringAttribute{
			Optional:    true,
			Description: docs.ParameterDescription("image", "create_image", "path_volume_id"),
			Validators:  common.UuidValidator(),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"size": rschema.Int64Attribute{
			Computed:    true,
			Description: desc.String("size"),
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
			},
		},
		"status": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("status"),
		},
		"owner": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("owner"),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"visibility": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("visibility"),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"is_shared": rschema.BoolAttribute{
			Computed:    true,
			Description: desc.String("is_shared"),
		},
		"disk_format": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("disk_format"),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"container_format": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("container_format"),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"min_disk": rschema.Int32Attribute{
			Computed:    true,
			Description: desc.String("min_disk"),
			PlanModifiers: []planmodifier.Int32{
				int32planmodifier.UseStateForUnknown(),
			},
		},
		"min_ram": rschema.Int32Attribute{
			Computed:    true,
			Description: desc.String("min_ram"),
			PlanModifiers: []planmodifier.Int32{
				int32planmodifier.UseStateForUnknown(),
			},
		},
		"virtual_size": rschema.Int64Attribute{
			Computed:    true,
			Description: desc.String("virtual_size"),
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
			},
		},
		"instance_type": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("instance_type"),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"image_member_status": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("image_member_status"),
		},
		"project_id": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("project_id"),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"created_at": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("created_at"),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"updated_at": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("updated_at"),
		},
		"os_info": rschema.ObjectAttribute{
			Computed:    true,
			Description: desc.String("os_info"),
			AttributeTypes: map[string]attr.Type{
				"type":         types.StringType,
				"distro":       types.StringType,
				"architecture": types.StringType,
				"admin_user":   types.StringType,
				"is_hidden":    types.BoolType,
			},
			PlanModifiers: []planmodifier.Object{
				objectplanmodifier.UseStateForUnknown(),
			},
		},
	}
}

func getImageDataSourceSchema() map[string]dschema.Attribute {
	desc := docs.Image("bcs_image__v1__api__get_image__model__ImageModel")

	return map[string]dschema.Attribute{
		"name": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("name"),
		},
		"size": dschema.Int64Attribute{
			Computed:    true,
			Description: desc.String("size"),
		},
		"status": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("status"),
		},
		"owner": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("owner"),
		},
		"visibility": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("visibility"),
		},
		"description": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("description"),
		},
		"is_shared": dschema.BoolAttribute{
			Computed:    true,
			Description: desc.String("is_shared"),
		},
		"disk_format": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("disk_format"),
		},
		"container_format": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("container_format"),
		},
		"min_disk": dschema.Int32Attribute{
			Computed:    true,
			Description: desc.String("min_disk"),
		},
		"min_ram": dschema.Int32Attribute{
			Computed:    true,
			Description: desc.String("min_ram"),
		},
		"virtual_size": dschema.Int64Attribute{
			Computed:    true,
			Description: desc.String("virtual_size"),
		},
		"instance_type": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("instance_type"),
		},
		"image_member_status": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("image_member_status"),
		},
		"project_id": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("project_id"),
		},
		"created_at": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("created_at"),
		},
		"updated_at": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("updated_at"),
		},
		"os_info": dschema.SingleNestedAttribute{
			Computed:    true,
			Description: desc.String("os_info"),
			Attributes:  getImageOsInfoSchemaAttributes(),
		},
	}
}

func getImageOsInfoSchemaAttributes() map[string]dschema.Attribute {
	desc := docs.Image("bcs_image__v1__api__get_image__model__OsInfoModel")

	return map[string]dschema.Attribute{
		"type": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("type"),
		},
		"distro": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("distro"),
		},
		"architecture": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("architecture"),
		},
		"admin_user": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("admin_user"),
		},
		"is_hidden": dschema.BoolAttribute{
			Computed:    true,
			Description: desc.String("is_hidden"),
		},
	}
}

var imageResourceSchemaAttributes = getImageResourceSchema()
var imageDataSourceSchemaAttributes = getImageDataSourceSchema()
var imageOsInfoSchemaAttributes = getImageOsInfoSchemaAttributes()

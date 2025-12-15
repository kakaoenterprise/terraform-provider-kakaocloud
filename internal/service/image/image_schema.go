// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package image

import (
	"terraform-provider-kakaocloud/internal/common"

	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func getImageResourceSchema() map[string]rschema.Attribute {
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
		"volume_id": rschema.StringAttribute{
			Optional:   true,
			Validators: common.UuidValidator(),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"size": rschema.Int64Attribute{
			Computed: true,
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
			},
		},
		"status": rschema.StringAttribute{
			Computed: true,
		},
		"owner": rschema.StringAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"visibility": rschema.StringAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"is_shared": rschema.BoolAttribute{
			Computed: true,
		},
		"disk_format": rschema.StringAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"container_format": rschema.StringAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"min_disk": rschema.Int32Attribute{
			Computed: true,
			PlanModifiers: []planmodifier.Int32{
				int32planmodifier.UseStateForUnknown(),
			},
		},
		"min_ram": rschema.Int32Attribute{
			Computed: true,
			PlanModifiers: []planmodifier.Int32{
				int32planmodifier.UseStateForUnknown(),
			},
		},
		"virtual_size": rschema.Int64Attribute{
			Computed: true,
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
			},
		},
		"instance_type": rschema.StringAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"image_member_status": rschema.StringAttribute{
			Computed: true,
		},
		"project_id": rschema.StringAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"created_at": rschema.StringAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"updated_at": rschema.StringAttribute{
			Computed: true,
		},
		"os_info": rschema.SingleNestedAttribute{
			Computed:   true,
			Attributes: getImageOsInfoResourceSchemaAttributes(),
			PlanModifiers: []planmodifier.Object{
				objectplanmodifier.UseStateForUnknown(),
			},
		},
	}
}

func getImageDataSourceSchema() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"name": dschema.StringAttribute{
			Computed: true,
		},
		"size": dschema.Int64Attribute{
			Computed: true,
		},
		"status": dschema.StringAttribute{
			Computed: true,
		},
		"owner": dschema.StringAttribute{
			Computed: true,
		},
		"visibility": dschema.StringAttribute{
			Computed: true,
		},
		"description": dschema.StringAttribute{
			Computed: true,
		},
		"is_shared": dschema.BoolAttribute{
			Computed: true,
		},
		"disk_format": dschema.StringAttribute{
			Computed: true,
		},
		"container_format": dschema.StringAttribute{
			Computed: true,
		},
		"min_disk": dschema.Int32Attribute{
			Computed: true,
		},
		"min_ram": dschema.Int32Attribute{
			Computed: true,
		},
		"virtual_size": dschema.Int64Attribute{
			Computed: true,
		},
		"instance_type": dschema.StringAttribute{
			Computed: true,
		},
		"image_member_status": dschema.StringAttribute{
			Computed: true,
		},
		"project_id": dschema.StringAttribute{
			Computed: true,
		},
		"created_at": dschema.StringAttribute{
			Computed: true,
		},
		"updated_at": dschema.StringAttribute{
			Computed: true,
		},
		"os_info": dschema.SingleNestedAttribute{
			Computed:   true,
			Attributes: getImageOsInfoSchemaAttributes(),
		},
	}
}

func getImageOsInfoResourceSchemaAttributes() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"type": rschema.StringAttribute{
			Computed: true,
		},
		"distro": rschema.StringAttribute{
			Computed: true,
		},
		"architecture": rschema.StringAttribute{
			Computed: true,
		},
		"admin_user": rschema.StringAttribute{
			Computed: true,
		},
		"is_hidden": rschema.BoolAttribute{
			Computed: true,
		},
	}
}

func getImageOsInfoSchemaAttributes() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"type": dschema.StringAttribute{
			Computed: true,
		},
		"distro": dschema.StringAttribute{
			Computed: true,
		},
		"architecture": dschema.StringAttribute{
			Computed: true,
		},
		"admin_user": dschema.StringAttribute{
			Computed: true,
		},
		"is_hidden": dschema.BoolAttribute{
			Computed: true,
		},
	}
}

var imageResourceSchemaAttributes = getImageResourceSchema()
var imageDataSourceSchemaAttributes = getImageDataSourceSchema()

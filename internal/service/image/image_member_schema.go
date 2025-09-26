// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package image

import (
	"terraform-provider-kakaocloud/internal/docs"

	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func getImageMemberResourceSchema() map[string]rschema.Attribute {
	desc := docs.Image("bcs_image__v1__api__add_image_share__model__ImageMemberModel")

	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("id"),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"status": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("status"),
		},
		"image_id": rschema.StringAttribute{
			Required:    true,
			Description: desc.String("image_id"),
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
		"member_id": rschema.StringAttribute{
			Required:    true,
			Description: docs.ParameterDescription("image", "add_image_share", "path_member_id"),
		},
	}
}

func getImageMemberDataSourceSchema() map[string]dschema.Attribute {
	desc := docs.Image("bcs_image__v1__api__list_image_shared_projects__model__ImageMemberModel")

	return map[string]dschema.Attribute{
		"created_at": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("created_at"),
		},
		"updated_at": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("updated_at"),
		},
		"image_id": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("image_id"),
		},
		"status": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("status"),
		},
		"is_shared": dschema.BoolAttribute{
			Computed:    true,
			Description: desc.String("is_shared"),
		},
	}
}

var imageMemberResourceSchema = getImageMemberResourceSchema()
var imageMemberDataSourceSchemaAttributes = getImageMemberDataSourceSchema()

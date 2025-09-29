// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package image

import (
	"terraform-provider-kakaocloud/internal/common"
	"terraform-provider-kakaocloud/internal/docs"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func getImageMemberResourceSchema() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			Required:    true,
			Description: "Unique ID of the image",
			Validators:  common.UuidValidator(),
		},
		"members": rschema.ListNestedAttribute{
			Computed: true,
			NestedObject: rschema.NestedAttributeObject{
				Attributes: getImageMembersResourceAttributes(),
			},
		},
		"shared_member_ids": rschema.SetAttribute{
			Required:    true,
			ElementType: types.StringType,
			Description: "Image shared member ID List",
			Validators: []validator.Set{
				setvalidator.ValueStringsAre(common.UuidNoHyphenValidator()...),
			},
		},
	}
}

func getImageMembersResourceAttributes() map[string]rschema.Attribute {
	desc := docs.Image("bcs_image__v1__api__add_image_share__model__ImageMemberModel")

	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("id"),
		},
		"created_at": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("created_at"),
		},
		"updated_at": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("updated_at"),
		},
		"image_id": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("image_id"),
		},
		"status": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("status"),
		},
		"is_shared": rschema.BoolAttribute{
			Computed:    true,
			Description: desc.String("is_shared"),
		},
	}
}

func getImageMemberDataSourceSchema() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"id": dschema.StringAttribute{
			Required:    true,
			Description: "Unique ID of the image",
			Validators:  common.UuidValidator(),
		},
		"members": dschema.ListNestedAttribute{
			Computed: true,
			NestedObject: dschema.NestedAttributeObject{
				Attributes: getImageMembersDataSourceAttributes(),
			},
		},
	}
}

func getImageMembersDataSourceAttributes() map[string]dschema.Attribute {
	desc := docs.Image("bcs_image__v1__api__add_image_share__model__ImageMemberModel")

	return map[string]dschema.Attribute{
		"id": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("id"),
		},
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
var imageMemberDataSourceSchema = getImageMemberDataSourceSchema()

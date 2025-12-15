// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package image

import (
	"terraform-provider-kakaocloud/internal/common"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func getImageMemberResourceSchema() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			Required:   true,
			Validators: common.UuidValidator(),
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
			Validators: []validator.Set{
				setvalidator.ValueStringsAre(common.UuidNoHyphenValidator()...),
			},
		},
	}
}

func getImageMembersResourceAttributes() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			Computed: true,
		},
		"created_at": rschema.StringAttribute{
			Computed: true,
		},
		"updated_at": rschema.StringAttribute{
			Computed: true,
		},
		"image_id": rschema.StringAttribute{
			Computed: true,
		},
		"status": rschema.StringAttribute{
			Computed: true,
		},
		"is_shared": rschema.BoolAttribute{
			Computed: true,
		},
	}
}

func getImageMemberDataSourceSchema() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"id": dschema.StringAttribute{
			Required:   true,
			Validators: common.UuidValidator(),
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
	return map[string]dschema.Attribute{
		"id": dschema.StringAttribute{
			Computed: true,
		},
		"created_at": dschema.StringAttribute{
			Computed: true,
		},
		"updated_at": dschema.StringAttribute{
			Computed: true,
		},
		"image_id": dschema.StringAttribute{
			Computed: true,
		},
		"status": dschema.StringAttribute{
			Computed: true,
		},
		"is_shared": dschema.BoolAttribute{
			Computed: true,
		},
	}
}

var imageMemberResourceSchema = getImageMemberResourceSchema()
var imageMemberDataSourceSchema = getImageMemberDataSourceSchema()

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package tgw

import (
	"terraform-provider-kakaocloud/internal/common"

	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func getTransitGatewayShareResourceSchema() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"tgw_id": rschema.StringAttribute{
			Required:   true,
			Validators: common.UuidValidator(),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"target_project_id": rschema.StringAttribute{
			Required:   true,
			Validators: common.UuidNoHyphenValidator(),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
	}
}

func getSharedProjectDataSourceSchema() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"id": dschema.StringAttribute{
			Computed: true,
		},
		"name": dschema.StringAttribute{
			Computed: true,
		},
		"nickname": dschema.StringAttribute{
			Computed: true,
		},
		"description": dschema.StringAttribute{
			Computed: true,
		},
		"domain_id": dschema.StringAttribute{
			Computed: true,
		},
		"is_enabled": dschema.BoolAttribute{
			Computed: true,
		},
		"created_at": dschema.StringAttribute{
			Computed: true,
		},
		"disabled_at": dschema.StringAttribute{
			Computed: true,
		},
	}
}

func getTransitGatewaySharedProjectsDataSourceSchema() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"tgw_id": dschema.StringAttribute{
			Required:   true,
			Validators: common.UuidValidator(),
		},
		"shared_projects": dschema.ListNestedAttribute{
			Computed: true,
			NestedObject: dschema.NestedAttributeObject{
				Attributes: getSharedProjectDataSourceSchema(),
			},
		},
	}
}

var transitGatewayShareResourceSchemaAttributes = getTransitGatewayShareResourceSchema()
var transitGatewaySharedProjectsDataSourceSchemaAttributes = getTransitGatewaySharedProjectsDataSourceSchema()

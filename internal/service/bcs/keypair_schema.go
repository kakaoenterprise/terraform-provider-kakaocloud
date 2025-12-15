// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package bcs

import (
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

var keypairDataSourceSchemaAttributes = map[string]dschema.Attribute{
	"id":          dschema.StringAttribute{Computed: true},
	"user_id":     dschema.StringAttribute{Computed: true},
	"fingerprint": dschema.StringAttribute{Computed: true},
	"public_key":  dschema.StringAttribute{Computed: true},
	"type":        dschema.StringAttribute{Computed: true},
	"created_at":  dschema.StringAttribute{Computed: true},
}

var keypairResourceSchemaAttributes = map[string]rschema.Attribute{
	"id": rschema.StringAttribute{
		Computed: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	},
	"name": rschema.StringAttribute{
		Required: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	},
	"public_key": rschema.StringAttribute{
		Optional: true,
		Computed: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	},
	"private_key": rschema.StringAttribute{
		Computed:  true,
		Sensitive: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	},
	"fingerprint": rschema.StringAttribute{
		Computed: true,
	},
	"user_id": rschema.StringAttribute{
		Computed: true,
	},
	"type": rschema.StringAttribute{
		Computed: true,
	},
	"created_at": rschema.StringAttribute{
		Computed: true,
	},
}

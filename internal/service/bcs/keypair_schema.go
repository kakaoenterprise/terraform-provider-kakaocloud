// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package bcs

import (
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"

	"terraform-provider-kakaocloud/internal/docs"
)

var (
	keypairDesc = docs.Bcs("bcs_instance__v1__api__get_keypair__model__KeypairModel")
)

var keypairDataSourceSchemaAttributes = map[string]dschema.Attribute{
	"id":          dschema.StringAttribute{Computed: true, Description: keypairDesc.String("id")},
	"user_id":     dschema.StringAttribute{Computed: true, Description: keypairDesc.String("user_id")},
	"fingerprint": dschema.StringAttribute{Computed: true, Description: keypairDesc.String("fingerprint")},
	"public_key":  dschema.StringAttribute{Computed: true, Description: keypairDesc.String("public_key")},
	"type":        dschema.StringAttribute{Computed: true, Description: keypairDesc.String("type")},
	"created_at":  dschema.StringAttribute{Computed: true, Description: keypairDesc.String("created_at")},
}

var keypairResourceSchemaAttributes = map[string]rschema.Attribute{
	"id": rschema.StringAttribute{
		Computed: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
		Description: keypairDesc.String("id"),
	},
	"name": rschema.StringAttribute{
		Required: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
		Description: keypairDesc.String("name"),
	},
	"public_key": rschema.StringAttribute{
		Optional: true,
		Computed: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
		Description: keypairDesc.String("public_key"),
	},
	"private_key": rschema.StringAttribute{
		Computed:  true,
		Sensitive: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
		Description: docs.Description("bcs", "bcs_instance__v1__api__create_keypair__model__KeypairModel", "private_key"),
	},
	"fingerprint": rschema.StringAttribute{
		Computed:    true,
		Description: keypairDesc.String("fingerprint"),
	},
	"user_id": rschema.StringAttribute{
		Computed:    true,
		Description: keypairDesc.String("user_id"),
	},
	"type": rschema.StringAttribute{
		Computed:    true,
		Description: keypairDesc.String("type"),
	},
	"created_at": rschema.StringAttribute{
		Computed:    true,
		Description: keypairDesc.String("created_at"),
	},
}

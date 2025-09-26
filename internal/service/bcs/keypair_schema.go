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
	"id":          dschema.StringAttribute{Computed: true, Description: "키페어 ID"},
	"user_id":     dschema.StringAttribute{Computed: true, Description: "사용자 ID"},
	"fingerprint": dschema.StringAttribute{Computed: true, Description: "핑거프린트"},
	"public_key":  dschema.StringAttribute{Computed: true, Description: "공개키"},
	"type":        dschema.StringAttribute{Computed: true, Description: "키페어 유형"},
	"created_at":  dschema.StringAttribute{Computed: true, Description: "생성 시간"},
}

var keypairResourceSchemaAttributes = map[string]rschema.Attribute{
	"id": rschema.StringAttribute{
		Computed: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
		Description: "The unique ID of the keypair.",
	},
	"name": rschema.StringAttribute{
		Required: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
		Description: "The unique name for the keypair.",
	},
	"public_key": rschema.StringAttribute{
		Optional: true,
		Computed: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
		Description: "The public key material",
	},
	"private_key": rschema.StringAttribute{
		Computed:  true,
		Sensitive: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
		Description: "The private key material (only returned on create).",
	},
	"fingerprint": rschema.StringAttribute{
		Computed:    true,
		Description: "The fingerprint of the public key.",
	},
	"user_id": rschema.StringAttribute{
		Computed:    true,
		Description: "The user ID of the keypair owner.",
	},
	"type": rschema.StringAttribute{
		Computed:    true,
		Description: "The type of the keypair (e.g., ssh).",
	},
	"created_at": rschema.StringAttribute{
		Computed:    true,
		Description: "The creation time of the keypair.",
	},
}

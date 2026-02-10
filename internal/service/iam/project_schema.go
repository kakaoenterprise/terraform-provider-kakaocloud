// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package iam

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

var projectDataSourceSchemaAttributes = map[string]schema.Attribute{
	"domain": schema.SingleNestedAttribute{
		Computed: true,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Computed: true,
			},
		},
	},
	"id": schema.StringAttribute{
		Computed: true,
	},
	"name": schema.StringAttribute{
		Computed: true,
	},
}

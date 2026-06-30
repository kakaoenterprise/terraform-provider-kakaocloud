// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package mysql

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var flavorSchemaAttributes = map[string]schema.Attribute{
	"id":        schema.StringAttribute{Computed: true},
	"name":      schema.StringAttribute{Computed: true},
	"type":      schema.StringAttribute{Computed: true},
	"vcpus":     schema.Int32Attribute{Computed: true},
	"memory":    schema.Int32Attribute{Computed: true},
	"memory_mb": schema.Int32Attribute{Computed: true},
	"group":     schema.StringAttribute{Computed: true},
	"family":    schema.StringAttribute{Computed: true},
	"availability_zones": schema.ListAttribute{
		Computed:    true,
		ElementType: types.StringType,
	},
	"deprecated": schema.BoolAttribute{Computed: true},
}

var engineVersionSchemaAttributes = map[string]schema.Attribute{
	"engine_version": schema.StringAttribute{Computed: true},
	"license":        schema.StringAttribute{Computed: true},
}

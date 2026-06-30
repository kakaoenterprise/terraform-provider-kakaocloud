// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package mysql

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var restorableTimeSchemaAttributes = map[string]schema.Attribute{
	"from_time": schema.StringAttribute{Computed: true},
	"to_time":   schema.StringAttribute{Computed: true},
}

var restorableTimeAttrTypes = map[string]attr.Type{
	"from_time": types.StringType,
	"to_time":   types.StringType,
}

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kubernetesengine

import "github.com/hashicorp/terraform-plugin-framework/datasource/schema"

var kubernetesVersionDataSourceSchemaAttributes = map[string]schema.Attribute{
	"is_deprecated": schema.BoolAttribute{Computed: true},
	"minor_version": schema.StringAttribute{Computed: true},
	"patch_version": schema.StringAttribute{Computed: true},
	"eol":           schema.StringAttribute{Computed: true},
	"next_version":  schema.StringAttribute{Computed: true},
}

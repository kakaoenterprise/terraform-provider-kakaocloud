// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package kubernetesengine

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func getUpgradableVersionsDataSourceSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"eol": schema.StringAttribute{
			Computed: true,
		},
		"is_deprecated": schema.BoolAttribute{
			Computed: true,
		},
		"minor_version": schema.StringAttribute{
			Computed: true,
		},
		"patch_version": schema.StringAttribute{
			Computed: true,
		},
		"next_version": schema.StringAttribute{
			Computed: true,
		},
	}
}

var upgradableVersionsDataSourceSchemaAttributes = getUpgradableVersionsDataSourceSchema()

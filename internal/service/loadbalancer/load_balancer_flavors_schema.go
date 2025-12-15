// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package loadbalancer

import (
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func getLoadBalancerFlavorSchema() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"id": dschema.StringAttribute{
			Computed: true,
		},
		"name": dschema.StringAttribute{
			Computed: true,
		},
		"description": dschema.StringAttribute{
			Computed: true,
		},
		"is_enabled": dschema.BoolAttribute{
			Computed: true,
		},
	}
}

var loadBalancerFlavorBaseSchemaAttributes = getLoadBalancerFlavorSchema()

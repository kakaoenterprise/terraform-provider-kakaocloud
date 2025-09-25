// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package loadbalancer

import (
	"terraform-provider-kakaocloud/internal/docs"

	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func getLoadBalancerFlavorSchema() map[string]dschema.Attribute {
	desc := docs.Loadbalancer("FlavorModel")

	return map[string]dschema.Attribute{
		"id": dschema.StringAttribute{
			Description: desc.String("id"),
			Computed:    true,
		},
		"name": dschema.StringAttribute{
			Description: desc.String("name"),
			Computed:    true,
		},
		"description": dschema.StringAttribute{
			Description: desc.String("description"),
			Computed:    true,
		},
		"is_enabled": dschema.BoolAttribute{
			Description: desc.String("is_enabled"),
			Computed:    true,
		},
	}
}

// loadBalancerFlavorBaseSchemaAttributes defines the attributes for a single load balancer flavor.
var loadBalancerFlavorBaseSchemaAttributes = getLoadBalancerFlavorSchema()

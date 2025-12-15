// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package loadbalancer

import (
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func getLoadBalancerSecretContentTypeSchema() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"default": dschema.StringAttribute{
			Computed: true,
		},
	}
}

func getLoadBalancerSecretBaseSchema() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"created_at": dschema.StringAttribute{
			Computed: true,
		},
		"updated_at": dschema.StringAttribute{
			Computed: true,
		},
		"status": dschema.StringAttribute{
			Computed: true,
		},
		"name": dschema.StringAttribute{
			Computed: true,
		},
		"secret_type": dschema.StringAttribute{
			Computed: true,
		},
		"expiration": dschema.StringAttribute{
			Computed: true,
		},
		"creator_id": dschema.StringAttribute{
			Computed: true,
		},
		"content_types": dschema.SingleNestedAttribute{
			Computed:   true,
			Attributes: getLoadBalancerSecretContentTypeSchema(),
		},
		"secret_ref": dschema.StringAttribute{
			Computed: true,
		},
	}
}

var loadBalancerSecretContentTypeSchemaAttributes = getLoadBalancerSecretContentTypeSchema()

var loadBalancerSecretBaseSchemaAttributes = getLoadBalancerSecretBaseSchema()

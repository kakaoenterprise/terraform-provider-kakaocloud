// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package loadbalancer

import (
	"terraform-provider-kakaocloud/internal/docs"

	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func getLoadBalancerSecretContentTypeSchema() map[string]dschema.Attribute {
	desc := docs.Loadbalancer("ContentType")

	return map[string]dschema.Attribute{
		"default": dschema.StringAttribute{
			Description: desc.String("default"),
			Computed:    true,
		},
	}
}

func getLoadBalancerSecretBaseSchema() map[string]dschema.Attribute {
	desc := docs.Loadbalancer("bns_load_balancer__v1__api__list_tls_certificates__model__SecretModel")

	return map[string]dschema.Attribute{
		"created_at": dschema.StringAttribute{
			Description: desc.String("created_at"),
			Computed:    true,
		},
		"updated_at": dschema.StringAttribute{
			Description: desc.String("updated_at"),
			Computed:    true,
		},
		"status": dschema.StringAttribute{
			Description: desc.String("status"),
			Computed:    true,
		},
		"name": dschema.StringAttribute{
			Description: desc.String("name"),
			Computed:    true,
		},
		"secret_type": dschema.StringAttribute{
			Description: desc.String("secret_type"),
			Computed:    true,
		},
		"expiration": dschema.StringAttribute{
			Description: desc.String("expiration"),
			Computed:    true,
		},
		"creator_id": dschema.StringAttribute{
			Description: desc.String("creator_id"),
			Computed:    true,
		},
		"content_types": dschema.SingleNestedAttribute{
			Description: desc.String("content_types"),
			Computed:    true,
			Attributes:  getLoadBalancerSecretContentTypeSchema(),
		},
		"secret_ref": dschema.StringAttribute{
			Description: desc.String("secret_ref"),
			Computed:    true,
		},
	}
}

// loadBalancerSecretContentTypeSchemaAttributes defines the attributes for a single content type.
var loadBalancerSecretContentTypeSchemaAttributes = getLoadBalancerSecretContentTypeSchema()

// loadBalancerSecretBaseSchemaAttributes defines the attributes for a single load balancer secret.
var loadBalancerSecretBaseSchemaAttributes = getLoadBalancerSecretBaseSchema()

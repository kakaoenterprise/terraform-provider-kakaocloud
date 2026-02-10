// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package tgw

import (
	"terraform-provider-kakaocloud/internal/common"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func getTransitGatewayRoutesDataSourceSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"route_table_id": schema.StringAttribute{
			Required:   true,
			Validators: common.UuidValidator(),
		},
		"filter": schema.ListNestedAttribute{
			Optional: true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						Required: true,
					},
					"value": schema.StringAttribute{
						Optional: true,
					},
				},
			},
		},
		"transit_gateway_routes": schema.ListNestedAttribute{
			Computed: true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: getTransitGatewayRouteDataSourceSchemaAttributes(),
			},
		},
	}
}

func getTransitGatewayRouteDataSourceSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed: true,
		},
		"route_table_id": schema.StringAttribute{
			Computed: true,
		},
		"route_type": schema.StringAttribute{
			Computed: true,
		},
		"destination_cidr_block": schema.StringAttribute{
			Computed: true,
		},
		"resource_attachment_id": schema.StringAttribute{
			Computed: true,
		},
		"resource_id": schema.StringAttribute{
			Computed: true,
		},
		"resource_type": schema.StringAttribute{
			Computed: true,
		},
		"tgw_attachment_id": schema.StringAttribute{
			Computed: true,
		},
		"tgw_route_table_id": schema.StringAttribute{
			Computed: true,
		},
		"provisioning_status": schema.StringAttribute{
			Computed: true,
		},
		"resource": schema.SingleNestedAttribute{
			Computed: true,
			Attributes: map[string]schema.Attribute{
				"id": schema.StringAttribute{
					Computed: true,
				},
				"name": schema.StringAttribute{
					Computed: true,
				},
				"cidr_block": schema.StringAttribute{
					Computed: true,
				},
				"project_id": schema.StringAttribute{
					Computed: true,
				},
				"project_name": schema.StringAttribute{
					Computed: true,
				},
				"provisioning_status": schema.StringAttribute{
					Computed: true,
				},
			},
		},
	}
}

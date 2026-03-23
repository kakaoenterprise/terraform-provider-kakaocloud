// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package tgw

import (
	"terraform-provider-kakaocloud/internal/common"

	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func getTransitGatewayRouteTableAssociationsDataSourceSchema() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"route_table_id": dschema.StringAttribute{
			Required:   true,
			Validators: common.UuidValidator(),
		},
		"filter": dschema.ListNestedAttribute{
			Optional: true,
			NestedObject: dschema.NestedAttributeObject{
				Attributes: map[string]dschema.Attribute{
					"name": dschema.StringAttribute{
						Required: true,
					},
					"value": dschema.StringAttribute{
						Optional: true,
					},
				},
			},
		},
		"associations": dschema.ListNestedAttribute{
			Computed: true,
			NestedObject: dschema.NestedAttributeObject{
				Attributes: getTransitGatewayRouteTableAssociationDataSourceSchema(),
			},
		},
	}
}

func getTransitGatewayRouteTableAssociationDataSourceSchema() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"id": dschema.StringAttribute{
			Computed: true,
		},
		"route_table_id": dschema.StringAttribute{
			Computed: true,
		},
		"resource_attachment_id": dschema.StringAttribute{
			Computed: true,
		},
		"resource_id": dschema.StringAttribute{
			Computed: true,
		},
		"resource_type": dschema.StringAttribute{
			Computed: true,
		},
		"tgw_route_table_id": dschema.StringAttribute{
			Computed: true,
		},
		"provisioning_status": dschema.StringAttribute{
			Computed: true,
		},
		"resource": dschema.SingleNestedAttribute{
			Computed: true,
			Attributes: map[string]dschema.Attribute{
				"id": dschema.StringAttribute{
					Computed: true,
				},
				"name": dschema.StringAttribute{
					Computed: true,
				},
				"cidr_block": dschema.StringAttribute{
					Computed: true,
				},
				"project_id": dschema.StringAttribute{
					Computed: true,
				},
				"project_name": dschema.StringAttribute{
					Computed: true,
				},
				"provisioning_status": dschema.StringAttribute{
					Computed: true,
				},
			},
		},
	}
}

func getTransitGatewayRouteTableAssociationResourceSchema() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"route_table_id": rschema.StringAttribute{
			Required:   true,
			Validators: common.UuidValidator(),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"tgw_attachment_id": rschema.StringAttribute{
			Required:   true,
			Validators: common.UuidValidator(),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"provisioning_status": rschema.StringAttribute{
			Computed: true,
		},
	}
}

var transitGatewayRouteTableAssociationsDataSourceSchemaAttributes = getTransitGatewayRouteTableAssociationsDataSourceSchema()
var transitGatewayRouteTableAssociationsResourceSchemaAttributes = getTransitGatewayRouteTableAssociationResourceSchema()

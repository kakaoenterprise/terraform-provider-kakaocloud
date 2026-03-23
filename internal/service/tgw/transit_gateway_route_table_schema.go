// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package tgw

import (
	"terraform-provider-kakaocloud/internal/common"

	"github.com/hashicorp/terraform-plugin-framework-nettypes/cidrtypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func getTransitGatewayRouteTableResourceSchema() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"name": rschema.StringAttribute{
			Required:   true,
			Validators: common.NameValidator(250),
		},
		"tgw_id": rschema.StringAttribute{
			Required:   true,
			Validators: common.UuidValidator(),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"region": rschema.StringAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"project_id": rschema.StringAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"project_name": rschema.StringAttribute{
			Computed: true,
		},
		"tgw_name": rschema.StringAttribute{
			Computed: true,
		},
		"is_default_association_route_table": rschema.BoolAttribute{
			Computed: true,
		},
		"is_default_propagation_route_table": rschema.BoolAttribute{
			Computed: true,
		},
		"provisioning_status": rschema.StringAttribute{
			Computed: true,
		},
		"created_at": rschema.StringAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"updated_at": rschema.StringAttribute{
			Computed: true,
		},
		"routes": rschema.SetNestedAttribute{
			Optional: true,
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
			},
			NestedObject: rschema.NestedAttributeObject{
				Attributes: map[string]rschema.Attribute{
					"id": rschema.StringAttribute{
						Computed: true,
					},
					"destination_cidr_block": rschema.StringAttribute{
						Required:   true,
						CustomType: cidrtypes.IPPrefixType{},
					},
					"tgw_attachment_id": rschema.StringAttribute{
						Required:   true,
						Validators: common.UuidValidator(),
					},
				},
			},
		},
		"associations": rschema.ListNestedAttribute{
			Computed: true,
			NestedObject: rschema.NestedAttributeObject{
				Attributes: map[string]rschema.Attribute{
					"id": rschema.StringAttribute{
						Computed: true,
					},
					"tgw_attachment_id": rschema.StringAttribute{
						Computed: true,
					},
				},
			},
		},
	}
}

func getTransitGatewayRouteTableDataSourceSchema() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"name": dschema.StringAttribute{
			Computed: true,
		},
		"tgw_id": dschema.StringAttribute{
			Computed: true,
		},
		"region": dschema.StringAttribute{
			Computed: true,
		},
		"project_id": dschema.StringAttribute{
			Computed: true,
		},
		"project_name": dschema.StringAttribute{
			Computed: true,
		},
		"tgw_name": dschema.StringAttribute{
			Computed: true,
		},
		"is_default_association_route_table": dschema.BoolAttribute{
			Computed: true,
		},
		"is_default_propagation_route_table": dschema.BoolAttribute{
			Computed: true,
		},
		"provisioning_status": dschema.StringAttribute{
			Computed: true,
		},
		"created_at": dschema.StringAttribute{
			Computed: true,
		},
		"updated_at": dschema.StringAttribute{
			Computed: true,
		},
		"routes": dschema.ListNestedAttribute{
			Computed: true,
			NestedObject: dschema.NestedAttributeObject{
				Attributes: map[string]dschema.Attribute{
					"id": dschema.StringAttribute{
						Computed: true,
					},
					"route_type": dschema.StringAttribute{
						Computed: true,
					},
					"destination_cidr_block": dschema.StringAttribute{
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
				},
			},
		},
		"associations": dschema.ListNestedAttribute{
			Computed: true,
			NestedObject: dschema.NestedAttributeObject{
				Attributes: map[string]dschema.Attribute{
					"id": dschema.StringAttribute{
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
				},
			},
		},
	}
}

var transitGatewayRouteTableResourceSchemaAttributes = getTransitGatewayRouteTableResourceSchema()
var transitGatewayRouteTableDataSourceSchemaAttributes = getTransitGatewayRouteTableDataSourceSchema()

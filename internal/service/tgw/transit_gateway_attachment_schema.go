// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package tgw

import (
	"terraform-provider-kakaocloud/internal/common"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var transitGatewayAttachmentResourceSchemaAttributes = map[string]rschema.Attribute{
	"id": rschema.StringAttribute{
		Computed: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	},
	"tgw_id": rschema.StringAttribute{
		Required:   true,
		Validators: common.UuidValidator(),
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	},
	"resource_id": rschema.StringAttribute{
		Required:   true,
		Validators: common.UuidValidator(),
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	},
	"subnet_ids": rschema.SetAttribute{
		ElementType: types.StringType,
		Required:    true,
		Validators: []validator.Set{
			setvalidator.SizeAtLeast(1),
			setvalidator.ValueStringsAre(common.UuidValidator()...),
		},
	},
	"provisioning_status": rschema.StringAttribute{
		Computed: true,
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
	"resource_type": rschema.StringAttribute{
		Computed: true,
	},
	"resource_name": rschema.StringAttribute{
		Computed: true,
	},
	"resource_cidr_block": rschema.StringAttribute{
		Computed: true,
	},
	"tgw": rschema.SingleNestedAttribute{
		Computed:    true,
		Description: "Connected Transit Gateway information",
		Attributes: map[string]rschema.Attribute{
			"id": rschema.StringAttribute{
				Computed: true,
			},
			"name": rschema.StringAttribute{
				Computed: true,
			},
			"project_id": rschema.StringAttribute{
				Computed: true,
			},
			"project_name": rschema.StringAttribute{
				Computed: true,
			},
		},
	},
	"resources": rschema.ListNestedAttribute{
		Computed:    true,
		Description: "List of attached resources (subnets)",
		NestedObject: rschema.NestedAttributeObject{
			Attributes: map[string]rschema.Attribute{
				"id": rschema.StringAttribute{
					Computed: true,
				},
				"name": rschema.StringAttribute{
					Computed: true,
				},
				"description": rschema.StringAttribute{
					Computed: true,
				},
				"availability_zone": rschema.StringAttribute{
					Computed: true,
				},
				"cidr_block": rschema.StringAttribute{
					Computed: true,
				},
				"operating_status": rschema.StringAttribute{
					Computed: true,
				},
				"provisioning_status": rschema.StringAttribute{
					Computed: true,
				},
				"vpc_id": rschema.StringAttribute{
					Computed: true,
				},
				"created_at": rschema.StringAttribute{
					Computed: true,
				},
				"updated_at": rschema.StringAttribute{
					Computed: true,
				},
			},
		},
	},
	"route_table": rschema.SingleNestedAttribute{
		Computed:    true,
		Description: "Route table information",
		Attributes: map[string]rschema.Attribute{
			"id": rschema.StringAttribute{
				Computed: true,
			},
			"name": rschema.StringAttribute{
				Computed: true,
			},
			"project_id": rschema.StringAttribute{
				Computed: true,
			},
			"region": rschema.StringAttribute{
				Computed: true,
			},
			"tgw_id": rschema.StringAttribute{
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
			},
			"updated_at": rschema.StringAttribute{
				Computed: true,
			},
		},
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
}

var transitGatewayAttachmentDataSourceSchemaAttributes = map[string]dschema.Attribute{
	"provisioning_status": dschema.StringAttribute{
		Computed: true,
	},
	"project_id": dschema.StringAttribute{
		Computed: true,
	},
	"project_name": dschema.StringAttribute{
		Computed: true,
	},
	"resource_type": dschema.StringAttribute{
		Computed: true,
	},
	"resource_id": dschema.StringAttribute{
		Computed: true,
	},
	"resource_name": dschema.StringAttribute{
		Computed: true,
	},
	"resource_cidr_block": dschema.StringAttribute{
		Computed: true,
	},
	"tgw": dschema.SingleNestedAttribute{
		Computed:    true,
		Description: "Connected Transit Gateway information",
		Attributes: map[string]dschema.Attribute{
			"id": dschema.StringAttribute{
				Computed: true,
			},
			"name": dschema.StringAttribute{
				Computed: true,
			},
			"project_id": dschema.StringAttribute{
				Computed: true,
			},
			"project_name": dschema.StringAttribute{
				Computed: true,
			},
		},
	},
	"resources": dschema.ListNestedAttribute{
		Computed:    true,
		Description: "List of attached resources (subnets)",
		NestedObject: dschema.NestedAttributeObject{
			Attributes: map[string]dschema.Attribute{
				"id": dschema.StringAttribute{
					Computed: true,
				},
				"name": dschema.StringAttribute{
					Computed: true,
				},
				"description": dschema.StringAttribute{
					Computed: true,
				},
				"availability_zone": dschema.StringAttribute{
					Computed: true,
				},
				"cidr_block": dschema.StringAttribute{
					Computed: true,
				},
				"operating_status": dschema.StringAttribute{
					Computed: true,
				},
				"provisioning_status": dschema.StringAttribute{
					Computed: true,
				},
				"vpc_id": dschema.StringAttribute{
					Computed: true,
				},
				"created_at": dschema.StringAttribute{
					Computed: true,
				},
				"updated_at": dschema.StringAttribute{
					Computed: true,
				},
			},
		},
	},
	"route_table": dschema.SingleNestedAttribute{
		Computed:    true,
		Description: "Route table information",
		Attributes: map[string]dschema.Attribute{
			"id": dschema.StringAttribute{
				Computed: true,
			},
			"name": dschema.StringAttribute{
				Computed: true,
			},
			"project_id": dschema.StringAttribute{
				Computed: true,
			},
			"region": dschema.StringAttribute{
				Computed: true,
			},
			"tgw_id": dschema.StringAttribute{
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
		},
	},
	"created_at": dschema.StringAttribute{
		Computed: true,
	},
	"updated_at": dschema.StringAttribute{
		Computed: true,
	},
}

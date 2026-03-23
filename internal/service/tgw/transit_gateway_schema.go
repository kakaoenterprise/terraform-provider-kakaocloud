package // Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
tgw

import (
	"terraform-provider-kakaocloud/internal/common"

	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func getOptionsResourceSchema() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"is_auto_accept_shared_attachments": rschema.BoolAttribute{
			Required: true,
		},
		"is_default_route_table_association": rschema.BoolAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.Bool{
				boolplanmodifier.UseStateForUnknown(),
			},
		},
		"association_default_route_table_id": rschema.StringAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
	}
}

func getTransitGatewayResourceSchema() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"name": rschema.StringAttribute{
			Required:   true,
			Validators: common.NameValidator(200),
		},
		"region": rschema.StringAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"is_shared": rschema.BoolAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.Bool{
				boolplanmodifier.UseStateForUnknown(),
			},
		},
		"options": rschema.SingleNestedAttribute{
			Required:   true,
			Attributes: getOptionsResourceSchema(),
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
		"owner_project_id": rschema.StringAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"owner_project_name": rschema.StringAttribute{
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
	}
}

func getOptionsDataSourceSchema() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"is_auto_accept_shared_attachments": dschema.BoolAttribute{
			Computed: true,
		},
		"is_default_route_table_association": dschema.BoolAttribute{
			Computed: true,
		},
		"association_default_route_table_id": dschema.StringAttribute{
			Computed: true,
		},
	}
}

func getTransitGatewayDataSourceSchema() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"name": dschema.StringAttribute{
			Computed: true,
		},
		"region": dschema.StringAttribute{
			Computed: true,
		},
		"is_shared": dschema.BoolAttribute{
			Computed: true,
		},
		"attachments": dschema.ListNestedAttribute{
			Computed: true,
			NestedObject: dschema.NestedAttributeObject{
				Attributes: map[string]dschema.Attribute{
					"id": dschema.StringAttribute{
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
					"tgw_id": dschema.StringAttribute{
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
		},
		"options": dschema.SingleNestedAttribute{
			Computed:   true,
			Attributes: getOptionsDataSourceSchema(),
		},
		"provisioning_status": dschema.StringAttribute{
			Computed: true,
		},
		"project_id": dschema.StringAttribute{
			Computed: true,
		},
		"project_name": dschema.StringAttribute{
			Computed: true,
		},
		"owner_project_id": dschema.StringAttribute{
			Computed: true,
		},
		"owner_project_name": dschema.StringAttribute{
			Computed: true,
		},
		"route_tables": dschema.ListNestedAttribute{
			Computed: true,
			NestedObject: dschema.NestedAttributeObject{
				Attributes: map[string]dschema.Attribute{
					"id": dschema.StringAttribute{
						Computed: true,
					},
					"name": dschema.StringAttribute{
						Computed: true,
					},
					"region": dschema.StringAttribute{
						Computed: true,
					},
					"project_id": dschema.StringAttribute{
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
		},
		"created_at": dschema.StringAttribute{
			Computed: true,
		},
		"updated_at": dschema.StringAttribute{
			Computed: true,
		},
	}
}

var transitGatewayResourceSchemaAttributes = getTransitGatewayResourceSchema()
var transitGatewayDataSourceSchemaAttributes = getTransitGatewayDataSourceSchema()

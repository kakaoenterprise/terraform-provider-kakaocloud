// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package vpc

import (
	"terraform-provider-kakaocloud/internal/common"

	"github.com/hashicorp/terraform-plugin-framework-nettypes/cidrtypes"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func getIgwDataSourceSchema() map[string]dschema.Attribute {
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
		"region": dschema.StringAttribute{
			Computed: true,
		},
		"project_id": dschema.StringAttribute{
			Computed: true,
		},
		"operating_status": dschema.StringAttribute{
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
	}
}

func getDefaultRouteTableDataSourceSchema() map[string]dschema.Attribute {
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
		"provisioning_status": dschema.StringAttribute{
			Computed: true,
		},
		"created_at": dschema.StringAttribute{
			Computed: true,
		},
		"updated_at": dschema.StringAttribute{
			Computed: true,
		},
	}
}

func getVpcDataSourceSchema() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"name": dschema.StringAttribute{
			Computed: true,
		},
		"description": dschema.StringAttribute{
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
		"cidr_block": dschema.StringAttribute{
			Computed:   true,
			CustomType: cidrtypes.IPPrefixType{},
		},
		"is_default": dschema.BoolAttribute{
			Computed: true,
		},
		"provisioning_status": dschema.StringAttribute{
			Computed: true,
		},
		"is_enable_dns_support": dschema.BoolAttribute{
			Computed: true,
		},
		"created_at": dschema.StringAttribute{
			Computed: true,
		},
		"updated_at": dschema.StringAttribute{
			Computed: true,
		},
		"igw": dschema.SingleNestedAttribute{
			Computed:   true,
			Attributes: igwDataSourceSchema,
		},
		"default_route_table": dschema.SingleNestedAttribute{
			Computed:   true,
			Attributes: defaultRouteTableDataSourceSchema,
		},
	}
}

func getIgwResourceSchema() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			Computed: true,
		},
		"name": rschema.StringAttribute{
			Computed: true,
		},
		"description": rschema.StringAttribute{
			Computed: true,
		},
		"region": rschema.StringAttribute{
			Computed: true,
		},
		"project_id": rschema.StringAttribute{
			Computed: true,
		},
		"operating_status": rschema.StringAttribute{
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
	}
}

func getDefaultRouteTableResourceSchema() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			Computed: true,
		},
		"name": rschema.StringAttribute{
			Computed: true,
		},
		"description": rschema.StringAttribute{
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
	}
}

func getSubnetResourceSchema() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"cidr_block": rschema.StringAttribute{
			Required:   true,
			CustomType: cidrtypes.IPPrefixType{},
			Validators: []validator.String{
				common.NewCIDRPrefixLengthValidator(20, 26),
			},
		},
		"availability_zone": rschema.StringAttribute{
			Required: true,
		},
	}
}

func getVpcResourceSchema() map[string]rschema.Attribute {
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
		"description": rschema.StringAttribute{
			Computed: true,
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
		"cidr_block": rschema.StringAttribute{
			Required:   true,
			CustomType: cidrtypes.IPPrefixType{},
			Validators: []validator.String{
				common.NewCIDRPrefixLengthValidator(16, 24),
			},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"is_default": rschema.BoolAttribute{
			Computed: true,
		},
		"provisioning_status": rschema.StringAttribute{
			Computed: true,
		},
		"is_enable_dns_support": rschema.BoolAttribute{
			Computed: true,
		},
		"created_at": rschema.StringAttribute{
			Computed: true,
		},
		"updated_at": rschema.StringAttribute{
			Computed: true,
		},
		"igw": rschema.SingleNestedAttribute{
			Computed:   true,
			Attributes: igwResourceSchema,
		},
		"default_route_table": rschema.SingleNestedAttribute{
			Computed:   true,
			Attributes: defaultRouteTableResourceSchema,
		},
		"subnet": rschema.SingleNestedAttribute{
			Optional: true,
			PlanModifiers: []planmodifier.Object{
				objectplanmodifier.RequiresReplace(),
			},
			Attributes: subnetResourceSchema,
		},
	}
}

var igwDataSourceSchema = getIgwDataSourceSchema()
var defaultRouteTableDataSourceSchema = getDefaultRouteTableDataSourceSchema()
var vpcDataSourceSchemaAttributes = getVpcDataSourceSchema()

var igwResourceSchema = getIgwResourceSchema()
var defaultRouteTableResourceSchema = getDefaultRouteTableResourceSchema()
var subnetResourceSchema = getSubnetResourceSchema()
var vpcResourceSchemaAttributes = getVpcResourceSchema()

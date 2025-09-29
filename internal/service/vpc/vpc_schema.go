// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package vpc

import (
	"terraform-provider-kakaocloud/internal/common"
	"terraform-provider-kakaocloud/internal/docs"

	"github.com/hashicorp/terraform-plugin-framework-nettypes/cidrtypes"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func getIgwDataSourceSchema() map[string]dschema.Attribute {
	desc := docs.Vpc("bns_vpc__v1__api__get_vpc__model__IgwModel")

	return map[string]dschema.Attribute{
		"id": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("id"),
		},
		"name": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("name"),
		},
		"description": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("description"),
		},
		"region": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("region"),
		},
		"project_id": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("project_id"),
		},
		"operating_status": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("operating_status"),
		},
		"provisioning_status": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("provisioning_status"),
		},
		"created_at": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("created_at"),
		},
		"updated_at": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("updated_at"),
		},
	}
}

func getDefaultRouteTableDataSourceSchema() map[string]dschema.Attribute {
	desc := docs.Vpc("bns_vpc__v1__api__get_vpc__model__RouteTableModel")

	return map[string]dschema.Attribute{
		"id": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("id"),
		},
		"name": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("name"),
		},
		"description": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("description"),
		},
		"provisioning_status": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("provisioning_status"),
		},
		"created_at": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("created_at"),
		},
		"updated_at": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("updated_at"),
		},
	}
}

func getVpcDataSourceSchema() map[string]dschema.Attribute {
	desc := docs.Vpc("bns_vpc__v1__api__get_vpc__model__VpcModel")

	return map[string]dschema.Attribute{
		"name": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("name"),
		},
		"description": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("description"),
		},
		"region": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("region"),
		},
		"project_id": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("project_id"),
		},
		"project_name": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("project_name"),
		},
		"cidr_block": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("cidr_block"),
			CustomType:  cidrtypes.IPPrefixType{},
		},
		"is_default": dschema.BoolAttribute{
			Computed:    true,
			Description: desc.String("is_default"),
		},
		"provisioning_status": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("provisioning_status"),
		},
		"is_enable_dns_support": dschema.BoolAttribute{
			Computed:    true,
			Description: desc.String("is_enable_dns_support"),
		},
		"created_at": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("created_at"),
		},
		"updated_at": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("updated_at"),
		},
		"igw": dschema.SingleNestedAttribute{
			Computed:    true,
			Description: desc.String("igw"),
			Attributes:  igwDataSourceSchema,
		},
		"default_route_table": dschema.SingleNestedAttribute{
			Computed:    true,
			Description: desc.String("default_route_table"),
			Attributes:  defaultRouteTableDataSourceSchema,
		},
	}
}

func getIgwResourceSchema() map[string]rschema.Attribute {
	desc := docs.Vpc("bns_vpc__v1__api__get_vpc__model__IgwModel")

	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("id"),
		},
		"name": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("name"),
		},
		"description": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("description"),
		},
		"region": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("region"),
		},
		"project_id": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("project_id"),
		},
		"operating_status": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("operating_status"),
		},
		"provisioning_status": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("provisioning_status"),
		},
		"created_at": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("created_at"),
		},
		"updated_at": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("updated_at"),
		},
	}
}

func getDefaultRouteTableResourceSchema() map[string]rschema.Attribute {
	desc := docs.Vpc("bns_vpc__v1__api__get_vpc__model__RouteTableModel")

	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("id"),
		},
		"name": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("name"),
		},
		"description": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("description"),
		},
		"provisioning_status": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("provisioning_status"),
		},
		"created_at": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("created_at"),
		},
		"updated_at": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("updated_at"),
		},
	}
}

func getSubnetResourceSchema() map[string]rschema.Attribute {
	desc := docs.Vpc("CreateSubnetModel")

	return map[string]rschema.Attribute{
		"cidr_block": rschema.StringAttribute{
			Required:    true,
			Description: desc.String("cidr_block"),
			CustomType:  cidrtypes.IPPrefixType{},
			Validators: []validator.String{
				common.NewCIDRPrefixLengthValidator(20, 26),
			},
		},
		"availability_zone": rschema.StringAttribute{
			Required:    true,
			Description: desc.String("availability_zone"),
		},
	}
}

func getVpcResourceSchema() map[string]rschema.Attribute {
	desc := docs.Vpc("bns_vpc__v1__api__get_vpc__model__VpcModel")
	createDesc := docs.Vpc("CreateVPCModel")

	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("id"),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"name": rschema.StringAttribute{
			Required:    true,
			Description: desc.String("name"),
			Validators:  common.NameValidator(200),
		},
		"description": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("description"),
		},
		"region": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("region"),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"project_id": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("project_id"),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"project_name": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("project_name"),
		},
		"cidr_block": rschema.StringAttribute{
			Required:    true,
			Description: createDesc.String("cidr_block"),
			CustomType:  cidrtypes.IPPrefixType{},
			Validators: []validator.String{
				common.NewCIDRPrefixLengthValidator(16, 24),
			},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"is_default": rschema.BoolAttribute{
			Computed:    true,
			Description: desc.String("is_default"),
		},
		"provisioning_status": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("provisioning_status"),
		},
		"is_enable_dns_support": rschema.BoolAttribute{
			Computed:    true,
			Description: desc.String("is_enable_dns_support"),
		},
		"created_at": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("created_at"),
		},
		"updated_at": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("updated_at"),
		},
		"igw": rschema.SingleNestedAttribute{
			Computed:    true,
			Description: desc.String("igw"),
			Attributes:  igwResourceSchema,
		},
		"default_route_table": rschema.SingleNestedAttribute{
			Computed:    true,
			Description: desc.String("default_route_table"),
			Attributes:  defaultRouteTableResourceSchema,
		},
		"subnet": rschema.SingleNestedAttribute{
			Optional:    true,
			Description: createDesc.String("subnet"),
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

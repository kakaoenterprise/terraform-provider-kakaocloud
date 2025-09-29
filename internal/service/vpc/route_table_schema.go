// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package vpc

import (
	"terraform-provider-kakaocloud/internal/common"
	"terraform-provider-kakaocloud/internal/docs"

	"github.com/hashicorp/terraform-plugin-framework-nettypes/cidrtypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/kakaoenterprise/kc-sdk-go/services/vpc"
)

func getAssociationDataSourceSchema() map[string]dschema.Attribute {
	desc := docs.Vpc("bns_vpc__v1__api__get_route_table__model__AssociationModel")

	return map[string]dschema.Attribute{
		"id": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("id"),
		},
		"provisioning_status": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("provisioning_status"),
		},
		"vpc_id": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("vpc_id"),
		},
		"vpc_name": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("vpc_name"),
		},
		"subnet_id": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("subnet_id"),
		},
		"subnet_name": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("subnet_name"),
		},
		"subnet_cidr_block": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("subnet_cidr_block"),
		},
		"availability_zone": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("availability_zone"),
		},
	}
}

func getRouteDataSourceSchema() map[string]dschema.Attribute {
	desc := docs.Vpc("bns_vpc__v1__api__get_route_table__model__RouteModel")

	return map[string]dschema.Attribute{
		"id": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("id"),
		},
		"destination": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("destination"),
		},
		"provisioning_status": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("provisioning_status"),
		},
		"target_type": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("target_type"),
		},
		"target_name": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("target_name"),
		},
		"is_local_route": dschema.BoolAttribute{
			Computed:    true,
			Description: desc.String("is_local_route"),
		},
		"target_id": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("target_id"),
		},
	}
}

func getRouteTableDataSourceSchema() map[string]dschema.Attribute {
	desc := docs.Vpc("bns_vpc__v1__api__get_route_table__model__RouteTableModel")

	return map[string]dschema.Attribute{
		"name": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("name"),
		},
		"associations": dschema.ListNestedAttribute{
			Computed:    true,
			Description: desc.String("associations"),
			NestedObject: dschema.NestedAttributeObject{
				Attributes: associationDataSourceSchema,
			},
		},
		"routes": dschema.ListNestedAttribute{
			Computed:    true,
			Description: desc.String("routes"),
			NestedObject: dschema.NestedAttributeObject{
				Attributes: routeDataSourceSchema,
			},
		},
		"vpc_id": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("vpc_id"),
		},
		"provisioning_status": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("provisioning_status"),
		},
		"vpc_name": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("vpc_name"),
		},
		"vpc_provisioning_status": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("vpc_provisioning_status"),
		},
		"project_id": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("project_id"),
		},
		"project_name": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("project_name"),
		},
		"is_main": dschema.BoolAttribute{
			Computed:    true,
			Description: desc.String("is_main"),
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

func getAssociationResourceSchema() map[string]rschema.Attribute {
	desc := docs.Vpc("bns_vpc__v1__api__get_route_table__model__AssociationModel")

	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("id"),
		},
		"provisioning_status": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("provisioning_status"),
		},
		"vpc_id": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("vpc_id"),
		},
		"vpc_name": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("vpc_name"),
		},
		"subnet_id": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("subnet_id"),
		},
		"subnet_name": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("subnet_name"),
		},
		"subnet_cidr_block": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("subnet_cidr_block"),
		},
		"availability_zone": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("availability_zone"),
		},
	}
}

func getRouteResourceSchema() map[string]rschema.Attribute {
	desc := docs.Vpc("bns_vpc__v1__api__get_route_table__model__RouteModel")

	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("id"),
		},
		"destination": rschema.StringAttribute{
			Required:    true,
			Description: desc.String("destination"),
			CustomType:  cidrtypes.IPPrefixType{},
		},
		"provisioning_status": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("provisioning_status"),
		},
		"target_type": rschema.StringAttribute{
			Required:    true,
			Description: desc.String("target_type"),
		},
		"target_name": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("target_name"),
		},
		"is_local_route": rschema.BoolAttribute{
			Computed:    true,
			Description: desc.String("is_local_route"),
		},
		"target_id": rschema.StringAttribute{
			Required:    true,
			Description: desc.String("target_id"),
			Validators:  common.UuidValidator(),
		},
	}
}

func getRequestRouteResourceSchema() map[string]rschema.Attribute {
	desc := docs.Vpc("bns_vpc__v1__api__get_route_table__model__RouteModel")

	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("id"),
		},
		"destination": rschema.StringAttribute{
			Required:    true,
			Description: desc.String("destination"),
			CustomType:  cidrtypes.IPPrefixType{},
		},
		"target_type": rschema.StringAttribute{
			Required:    true,
			Description: desc.String("target_type"),
			Validators: []validator.String{
				stringvalidator.OneOf(
					string(vpc.ROUTETABLEROUTETYPE_INSTANCE),
					string(vpc.ROUTETABLEROUTETYPE_IGW),
					string(vpc.ROUTETABLEROUTETYPE_TGW),
				),
			},
		},
		"target_id": rschema.StringAttribute{
			Required:    true,
			Description: desc.String("target_id"),
			Validators:  common.UuidValidator(),
		},
	}
}

func getRouteTableResourceSchema() map[string]rschema.Attribute {
	desc := docs.Vpc("bns_vpc__v1__api__get_route_table__model__RouteTableModel")

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
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"associations": rschema.ListNestedAttribute{
			Computed:    true,
			Description: desc.String("associations"),
			NestedObject: rschema.NestedAttributeObject{
				Attributes: associationResourceSchema,
			},
		},
		"routes": rschema.ListNestedAttribute{
			Computed:    true,
			Description: desc.String("routes"),
			NestedObject: rschema.NestedAttributeObject{
				Attributes: routeResourceSchema,
			},
		},
		"vpc_id": rschema.StringAttribute{
			Required:    true,
			Description: desc.String("vpc_id"),
			Validators:  common.UuidValidator(),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"provisioning_status": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("provisioning_status"),
		},
		"vpc_name": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("vpc_name"),
		},
		"vpc_provisioning_status": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("vpc_provisioning_status"),
		},
		"project_id": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("project_id"),
		},
		"project_name": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("project_name"),
		},
		"is_main": rschema.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Description: desc.String("is_main"),
		},
		"request_routes": rschema.ListNestedAttribute{
			Optional:    true,
			Description: desc.String("routes"),
			Validators: []validator.List{
				listvalidator.SizeAtLeast(1),
			},
			NestedObject: rschema.NestedAttributeObject{
				Attributes: requestRouteResourceSchema,
			},
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

var associationDataSourceSchema = getAssociationDataSourceSchema()
var routeDataSourceSchema = getRouteDataSourceSchema()
var routeTableDataSourceSchemaAttributes = getRouteTableDataSourceSchema()

var associationResourceSchema = getAssociationResourceSchema()
var routeResourceSchema = getRouteResourceSchema()
var requestRouteResourceSchema = getRequestRouteResourceSchema()
var routeTableResourceSchemaAttributes = getRouteTableResourceSchema()

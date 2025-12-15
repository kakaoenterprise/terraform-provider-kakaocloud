// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package vpc

import (
	"terraform-provider-kakaocloud/internal/common"

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

	return map[string]dschema.Attribute{
		"id": dschema.StringAttribute{
			Computed: true,
		},
		"provisioning_status": dschema.StringAttribute{
			Computed: true,
		},
		"vpc_id": dschema.StringAttribute{
			Computed: true,
		},
		"vpc_name": dschema.StringAttribute{
			Computed: true,
		},
		"subnet_id": dschema.StringAttribute{
			Computed: true,
		},
		"subnet_name": dschema.StringAttribute{
			Computed: true,
		},
		"subnet_cidr_block": dschema.StringAttribute{
			Computed: true,
		},
		"availability_zone": dschema.StringAttribute{
			Computed: true,
		},
	}
}

func getRouteDataSourceSchema() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"id": dschema.StringAttribute{
			Computed: true,
		},
		"destination": dschema.StringAttribute{
			Computed: true,
		},
		"provisioning_status": dschema.StringAttribute{
			Computed: true,
		},
		"target_type": dschema.StringAttribute{
			Computed: true,
		},
		"target_name": dschema.StringAttribute{
			Computed: true,
		},
		"is_local_route": dschema.BoolAttribute{
			Computed: true,
		},
		"target_id": dschema.StringAttribute{
			Computed: true,
		},
	}
}

func getRouteTableDataSourceSchema() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"name": dschema.StringAttribute{
			Computed: true,
		},
		"associations": dschema.ListNestedAttribute{
			Computed: true,
			NestedObject: dschema.NestedAttributeObject{
				Attributes: associationDataSourceSchema,
			},
		},
		"routes": dschema.ListNestedAttribute{
			Computed: true,
			NestedObject: dschema.NestedAttributeObject{
				Attributes: routeDataSourceSchema,
			},
		},
		"vpc_id": dschema.StringAttribute{
			Computed: true,
		},
		"provisioning_status": dschema.StringAttribute{
			Computed: true,
		},
		"vpc_name": dschema.StringAttribute{
			Computed: true,
		},
		"vpc_provisioning_status": dschema.StringAttribute{
			Computed: true,
		},
		"project_id": dschema.StringAttribute{
			Computed: true,
		},
		"project_name": dschema.StringAttribute{
			Computed: true,
		},
		"is_main": dschema.BoolAttribute{
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

func getAssociationResourceSchema() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			Computed: true,
		},
		"provisioning_status": rschema.StringAttribute{
			Computed: true,
		},
		"vpc_id": rschema.StringAttribute{
			Computed: true,
		},
		"vpc_name": rschema.StringAttribute{
			Computed: true,
		},
		"subnet_id": rschema.StringAttribute{
			Computed: true,
		},
		"subnet_name": rschema.StringAttribute{
			Computed: true,
		},
		"subnet_cidr_block": rschema.StringAttribute{
			Computed: true,
		},
		"availability_zone": rschema.StringAttribute{
			Computed: true,
		},
	}
}

func getRouteResourceSchema() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			Computed: true,
		},
		"destination": rschema.StringAttribute{
			Required:   true,
			CustomType: cidrtypes.IPPrefixType{},
		},
		"provisioning_status": rschema.StringAttribute{
			Computed: true,
		},
		"target_type": rschema.StringAttribute{
			Required: true,
		},
		"target_name": rschema.StringAttribute{
			Computed: true,
		},
		"is_local_route": rschema.BoolAttribute{
			Computed: true,
		},
		"target_id": rschema.StringAttribute{
			Required:   true,
			Validators: common.UuidValidator(),
		},
	}
}

func getRequestRouteResourceSchema() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			Computed: true,
		},
		"destination": rschema.StringAttribute{
			Required:   true,
			CustomType: cidrtypes.IPPrefixType{},
		},
		"target_type": rschema.StringAttribute{
			Required: true,
			Validators: []validator.String{
				stringvalidator.OneOf(
					string(vpc.ROUTETABLEROUTETYPE_INSTANCE),
					string(vpc.ROUTETABLEROUTETYPE_IGW),
					string(vpc.ROUTETABLEROUTETYPE_TGW),
				),
			},
		},
		"target_id": rschema.StringAttribute{
			Required:   true,
			Validators: common.UuidValidator(),
		},
	}
}

func getRouteTableResourceSchema() map[string]rschema.Attribute {
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
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"associations": rschema.ListNestedAttribute{
			Computed: true,
			NestedObject: rschema.NestedAttributeObject{
				Attributes: associationResourceSchema,
			},
		},
		"routes": rschema.ListNestedAttribute{
			Computed: true,
			NestedObject: rschema.NestedAttributeObject{
				Attributes: routeResourceSchema,
			},
		},
		"vpc_id": rschema.StringAttribute{
			Required:   true,
			Validators: common.UuidValidator(),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"provisioning_status": rschema.StringAttribute{
			Computed: true,
		},
		"vpc_name": rschema.StringAttribute{
			Computed: true,
		},
		"vpc_provisioning_status": rschema.StringAttribute{
			Computed: true,
		},
		"project_id": rschema.StringAttribute{
			Computed: true,
		},
		"project_name": rschema.StringAttribute{
			Computed: true,
		},
		"is_main": rschema.BoolAttribute{
			Optional: true,
			Computed: true,
		},
		"request_routes": rschema.ListNestedAttribute{
			Optional: true,
			Validators: []validator.List{
				listvalidator.SizeAtLeast(1),
			},
			NestedObject: rschema.NestedAttributeObject{
				Attributes: requestRouteResourceSchema,
			},
		},
		"created_at": rschema.StringAttribute{
			Computed: true,
		},
		"updated_at": rschema.StringAttribute{
			Computed: true,
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

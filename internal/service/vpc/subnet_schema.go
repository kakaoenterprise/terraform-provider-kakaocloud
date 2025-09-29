// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package vpc

import (
	"terraform-provider-kakaocloud/internal/common"
	"terraform-provider-kakaocloud/internal/docs"

	"github.com/hashicorp/terraform-plugin-framework-nettypes/cidrtypes"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func getSubnetDataSourceSchema() map[string]dschema.Attribute {
	desc := docs.Vpc("bns_vpc__v1__api__get_subnet__model__SubnetModel")

	return map[string]dschema.Attribute{
		"name": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("name"),
		},
		"is_shared": dschema.BoolAttribute{
			Computed:    true,
			Description: desc.String("is_shared"),
		},
		"availability_zone": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("availability_zone"),
		},
		"cidr_block": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("cidr_block"),
			CustomType:  cidrtypes.IPPrefixType{},
		},
		"project_id": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("project_id"),
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
		"project_name": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("project_name"),
		},
		"owner_project_id": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("owner_project_id"),
		},
		"route_table_id": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("route_table_id"),
		},
		"route_table_name": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("route_table_name"),
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

func getSubnetResourceSchemaAttributes() map[string]rschema.Attribute {
	desc := docs.Vpc("bns_vpc__v1__api__get_subnet__model__SubnetModel")

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
		"is_shared": rschema.BoolAttribute{
			Computed:    true,
			Description: desc.String("is_shared"),
		},
		"availability_zone": rschema.StringAttribute{
			Required:    true,
			Description: desc.String("availability_zone"),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"cidr_block": rschema.StringAttribute{
			Required:    true,
			Description: desc.String("cidr_block"),
			CustomType:  cidrtypes.IPPrefixType{},
			Validators: []validator.String{
				common.NewCIDRPrefixLengthValidator(20, 26),
			},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"project_id": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("project_id"),
		},
		"provisioning_status": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("provisioning_status"),
		},
		"vpc_id": rschema.StringAttribute{
			Required:    true,
			Description: desc.String("vpc_id"),
			Validators:  common.UuidValidator(),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"vpc_name": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("vpc_name"),
		},
		"project_name": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("project_name"),
		},
		"owner_project_id": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("owner_project_id"),
		},
		"route_table_id": rschema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: desc.String("route_table_id"),
			Validators:  common.UuidValidator(),
		},
		"route_table_name": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("route_table_name"),
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

var subnetDataSourceSchemaAttributes = getSubnetDataSourceSchema()
var subnetResourceSchemaAttributes = getSubnetResourceSchemaAttributes()

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package loadbalancer

import (
	"terraform-provider-kakaocloud/internal/common"
	"terraform-provider-kakaocloud/internal/docs"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kakaoenterprise/kc-sdk-go/services/loadbalancer"
)

func getBeyondLoadBalancerResourceSchema() map[string]rschema.Attribute {
	desc := docs.Loadbalancer("bns_load_balancer__v1__api__get_ha_group__model__BeyondLoadBalancerModel")

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
			Validators:  common.NameValidator(32),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"description": rschema.StringAttribute{
			Optional:    true,
			Description: desc.String("description"),
			Validators:  common.DescriptionValidator(),
		},
		"provider_name": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("provider"),
		},
		"scheme": rschema.StringAttribute{
			Required:    true,
			Description: desc.String("scheme"),
			Validators: []validator.String{stringvalidator.OneOf(
				string(loadbalancer.SCHEME_INTERNET_FACING),
				string(loadbalancer.SCHEME_INTERNAL),
			)},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"project_id": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("project_id"),
		},
		"dns_name": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("dns_name"),
		},
		"type_id": rschema.StringAttribute{
			Required:    true,
			Description: desc.String("type_id"),
			Validators:  common.UuidValidator(),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
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
		"provisioning_status": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("provisioning_status"),
		},
		"operating_status": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("operating_status"),
		},
		"vpc_id": rschema.StringAttribute{
			Required:    true,
			Description: desc.String("vpc_id"),
			Validators:  common.UuidValidator(),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"type": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("type"),
		},
		"vpc_name": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("vpc_name"),
		},
		"vpc_cidr_block": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("vpc_cidr_block"),
		},
		"availability_zones": rschema.ListAttribute{
			ElementType: types.StringType,
			Computed:    true,
			Description: desc.String("availability_zones"),
		},
		"load_balancers": rschema.ListNestedAttribute{
			Required:    true,
			Description: desc.String("load_balancers"),
			NestedObject: rschema.NestedAttributeObject{
				Attributes: getBeyondLoadBalancerLoadBalancerResourceSchemaAttributes(),
			},
		},
		"attached_load_balancers": rschema.SetNestedAttribute{
			Required:    true,
			Description: "Request List of load balancers belonging to the HA group",
			NestedObject: rschema.NestedAttributeObject{
				Attributes: map[string]rschema.Attribute{
					"id": rschema.StringAttribute{
						Required:    true,
						Description: "Load balancer ID",
						Validators:  common.UuidValidator(),
					},
					"availability_zone": rschema.StringAttribute{
						Required:    true,
						Description: "Availability zone",
					},
				},
			},
		},
	}
}

func getBeyondLoadBalancerLoadBalancerResourceSchemaAttributes() map[string]rschema.Attribute {
	desc := docs.Loadbalancer("bns_load_balancer__v1__api__get_ha_group__model__LoadBalancerModel")

	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("id"),
			Validators:  common.UuidValidator(),
		},
		"name": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("name"),
		},
		"description": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("description"),
		},
		"type": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("type"),
		},
		"provisioning_status": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("provisioning_status"),
		},
		"operating_status": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("operating_status"),
		},
		"availability_zone": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("availability_zone"),
		},
		"type_id": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("type_id"),
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

func getBeyondLoadBalancerDataSourceSchema() map[string]dschema.Attribute {
	desc := docs.Loadbalancer("bns_load_balancer__v1__api__get_ha_group__model__BeyondLoadBalancerModel")

	return map[string]dschema.Attribute{
		"name": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("name"),
		},
		"description": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("description"),
		},
		"provider_name": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("provider"),
		},
		"scheme": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("scheme"),
		},
		"project_id": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("project_id"),
		},
		"dns_name": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("dns_name"),
		},
		"type_id": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("type_id"),
		},
		"created_at": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("created_at"),
		},
		"updated_at": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("updated_at"),
		},
		"provisioning_status": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("provisioning_status"),
		},
		"operating_status": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("operating_status"),
		},
		"vpc_id": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("vpc_id"),
		},
		"type": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("type"),
		},
		"vpc_name": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("vpc_name"),
		},
		"vpc_cidr_block": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("vpc_cidr_block"),
		},
		"availability_zones": dschema.ListAttribute{
			ElementType: types.StringType,
			Computed:    true,
			Description: desc.String("availability_zones"),
		},
		"load_balancers": dschema.ListNestedAttribute{
			Computed:    true,
			Description: desc.String("load_balancers"),
			NestedObject: dschema.NestedAttributeObject{
				Attributes: getBeyondLoadBalancerLoadBalancerDataSourceSchemaAttributes(),
			},
		},
	}
}

func getBeyondLoadBalancerLoadBalancerDataSourceSchemaAttributes() map[string]dschema.Attribute {
	desc := docs.Loadbalancer("bns_load_balancer__v1__api__get_ha_group__model__LoadBalancerModel")

	return map[string]dschema.Attribute{
		"id":                  dschema.StringAttribute{Computed: true, Description: desc.String("id")},
		"name":                dschema.StringAttribute{Computed: true, Description: desc.String("name")},
		"description":         dschema.StringAttribute{Computed: true, Description: desc.String("description")},
		"type":                dschema.StringAttribute{Computed: true, Description: desc.String("type")},
		"provisioning_status": dschema.StringAttribute{Computed: true, Description: desc.String("provisioning_status")},
		"operating_status":    dschema.StringAttribute{Computed: true, Description: desc.String("operating_status")},
		"availability_zone":   dschema.StringAttribute{Computed: true, Description: desc.String("availability_zone")},
		"type_id":             dschema.StringAttribute{Computed: true, Description: desc.String("type_id")},
		"subnet_id":           dschema.StringAttribute{Computed: true, Description: desc.String("subnet_id")},
		"subnet_name":         dschema.StringAttribute{Computed: true, Description: desc.String("subnet_name")},
		"subnet_cidr_block":   dschema.StringAttribute{Computed: true, Description: desc.String("subnet_cidr_block")},
		"created_at":          dschema.StringAttribute{Computed: true, Description: desc.String("created_at")},
		"updated_at":          dschema.StringAttribute{Computed: true, Description: desc.String("updated_at")},
	}
}

var beyondLoadBalancerResourceSchema = getBeyondLoadBalancerResourceSchema()
var beyondLoadBalancerDatasourceSchema = getBeyondLoadBalancerDataSourceSchema()

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package loadbalancer

import (
	"terraform-provider-kakaocloud/internal/common"
	"terraform-provider-kakaocloud/internal/docs"

	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func getAccessLogsResourceSchema() map[string]rschema.Attribute {
	desc := docs.Loadbalancer("AccessLogsModel")

	return map[string]rschema.Attribute{
		"bucket": rschema.StringAttribute{
			Required:    true,
			Description: desc.String("bucket"),
		},
		"access_key": rschema.StringAttribute{
			Required:    true,
			Sensitive:   true,
			Description: desc.String("access_key"),
		},
		"secret_key": rschema.StringAttribute{
			Required:    true,
			Sensitive:   true,
			Description: desc.String("secret_key"),
		},
	}
}

func getLoadBalancerResourceSchema() map[string]rschema.Attribute {
	desc := docs.Loadbalancer("bns_load_balancer__v1__api__create_load_balancer__model__LoadBalancerModel")
	getDesc := docs.Loadbalancer("bns_load_balancer__v1__api__get_load_balancer__model__LoadBalancerModel")

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
			Validators:  common.NameValidator(250),
		},
		"description": rschema.StringAttribute{
			Optional:    true,
			Description: desc.String("description"),
			Validators:  common.DescriptionValidator(),
		},
		"subnet_id": rschema.StringAttribute{
			Required:    true,
			Description: desc.String("subnet_id"),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"availability_zone": rschema.StringAttribute{
			Required:    true,
			Description: desc.String("availability_zone"),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"flavor_id": rschema.StringAttribute{
			Required:    true,
			Description: desc.String("flavor_id"),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},

		// Computed Attributes
		"type": rschema.StringAttribute{
			Computed:    true,
			Description: getDesc.String("type"),
		},
		"listener_ids": rschema.ListAttribute{
			ElementType: types.StringType,
			Computed:    true,
			Description: getDesc.String("listener_ids"),
		},
		"project_id": rschema.StringAttribute{
			Computed:    true,
			Description: getDesc.String("project_id"),
		},
		"provisioning_status": rschema.StringAttribute{
			Computed:    true,
			Description: getDesc.String("provisioning_status"),
		},
		"operating_status": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("operating_status"),
		},
		"created_at": rschema.StringAttribute{
			Computed:    true,
			Description: getDesc.String("created_at"),
		},
		"updated_at": rschema.StringAttribute{
			Computed:    true,
			Description: getDesc.String("updated_at"),
		},
		"access_logs": rschema.SingleNestedAttribute{
			Optional:    true,
			Description: getDesc.String("access_logs"),
			Attributes:  getAccessLogsResourceSchema(),
		},
		"beyond_load_balancer_id": rschema.StringAttribute{
			Computed:    true,
			Description: getDesc.String("beyond_load_balancer_id"),
		},
		"beyond_load_balancer_name": rschema.StringAttribute{
			Computed:    true,
			Description: getDesc.String("beyond_load_balancer_name"),
		},
		"beyond_load_balancer_dns_name": rschema.StringAttribute{
			Computed:    true,
			Description: getDesc.String("beyond_load_balancer_dns_name"),
		},
		"target_group_count": rschema.Int64Attribute{
			Computed:    true,
			Description: desc.String("target_group_count"),
		},
		"listener_count": rschema.Int64Attribute{
			Computed:    true,
			Description: getDesc.String("listener_count"),
		},
		"private_vip": rschema.StringAttribute{
			Computed:    true,
			Description: getDesc.String("private_vip"),
		},
		"public_vip": rschema.StringAttribute{
			Computed:    true,
			Description: getDesc.String("public_vip"),
		},
		"subnet_name": rschema.StringAttribute{
			Computed:    true,
			Description: getDesc.String("subnet_name"),
		},
		"subnet_cidr_block": rschema.StringAttribute{
			Computed:    true,
			Description: getDesc.String("subnet_cidr_block"),
		},
		"vpc_id": rschema.StringAttribute{
			Computed:    true,
			Description: getDesc.String("vpc_id"),
		},
		"vpc_name": rschema.StringAttribute{
			Computed:    true,
			Description: getDesc.String("vpc_name"),
		},
	}
}

func getLoadBalancerDataSourceSchema() map[string]dschema.Attribute {
	desc := docs.Loadbalancer("bns_load_balancer__v1__api__get_load_balancer__model__LoadBalancerModel")

	return map[string]dschema.Attribute{
		"name": dschema.StringAttribute{
			Description: desc.String("name"),
			Computed:    true,
		},
		"description": dschema.StringAttribute{
			Description: desc.String("description"),
			Computed:    true,
		},
		"type": dschema.StringAttribute{
			Description: desc.String("type"),
			Computed:    true,
		},
		"listener_ids": dschema.ListAttribute{
			Description: desc.String("listener_ids"),
			ElementType: types.StringType,
			Computed:    true,
		},
		"project_id": dschema.StringAttribute{
			Description: desc.String("project_id"),
			Computed:    true,
		},
		"provisioning_status": dschema.StringAttribute{
			Description: desc.String("provisioning_status"),
			Computed:    true,
		},
		"operating_status": dschema.StringAttribute{
			Description: desc.String("operating_status"),
			Computed:    true,
		},
		"created_at": dschema.StringAttribute{
			Description: desc.String("created_at"),
			Computed:    true,
		},
		"updated_at": dschema.StringAttribute{
			Description: desc.String("updated_at"),
			Computed:    true,
		},
		"availability_zone": dschema.StringAttribute{
			Description: desc.String("availability_zone"),
			Computed:    true,
		},
		"access_logs": dschema.StringAttribute{
			Description: desc.String("access_logs"),
			Computed:    true,
		},
		"beyond_load_balancer_id": dschema.StringAttribute{
			Description: desc.String("beyond_load_balancer_id"),
			Computed:    true,
		},
		"beyond_load_balancer_name": dschema.StringAttribute{
			Description: desc.String("beyond_load_balancer_name"),
			Computed:    true,
		},
		"beyond_load_balancer_dns_name": dschema.StringAttribute{
			Description: desc.String("beyond_load_balancer_dns_name"),
			Computed:    true,
		},
		"target_group_count": dschema.Int64Attribute{
			Description: desc.String("target_group_count"),
			Computed:    true,
		},
		"listener_count": dschema.Int64Attribute{
			Description: desc.String("listener_count"),
			Computed:    true,
		},
		"private_vip": dschema.StringAttribute{
			Description: desc.String("private_vip"),
			Computed:    true,
		},
		"public_vip": dschema.StringAttribute{
			Description: desc.String("public_vip"),
			Computed:    true,
		},
		"subnet_name": dschema.StringAttribute{
			Description: desc.String("subnet_name"),
			Computed:    true,
		},
		"subnet_cidr_block": dschema.StringAttribute{
			Description: desc.String("subnet_cidr_block"),
			Computed:    true,
		},
		"vpc_id": dschema.StringAttribute{
			Description: desc.String("vpc_id"),
			Computed:    true,
		},
		"vpc_name": dschema.StringAttribute{
			Description: desc.String("vpc_name"),
			Computed:    true,
		},
		"subnet_id": dschema.StringAttribute{
			Description: desc.String("subnet_id"),
			Computed:    true,
		},
	}
}

// loadBalancerResourceSchemaAttributes defines the resource schema for the load balancer.
var loadBalancerResourceSchemaAttributes = getLoadBalancerResourceSchema()

// loadBalancerDataSourceSchemaAttributes defines the computed attributes for the load balancer data source.
var loadBalancerDataSourceSchemaAttributes = getLoadBalancerDataSourceSchema()

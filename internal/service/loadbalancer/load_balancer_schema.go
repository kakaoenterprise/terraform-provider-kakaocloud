// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package loadbalancer

import (
	"terraform-provider-kakaocloud/internal/common"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func getAccessLogsResourceSchema() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"bucket": rschema.StringAttribute{
			Required: true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
		"access_key": rschema.StringAttribute{
			Required:  true,
			Sensitive: true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
		"secret_key": rschema.StringAttribute{
			Required:  true,
			Sensitive: true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
	}
}

func getLoadBalancerResourceSchema() map[string]rschema.Attribute {
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
		"description": rschema.StringAttribute{
			Optional:   true,
			Validators: common.DescriptionValidator(),
		},
		"subnet_id": rschema.StringAttribute{
			Required:   true,
			Validators: common.UuidValidator(),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"availability_zone": rschema.StringAttribute{
			Required: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"flavor_id": rschema.StringAttribute{
			Required:   true,
			Validators: common.UuidValidator(),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},

		"type": rschema.StringAttribute{
			Computed: true,
		},
		"listener_ids": rschema.ListAttribute{
			ElementType: types.StringType,
			Computed:    true,
		},
		"project_id": rschema.StringAttribute{
			Computed: true,
		},
		"provisioning_status": rschema.StringAttribute{
			Computed: true,
		},
		"operating_status": rschema.StringAttribute{
			Computed: true,
		},
		"created_at": rschema.StringAttribute{
			Computed: true,
		},
		"updated_at": rschema.StringAttribute{
			Computed: true,
		},
		"access_logs": rschema.SingleNestedAttribute{
			Optional:   true,
			Attributes: getAccessLogsResourceSchema(),
		},
		"beyond_load_balancer_id": rschema.StringAttribute{
			Computed: true,
		},
		"beyond_load_balancer_name": rschema.StringAttribute{
			Computed: true,
		},
		"beyond_load_balancer_dns_name": rschema.StringAttribute{
			Computed: true,
		},
		"target_group_count": rschema.Int64Attribute{
			Computed: true,
		},
		"listener_count": rschema.Int64Attribute{
			Computed: true,
		},
		"private_vip": rschema.StringAttribute{
			Computed: true,
		},
		"public_vip": rschema.StringAttribute{
			Computed: true,
		},
		"subnet_name": rschema.StringAttribute{
			Computed: true,
		},
		"subnet_cidr_block": rschema.StringAttribute{
			Computed: true,
		},
		"vpc_id": rschema.StringAttribute{
			Computed: true,
		},
		"vpc_name": rschema.StringAttribute{
			Computed: true,
		},
	}
}

func getLoadBalancerDataSourceSchema() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"name": dschema.StringAttribute{
			Computed: true,
		},
		"description": dschema.StringAttribute{
			Computed: true,
		},
		"type": dschema.StringAttribute{
			Computed: true,
		},
		"listener_ids": dschema.ListAttribute{
			ElementType: types.StringType,
			Computed:    true,
		},
		"project_id": dschema.StringAttribute{
			Computed: true,
		},
		"provisioning_status": dschema.StringAttribute{
			Computed: true,
		},
		"operating_status": dschema.StringAttribute{
			Computed: true,
		},
		"created_at": dschema.StringAttribute{
			Computed: true,
		},
		"updated_at": dschema.StringAttribute{
			Computed: true,
		},
		"availability_zone": dschema.StringAttribute{
			Computed: true,
		},
		"access_logs": dschema.StringAttribute{
			Computed: true,
		},
		"beyond_load_balancer_id": dschema.StringAttribute{
			Computed: true,
		},
		"beyond_load_balancer_name": dschema.StringAttribute{
			Computed: true,
		},
		"beyond_load_balancer_dns_name": dschema.StringAttribute{
			Computed: true,
		},
		"target_group_count": dschema.Int64Attribute{
			Computed: true,
		},
		"listener_count": dschema.Int64Attribute{
			Computed: true,
		},
		"private_vip": dschema.StringAttribute{
			Computed: true,
		},
		"public_vip": dschema.StringAttribute{
			Computed: true,
		},
		"subnet_name": dschema.StringAttribute{
			Computed: true,
		},
		"subnet_cidr_block": dschema.StringAttribute{
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
	}
}

var loadBalancerResourceSchemaAttributes = getLoadBalancerResourceSchema()

var loadBalancerDataSourceSchemaAttributes = getLoadBalancerDataSourceSchema()

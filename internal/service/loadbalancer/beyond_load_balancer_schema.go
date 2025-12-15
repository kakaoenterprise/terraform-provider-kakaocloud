// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package loadbalancer

import (
	"terraform-provider-kakaocloud/internal/common"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
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
	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"name": rschema.StringAttribute{
			Required:   true,
			Validators: common.NameValidator(32),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"description": rschema.StringAttribute{
			Optional:   true,
			Validators: common.DescriptionValidator(),
		},
		"provider_name": rschema.StringAttribute{
			Computed: true,
		},
		"scheme": rschema.StringAttribute{
			Required: true,
			Validators: []validator.String{stringvalidator.OneOf(
				string(loadbalancer.SCHEME_INTERNET_FACING),
				string(loadbalancer.SCHEME_INTERNAL),
			)},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"project_id": rschema.StringAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"dns_name": rschema.StringAttribute{
			Computed: true,
		},
		"type_id": rschema.StringAttribute{
			Required:   true,
			Validators: common.UuidValidator(),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
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
		"provisioning_status": rschema.StringAttribute{
			Computed: true,
		},
		"operating_status": rschema.StringAttribute{
			Computed: true,
		},
		"vpc_id": rschema.StringAttribute{
			Required:   true,
			Validators: common.UuidValidator(),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"type": rschema.StringAttribute{
			Computed: true,
		},
		"vpc_name": rschema.StringAttribute{
			Computed: true,
		},
		"vpc_cidr_block": rschema.StringAttribute{
			Computed: true,
		},
		"availability_zones": rschema.ListAttribute{
			ElementType: types.StringType,
			Computed:    true,
		},
		"load_balancers": rschema.ListNestedAttribute{
			Computed: true,
			NestedObject: rschema.NestedAttributeObject{
				Attributes: getBeyondLoadBalancerLoadBalancerResourceSchemaAttributes(),
			},
		},
		"attached_load_balancers": rschema.SetNestedAttribute{
			Required: true,
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
			},
			NestedObject: rschema.NestedAttributeObject{
				Attributes: map[string]rschema.Attribute{
					"id": rschema.StringAttribute{
						Required:   true,
						Validators: common.UuidValidator(),
					},
					"availability_zone": rschema.StringAttribute{
						Required: true,
					},
				},
			},
		},
	}
}

func getBeyondLoadBalancerLoadBalancerResourceSchemaAttributes() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			Computed:   true,
			Validators: common.UuidValidator(),
		},
		"name": rschema.StringAttribute{
			Computed: true,
		},
		"description": rschema.StringAttribute{
			Computed: true,
		},
		"type": rschema.StringAttribute{
			Computed: true,
		},
		"provisioning_status": rschema.StringAttribute{
			Computed: true,
		},
		"operating_status": rschema.StringAttribute{
			Computed: true,
		},
		"availability_zone": rschema.StringAttribute{
			Computed: true,
		},
		"type_id": rschema.StringAttribute{
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
		"created_at": rschema.StringAttribute{
			Computed: true,
		},
		"updated_at": rschema.StringAttribute{
			Computed: true,
		},
	}
}

func getBeyondLoadBalancerDataSourceSchema() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"name": dschema.StringAttribute{
			Computed: true,
		},
		"description": dschema.StringAttribute{
			Computed: true,
		},
		"provider_name": dschema.StringAttribute{
			Computed: true,
		},
		"scheme": dschema.StringAttribute{
			Computed: true,
		},
		"project_id": dschema.StringAttribute{
			Computed: true,
		},
		"dns_name": dschema.StringAttribute{
			Computed: true,
		},
		"type_id": dschema.StringAttribute{
			Computed: true,
		},
		"created_at": dschema.StringAttribute{
			Computed: true,
		},
		"updated_at": dschema.StringAttribute{
			Computed: true,
		},
		"provisioning_status": dschema.StringAttribute{
			Computed: true,
		},
		"operating_status": dschema.StringAttribute{
			Computed: true,
		},
		"vpc_id": dschema.StringAttribute{
			Computed: true,
		},
		"type": dschema.StringAttribute{
			Computed: true,
		},
		"vpc_name": dschema.StringAttribute{
			Computed: true,
		},
		"vpc_cidr_block": dschema.StringAttribute{
			Computed: true,
		},
		"availability_zones": dschema.ListAttribute{
			ElementType: types.StringType,
			Computed:    true,
		},
		"load_balancers": dschema.ListNestedAttribute{
			Computed: true,
			NestedObject: dschema.NestedAttributeObject{
				Attributes: getBeyondLoadBalancerLoadBalancerDataSourceSchemaAttributes(),
			},
		},
	}
}

func getBeyondLoadBalancerLoadBalancerDataSourceSchemaAttributes() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"id":                  dschema.StringAttribute{Computed: true},
		"name":                dschema.StringAttribute{Computed: true},
		"description":         dschema.StringAttribute{Computed: true},
		"type":                dschema.StringAttribute{Computed: true},
		"provisioning_status": dschema.StringAttribute{Computed: true},
		"operating_status":    dschema.StringAttribute{Computed: true},
		"availability_zone":   dschema.StringAttribute{Computed: true},
		"type_id":             dschema.StringAttribute{Computed: true},
		"subnet_id":           dschema.StringAttribute{Computed: true},
		"subnet_name":         dschema.StringAttribute{Computed: true},
		"subnet_cidr_block":   dschema.StringAttribute{Computed: true},
		"created_at":          dschema.StringAttribute{Computed: true},
		"updated_at":          dschema.StringAttribute{Computed: true},
	}
}

var beyondLoadBalancerResourceSchema = getBeyondLoadBalancerResourceSchema()
var beyondLoadBalancerDatasourceSchema = getBeyondLoadBalancerDataSourceSchema()

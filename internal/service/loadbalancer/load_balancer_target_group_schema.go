// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package loadbalancer

import (
	"terraform-provider-kakaocloud/internal/common"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/kakaoenterprise/kc-sdk-go/services/loadbalancer"
)

var loadBalancerTargetGroupResourceSchema = map[string]rschema.Attribute{
	"id": rschema.StringAttribute{
		Computed: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	},
	"name": rschema.StringAttribute{
		Required:   true,
		Validators: common.NameValidator(255),
	},
	"description": rschema.StringAttribute{
		Optional:   true,
		Computed:   true,
		Validators: common.DescriptionValidator(),
	},
	"protocol": rschema.StringAttribute{
		Required: true,
		Validators: []validator.String{
			stringvalidator.OneOf(
				string(loadbalancer.TARGETGROUPPROTOCOL_HTTP),
				string(loadbalancer.TARGETGROUPPROTOCOL_HTTPS),
				string(loadbalancer.TARGETGROUPPROTOCOL_TCP),
				string(loadbalancer.TARGETGROUPPROTOCOL_UDP),
				string(loadbalancer.TARGETGROUPPROTOCOL_PROXY),
			),
		},
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	},
	"load_balancer_algorithm": rschema.StringAttribute{
		Required: true,
		Validators: []validator.String{
			stringvalidator.OneOf(
				string(loadbalancer.TARGETGROUPALGORITHM_ROUND_ROBIN),
				string(loadbalancer.TARGETGROUPALGORITHM_LEAST_CONNECTIONS),
				string(loadbalancer.TARGETGROUPALGORITHM_SOURCE_IP),
			),
		},
	},
	"listener_id": rschema.StringAttribute{
		Optional:   true,
		Validators: common.UuidValidator(),
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	},
	"load_balancer_id": rschema.StringAttribute{
		Required:   true,
		Validators: common.UuidValidator(),
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	},
	"subnet_id": rschema.StringAttribute{
		Computed: true,
	},
	"vpc_id": rschema.StringAttribute{
		Computed: true,
	},
	"availability_zone": rschema.StringAttribute{
		Computed: true,
	},
	"provisioning_status": rschema.StringAttribute{
		Computed: true,
	},
	"operating_status": rschema.StringAttribute{
		Computed: true,
	},
	"project_id": rschema.StringAttribute{
		Computed: true,
	},
	"created_at": rschema.StringAttribute{
		Computed: true,
	},
	"updated_at": rschema.StringAttribute{
		Computed: true,
	},
	"load_balancer_name": rschema.StringAttribute{
		Computed: true,
	},
	"load_balancer_provisioning_status": rschema.StringAttribute{
		Computed: true,
	},
	"load_balancer_type": rschema.StringAttribute{
		Computed: true,
	},
	"subnet_name": rschema.StringAttribute{
		Computed: true,
	},
	"vpc_name": rschema.StringAttribute{
		Computed: true,
	},
	"member_count": rschema.Int64Attribute{
		Computed: true,
	},
	"health_monitor": rschema.SingleNestedAttribute{
		Computed: true,
		Attributes: map[string]rschema.Attribute{
			"id": rschema.StringAttribute{
				Computed: true,
			},
			"type": rschema.StringAttribute{
				Computed: true,
			},
			"delay": rschema.Int64Attribute{
				Computed: true,
			},
			"timeout": rschema.Int64Attribute{
				Computed: true,
			},
			"fall_threshold": rschema.Int64Attribute{
				Computed: true,
			},
			"rise_threshold": rschema.Int64Attribute{
				Computed: true,
			},
			"http_method": rschema.StringAttribute{
				Computed: true,
			},
			"http_version": rschema.StringAttribute{
				Computed: true,
			},
			"expected_codes": rschema.StringAttribute{
				Computed: true,
			},
			"url_path": rschema.StringAttribute{
				Computed: true,
			},
			"operating_status": rschema.StringAttribute{
				Computed: true,
			},
			"provisioning_status": rschema.StringAttribute{
				Computed: true,
			},
			"project_id": rschema.StringAttribute{
				Computed: true,
			},
		},
	},
	"session_persistence": rschema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]rschema.Attribute{
			"type": rschema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						"APP_COOKIE",
						"HTTP_COOKIE",
						"SOURCE_IP",
					),
				},
			},
			"cookie_name": rschema.StringAttribute{
				Optional: true,
			},
			"persistence_timeout": rschema.Int64Attribute{
				Required: true,
				Validators: []validator.Int64{
					int64validator.Between(1, 604800),
				},
			},
			"persistence_granularity": rschema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					common.NewIPv4OrIPv6Validator(),
				},
			},
		},
	},
	"listeners": rschema.ListNestedAttribute{
		Computed: true,
		NestedObject: rschema.NestedAttributeObject{
			Attributes: map[string]rschema.Attribute{
				"id": rschema.StringAttribute{
					Computed: true,
				},
				"protocol": rschema.StringAttribute{
					Computed: true,
				},
				"protocol_port": rschema.Int64Attribute{
					Computed: true,
				},
			},
		},
	},
}

var loadBalancerTargetGroupDataSourceSchemaAttributes = map[string]dschema.Attribute{
	"name": dschema.StringAttribute{
		Computed: true,
	},
	"description": dschema.StringAttribute{
		Computed: true,
	},
	"protocol": dschema.StringAttribute{
		Computed: true,
	},
	"load_balancer_algorithm": dschema.StringAttribute{
		Computed: true,
	},
	"subnet_id": dschema.StringAttribute{
		Computed: true,
	},
	"vpc_id": dschema.StringAttribute{
		Computed: true,
	},
	"availability_zone": dschema.StringAttribute{
		Computed: true,
	},
	"provisioning_status": dschema.StringAttribute{
		Computed: true,
	},
	"operating_status": dschema.StringAttribute{
		Computed: true,
	},
	"project_id": dschema.StringAttribute{
		Computed: true,
	},
	"created_at": dschema.StringAttribute{
		Computed: true,
	},
	"updated_at": dschema.StringAttribute{
		Computed: true,
	},
	"load_balancer_id": dschema.StringAttribute{
		Computed: true,
	},
	"load_balancer_name": dschema.StringAttribute{
		Computed: true,
	},
	"load_balancer_provisioning_status": dschema.StringAttribute{
		Computed: true,
	},
	"load_balancer_type": dschema.StringAttribute{
		Computed: true,
	},
	"subnet_name": dschema.StringAttribute{
		Computed: true,
	},
	"vpc_name": dschema.StringAttribute{
		Computed: true,
	},
	"member_count": dschema.Int64Attribute{
		Computed: true,
	},
	"listeners": dschema.ListNestedAttribute{
		Computed: true,
		NestedObject: dschema.NestedAttributeObject{
			Attributes: map[string]dschema.Attribute{
				"id": dschema.StringAttribute{
					Computed: true,
				},
				"protocol": dschema.StringAttribute{
					Computed: true,
				},
				"protocol_port": dschema.Int64Attribute{
					Computed: true,
				},
			},
		},
	},
	"health_monitor": dschema.SingleNestedAttribute{
		Computed: true,
		Attributes: map[string]dschema.Attribute{
			"id": dschema.StringAttribute{
				Computed: true,
			},
			"type": dschema.StringAttribute{
				Computed: true,
			},
			"delay": dschema.Int64Attribute{
				Computed: true,
			},
			"timeout": dschema.Int64Attribute{
				Computed: true,
			},
			"fall_threshold": dschema.Int64Attribute{
				Computed: true,
			},
			"rise_threshold": dschema.Int64Attribute{
				Computed: true,
			},
			"http_method": dschema.StringAttribute{
				Computed: true,
			},
			"http_version": dschema.StringAttribute{
				Computed: true,
			},
			"expected_codes": dschema.StringAttribute{
				Computed: true,
			},
			"url_path": dschema.StringAttribute{
				Computed: true,
			},
			"operating_status": dschema.StringAttribute{
				Computed: true,
			},
			"project_id": dschema.StringAttribute{
				Computed: true,
			},
			"provisioning_status": dschema.StringAttribute{
				Computed: true,
			},
		},
	},
	"session_persistence": dschema.SingleNestedAttribute{
		Computed: true,
		Attributes: map[string]dschema.Attribute{
			"type": dschema.StringAttribute{
				Computed: true,
			},
			"cookie_name": dschema.StringAttribute{
				Computed: true,
			},
			"persistence_timeout": dschema.Int64Attribute{
				Computed: true,
			},
			"persistence_granularity": dschema.StringAttribute{
				Computed: true,
			},
		},
	},
}

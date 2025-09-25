// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package loadbalancer

import (
	"regexp"
	"terraform-provider-kakaocloud/internal/common"
	"terraform-provider-kakaocloud/internal/docs"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// Common description variables
var (
	targetGroupDesc        = docs.Loadbalancer("bns_load_balancer__v1__api__create_target_group__model__TargetGroupModel")
	listTargetGroupDesc    = docs.Loadbalancer("bns_load_balancer__v1__api__list_target_groups__model__TargetGroupModel")
	healthMonitorDesc      = docs.Loadbalancer("bns_load_balancer__v1__api__get_target_group__model__HealthMonitorModel")
	sessionPersistenceDesc = docs.Loadbalancer("bns_load_balancer__v1__api__create_target_group__model__SessionPersistenceModel")
)

// Resource schema attributes
var loadBalancerTargetGroupResourceSchema = map[string]rschema.Attribute{
	"id": rschema.StringAttribute{
		Computed:    true,
		Description: targetGroupDesc.String("id"),
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	},
	"name": rschema.StringAttribute{
		Required:    true,
		Description: targetGroupDesc.String("name"),
		Validators:  common.NameValidator(255),
	},
	"description": rschema.StringAttribute{
		Optional:    true,
		Description: targetGroupDesc.String("description"),
		Validators:  common.DescriptionValidator(),
	},
	"protocol": rschema.StringAttribute{
		Required:    true,
		Description: targetGroupDesc.String("protocol"),
		Validators: []validator.String{
			stringvalidator.OneOf(
				"HTTP",
				"HTTPS",
				"TCP",
				"UDP",
				"PROXY",
			),
		},
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	},
	"load_balancer_algorithm": rschema.StringAttribute{
		Required:    true,
		Description: targetGroupDesc.String("load_balancer_algorithm"),
		Validators: []validator.String{
			stringvalidator.OneOf(
				"ROUND_ROBIN",
				"LEAST_CONNECTIONS",
				"SOURCE_IP",
			),
		},
	},
	"listener_id": rschema.StringAttribute{
		Optional:    true,
		Computed:    true,
		Description: "ID of the listener associated with this target group",
		Validators:  common.UuidValidator(),
	},
	"load_balancer_id": rschema.StringAttribute{
		Required:    true,
		Description: "Load balancer ID for the target group",
		Validators:  common.UuidValidator(),
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	},
	"subnet_id": rschema.StringAttribute{
		Computed:    true,
		Description: listTargetGroupDesc.String("subnet_id"),
	},
	"vpc_id": rschema.StringAttribute{
		Computed:    true,
		Description: listTargetGroupDesc.String("vpc_id"),
	},
	"availability_zone": rschema.StringAttribute{
		Computed:    true,
		Description: listTargetGroupDesc.String("availability_zone"),
	},
	"provisioning_status": rschema.StringAttribute{
		Computed:    true,
		Description: targetGroupDesc.String("provisioning_status"),
	},
	"operating_status": rschema.StringAttribute{
		Computed:    true,
		Description: targetGroupDesc.String("operating_status"),
	},
	"project_id": rschema.StringAttribute{
		Computed:    true,
		Description: targetGroupDesc.String("project_id"),
	},
	"created_at": rschema.StringAttribute{
		Computed:    true,
		Description: targetGroupDesc.String("created_at"),
	},
	"updated_at": rschema.StringAttribute{
		Computed:    true,
		Description: targetGroupDesc.String("updated_at"),
	},
	"load_balancer_name": rschema.StringAttribute{
		Computed:    true,
		Description: listTargetGroupDesc.String("load_balancer_name"),
	},
	"load_balancer_provisioning_status": rschema.StringAttribute{
		Computed:    true,
		Description: listTargetGroupDesc.String("load_balancer_provisioning_status"),
	},
	"load_balancer_type": rschema.StringAttribute{
		Computed:    true,
		Description: listTargetGroupDesc.String("load_balancer_type"),
	},
	"subnet_name": rschema.StringAttribute{
		Computed:    true,
		Description: listTargetGroupDesc.String("subnet_name"),
	},
	"vpc_name": rschema.StringAttribute{
		Computed:    true,
		Description: listTargetGroupDesc.String("vpc_name"),
	},
	"member_count": rschema.Int64Attribute{
		Computed:    true,
		Description: listTargetGroupDesc.String("member_count"),
	},
	"health_monitor": rschema.SingleNestedAttribute{
		Optional:    true,
		Computed:    true,
		Description: listTargetGroupDesc.String("health_monitor"),
		Attributes: map[string]rschema.Attribute{
			"id": rschema.StringAttribute{
				Computed:    true,
				Description: healthMonitorDesc.String("id"),
			},
			"type": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: healthMonitorDesc.String("type"),
				Validators: []validator.String{
					stringvalidator.OneOf(
						"HTTP",
						"HTTPS",
						"TCP",
						"PING",
					),
				},
			},
			"delay": rschema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: healthMonitorDesc.String("delay"),
				Validators: []validator.Int64{
					int64validator.Between(0, 3600),
				},
			},
			"timeout": rschema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: healthMonitorDesc.String("timeout"),
				Validators: []validator.Int64{
					int64validator.Between(0, 900),
				},
			},
			"fall_threshold": rschema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: healthMonitorDesc.String("fall_threshold"),
				Validators: []validator.Int64{
					int64validator.Between(1, 10),
				},
			},
			"rise_threshold": rschema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: healthMonitorDesc.String("rise_threshold"),
				Validators: []validator.Int64{
					int64validator.Between(1, 10),
				},
			},
			"http_method": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: healthMonitorDesc.String("http_method"),
				Validators: []validator.String{
					stringvalidator.OneOf(
						"CONNECT",
						"GET",
					),
				},
			},
			"http_version": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: healthMonitorDesc.String("http_version"),
				Validators: []validator.String{
					stringvalidator.OneOf(
						"1.0",
						"1.1",
					),
				},
			},
			"expected_codes": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: healthMonitorDesc.String("expected_codes"),
			},
			"url_path": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: healthMonitorDesc.String("url_path"),
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^/(.{0,120})$`),
						"Must start with a forward slash (/) and be at most 120 characters long excluding the slash",
					),
				},
			},
			"operating_status": rschema.StringAttribute{
				Computed:    true,
				Description: healthMonitorDesc.String("operating_status"),
			},
			"provisioning_status": rschema.StringAttribute{
				Computed:    true,
				Description: healthMonitorDesc.String("provisioning_status"),
			},
			"project_id": rschema.StringAttribute{
				Computed:    true,
				Description: healthMonitorDesc.String("project_id"),
			},
		},
	},
	"session_persistence": rschema.SingleNestedAttribute{
		Optional:    true,
		Computed:    true,
		Description: listTargetGroupDesc.String("session_persistence"),
		Attributes: map[string]rschema.Attribute{
			"type": rschema.StringAttribute{
				Required:    true,
				Description: sessionPersistenceDesc.String("type"),
				Validators: []validator.String{
					stringvalidator.OneOf(
						"APP_COOKIE",
						"HTTP_COOKIE",
						"SOURCE_IP",
					),
				},
			},
			"cookie_name": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: sessionPersistenceDesc.String("cookie_name"),
			},
			"persistence_timeout": rschema.Int64Attribute{
				Required:    true,
				Description: sessionPersistenceDesc.String("persistence_timeout"),
				Validators: []validator.Int64{
					int64validator.Between(1, 604800),
				},
			},
			"persistence_granularity": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: sessionPersistenceDesc.String("persistence_granularity"),
				Validators: []validator.String{
					common.NewIPv4OrIPv6Validator(),
				},
			},
		},
	},
	"listeners": rschema.ListNestedAttribute{
		Computed:    true,
		Description: listTargetGroupDesc.String("listeners"),
		NestedObject: rschema.NestedAttributeObject{
			Attributes: map[string]rschema.Attribute{
				"id": rschema.StringAttribute{
					Computed:    true,
					Description: "Listener ID",
				},
			},
		},
	},
	"load_balancers": rschema.ListNestedAttribute{
		Computed:    true,
		Description: targetGroupDesc.String("load_balancers"),
		NestedObject: rschema.NestedAttributeObject{
			Attributes: map[string]rschema.Attribute{
				"id": rschema.StringAttribute{
					Computed:    true,
					Description: "Load balancer ID",
				},
			},
		},
	},
}

// Data source schema attributes (single target group)
var loadBalancerTargetGroupDataSourceSchemaAttributes = map[string]dschema.Attribute{
	"id": dschema.StringAttribute{
		Required:    true,
		Description: targetGroupDesc.String("id"),
	},
	"name": dschema.StringAttribute{
		Computed:    true,
		Description: targetGroupDesc.String("name"),
	},
	"description": dschema.StringAttribute{
		Computed:    true,
		Description: targetGroupDesc.String("description"),
	},
	"protocol": dschema.StringAttribute{
		Computed:    true,
		Description: targetGroupDesc.String("protocol"),
	},
	"load_balancer_algorithm": dschema.StringAttribute{
		Computed:    true,
		Description: targetGroupDesc.String("load_balancer_algorithm"),
	},
	"subnet_id": dschema.StringAttribute{
		Computed:    true,
		Description: listTargetGroupDesc.String("subnet_id"),
	},
	"vpc_id": dschema.StringAttribute{
		Computed:    true,
		Description: listTargetGroupDesc.String("vpc_id"),
	},
	"availability_zone": dschema.StringAttribute{
		Computed:    true,
		Description: listTargetGroupDesc.String("availability_zone"),
	},
	"provisioning_status": dschema.StringAttribute{
		Computed:    true,
		Description: targetGroupDesc.String("provisioning_status"),
	},
	"operating_status": dschema.StringAttribute{
		Computed:    true,
		Description: targetGroupDesc.String("operating_status"),
	},
	"project_id": dschema.StringAttribute{
		Computed:    true,
		Description: targetGroupDesc.String("project_id"),
	},
	"created_at": dschema.StringAttribute{
		Computed:    true,
		Description: targetGroupDesc.String("created_at"),
	},
	"updated_at": dschema.StringAttribute{
		Computed:    true,
		Description: targetGroupDesc.String("updated_at"),
	},
	"load_balancer_id": dschema.StringAttribute{
		Computed:    true,
		Description: listTargetGroupDesc.String("load_balancer_id"),
	},
	"load_balancer_name": dschema.StringAttribute{
		Computed:    true,
		Description: listTargetGroupDesc.String("load_balancer_name"),
	},
	"load_balancer_provisioning_status": dschema.StringAttribute{
		Computed:    true,
		Description: listTargetGroupDesc.String("load_balancer_provisioning_status"),
	},
	"load_balancer_type": dschema.StringAttribute{
		Computed:    true,
		Description: listTargetGroupDesc.String("load_balancer_type"),
	},
	"subnet_name": dschema.StringAttribute{
		Computed:    true,
		Description: listTargetGroupDesc.String("subnet_name"),
	},
	"vpc_name": dschema.StringAttribute{
		Computed:    true,
		Description: listTargetGroupDesc.String("vpc_name"),
	},
	"member_count": dschema.Int64Attribute{
		Computed:    true,
		Description: listTargetGroupDesc.String("member_count"),
	},
	"listeners": dschema.ListNestedAttribute{
		Computed:    true,
		Description: listTargetGroupDesc.String("listeners"),
		NestedObject: dschema.NestedAttributeObject{
			Attributes: map[string]dschema.Attribute{
				"id": dschema.StringAttribute{
					Computed:    true,
					Description: "Listener ID",
				},
				"protocol": dschema.StringAttribute{
					Computed:    true,
					Description: "Listener protocol",
				},
				"protocol_port": dschema.Int64Attribute{
					Computed:    true,
					Description: "Listener protocol port",
				},
			},
		},
	},
	"health_monitor": dschema.SingleNestedAttribute{
		Computed:    true,
		Description: listTargetGroupDesc.String("health_monitor"),
		Attributes: map[string]dschema.Attribute{
			"id": dschema.StringAttribute{
				Computed:    true,
				Description: healthMonitorDesc.String("id"),
			},
			"type": dschema.StringAttribute{
				Computed:    true,
				Description: healthMonitorDesc.String("type"),
			},
			"delay": dschema.Int64Attribute{
				Computed:    true,
				Description: healthMonitorDesc.String("delay"),
			},
			"timeout": dschema.Int64Attribute{
				Computed:    true,
				Description: healthMonitorDesc.String("timeout"),
			},
			"fall_threshold": dschema.Int64Attribute{
				Computed:    true,
				Description: healthMonitorDesc.String("fall_threshold"),
			},
			"rise_threshold": dschema.Int64Attribute{
				Computed:    true,
				Description: healthMonitorDesc.String("rise_threshold"),
			},
			"http_method": dschema.StringAttribute{
				Computed:    true,
				Description: healthMonitorDesc.String("http_method"),
			},
			"http_version": dschema.StringAttribute{
				Computed:    true,
				Description: healthMonitorDesc.String("http_version"),
			},
			"expected_codes": dschema.StringAttribute{
				Computed:    true,
				Description: healthMonitorDesc.String("expected_codes"),
			},
			"url_path": dschema.StringAttribute{
				Computed:    true,
				Description: healthMonitorDesc.String("url_path"),
			},
			"operating_status": dschema.StringAttribute{
				Computed:    true,
				Description: healthMonitorDesc.String("operating_status"),
			},
			"project_id": dschema.StringAttribute{
				Computed:    true,
				Description: healthMonitorDesc.String("project_id"),
			},
			"provisioning_status": dschema.StringAttribute{
				Computed:    true,
				Description: healthMonitorDesc.String("provisioning_status"),
			},
		},
	},
	"session_persistence": dschema.SingleNestedAttribute{
		Computed:    true,
		Description: listTargetGroupDesc.String("session_persistence"),
		Attributes: map[string]dschema.Attribute{
			"type": dschema.StringAttribute{
				Computed:    true,
				Description: sessionPersistenceDesc.String("type"),
			},
			"cookie_name": dschema.StringAttribute{
				Computed:    true,
				Description: sessionPersistenceDesc.String("cookie_name"),
			},
			"persistence_timeout": dschema.Int64Attribute{
				Computed:    true,
				Description: sessionPersistenceDesc.String("persistence_timeout"),
			},
			"persistence_granularity": dschema.StringAttribute{
				Computed:    true,
				Description: sessionPersistenceDesc.String("persistence_granularity"),
			},
		},
	},
}

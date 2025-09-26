// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package loadbalancer

import (
	"terraform-provider-kakaocloud/internal/common"
	"terraform-provider-kakaocloud/internal/docs"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func getL7PolicyResourceSchema() map[string]rschema.Attribute {
	desc := docs.Loadbalancer("bns_load_balancer__v1__api__create_l7_policy__model__L7PolicyModel")
	getDesc := docs.Loadbalancer("bns_load_balancer__v1__api__get_l7_policy__model__l7PolicyModel")

	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("id"),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"name": rschema.StringAttribute{
			Optional:    true,
			Description: desc.String("name"),
			Validators:  common.NameValidator(255),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"description": rschema.StringAttribute{
			Optional:    true,
			Description: desc.String("description"),
			Validators:  common.DescriptionValidator(),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"listener_id": rschema.StringAttribute{
			Required:    true,
			Description: desc.String("listener_id"),
			Validators:  common.UuidValidator(),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"action": rschema.StringAttribute{
			Required:    true,
			Description: desc.String("action"),
			Validators: []validator.String{
				stringvalidator.OneOf("REDIRECT_PREFIX", "REDIRECT_TO_POOL", "REDIRECT_TO_URL"),
			},
		},
		"position": rschema.Int64Attribute{
			Optional:    true,
			Computed:    true,
			Description: desc.String("position"),
			Validators: []validator.Int64{
				int64validator.Between(1, 1000),
			},
		},
		"redirect_target_group_id": rschema.StringAttribute{
			Optional:    true,
			Description: desc.String("redirect_target_group_id"),
			Validators:  common.UuidValidator(),
		},
		"redirect_url": rschema.StringAttribute{
			Optional:    true,
			Description: desc.String("redirect_url"),
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
		"redirect_prefix": rschema.StringAttribute{
			Optional:    true,
			Description: desc.String("redirect_prefix"),
		},
		"redirect_http_code": rschema.Int64Attribute{
			Computed:    true,
			Description: getDesc.String("redirect_http_code"),
		},
		"provisioning_status": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("provisioning_status"),
		},
		"operating_status": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("operating_status"),
		},
		"project_id": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("project_id"),
		},
		"rules": rschema.ListNestedAttribute{
			Computed:    true,
			Description: desc.String("rules"),
			PlanModifiers: []planmodifier.List{
				listplanmodifier.UseStateForUnknown(),
			},
			NestedObject: rschema.NestedAttributeObject{
				Attributes: getL7PolicyRuleSchemaAttributes(),
			},
		},
	}
}

func getL7PolicyRuleSchemaAttributes() map[string]rschema.Attribute {
	ruleDesc := docs.Loadbalancer("bns_load_balancer__v1__api__create_l7_policy__model__RuleModel")

	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			Computed:    true,
			Description: ruleDesc.String("id"),
		},
		"type": rschema.StringAttribute{
			Required:    true,
			Description: ruleDesc.String("type"),
			Validators: []validator.String{
				stringvalidator.OneOf("PATH", "HEADER", "HOST_NAME", "FILE_TYPE", "COOKIE"),
			},
		},
		"compare_type": rschema.StringAttribute{
			Required:    true,
			Description: ruleDesc.String("compare_type"),
			Validators: []validator.String{
				stringvalidator.OneOf("EQUAL_TO", "STARTS_WITH", "ENDS_WITH", "CONTAINS"),
			},
		},
		"key": rschema.StringAttribute{
			Optional:    true,
			Description: ruleDesc.String("key"),
		},
		"value": rschema.StringAttribute{
			Required:    true,
			Description: ruleDesc.String("value"),
		},
		"is_inverted": rschema.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Description: ruleDesc.String("is_inverted"),
		},
		"provisioning_status": rschema.StringAttribute{
			Computed:    true,
			Description: ruleDesc.String("provisioning_status"),
		},
		"operating_status": rschema.StringAttribute{
			Computed:    true,
			Description: ruleDesc.String("operating_status"),
		},
		"project_id": rschema.StringAttribute{
			Computed:    true,
			Description: ruleDesc.String("project_id"),
		},
	}
}

func getL7PolicyDataSourceSchema() map[string]dschema.Attribute {
	desc := docs.Loadbalancer("bns_load_balancer__v1__api__create_l7_policy__model__L7PolicyModel")
	getDesc := docs.Loadbalancer("bns_load_balancer__v1__api__get_l7_policy__model__l7PolicyModel")

	return map[string]dschema.Attribute{
		"id": dschema.StringAttribute{
			Required:    true,
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
		"listener_id": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("listener_id"),
		},
		"action": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("action"),
		},
		"position": dschema.Int64Attribute{
			Computed:    true,
			Description: desc.String("position"),
		},
		"redirect_target_group_id": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("redirect_target_group_id"),
		},
		"redirect_url": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("redirect_url"),
		},
		"redirect_prefix": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("redirect_prefix"),
		},
		"redirect_http_code": dschema.Int64Attribute{
			Computed:    true,
			Description: getDesc.String("redirect_http_code"),
		},
		"provisioning_status": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("provisioning_status"),
		},
		"operating_status": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("operating_status"),
		},
		"project_id": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("project_id"),
		},
		"rules": dschema.ListNestedAttribute{
			Computed:    true,
			Description: desc.String("rules"),
			NestedObject: dschema.NestedAttributeObject{
				Attributes: getL7PolicyRuleDataSourceSchemaAttributes(),
			},
		},
	}
}

func getL7PolicyRuleDataSourceSchemaAttributes() map[string]dschema.Attribute {
	ruleDesc := docs.Loadbalancer("bns_load_balancer__v1__api__create_l7_policy__model__RuleModel")

	return map[string]dschema.Attribute{
		"id": dschema.StringAttribute{
			Computed:    true,
			Description: ruleDesc.String("id"),
		},
		"type": dschema.StringAttribute{
			Computed:    true,
			Description: ruleDesc.String("type"),
		},
		"compare_type": dschema.StringAttribute{
			Computed:    true,
			Description: ruleDesc.String("compare_type"),
		},
		"key": dschema.StringAttribute{
			Computed:    true,
			Description: ruleDesc.String("key"),
		},
		"value": dschema.StringAttribute{
			Computed:    true,
			Description: ruleDesc.String("value"),
		},
		"is_inverted": dschema.BoolAttribute{
			Computed:    true,
			Description: ruleDesc.String("is_inverted"),
		},
		"provisioning_status": dschema.StringAttribute{
			Computed:    true,
			Description: ruleDesc.String("provisioning_status"),
		},
		"operating_status": dschema.StringAttribute{
			Computed:    true,
			Description: ruleDesc.String("operating_status"),
		},
		"project_id": dschema.StringAttribute{
			Computed:    true,
			Description: ruleDesc.String("project_id"),
		},
	}
}

var loadBalancerL7PolicyResourceSchemaAttributes = getL7PolicyResourceSchema()

var loadBalancerL7PolicyDataSourceSchemaAttributes = getL7PolicyDataSourceSchema()

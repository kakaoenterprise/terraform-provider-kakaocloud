// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package loadbalancer

import (
	"regexp"
	"terraform-provider-kakaocloud/internal/common"
	"terraform-provider-kakaocloud/internal/docs"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/kakaoenterprise/kc-sdk-go/services/loadbalancer"
)

func getL7PolicyRuleDescriptions() map[string]string {
	desc := docs.Loadbalancer("bns_load_balancer__v1__api__add_l7_policy_rule__model__L7PolicyRuleModel")

	return map[string]string{
		"id":                  desc.String("id"),
		"l7_policy_id":        docs.ParameterDescription("loadbalancer", "add_l7_policy_rule", "path_l7_policy_id"),
		"type":                desc.String("type"),
		"compare_type":        desc.String("compare_type"),
		"key":                 desc.String("key"),
		"value":               desc.String("value"),
		"is_inverted":         desc.String("is_inverted"),
		"provisioning_status": desc.String("provisioning_status"),
		"operating_status":    desc.String("operating_status"),
		"project_id":          desc.String("project_id"),
	}
}

func getL7PolicyRuleResourceSchema() map[string]rschema.Attribute {
	descriptions := getL7PolicyRuleDescriptions()

	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			Computed:            true,
			Description:         descriptions["id"],
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			MarkdownDescription: descriptions["id"],
		},
		"l7_policy_id": rschema.StringAttribute{
			Required:            true,
			Description:         descriptions["l7_policy_id"],
			MarkdownDescription: descriptions["l7_policy_id"],
			Validators:          common.UuidValidator(),
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
		},
		"type": rschema.StringAttribute{
			Required:            true,
			Description:         descriptions["type"],
			MarkdownDescription: descriptions["type"],
			Validators: []validator.String{
				stringvalidator.OneOf(
					string(loadbalancer.L7RULETYPE_COOKIE),
					string(loadbalancer.L7RULETYPE_FILE_TYPE),
					string(loadbalancer.L7RULETYPE_HEADER),
					string(loadbalancer.L7RULETYPE_HOST_NAME),
					string(loadbalancer.L7RULETYPE_PATH),
				),
			},
		},
		"compare_type": rschema.StringAttribute{
			Required:            true,
			Description:         descriptions["compare_type"],
			MarkdownDescription: descriptions["compare_type"],
			Validators: []validator.String{
				stringvalidator.OneOf(
					string(loadbalancer.L7RULECOMPARETYPE_CONTAINS),
					string(loadbalancer.L7RULECOMPARETYPE_ENDS_WITH),
					string(loadbalancer.L7RULECOMPARETYPE_EQUAL_TO),
					string(loadbalancer.L7RULECOMPARETYPE_STARTS_WITH),
				),
			},
		},
		"key": rschema.StringAttribute{
			Optional:            true,
			Description:         descriptions["key"],
			MarkdownDescription: descriptions["key"],
			Validators: []validator.String{
				stringvalidator.LengthBetween(1, 255),
				stringvalidator.RegexMatches(
					regexp.MustCompile(`^[^\s]+$`),
					"Key cannot contain spaces",
				),
			},
		},
		"value": rschema.StringAttribute{
			Required:            true,
			Description:         descriptions["value"],
			MarkdownDescription: descriptions["value"],
			Validators: []validator.String{
				stringvalidator.LengthBetween(1, 255),
			},
		},
		"is_inverted": rschema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			Description:         descriptions["is_inverted"],
			MarkdownDescription: descriptions["is_inverted"],
			PlanModifiers: []planmodifier.Bool{
				boolplanmodifier.UseStateForUnknown(),
			},
		},
		"provisioning_status": rschema.StringAttribute{
			Computed:            true,
			Description:         descriptions["provisioning_status"],
			MarkdownDescription: descriptions["provisioning_status"],
		},
		"operating_status": rschema.StringAttribute{
			Computed:            true,
			Description:         descriptions["operating_status"],
			MarkdownDescription: descriptions["operating_status"],
		},
		"project_id": rschema.StringAttribute{
			Computed:            true,
			Description:         descriptions["project_id"],
			MarkdownDescription: descriptions["project_id"],
		},
	}
}

func getL7PolicyRuleDataSourceSchema() map[string]dschema.Attribute {
	descriptions := getL7PolicyRuleDescriptions()

	return map[string]dschema.Attribute{
		"id": dschema.StringAttribute{
			Required:    true,
			Description: descriptions["id"],
			Validators:  common.UuidValidator(),
		},
		"l7_policy_id": dschema.StringAttribute{
			Required:    true,
			Description: descriptions["l7_policy_id"],
			Validators:  common.UuidValidator(),
		},
		"type": dschema.StringAttribute{
			Computed:    true,
			Description: descriptions["type"],
		},
		"compare_type": dschema.StringAttribute{
			Computed:    true,
			Description: descriptions["compare_type"],
		},
		"key": dschema.StringAttribute{
			Computed:    true,
			Description: descriptions["key"],
		},
		"value": dschema.StringAttribute{
			Computed:    true,
			Description: descriptions["value"],
		},
		"is_inverted": dschema.BoolAttribute{
			Computed:    true,
			Description: descriptions["is_inverted"],
		},
		"provisioning_status": dschema.StringAttribute{
			Computed:    true,
			Description: descriptions["provisioning_status"],
		},
		"operating_status": dschema.StringAttribute{
			Computed:    true,
			Description: descriptions["operating_status"],
		},
		"project_id": dschema.StringAttribute{
			Computed:    true,
			Description: descriptions["project_id"],
		},
	}
}

func getL7PolicyRuleListDataSourceSchema() map[string]dschema.Attribute {
	descriptions := getL7PolicyRuleDescriptions()

	return map[string]dschema.Attribute{
		"id":                  dschema.StringAttribute{Computed: true, Description: descriptions["id"]},
		"l7_policy_id":        dschema.StringAttribute{Computed: true, Description: descriptions["l7_policy_id"]},
		"type":                dschema.StringAttribute{Computed: true, Description: descriptions["type"]},
		"compare_type":        dschema.StringAttribute{Computed: true, Description: descriptions["compare_type"]},
		"key":                 dschema.StringAttribute{Computed: true, Description: descriptions["key"]},
		"value":               dschema.StringAttribute{Computed: true, Description: descriptions["value"]},
		"is_inverted":         dschema.BoolAttribute{Computed: true, Description: descriptions["is_inverted"]},
		"provisioning_status": dschema.StringAttribute{Computed: true, Description: descriptions["provisioning_status"]},
		"operating_status":    dschema.StringAttribute{Computed: true, Description: descriptions["operating_status"]},
		"project_id":          dschema.StringAttribute{Computed: true, Description: descriptions["project_id"]},
	}
}

// loadBalancerL7PolicyRuleResourceSchema defines the schema for the L7 policy rule resource
var loadBalancerL7PolicyRuleResourceSchema = getL7PolicyRuleResourceSchema()

// loadBalancerL7PolicyRuleDataSourceSchema defines the schema for the L7 policy rule data source
var loadBalancerL7PolicyRuleDataSourceSchema = getL7PolicyRuleDataSourceSchema()

// loadBalancerL7PolicyRuleListDataSourceSchema defines the schema for the L7 policy rule list data source
var loadBalancerL7PolicyRuleListDataSourceSchema = getL7PolicyRuleListDataSourceSchema()

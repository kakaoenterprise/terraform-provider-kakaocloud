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
	"github.com/kakaoenterprise/kc-sdk-go/services/loadbalancer"
)

func getL7PolicyRuleResourceSchema() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			Computed:      true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"l7_policy_id": rschema.StringAttribute{
			Required:      true,
			Validators:    common.UuidValidator(),
			PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
		},
		"type": rschema.StringAttribute{
			Required: true,
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
			Required: true,
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
			Optional: true,
			Validators: []validator.String{
				stringvalidator.LengthBetween(1, 255),
			},
		},
		"value": rschema.StringAttribute{
			Required: true,
			Validators: []validator.String{
				stringvalidator.LengthBetween(1, 255),
			},
		},
		"is_inverted": rschema.BoolAttribute{
			Optional: true,
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
	}
}

func getL7PolicyRuleDataSourceSchema() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"id": dschema.StringAttribute{
			Required:   true,
			Validators: common.UuidValidator(),
		},
		"l7_policy_id": dschema.StringAttribute{
			Required:   true,
			Validators: common.UuidValidator(),
		},
		"type": dschema.StringAttribute{
			Computed: true,
		},
		"compare_type": dschema.StringAttribute{
			Computed: true,
		},
		"key": dschema.StringAttribute{
			Computed: true,
		},
		"value": dschema.StringAttribute{
			Computed: true,
		},
		"is_inverted": dschema.BoolAttribute{
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
	}
}

func getL7PolicyRuleListDataSourceSchema() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"id":                  dschema.StringAttribute{Computed: true},
		"l7_policy_id":        dschema.StringAttribute{Computed: true},
		"type":                dschema.StringAttribute{Computed: true},
		"compare_type":        dschema.StringAttribute{Computed: true},
		"key":                 dschema.StringAttribute{Computed: true},
		"value":               dschema.StringAttribute{Computed: true},
		"is_inverted":         dschema.BoolAttribute{Computed: true},
		"provisioning_status": dschema.StringAttribute{Computed: true},
		"operating_status":    dschema.StringAttribute{Computed: true},
		"project_id":          dschema.StringAttribute{Computed: true},
	}
}

var loadBalancerL7PolicyRuleResourceSchema = getL7PolicyRuleResourceSchema()

var loadBalancerL7PolicyRuleDataSourceSchema = getL7PolicyRuleDataSourceSchema()

var loadBalancerL7PolicyRuleListDataSourceSchema = getL7PolicyRuleListDataSourceSchema()

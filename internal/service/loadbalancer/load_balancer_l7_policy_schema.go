// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package loadbalancer

import (
	"terraform-provider-kakaocloud/internal/common"

	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/kakaoenterprise/kc-sdk-go/services/loadbalancer"
)

func getL7PolicyResourceSchema() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"name": rschema.StringAttribute{
			Optional:   true,
			Validators: common.NameValidator(255),
		},
		"description": rschema.StringAttribute{
			Optional:   true,
			Validators: common.DescriptionValidator(),
		},
		"listener_id": rschema.StringAttribute{
			Required:   true,
			Validators: common.UuidValidator(),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"action": rschema.StringAttribute{
			Required: true,
			Validators: []validator.String{
				stringvalidator.OneOf(
					string(loadbalancer.L7POLICYACTION_REDIRECT_PREFIX),
					string(loadbalancer.L7POLICYACTION_REDIRECT_TO_POOL),
					string(loadbalancer.L7POLICYACTION_REDIRECT_TO_URL),
				),
			},
		},
		"position": rschema.Int32Attribute{
			Optional: true,
			Computed: true,
			Validators: []validator.Int32{
				int32validator.AtLeast(1),
			},
		},
		"redirect_target_group_id": rschema.StringAttribute{
			Optional:   true,
			Validators: common.UuidValidator(),
		},
		"redirect_url": rschema.StringAttribute{
			Optional:   true,
			Validators: common.UrlValidator(),
		},
		"redirect_prefix": rschema.StringAttribute{
			Optional:   true,
			Validators: common.UrlValidator(),
		},
		"redirect_http_code": rschema.Int32Attribute{
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
		"rules": rschema.ListNestedAttribute{
			Computed: true,
			NestedObject: rschema.NestedAttributeObject{
				Attributes: getL7PolicyRuleSchemaAttributes(),
			},
		},
	}
}

func getL7PolicyRuleSchemaAttributes() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			Computed: true,
		},
		"type": rschema.StringAttribute{
			Computed: true,
		},
		"compare_type": rschema.StringAttribute{
			Computed: true,
		},
		"key": rschema.StringAttribute{
			Computed: true,
		},
		"value": rschema.StringAttribute{
			Computed: true,
		},
		"is_inverted": rschema.BoolAttribute{
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

func getL7PolicyDataSourceSchema() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"name": dschema.StringAttribute{
			Computed: true,
		},
		"description": dschema.StringAttribute{
			Computed: true,
		},
		"action": dschema.StringAttribute{
			Computed: true,
		},
		"position": dschema.Int32Attribute{
			Computed: true,
		},
		"redirect_target_group_id": dschema.StringAttribute{
			Computed: true,
		},
		"redirect_url": dschema.StringAttribute{
			Computed: true,
		},
		"redirect_prefix": dschema.StringAttribute{
			Computed: true,
		},
		"redirect_http_code": dschema.Int32Attribute{
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
		"rules": dschema.ListNestedAttribute{
			Computed: true,
			NestedObject: dschema.NestedAttributeObject{
				Attributes: getL7PolicyRuleDataSourceSchemaAttributes(),
			},
		},
	}
}

func getL7PolicyRuleDataSourceSchemaAttributes() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"id": dschema.StringAttribute{
			Computed: true,
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

var loadBalancerL7PolicyResourceSchemaAttributes = getL7PolicyResourceSchema()

var loadBalancerL7PolicyDataSourceSchemaAttributes = getL7PolicyDataSourceSchema()

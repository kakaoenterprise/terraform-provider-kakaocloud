// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package network

import (
	"terraform-provider-kakaocloud/internal/common"

	"github.com/hashicorp/terraform-plugin-framework-nettypes/cidrtypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/kakaoenterprise/kc-sdk-go/services/network"
)

func getSecurityGroupRuleInlineResourceSchema() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			Computed: true,
		},
		"remote_group_name": rschema.StringAttribute{
			Computed: true,
		},
		"created_at": rschema.StringAttribute{
			Computed: true,
		},
		"updated_at": rschema.StringAttribute{
			Computed: true,
		},
		"direction": rschema.StringAttribute{
			Required: true,
			Validators: []validator.String{
				stringvalidator.OneOf(
					string(network.SECURITYGROUPRULEDIRECTION_INGRESS),
					string(network.SECURITYGROUPRULEDIRECTION_EGRESS),
				),
			},
		},
		"protocol": rschema.StringAttribute{
			Required: true,
			Validators: []validator.String{
				stringvalidator.OneOf(
					string(network.SECURITYGROUPRULEPROTOCOL_TCP),
					string(network.SECURITYGROUPRULEPROTOCOL_UDP),
					string(network.SECURITYGROUPRULEPROTOCOL_ICMP),
					string(network.SECURITYGROUPRULEPROTOCOL_ALL),
				),
			},
		},
		"port_range_min": rschema.Int32Attribute{
			Optional: true,
			Computed: true,
			Validators: []validator.Int32{
				int32validator.Between(1, 65535),
			},
		},
		"port_range_max": rschema.Int32Attribute{
			Optional: true,
			Computed: true,
			Validators: []validator.Int32{
				int32validator.Between(1, 65535),
			},
		},
		"remote_ip_prefix": rschema.StringAttribute{
			Optional:   true,
			Computed:   true,
			CustomType: cidrtypes.IPPrefixType{},
		},
		"remote_group_id": rschema.StringAttribute{
			Optional:   true,
			Computed:   true,
			Validators: common.UuidValidator(),
		},
		"description": rschema.StringAttribute{
			Optional:   true,
			Computed:   true,
			Validators: common.DescriptionValidatorWithMaxLength(50),
		},
	}
}

func getSecurityGroupResourceSchema() map[string]rschema.Attribute {
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
			Computed:   true,
			Validators: common.DescriptionValidator(),
		},
		"project_id": rschema.StringAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"project_name": rschema.StringAttribute{
			Computed: true,
		},
		"is_stateful": rschema.BoolAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.Bool{
				boolplanmodifier.UseStateForUnknown(),
			},
		},
		"rules": rschema.SetNestedAttribute{
			Optional: true,
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
			},
			NestedObject: rschema.NestedAttributeObject{
				Attributes: securityGroupRuleInlineResourceSchema,
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
	}
}

func getSecurityGroupRuleDataSourceSchema() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"id": dschema.StringAttribute{
			Computed: true,
		},
		"description": dschema.StringAttribute{
			Computed: true,
		},
		"remote_group_id": dschema.StringAttribute{
			Computed: true,
		},
		"remote_group_name": dschema.StringAttribute{
			Computed: true,
		},
		"direction": dschema.StringAttribute{
			Computed: true,
		},
		"protocol": dschema.StringAttribute{
			Computed: true,
		},
		"port_range_min": dschema.Int32Attribute{
			Computed: true,
		},
		"port_range_max": dschema.Int32Attribute{
			Computed: true,
		},
		"remote_ip_prefix": dschema.StringAttribute{
			Computed:   true,
			CustomType: cidrtypes.IPPrefixType{},
		},
		"created_at": dschema.StringAttribute{
			Computed: true,
		},
		"updated_at": dschema.StringAttribute{
			Computed: true,
		},
	}
}

func getSecurityGroupDataSourceSchema() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"name": dschema.StringAttribute{
			Computed: true,
		},
		"description": dschema.StringAttribute{
			Computed: true,
		},
		"project_id": dschema.StringAttribute{
			Computed: true,
		},
		"project_name": dschema.StringAttribute{
			Computed: true,
		},
		"is_stateful": dschema.BoolAttribute{
			Computed: true,
		},
		"rules": dschema.SetNestedAttribute{
			Computed: true,
			NestedObject: dschema.NestedAttributeObject{
				Attributes: securityGroupRuleDataSourceSchema,
			},
		},
		"created_at": dschema.StringAttribute{
			Computed: true,
		},
		"updated_at": dschema.StringAttribute{
			Computed: true,
		},
	}
}

var securityGroupResourceSchemaAttributes = getSecurityGroupResourceSchema()
var securityGroupRuleInlineResourceSchema = getSecurityGroupRuleInlineResourceSchema()

var securityGroupDataSourceSchemaAttributes = getSecurityGroupDataSourceSchema()
var securityGroupRuleDataSourceSchema = getSecurityGroupRuleDataSourceSchema()

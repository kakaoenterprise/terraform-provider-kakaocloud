// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package network

import (
	"terraform-provider-kakaocloud/internal/docs"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/kakaoenterprise/kc-sdk-go/services/network"
)

func getSecurityGroupRuleInlineResourceSchema() map[string]rschema.Attribute {
	desc := docs.Network("bns_network__v1__api__get_security_group__model__SecurityGroupRuleModel")

	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("id"),
		},
		"remote_group_name": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("remote_group_name"),
		},
		"created_at": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("created_at"),
		},
		"updated_at": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("updated_at"),
		},
		"direction": rschema.StringAttribute{
			Required:    true,
			Description: desc.String("direction"),
			Validators: []validator.String{
				stringvalidator.OneOf(
					string(network.SECURITYGROUPRULEDIRECTION_INGRESS),
					string(network.SECURITYGROUPRULEDIRECTION_EGRESS),
				),
			},
		},
		"protocol": rschema.StringAttribute{
			Required:    true,
			Description: desc.String("protocol"),
			Validators: []validator.String{
				stringvalidator.OneOf(
					string(network.SECURITYGROUPRULEPROTOCOL_TCP),
					string(network.SECURITYGROUPRULEPROTOCOL_UDP),
					string(network.SECURITYGROUPRULEPROTOCOL_ICMP),
					string(network.SECURITYGROUPRULEPROTOCOL_ALL),
				),
			},
		},
		"port_range_min": rschema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: desc.String("port_range_min"),
		},
		"port_range_max": rschema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: desc.String("port_range_max"),
		},
		"remote_ip_prefix": rschema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: desc.String("remote_ip_prefix"),
		},
		"remote_group_id": rschema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: desc.String("remote_group_id"),
		},
		"description": rschema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: desc.String("description"),
		},
	}
}

func getSecurityGroupResourceSchema() map[string]rschema.Attribute {
	desc := docs.Network("bns_network__v1__api__get_security_group__model__SecurityGroupModel")

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
		},
		"description": rschema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: desc.String("description"),
		},
		"project_id": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("project_id"),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"project_name": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("project_name"),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"is_stateful": rschema.BoolAttribute{
			Computed:    true,
			Description: desc.String("is_stateful"),
			PlanModifiers: []planmodifier.Bool{
				boolplanmodifier.UseStateForUnknown(),
			},
		},
		"rules": rschema.SetNestedAttribute{
			Optional:    true,
			Computed:    true,
			Description: desc.String("rules"),
			PlanModifiers: []planmodifier.Set{
				setplanmodifier.UseStateForUnknown(),
			},
			NestedObject: rschema.NestedAttributeObject{
				Attributes: securityGroupRuleInlineResourceSchema,
			},
		},
		"created_at": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("created_at"),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"updated_at": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("updated_at"),
		},
	}
}

func getSecurityGroupRuleDataSourceSchema() map[string]dschema.Attribute {
	desc := docs.Network("bns_network__v1__api__get_security_group__model__SecurityGroupRuleModel")

	return map[string]dschema.Attribute{
		"id": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("id"),
		},
		"description": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("description"),
		},
		"remote_group_id": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("remote_group_id"),
		},
		"remote_group_name": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("remote_group_name"),
		},
		"direction": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("direction"),
		},
		"protocol": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("protocol"),
		},
		"port_range_min": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("port_range_min"),
		},
		"port_range_max": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("port_range_max"),
		},
		"remote_ip_prefix": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("remote_ip_prefix"),
		},
		"created_at": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("created_at"),
		},
		"updated_at": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("updated_at"),
		},
	}
}

func getSecurityGroupDataSourceSchema() map[string]dschema.Attribute {
	desc := docs.Network("bns_network__v1__api__get_security_group__model__SecurityGroupModel")

	return map[string]dschema.Attribute{
		"name": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("name"),
		},
		"description": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("description"),
		},
		"project_id": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("project_id"),
		},
		"project_name": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("project_name"),
		},
		"is_stateful": dschema.BoolAttribute{
			Computed:    true,
			Description: desc.String("is_stateful"),
		},
		"rules": dschema.SetNestedAttribute{
			Computed: true,
			NestedObject: dschema.NestedAttributeObject{
				Attributes: securityGroupRuleDataSourceSchema,
			},
		},
		"created_at": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("created_at"),
		},
		"updated_at": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("updated_at"),
		},
	}
}

// securityGroupResourceSchemaAttributes defines the schema attributes for the security group resource.
var securityGroupResourceSchemaAttributes = getSecurityGroupResourceSchema()
var securityGroupRuleInlineResourceSchema = getSecurityGroupRuleInlineResourceSchema()

// Schema for security group data source
var securityGroupDataSourceSchemaAttributes = getSecurityGroupDataSourceSchema()
var securityGroupRuleDataSourceSchema = getSecurityGroupRuleDataSourceSchema()

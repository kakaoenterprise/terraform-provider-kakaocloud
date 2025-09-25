// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package loadbalancer

import (
	"terraform-provider-kakaocloud/internal/common"
	"terraform-provider-kakaocloud/internal/docs"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func getInsertHeadersResourceSchema() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"x_forwarded_for": rschema.StringAttribute{
			Description: "Configures the X-Forwarded-For header.",
			Optional:    true,
			Computed:    true,
		},
		"x_forwarded_proto": rschema.StringAttribute{
			Description: "Configures the X-Forwarded-Proto header.",
			Optional:    true,
			Computed:    true,
		},
		"x_forwarded_port": rschema.StringAttribute{
			Description: "Configures the X-Forwarded-Port header.",
			Optional:    true,
			Computed:    true,
		},
	}
}

func getInsertHeadersDataSourceSchema() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"x_forwarded_for": dschema.StringAttribute{
			Description: "Indicates if the X-Forwarded-For header is inserted.",
			Computed:    true,
		},
		"x_forwarded_proto": dschema.StringAttribute{
			Description: "Indicates if the X-Forwarded-Proto header is inserted.",
			Computed:    true,
		},
		"x_forwarded_port": dschema.StringAttribute{
			Description: "Indicates if the X-Forwarded-Port header is inserted.",
			Computed:    true,
		},
	}
}

func getListenerResourceSchema() map[string]rschema.Attribute {
	desc := docs.Loadbalancer("bns_load_balancer__v1__api__create_listener__model__ListenerModel")
	getDesc := docs.Loadbalancer("bns_load_balancer__v1__api__get_listener__model__ListenerModel")
	createDesc := docs.Loadbalancer("CreateListener")

	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("id"),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"name": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("name"),
			Validators:  common.NameValidator(255),
		},
		"description": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("description"),
			Validators:  common.DescriptionValidator(),
		},
		"protocol": rschema.StringAttribute{
			Required:    true,
			Description: desc.String("protocol"),
			Validators:  common.ProtocolValidator(),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"is_enabled": rschema.BoolAttribute{
			Computed:    true,
			Description: getDesc.String("is_enabled"),
			PlanModifiers: []planmodifier.Bool{
				boolplanmodifier.UseStateForUnknown(),
			},
		},
		"tls_ciphers": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("tls_ciphers"),
		},
		"tls_versions": rschema.ListAttribute{
			ElementType: types.StringType,
			Computed:    true,
			Description: desc.String("tls_versions"),
		},
		"alpn_protocols": rschema.ListAttribute{
			ElementType: types.StringType,
			Computed:    true,
			Description: desc.String("alpn_protocols"),
			PlanModifiers: []planmodifier.List{
				listplanmodifier.UseStateForUnknown(),
			},
		},
		"project_id": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("project_id"),
		},
		"protocol_port": rschema.Int64Attribute{
			Required:    true,
			Description: desc.String("protocol_port"),
			Validators:  common.PortValidator(),
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.RequiresReplace(),
			},
		},
		"connection_limit": rschema.Int64Attribute{
			Optional:    true,
			Computed:    true,
			Description: desc.String("connection_limit"),
			Validators:  common.ConnectionLimitValidator(),
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
			},
		},
		"load_balancer_id": rschema.StringAttribute{
			Required:    true,
			Description: getDesc.String("load_balancer_id"),
			Validators:  common.UuidValidator(),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"tls_certificate_id": rschema.StringAttribute{
			Computed:    true,
			Description: getDesc.String("tls_certificate_id"),
			Validators:  common.UuidValidator(),
		},
		"provisioning_status": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("provisioning_status"),
		},
		"operating_status": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("operating_status"),
		},
		"insert_headers": rschema.SingleNestedAttribute{
			Description: desc.String("insert_headers"),
			Optional:    true,
			Computed:    true,
			Attributes:  getInsertHeadersResourceSchema(),
			PlanModifiers: []planmodifier.Object{
				objectplanmodifier.UseStateForUnknown(),
			},
		},
		"created_at": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("created_at"),
		},
		"updated_at": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("updated_at"),
		},
		"timeout_client_data": rschema.Int64Attribute{
			Optional:    true,
			Computed:    true,
			Description: desc.String("timeout_client_data"),
			Validators: []validator.Int64{
				int64validator.AtLeast(1000),
				int64validator.AtMost(4000000),
			},
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
			},
		},
		"default_target_group_id": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("default_target_group_id"),
			Validators:  common.UuidValidator(),
		},
		"default_target_group_name": rschema.StringAttribute{
			Computed:    true,
			Description: getDesc.String("default_target_group_name"),
		},
		"load_balancer_type": rschema.StringAttribute{
			Computed:    true,
			Description: getDesc.String("load_balancer_type"),
		},
		"sni_container_refs": rschema.ListAttribute{
			ElementType: types.StringType,
			Optional:    true,
			Description: desc.String("sni_container_refs"),
		},
		"default_tls_container_ref": rschema.StringAttribute{
			Optional:    true,
			Description: desc.String("default_tls_container_ref"),
		},
		"target_group_id": rschema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: docs.ParameterDescription("loadbalancer", "get_target_group", "path_target_group_id"),
			Validators:  common.UuidValidator(),
			PlanModifiers: []planmodifier.String{
				common.NewPreserveStateWhenNotSet(),
			},
		},
		"tls_min_version": rschema.StringAttribute{
			Optional:    true,
			Description: createDesc.String("tls_min_version"),
		},
		"secrets": rschema.ListNestedAttribute{
			Computed:    true,
			Description: getDesc.String("secrets"),
			NestedObject: rschema.NestedAttributeObject{
				Attributes: getListenerSecretSchema(),
			},
		},
		"l7_policies": rschema.ListNestedAttribute{
			Computed:    true,
			Description: desc.String("l7_policies"),
			NestedObject: rschema.NestedAttributeObject{
				Attributes: getListenerL7PolicySchema(),
			},
		},
	}
}

func getListenerSecretSchema() map[string]rschema.Attribute {
	secretDesc := docs.Loadbalancer("bns_load_balancer__v1__api__get_listener__model__SecretModel")

	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			Computed:    true,
			Description: secretDesc.String("id"),
		},
		"name": rschema.StringAttribute{
			Computed:    true,
			Description: secretDesc.String("name"),
		},
		"expiration": rschema.StringAttribute{
			Computed:    true,
			Description: secretDesc.String("expiration"),
		},
		"status": rschema.StringAttribute{
			Computed:    true,
			Description: secretDesc.String("status"),
		},
		"secret_type": rschema.StringAttribute{
			Computed:    true,
			Description: secretDesc.String("secret_type"),
		},
		"is_default": rschema.BoolAttribute{
			Computed:    true,
			Description: secretDesc.String("is_default"),
		},
		"creator_id": rschema.StringAttribute{
			Computed:    true,
			Description: secretDesc.String("creator_id"),
		},
	}
}

func getListenerL7PolicySchema() map[string]rschema.Attribute {
	policyDesc := docs.Loadbalancer("bns_load_balancer__v1__api__get_l7_policy__model__l7PolicyModel")
	ruleDesc := docs.Loadbalancer("bns_load_balancer__v1__api__get_listener__model__RuleModel")

	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			Computed:    true,
			Description: policyDesc.String("id"),
		},
		"name": rschema.StringAttribute{
			Computed:    true,
			Description: policyDesc.String("name"),
		},
		"description": rschema.StringAttribute{
			Computed:    true,
			Description: policyDesc.String("description"),
		},
		"provisioning_status": rschema.StringAttribute{
			Computed:    true,
			Description: policyDesc.String("provisioning_status"),
		},
		"operating_status": rschema.StringAttribute{
			Computed:    true,
			Description: policyDesc.String("operating_status"),
		},
		"project_id": rschema.StringAttribute{
			Computed:    true,
			Description: policyDesc.String("project_id"),
		},
		"action": rschema.StringAttribute{
			Computed:    true,
			Description: policyDesc.String("action"),
		},
		"position": rschema.Int64Attribute{
			Computed:    true,
			Description: policyDesc.String("position"),
		},
		"redirect_target_group_id": rschema.StringAttribute{
			Computed:    true,
			Description: policyDesc.String("redirect_target_group_id"),
		},
		"redirect_url": rschema.StringAttribute{
			Computed:    true,
			Description: policyDesc.String("redirect_url"),
		},
		"redirect_prefix": rschema.StringAttribute{
			Computed:    true,
			Description: policyDesc.String("redirect_prefix"),
		},
		"redirect_http_code": rschema.Int64Attribute{
			Computed:    true,
			Description: policyDesc.String("redirect_http_code"),
		},
		"rules": rschema.ListNestedAttribute{
			Computed:    true,
			Description: policyDesc.String("rules"),
			NestedObject: rschema.NestedAttributeObject{
				Attributes: map[string]rschema.Attribute{
					"id": rschema.StringAttribute{
						Computed:    true,
						Description: ruleDesc.String("id"),
					},
					"compare_type": rschema.StringAttribute{
						Computed:    true,
						Description: ruleDesc.String("compare_type"),
					},
					"is_inverted": rschema.BoolAttribute{
						Computed:    true,
						Description: ruleDesc.String("is_inverted"),
					},
					"key": rschema.StringAttribute{
						Computed:    true,
						Description: ruleDesc.String("key"),
					},
					"value": rschema.StringAttribute{
						Computed:    true,
						Description: ruleDesc.String("value"),
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
					"type": rschema.StringAttribute{
						Computed:    true,
						Description: ruleDesc.String("type"),
					},
				},
			},
		},
	}
}

func getListenerDataSourceSchema() map[string]dschema.Attribute {
	desc := docs.Loadbalancer("bns_load_balancer__v1__api__create_listener__model__ListenerModel")
	getDesc := docs.Loadbalancer("bns_load_balancer__v1__api__get_listener__model__ListenerModel")

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
		"protocol": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("protocol"),
		},
		"is_enabled": dschema.BoolAttribute{
			Computed:    true,
			Description: getDesc.String("is_enabled"),
		},
		"tls_ciphers": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("tls_ciphers"),
		},
		"tls_versions": dschema.ListAttribute{
			ElementType: types.StringType,
			Computed:    true,
			Description: desc.String("tls_versions"),
		},
		"alpn_protocols": dschema.ListAttribute{
			ElementType: types.StringType,
			Computed:    true,
			Description: desc.String("alpn_protocols"),
		},
		"project_id": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("project_id"),
		},
		"protocol_port": dschema.Int64Attribute{
			Computed:    true,
			Description: desc.String("protocol_port"),
		},
		"connection_limit": dschema.Int64Attribute{
			Computed:    true,
			Description: desc.String("connection_limit"),
		},
		"load_balancer_id": dschema.StringAttribute{
			Computed:    true,
			Description: getDesc.String("load_balancer_id"),
		},
		"tls_certificate_id": dschema.StringAttribute{
			Computed:    true,
			Description: getDesc.String("tls_certificate_id"),
		},
		"provisioning_status": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("provisioning_status"),
		},
		"operating_status": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("operating_status"),
		},
		"insert_headers": dschema.SingleNestedAttribute{
			Attributes:  getInsertHeadersDataSourceSchema(),
			Computed:    true,
			Description: desc.String("insert_headers"),
		},
		"created_at": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("created_at"),
		},
		"updated_at": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("updated_at"),
		},
		"timeout_client_data": dschema.Int64Attribute{
			Computed:    true,
			Description: desc.String("timeout_client_data"),
		},
		"default_target_group_name": dschema.StringAttribute{
			Computed:    true,
			Description: getDesc.String("default_target_group_name"),
		},
		"default_target_group_id": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("default_target_group_id"),
		},
		"load_balancer_type": dschema.StringAttribute{
			Computed:    true,
			Description: getDesc.String("load_balancer_type"),
		},
		"secrets": dschema.ListNestedAttribute{
			Computed:    true,
			Description: getDesc.String("secrets"),
			NestedObject: dschema.NestedAttributeObject{
				Attributes: getListenerSecretDataSourceSchema(),
			},
		},
		"l7_policies": dschema.ListNestedAttribute{
			Computed:    true,
			Description: desc.String("l7_policies"),
			NestedObject: dschema.NestedAttributeObject{
				Attributes: getListenerL7PolicyDataSourceSchema(),
			},
		},
	}
}

func getListenerSecretDataSourceSchema() map[string]dschema.Attribute {
	secretDesc := docs.Loadbalancer("bns_load_balancer__v1__api__get_listener__model__SecretModel")

	return map[string]dschema.Attribute{
		"id": dschema.StringAttribute{
			Computed:    true,
			Description: secretDesc.String("id"),
		},
		"name": dschema.StringAttribute{
			Computed:    true,
			Description: secretDesc.String("name"),
		},
		"expiration": dschema.StringAttribute{
			Computed:    true,
			Description: secretDesc.String("expiration"),
		},
		"status": dschema.StringAttribute{
			Computed:    true,
			Description: secretDesc.String("status"),
		},
		"secret_type": dschema.StringAttribute{
			Computed:    true,
			Description: secretDesc.String("secret_type"),
		},
		"is_default": dschema.BoolAttribute{
			Computed:    true,
			Description: secretDesc.String("is_default"),
		},
		"creator_id": dschema.StringAttribute{
			Computed:    true,
			Description: secretDesc.String("creator_id"),
		},
	}
}

func getListenerL7PolicyDataSourceSchema() map[string]dschema.Attribute {
	policyDesc := docs.Loadbalancer("bns_load_balancer__v1__api__get_l7_policy__model__l7PolicyModel")
	ruleDesc := docs.Loadbalancer("bns_load_balancer__v1__api__get_listener__model__RuleModel")

	return map[string]dschema.Attribute{
		"id": dschema.StringAttribute{
			Computed:    true,
			Description: policyDesc.String("id"),
		},
		"name": dschema.StringAttribute{
			Computed:    true,
			Description: policyDesc.String("name"),
		},
		"description": dschema.StringAttribute{
			Computed:    true,
			Description: policyDesc.String("description"),
		},
		"provisioning_status": dschema.StringAttribute{
			Computed:    true,
			Description: policyDesc.String("provisioning_status"),
		},
		"operating_status": dschema.StringAttribute{
			Computed:    true,
			Description: policyDesc.String("operating_status"),
		},
		"project_id": dschema.StringAttribute{
			Computed:    true,
			Description: policyDesc.String("project_id"),
		},
		"action": dschema.StringAttribute{
			Computed:    true,
			Description: policyDesc.String("action"),
		},
		"position": dschema.Int64Attribute{
			Computed:    true,
			Description: policyDesc.String("position"),
		},
		"redirect_target_group_id": dschema.StringAttribute{
			Computed:    true,
			Description: policyDesc.String("redirect_target_group_id"),
		},
		"redirect_url": dschema.StringAttribute{
			Computed:    true,
			Description: policyDesc.String("redirect_url"),
		},
		"redirect_prefix": dschema.StringAttribute{
			Computed:    true,
			Description: policyDesc.String("redirect_prefix"),
		},
		"redirect_http_code": dschema.Int64Attribute{
			Computed:    true,
			Description: policyDesc.String("redirect_http_code"),
		},
		"rules": dschema.ListNestedAttribute{
			Computed:    true,
			Description: policyDesc.String("rules"),
			NestedObject: dschema.NestedAttributeObject{
				Attributes: map[string]dschema.Attribute{
					"id": dschema.StringAttribute{
						Computed:    true,
						Description: ruleDesc.String("id"),
					},
					"compare_type": dschema.StringAttribute{
						Computed:    true,
						Description: ruleDesc.String("compare_type"),
					},
					"is_inverted": dschema.BoolAttribute{
						Computed:    true,
						Description: ruleDesc.String("is_inverted"),
					},
					"key": dschema.StringAttribute{
						Computed:    true,
						Description: ruleDesc.String("key"),
					},
					"value": dschema.StringAttribute{
						Computed:    true,
						Description: ruleDesc.String("value"),
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
					"type": dschema.StringAttribute{
						Computed:    true,
						Description: ruleDesc.String("type"),
					},
				},
			},
		},
	}
}

// listenerResourceSchemaAttributes defines the schema for the listener resource
var listenerResourceSchemaAttributes = getListenerResourceSchema()

// listenerDataSourceSchemaAttributes defines the schema for the listener data source
var listenerDataSourceSchemaAttributes = getListenerDataSourceSchema()

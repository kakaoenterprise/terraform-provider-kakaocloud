// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package loadbalancer

import (
	"terraform-provider-kakaocloud/internal/common"

	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kakaoenterprise/kc-sdk-go/services/loadbalancer"
)

func getInsertHeadersResourceSchema() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"x_forwarded_for": rschema.StringAttribute{
			Optional: true,
			Computed: true,
			Validators: []validator.String{
				stringvalidator.OneOf("true", "false", "remove"),
			},
		},
		"x_forwarded_proto": rschema.StringAttribute{
			Optional: true,
			Computed: true,
			Validators: []validator.String{
				stringvalidator.OneOf("true", "false"),
			},
		},
		"x_forwarded_port": rschema.StringAttribute{
			Optional: true,
			Computed: true,
			Validators: []validator.String{
				stringvalidator.OneOf("true", "false"),
			},
		},
	}
}

func getInsertHeadersDataSourceSchema() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"x_forwarded_for": dschema.StringAttribute{
			Computed: true,
		},
		"x_forwarded_proto": dschema.StringAttribute{
			Computed: true,
		},
		"x_forwarded_port": dschema.StringAttribute{
			Computed: true,
		},
	}
}

func getListenerResourceSchema() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"name": rschema.StringAttribute{
			Computed: true,
		},
		"description": rschema.StringAttribute{
			Computed: true,
		},
		"protocol": rschema.StringAttribute{
			Required:   true,
			Validators: common.ProtocolValidator(),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"is_enabled": rschema.BoolAttribute{
			Computed: true,
		},
		"tls_ciphers": rschema.StringAttribute{
			Computed: true,
		},
		"tls_versions": rschema.ListAttribute{
			ElementType: types.StringType,
			Computed:    true,
		},
		"alpn_protocols": rschema.ListAttribute{
			ElementType: types.StringType,
			Computed:    true,
			PlanModifiers: []planmodifier.List{
				listplanmodifier.UseStateForUnknown(),
			},
		},
		"project_id": rschema.StringAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"protocol_port": rschema.Int32Attribute{
			Required:   true,
			Validators: common.PortValidator(),
			PlanModifiers: []planmodifier.Int32{
				int32planmodifier.RequiresReplace(),
			},
		},
		"connection_limit": rschema.Int32Attribute{
			Optional:   true,
			Computed:   true,
			Validators: common.ConnectionLimitValidator(),
		},
		"load_balancer_id": rschema.StringAttribute{
			Required:   true,
			Validators: common.UuidValidator(),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"tls_certificate_id": rschema.StringAttribute{
			Computed: true,
		},
		"provisioning_status": rschema.StringAttribute{
			Computed: true,
		},
		"operating_status": rschema.StringAttribute{
			Computed: true,
		},
		"insert_headers": rschema.SingleNestedAttribute{
			Optional:   true,
			Computed:   true,
			Attributes: getInsertHeadersResourceSchema(),
		},
		"created_at": rschema.StringAttribute{
			Computed: true,
		},
		"updated_at": rschema.StringAttribute{
			Computed: true,
		},
		"timeout_client_data": rschema.Int32Attribute{
			Optional: true,
			Computed: true,
			Validators: []validator.Int32{
				int32validator.AtLeast(1000),
				int32validator.AtMost(4000000),
			},
		},
		"default_target_group_id": rschema.StringAttribute{
			Computed: true,
		},
		"default_target_group_name": rschema.StringAttribute{
			Computed: true,
		},
		"load_balancer_type": rschema.StringAttribute{
			Computed: true,
		},
		"sni_container_refs": rschema.ListAttribute{
			ElementType: types.StringType,
			Optional:    true,
		},
		"default_tls_container_ref": rschema.StringAttribute{
			Optional: true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
		"target_group_id": rschema.StringAttribute{
			Optional:   true,
			Validators: common.UuidValidator(),
		},
		"tls_min_version": rschema.StringAttribute{
			Optional: true,
			Validators: []validator.String{
				stringvalidator.OneOf(
					string(loadbalancer.TLSVERSION_TLSV1),
					string(loadbalancer.TLSVERSION_TLSV1_1),
					string(loadbalancer.TLSVERSION_TLSV1_2),
				),
			},
		},
		"secrets": rschema.ListNestedAttribute{
			Computed: true,
			NestedObject: rschema.NestedAttributeObject{
				Attributes: getListenerSecretSchema(),
			},
		},
		"l7_policies": rschema.ListNestedAttribute{
			Computed: true,
			NestedObject: rschema.NestedAttributeObject{
				Attributes: getListenerL7PolicySchema(),
			},
		},
	}
}

func getListenerSecretSchema() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			Computed: true,
		},
		"name": rschema.StringAttribute{
			Computed: true,
		},
		"expiration": rschema.StringAttribute{
			Computed: true,
		},
		"status": rschema.StringAttribute{
			Computed: true,
		},
		"secret_type": rschema.StringAttribute{
			Computed: true,
		},
		"is_default": rschema.BoolAttribute{
			Computed: true,
		},
		"creator_id": rschema.StringAttribute{
			Computed: true,
		},
	}
}

func getListenerL7PolicySchema() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			Computed: true,
		},
		"name": rschema.StringAttribute{
			Computed: true,
		},
		"description": rschema.StringAttribute{
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
		"action": rschema.StringAttribute{
			Computed: true,
		},
		"position": rschema.Int32Attribute{
			Computed: true,
		},
		"redirect_target_group_id": rschema.StringAttribute{
			Computed: true,
		},
		"redirect_url": rschema.StringAttribute{
			Computed: true,
		},
		"redirect_prefix": rschema.StringAttribute{
			Computed: true,
		},
		"redirect_http_code": rschema.Int32Attribute{
			Computed: true,
		},
		"rules": rschema.ListNestedAttribute{
			Computed: true,
			NestedObject: rschema.NestedAttributeObject{
				Attributes: map[string]rschema.Attribute{
					"id": rschema.StringAttribute{
						Computed: true,
					},
					"compare_type": rschema.StringAttribute{
						Computed: true,
					},
					"is_inverted": rschema.BoolAttribute{
						Computed: true,
					},
					"key": rschema.StringAttribute{
						Computed: true,
					},
					"value": rschema.StringAttribute{
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
					"type": rschema.StringAttribute{
						Computed: true,
					},
				},
			},
		},
	}
}

func getListenerDataSourceSchema() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"name": dschema.StringAttribute{
			Computed: true,
		},
		"description": dschema.StringAttribute{
			Computed: true,
		},
		"protocol": dschema.StringAttribute{
			Computed: true,
		},
		"is_enabled": dschema.BoolAttribute{
			Computed: true,
		},
		"tls_ciphers": dschema.StringAttribute{
			Computed: true,
		},
		"tls_versions": dschema.ListAttribute{
			ElementType: types.StringType,
			Computed:    true,
		},
		"alpn_protocols": dschema.ListAttribute{
			ElementType: types.StringType,
			Computed:    true,
		},
		"project_id": dschema.StringAttribute{
			Computed: true,
		},
		"protocol_port": dschema.Int32Attribute{
			Computed: true,
		},
		"connection_limit": dschema.Int32Attribute{
			Computed: true,
		},
		"load_balancer_id": dschema.StringAttribute{
			Computed: true,
		},
		"tls_certificate_id": dschema.StringAttribute{
			Computed: true,
		},
		"provisioning_status": dschema.StringAttribute{
			Computed: true,
		},
		"operating_status": dschema.StringAttribute{
			Computed: true,
		},
		"insert_headers": dschema.SingleNestedAttribute{
			Attributes: getInsertHeadersDataSourceSchema(),
			Computed:   true,
		},
		"created_at": dschema.StringAttribute{
			Computed: true,
		},
		"updated_at": dschema.StringAttribute{
			Computed: true,
		},
		"timeout_client_data": dschema.Int32Attribute{
			Computed: true,
		},
		"default_target_group_name": dschema.StringAttribute{
			Computed: true,
		},
		"default_target_group_id": dschema.StringAttribute{
			Computed: true,
		},
		"load_balancer_type": dschema.StringAttribute{
			Computed: true,
		},
		"secrets": dschema.ListNestedAttribute{
			Computed: true,
			NestedObject: dschema.NestedAttributeObject{
				Attributes: getListenerSecretDataSourceSchema(),
			},
		},
		"l7_policies": dschema.ListNestedAttribute{
			Computed: true,
			NestedObject: dschema.NestedAttributeObject{
				Attributes: getListenerL7PolicyDataSourceSchema(),
			},
		},
	}
}

func getListenerSecretDataSourceSchema() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"id": dschema.StringAttribute{
			Computed: true,
		},
		"name": dschema.StringAttribute{
			Computed: true,
		},
		"expiration": dschema.StringAttribute{
			Computed: true,
		},
		"status": dschema.StringAttribute{
			Computed: true,
		},
		"secret_type": dschema.StringAttribute{
			Computed: true,
		},
		"is_default": dschema.BoolAttribute{
			Computed: true,
		},
		"creator_id": dschema.StringAttribute{
			Computed: true,
		},
	}
}

func getListenerL7PolicyDataSourceSchema() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"id": dschema.StringAttribute{
			Computed: true,
		},
		"name": dschema.StringAttribute{
			Computed: true,
		},
		"description": dschema.StringAttribute{
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
		"rules": dschema.ListNestedAttribute{
			Computed: true,
			NestedObject: dschema.NestedAttributeObject{
				Attributes: map[string]dschema.Attribute{
					"id": dschema.StringAttribute{
						Computed: true,
					},
					"compare_type": dschema.StringAttribute{
						Computed: true,
					},
					"is_inverted": dschema.BoolAttribute{
						Computed: true,
					},
					"key": dschema.StringAttribute{
						Computed: true,
					},
					"value": dschema.StringAttribute{
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
					"type": dschema.StringAttribute{
						Computed: true,
					},
				},
			},
		},
	}
}

var listenerResourceSchemaAttributes = getListenerResourceSchema()

var listenerDataSourceSchemaAttributes = getListenerDataSourceSchema()

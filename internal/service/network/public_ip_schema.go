// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package network

import (
	"terraform-provider-kakaocloud/internal/common"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func getRelatedResourceSchema() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			Optional:   true,
			Computed:   true,
			Validators: common.UuidValidator(),
		},
		"name": rschema.StringAttribute{
			Computed: true,
		},
		"status": rschema.StringAttribute{
			Computed: true,
		},
		"type": rschema.StringAttribute{
			Computed: true,
		},
		"device_id": rschema.StringAttribute{
			Required:   true,
			Validators: common.UuidValidator(),
		},
		"device_owner": rschema.StringAttribute{
			Computed: true,
		},
		"device_type": rschema.StringAttribute{
			Required: true,
			Validators: []validator.String{
				stringvalidator.OneOf("instance", "load-balancer"),
			},
		},
		"subnet_id": rschema.StringAttribute{
			Computed: true,
		},
		"subnet_name": rschema.StringAttribute{
			Computed: true,
		},
		"subnet_cidr": rschema.StringAttribute{
			Computed: true,
		},
		"vpc_id": rschema.StringAttribute{
			Computed: true,
		},
		"vpc_name": rschema.StringAttribute{
			Computed: true,
		},
	}
}

func getPublicIpResourceSchema() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"status": rschema.StringAttribute{
			Computed: true,
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
		"public_ip": rschema.StringAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"private_ip": rschema.StringAttribute{
			Computed: true,
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
		"related_resource": rschema.SingleNestedAttribute{
			Optional:   true,
			Attributes: relatedResourceSchemaAttributes,
		},
	}
}

func getRelatedDataSourceSchema() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"id": dschema.StringAttribute{
			Computed: true,
		},
		"name": dschema.StringAttribute{
			Computed: true,
		},
		"status": dschema.StringAttribute{
			Computed: true,
		},
		"type": dschema.StringAttribute{
			Computed: true,
		},
		"device_id": dschema.StringAttribute{
			Computed: true,
		},
		"device_owner": dschema.StringAttribute{
			Computed: true,
		},
		"device_type": dschema.StringAttribute{
			Computed: true,
		},
		"subnet_id": dschema.StringAttribute{
			Computed: true,
		},
		"subnet_name": dschema.StringAttribute{
			Computed: true,
		},
		"subnet_cidr": dschema.StringAttribute{
			Computed: true,
		},
		"vpc_id": dschema.StringAttribute{
			Computed: true,
		},
		"vpc_name": dschema.StringAttribute{
			Computed: true,
		},
	}
}

func getPublicIpDataSourceSchema() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"status": dschema.StringAttribute{
			Computed: true,
		},
		"description": dschema.StringAttribute{
			Computed: true,
		},
		"project_id": dschema.StringAttribute{
			Computed: true,
		},
		"public_ip": dschema.StringAttribute{
			Computed: true,
		},
		"private_ip": dschema.StringAttribute{
			Computed: true,
		},
		"created_at": dschema.StringAttribute{
			Computed: true,
		},
		"updated_at": dschema.StringAttribute{
			Computed: true,
		},
		"related_resource": dschema.SingleNestedAttribute{
			Computed:   true,
			Attributes: relatedDataSourceSchemaAttributes,
		},
	}
}

var publicIpResourceSchema = getPublicIpResourceSchema()
var publicIpDataSourceSchemaAttributes = getPublicIpDataSourceSchema()
var relatedResourceSchemaAttributes = getRelatedResourceSchema()
var relatedDataSourceSchemaAttributes = getRelatedDataSourceSchema()

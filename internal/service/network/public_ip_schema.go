// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package network

import (
	"terraform-provider-kakaocloud/internal/common"
	"terraform-provider-kakaocloud/internal/docs"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func getRelatedResourceSchema() map[string]rschema.Attribute {
	desc := docs.Network("bns_network__v1__api__get_public_ip__model__RelatedResourceInfoModel")

	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: desc.String("id"),
			Validators:  common.UuidValidator(),
			PlanModifiers: []planmodifier.String{
				common.NullToUnknownString(),
			},
		},
		"name": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("name"),
		},
		"status": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("status"),
		},
		"type": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("type"),
		},
		"device_id": rschema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: desc.String("device_id"),
			Validators:  common.UuidValidator(),
		},
		"device_owner": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("device_owner"),
		},
		"device_type": rschema.StringAttribute{
			Required:    true,
			Description: "The type of device to associate with the public IP (instance, load-balancer).",
			Validators: []validator.String{
				stringvalidator.OneOf("instance", "load-balancer"),
			},
		},
		"subnet_id": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("subnet_id"),
		},
		"subnet_name": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("subnet_name"),
		},
		"subnet_cidr": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("subnet_cidr"),
		},
		"vpc_id": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("vpc_id"),
		},
		"vpc_name": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("vpc_name"),
		},
	}
}

func getPublicIpResourceSchema() map[string]rschema.Attribute {
	desc := docs.Network("bns_network__v1__api__get_public_ip__model__FloatingIpModel")

	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("id"),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"status": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("status"),
		},
		"description": rschema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: desc.String("description"),
			Validators:  common.DescriptionValidator(),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"project_id": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("project_id"),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"public_ip": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("public_ip"),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"private_ip": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("private_ip"),
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
		"related_resource": rschema.SingleNestedAttribute{
			Optional:    true,
			Computed:    true,
			Description: desc.String("related_resource"),
			Attributes:  relatedResourceSchemaAttributes,
		},
	}
}

func getRelatedDataSourceSchema() map[string]dschema.Attribute {
	desc := docs.Network("bns_network__v1__api__get_public_ip__model__RelatedResourceInfoModel")

	return map[string]dschema.Attribute{
		"id": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("id"),
		},
		"name": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("name"),
		},
		"status": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("status"),
		},
		"type": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("type"),
		},
		"device_id": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("device_id"),
		},
		"device_owner": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("device_owner"),
		},
		"device_type": dschema.StringAttribute{
			Computed:    true,
			Description: "The type of device associated with the public IP (instance, load-balancer).",
		},
		"subnet_id": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("subnet_id"),
		},
		"subnet_name": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("subnet_name"),
		},
		"subnet_cidr": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("subnet_cidr"),
		},
		"vpc_id": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("vpc_id"),
		},
		"vpc_name": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("vpc_name"),
		},
	}
}

func getPublicIpDataSourceSchema() map[string]dschema.Attribute {
	desc := docs.Network("bns_network__v1__api__get_public_ip__model__FloatingIpModel")

	return map[string]dschema.Attribute{
		"status": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("status"),
		},
		"description": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("description"),
		},
		"project_id": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("project_id"),
		},
		"public_ip": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("public_ip"),
		},
		"private_ip": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("private_ip"),
		},
		"created_at": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("created_at"),
		},
		"updated_at": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("updated_at"),
		},
		"related_resource": dschema.SingleNestedAttribute{
			Computed:    true,
			Description: desc.String("related_resource"),
			Attributes:  relatedDataSourceSchemaAttributes,
		},
	}
}

var publicIpResourceSchema = getPublicIpResourceSchema()
var publicIpDataSourceSchemaAttributes = getPublicIpDataSourceSchema()
var relatedResourceSchemaAttributes = getRelatedResourceSchema()
var relatedDataSourceSchemaAttributes = getRelatedDataSourceSchema()

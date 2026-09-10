// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package vpc

import (
	"terraform-provider-kakaocloud/internal/common"

	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func getIgwVpcDataSourceSchema() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"id": dschema.StringAttribute{
			Computed: true,
		},
		"name": dschema.StringAttribute{
			Computed: true,
		},
		"cidr_block": dschema.StringAttribute{
			Computed: true,
		},
	}
}

func getIgwAttachmentDataSourceSchema() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"id": dschema.StringAttribute{
			Computed: true,
		},
		"provisioning_status": dschema.StringAttribute{
			Computed: true,
		},
	}
}

func getInternetGatewayDataSourceSchemaAttributes() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"name": dschema.StringAttribute{
			Computed: true,
		},
		"description": dschema.StringAttribute{
			Computed: true,
		},
		"vpc": dschema.SingleNestedAttribute{
			Computed:   true,
			Attributes: igwVpcDataSourceSchema,
		},
		"attachment": dschema.SingleNestedAttribute{
			Computed:   true,
			Attributes: igwAttachmentDataSourceSchema,
		},
		"attachment_status": dschema.StringAttribute{
			Computed: true,
		},
		"nat_ip": dschema.StringAttribute{
			Computed: true,
		},
		"project_id": dschema.StringAttribute{
			Computed: true,
		},
		"project_name": dschema.StringAttribute{
			Computed: true,
		},
		"domain_id": dschema.StringAttribute{
			Computed: true,
		},
		"provisioning_status": dschema.StringAttribute{
			Computed: true,
		},
		"created_at": dschema.StringAttribute{
			Computed: true,
		},
		"updated_at": dschema.StringAttribute{
			Computed: true,
		},
	}
}

func getIgwVpcResourceSchema() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			Computed: true,
		},
		"name": rschema.StringAttribute{
			Computed: true,
		},
		"cidr_block": rschema.StringAttribute{
			Computed: true,
		},
	}
}

func getIgwAttachmentResourceSchema() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			Computed: true,
		},
		"provisioning_status": rschema.StringAttribute{
			Computed: true,
		},
	}
}

func getInternetGatewayResourceSchemaAttributes() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"name": rschema.StringAttribute{
			Optional:   true,
			Computed:   true,
			Validators: common.NameValidator(250),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"description": rschema.StringAttribute{
			Optional:   true,
			Computed:   true,
			Validators: common.DescriptionValidatorWithMaxLength(255),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"vpc": rschema.SingleNestedAttribute{
			Computed:   true,
			Attributes: igwVpcResourceSchema,
		},
		"attachment": rschema.SingleNestedAttribute{
			Computed:   true,
			Attributes: igwAttachmentResourceSchema,
		},
		"attachment_status": rschema.StringAttribute{
			Computed: true,
		},
		"nat_ip": rschema.StringAttribute{
			Computed: true,
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
		"domain_id": rschema.StringAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"provisioning_status": rschema.StringAttribute{
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
	}
}

var igwVpcResourceSchema = getIgwVpcResourceSchema()
var igwAttachmentResourceSchema = getIgwAttachmentResourceSchema()
var internetGatewayResourceSchemaAttributes = getInternetGatewayResourceSchemaAttributes()

var igwVpcDataSourceSchema = getIgwVpcDataSourceSchema()
var igwAttachmentDataSourceSchema = getIgwAttachmentDataSourceSchema()
var internetGatewayDataSourceSchemaAttributes = getInternetGatewayDataSourceSchemaAttributes()

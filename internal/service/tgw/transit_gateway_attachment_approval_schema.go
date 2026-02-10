// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package tgw

import (
	"terraform-provider-kakaocloud/internal/common"

	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

var transitGatewayAttachmentApprovalResourceSchemaAttributes = map[string]rschema.Attribute{
	"id": rschema.StringAttribute{
		Computed: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	},
	"attachment_id": rschema.StringAttribute{
		Required:   true,
		Validators: common.UuidValidator(),
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	},
	"tgw_id": rschema.StringAttribute{
		Computed: true,
	},
	"vpc_id": rschema.StringAttribute{
		Computed: true,
	},
	"provisioning_status": rschema.StringAttribute{
		Computed: true,
	},
	"tgw_project_id": rschema.StringAttribute{
		Computed: true,
	},
	"vpc_name": rschema.StringAttribute{
		Computed: true,
	},
	"cidr_block": rschema.StringAttribute{
		Computed: true,
	},
	"project_id": rschema.StringAttribute{
		Computed: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
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

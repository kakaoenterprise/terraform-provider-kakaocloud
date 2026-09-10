// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package vpc

import (
	"terraform-provider-kakaocloud/internal/common"

	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

var internetGatewayAttachmentResourceSchemaAttributes = map[string]rschema.Attribute{
	"id": rschema.StringAttribute{
		Computed:   true,
		Validators: common.UuidValidator(),
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	},
	"internet_gateway_id": rschema.StringAttribute{
		Required:   true,
		Validators: common.UuidValidator(),
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	},
	"vpc_id": rschema.StringAttribute{
		Required:   true,
		Validators: common.UuidValidator(),
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	},
	"provisioning_status": rschema.StringAttribute{
		Computed: true,
	},
}

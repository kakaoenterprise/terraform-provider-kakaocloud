// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package vpc

import (
	resourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type internetGatewayAttachmentModel struct {
	Id                 types.String `tfsdk:"id"`
	InternetGatewayId  types.String `tfsdk:"internet_gateway_id"`
	VpcId              types.String `tfsdk:"vpc_id"`
	ProvisioningStatus types.String `tfsdk:"provisioning_status"`
}

type internetGatewayAttachmentResourceModel struct {
	internetGatewayAttachmentModel
	Timeouts resourceTimeouts.Value `tfsdk:"timeouts"`
}

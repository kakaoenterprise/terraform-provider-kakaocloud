// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package tgw

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type transitGatewayAttachmentApprovalResourceModel struct {
	Id                 types.String   `tfsdk:"id"`
	AttachmentId       types.String   `tfsdk:"attachment_id"`
	TgwId              types.String   `tfsdk:"tgw_id"`
	VpcId              types.String   `tfsdk:"vpc_id"`
	ProvisioningStatus types.String   `tfsdk:"provisioning_status"`
	TgwProjectId       types.String   `tfsdk:"tgw_project_id"`
	VpcName            types.String   `tfsdk:"vpc_name"`
	CidrBlock          types.String   `tfsdk:"cidr_block"`
	ProjectId          types.String   `tfsdk:"project_id"`
	CreatedAt          types.String   `tfsdk:"created_at"`
	UpdatedAt          types.String   `tfsdk:"updated_at"`
	Timeouts           timeouts.Value `tfsdk:"timeouts"`
}

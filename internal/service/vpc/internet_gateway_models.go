// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package vpc

import (
	"terraform-provider-kakaocloud/internal/common"

	datasourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	resourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type internetGatewayBaseModel struct {
	Id                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	Description        types.String `tfsdk:"description"`
	Vpc                types.Object `tfsdk:"vpc"`
	Attachment         types.Object `tfsdk:"attachment"`
	AttachmentStatus   types.String `tfsdk:"attachment_status"`
	NatIp              types.String `tfsdk:"nat_ip"`
	ProjectId          types.String `tfsdk:"project_id"`
	ProjectName        types.String `tfsdk:"project_name"`
	DomainId           types.String `tfsdk:"domain_id"`
	ProvisioningStatus types.String `tfsdk:"provisioning_status"`
	CreatedAt          types.String `tfsdk:"created_at"`
	UpdatedAt          types.String `tfsdk:"updated_at"`
}

type internetGatewayResourceModel struct {
	internetGatewayBaseModel
	Timeouts resourceTimeouts.Value `tfsdk:"timeouts"`
}

type internetGatewayDataSourceModel struct {
	internetGatewayBaseModel
	Timeouts datasourceTimeouts.Value `tfsdk:"timeouts"`
}

type internetGatewaysDataSourceModel struct {
	Filter   []common.FilterModel       `tfsdk:"filter"`
	Igws     []internetGatewayBaseModel `tfsdk:"igws"`
	Timeouts datasourceTimeouts.Value   `tfsdk:"timeouts"`
}

type igwVpcModel struct {
	Id        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	CidrBlock types.String `tfsdk:"cidr_block"`
}

var igwVpcAttrType = map[string]attr.Type{
	"id":         types.StringType,
	"name":       types.StringType,
	"cidr_block": types.StringType,
}

type igwAttachmentModel struct {
	Id                 types.String `tfsdk:"id"`
	ProvisioningStatus types.String `tfsdk:"provisioning_status"`
}

var igwAttachmentAttrType = map[string]attr.Type{
	"id":                  types.StringType,
	"provisioning_status": types.StringType,
}

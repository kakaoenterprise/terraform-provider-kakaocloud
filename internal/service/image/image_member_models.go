// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package image

import (
	datasourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	resourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type imageMemberBaseModel struct {
	Id      types.String `tfsdk:"id"`
	Members types.List   `tfsdk:"members"`
}
type imageMemberMemberModel struct {
	Id        types.String `tfsdk:"id"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
	ImageId   types.String `tfsdk:"image_id"`
	Status    types.String `tfsdk:"status"`
	IsShared  types.Bool   `tfsdk:"is_shared"`
}

var imageMemberMembersAttrType = map[string]attr.Type{
	"id":         types.StringType,
	"created_at": types.StringType,
	"updated_at": types.StringType,
	"image_id":   types.StringType,
	"status":     types.StringType,
	"is_shared":  types.BoolType,
}

type imageMemberResourceModel struct {
	imageMemberBaseModel
	SharedMemberIds types.Set              `tfsdk:"shared_member_ids"`
	Timeouts        resourceTimeouts.Value `tfsdk:"timeouts"`
}

type imageMemberDataSourceModel struct {
	imageMemberBaseModel
	Timeouts datasourceTimeouts.Value `tfsdk:"timeouts"`
}

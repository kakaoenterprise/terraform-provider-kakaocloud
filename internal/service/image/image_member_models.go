// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package image

import (
	datasourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	resourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type imageMemberBaseModel struct {
	Id        types.String `tfsdk:"id"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
	ImageId   types.String `tfsdk:"image_id"`
	Status    types.String `tfsdk:"status"`
	IsShared  types.Bool   `tfsdk:"is_shared"`
}

type imageMemberResourceModel struct {
	MemberId  types.String           `tfsdk:"member_id"`
	Id        types.String           `tfsdk:"id"`
	CreatedAt types.String           `tfsdk:"created_at"`
	UpdatedAt types.String           `tfsdk:"updated_at"`
	ImageId   types.String           `tfsdk:"image_id"`
	Status    types.String           `tfsdk:"status"`
	Timeouts  resourceTimeouts.Value `tfsdk:"timeouts"`
}
type imageMemberDataSourceModel struct {
	ImageId  types.String             `tfsdk:"image_id"`
	Members  []imageMemberBaseModel   `tfsdk:"members"`
	Timeouts datasourceTimeouts.Value `tfsdk:"timeouts"`
}

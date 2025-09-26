// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package image

import (
	"terraform-provider-kakaocloud/internal/common"

	datasourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	resourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type imageBaseModel struct {
	Id                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	Size              types.Int64  `tfsdk:"size"`
	Status            types.String `tfsdk:"status"`
	Owner             types.String `tfsdk:"owner"`
	Visibility        types.String `tfsdk:"visibility"`
	Description       types.String `tfsdk:"description"`
	IsShared          types.Bool   `tfsdk:"is_shared"`
	DiskFormat        types.String `tfsdk:"disk_format"`
	ContainerFormat   types.String `tfsdk:"container_format"`
	MinDisk           types.Int32  `tfsdk:"min_disk"`
	MinRam            types.Int32  `tfsdk:"min_ram"`
	VirtualSize       types.Int64  `tfsdk:"virtual_size"`
	InstanceType      types.String `tfsdk:"instance_type"`
	ImageMemberStatus types.String `tfsdk:"image_member_status"`
	ProjectId         types.String `tfsdk:"project_id"`
	OsInfo            types.Object `tfsdk:"os_info"`
	CreatedAt         types.String `tfsdk:"created_at"`
	UpdatedAt         types.String `tfsdk:"updated_at"`
}

var osInfoAttrType = map[string]attr.Type{
	"type":         types.StringType,
	"distro":       types.StringType,
	"architecture": types.StringType,
	"admin_user":   types.StringType,
	"is_hidden":    types.BoolType,
}

type osInfoModel struct {
	Type         types.String `tfsdk:"type"`
	Distro       types.String `tfsdk:"distro"`
	Architecture types.String `tfsdk:"architecture"`
	AdminUser    types.String `tfsdk:"admin_user"`
	IsHidden     types.Bool   `tfsdk:"is_hidden"`
}

type imageDataSourceModel struct {
	imageBaseModel
	Timeouts datasourceTimeouts.Value `tfsdk:"timeouts"`
}

type imagesDataSourceModel struct {
	Filter   []common.FilterModel     `tfsdk:"filter"`
	Images   []imageBaseModel         `tfsdk:"images"`
	Timeouts datasourceTimeouts.Value `tfsdk:"timeouts"`
}

type imageResourceModel struct {
	imageBaseModel
	VolumeId types.String           `tfsdk:"volume_id"`
	Timeouts resourceTimeouts.Value `tfsdk:"timeouts"`
}

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package volume

import (
	"terraform-provider-kakaocloud/internal/common"

	datasourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	resourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type volumeBaseModel struct {
	Id               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	AvailabilityZone types.String `tfsdk:"availability_zone"`
	Status           types.String `tfsdk:"status"`
	MountPoint       types.String `tfsdk:"mount_point"`
	VolumeType       types.String `tfsdk:"volume_type"`
	Size             types.Int32  `tfsdk:"size"`
	IsRoot           types.Bool   `tfsdk:"is_root"`
	IsEncrypted      types.Bool   `tfsdk:"is_encrypted"`
	IsBootable       types.Bool   `tfsdk:"is_bootable"`
	Type             types.String `tfsdk:"type"`
	UserId           types.String `tfsdk:"user_id"`
	ProjectId        types.String `tfsdk:"project_id"`
	AttachStatus     types.String `tfsdk:"attach_status"`
	LaunchedAt       types.String `tfsdk:"launched_at"`
	EncryptionKeyId  types.String `tfsdk:"encryption_key_id"`
	PreviousStatus   types.String `tfsdk:"previous_status"`
	CreatedAt        types.String `tfsdk:"created_at"`
	UpdatedAt        types.String `tfsdk:"updated_at"`
	InstanceId       types.String `tfsdk:"instance_id"`
	InstanceName     types.String `tfsdk:"instance_name"`
	ImageMetadata    types.Object `tfsdk:"image_metadata"`
	Metadata         types.Map    `tfsdk:"metadata"`
}

type volumeResourceModel struct {
	volumeBaseModel
	VolumeTypeId       types.String           `tfsdk:"volume_type_id"`
	ImageId            types.String           `tfsdk:"image_id"`
	VolumeSnapshotId   types.String           `tfsdk:"volume_snapshot_id"`
	SourceVolumeId     types.String           `tfsdk:"source_volume_id"`
	EncryptionSecretId types.String           `tfsdk:"encryption_secret_id"`
	Timeouts           resourceTimeouts.Value `tfsdk:"timeouts"`
}

type volumeDataSourceModel struct {
	volumeBaseModel
	Timeouts datasourceTimeouts.Value `tfsdk:"timeouts"`
}

type volumesDataSourceModel struct {
	Filter   []common.FilterModel     `tfsdk:"filter"`
	Volumes  []volumeBaseModel        `tfsdk:"volumes"`
	Timeouts datasourceTimeouts.Value `tfsdk:"timeouts"`
}

var imageMetadataAttrType = map[string]attr.Type{
	"container_format": types.StringType,
	"disk_format":      types.StringType,
	"image_id":         types.StringType,
	"image_name":       types.StringType,
	"min_disk":         types.StringType,
	"os_type":          types.StringType,
	"min_ram":          types.StringType,
	"size":             types.StringType,
}

type imageMetadataModel struct {
	ContainerFormat types.String `tfsdk:"container_format"`
	DiskFormat      types.String `tfsdk:"disk_format"`
	ImageId         types.String `tfsdk:"image_id"`
	ImageName       types.String `tfsdk:"image_name"`
	MinDisk         types.String `tfsdk:"min_disk"`
	OsType          types.String `tfsdk:"os_type"`
	MinRam          types.String `tfsdk:"min_ram"`
	Size            types.String `tfsdk:"size"`
}

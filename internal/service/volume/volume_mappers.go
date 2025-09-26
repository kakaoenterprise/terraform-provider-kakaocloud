// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package volume

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kakaoenterprise/kc-sdk-go/services/volume"
	"golang.org/x/net/context"
	. "terraform-provider-kakaocloud/internal/utils"
)

func mapVolumeBaseModel(
	ctx context.Context,
	base *volumeBaseModel,
	volumeResult *volume.BcsVolumeV1ApiGetVolumeModelVolumeModel,
	respDiags *diag.Diagnostics,
) bool {
	imageMeta, diags := ConvertObjectFromModel(ctx, volumeResult.ImageMetadata, imageMetadataAttrType, func(src volume.BcsVolumeV1ApiGetVolumeModelImageMetaData) any {
		return imageMetadataModel{
			ContainerFormat: ConvertNullableString(src.ContainerFormat),
			DiskFormat:      ConvertNullableString(src.DiskFormat),
			ImageId:         ConvertNullableString(src.ImageId),
			ImageName:       ConvertNullableString(src.ImageName),
			MinDisk:         ConvertNullableString(src.MinDisk),
			OsType:          ConvertNullableString(src.OsType),
			MinRam:          ConvertNullableString(src.MinRam),
			Size:            ConvertNullableString(src.Size),
		}
	})
	respDiags.Append(diags...)

	metaMap, metaDiags := types.MapValueFrom(ctx, types.StringType, volumeResult.Metadata)
	respDiags.Append(metaDiags...)

	base.Id = types.StringValue(volumeResult.Id)
	base.Name = ConvertNullableString(volumeResult.Name)
	base.Description = ConvertNullableString(volumeResult.Description)
	base.AvailabilityZone = ConvertNullableString(volumeResult.AvailabilityZone)
	base.Status = ConvertNullableString(volumeResult.Status)
	base.MountPoint = ConvertNullableString(volumeResult.MountPoint)
	base.VolumeType = ConvertNullableString(volumeResult.VolumeType)
	base.Size = ConvertNullableInt32(volumeResult.Size)
	base.IsBootable = ConvertNullableBool(volumeResult.IsBootable)
	base.IsEncrypted = ConvertNullableBool(volumeResult.IsEncrypted)
	base.IsRoot = ConvertNullableBool(volumeResult.IsRoot)
	base.Type = ConvertNullableString(volumeResult.Type)
	base.UserId = ConvertNullableString(volumeResult.UserId)
	base.ProjectId = ConvertNullableString(volumeResult.ProjectId)
	base.AttachStatus = ConvertNullableString(volumeResult.AttachStatus)
	base.LaunchedAt = ConvertNullableTime(volumeResult.LaunchedAt)
	base.EncryptionKeyId = ConvertNullableString(volumeResult.EncryptionKeyId)
	base.PreviousStatus = ConvertNullableString(volumeResult.PreviousStatus)
	base.CreatedAt = ConvertNullableTime(volumeResult.CreatedAt)
	base.UpdatedAt = ConvertNullableTime(volumeResult.UpdatedAt)
	base.InstanceId = ConvertNullableString(volumeResult.InstanceId)
	base.InstanceName = ConvertNullableString(volumeResult.InstanceName)
	base.ImageMetadata = imageMeta
	base.Metadata = metaMap

	if respDiags.HasError() {
		return false
	}

	return true
}

func mapVolumeListModel(
	ctx context.Context,
	base *volumeBaseModel,
	volumeResult *volume.BcsVolumeV1ApiListVolumesModelVolumeModel,
	respDiags *diag.Diagnostics,
) bool {
	imageMeta, diags := ConvertObjectFromModel(ctx, volumeResult.ImageMetadata, imageMetadataAttrType, func(src volume.BcsVolumeV1ApiListVolumesModelImageMetaData) any {
		return imageMetadataModel{
			ContainerFormat: ConvertNullableString(src.ContainerFormat),
			DiskFormat:      ConvertNullableString(src.DiskFormat),
			ImageId:         ConvertNullableString(src.ImageId),
			ImageName:       ConvertNullableString(src.ImageName),
			MinDisk:         ConvertNullableString(src.MinDisk),
			OsType:          ConvertNullableString(src.OsType),
			MinRam:          ConvertNullableString(src.MinRam),
			Size:            ConvertNullableString(src.Size),
		}
	})
	respDiags.Append(diags...)

	metaMap, metaDiags := types.MapValueFrom(ctx, types.StringType, volumeResult.Metadata)
	respDiags.Append(metaDiags...)

	base.Id = types.StringValue(volumeResult.Id)
	base.Name = ConvertNullableString(volumeResult.Name)
	base.Description = ConvertNullableString(volumeResult.Description)
	base.AvailabilityZone = ConvertNullableString(volumeResult.AvailabilityZone)
	base.Status = ConvertNullableString(volumeResult.Status)
	base.MountPoint = ConvertNullableString(volumeResult.MountPoint)
	base.VolumeType = ConvertNullableString(volumeResult.VolumeType)
	base.Size = ConvertNullableInt32(volumeResult.Size)
	base.IsBootable = ConvertNullableBool(volumeResult.IsBootable)
	base.IsEncrypted = ConvertNullableBool(volumeResult.IsEncrypted)
	base.IsRoot = ConvertNullableBool(volumeResult.IsRoot)
	base.Type = ConvertNullableString(volumeResult.Type)
	base.UserId = ConvertNullableString(volumeResult.UserId)
	base.ProjectId = ConvertNullableString(volumeResult.ProjectId)
	base.AttachStatus = ConvertNullableString(volumeResult.AttachStatus)
	base.LaunchedAt = ConvertNullableTime(volumeResult.LaunchedAt)
	base.EncryptionKeyId = ConvertNullableString(volumeResult.EncryptionKeyId)
	base.PreviousStatus = ConvertNullableString(volumeResult.PreviousStatus)
	base.CreatedAt = ConvertNullableTime(volumeResult.CreatedAt)
	base.UpdatedAt = ConvertNullableTime(volumeResult.UpdatedAt)
	base.InstanceId = ConvertNullableString(volumeResult.InstanceId)
	base.InstanceName = ConvertNullableString(volumeResult.InstanceName)
	base.ImageMetadata = imageMeta
	base.Metadata = metaMap

	if respDiags.HasError() {
		return false
	}

	return true
}

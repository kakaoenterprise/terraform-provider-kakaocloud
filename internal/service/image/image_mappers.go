// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package image

import (
	"context"

	. "terraform-provider-kakaocloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kakaoenterprise/kc-sdk-go/services/image"
)

func (d *imageDataSource) mapImage(
	ctx context.Context,
	model *imageDataSourceModel,
	imageResult *image.BcsImageV1ApiGetImageModelImageModel,
	respDiags *diag.Diagnostics,
) bool {
	mapImageBaseModel(ctx, &model.imageBaseModel, imageResult, respDiags)

	if respDiags.HasError() {
		return false
	}

	return true
}

func (d *imagesDataSource) mapImages(
	ctx context.Context,
	base *imageBaseModel,
	imageResult *image.BcsImageV1ApiListImagesModelImageModel,
	respDiags *diag.Diagnostics,
) bool {
	base.Id = types.StringValue(imageResult.Id)
	base.Name = ConvertNullableString(imageResult.Name)
	base.Size = ConvertNullableInt64(imageResult.Size)
	base.Status = ConvertNullableString(imageResult.Status)
	base.Owner = ConvertNullableString(imageResult.Owner)
	base.Visibility = ConvertNullableString(imageResult.Visibility)
	base.Description = ConvertNullableString(imageResult.Description)
	base.IsShared = ConvertNullableBool(imageResult.IsShared)
	base.DiskFormat = ConvertNullableString(imageResult.DiskFormat)
	base.ContainerFormat = ConvertNullableString(imageResult.ContainerFormat)
	base.MinDisk = ConvertNullableInt32(imageResult.MinDisk)
	base.MinRam = ConvertNullableInt32(imageResult.MinRam)
	base.VirtualSize = ConvertNullableInt64(imageResult.VirtualSize)
	base.InstanceType = ConvertNullableString(imageResult.InstanceType)
	base.ImageMemberStatus = ConvertNullableString(imageResult.ImageMemberStatus)
	base.ProjectId = ConvertNullableString(imageResult.ProjectId)
	base.CreatedAt = ConvertNullableTime(imageResult.CreatedAt)
	base.UpdatedAt = ConvertNullableTime(imageResult.UpdatedAt)

	osInfoObj, osDiags := ConvertObjectFromModel(ctx, imageResult.OsInfo, osInfoAttrType,
		func(src image.BcsImageV1ApiListImagesModelOsInfoModel) any {
			return osInfoModel{
				Type:         ConvertNullableString(src.Type),
				Distro:       ConvertNullableString(src.Distro),
				Architecture: ConvertNullableString(src.Architecture),
				AdminUser:    ConvertNullableString(src.AdminUser),
				IsHidden:     ConvertNullableBool(src.IsHidden),
			}
		},
	)
	respDiags.Append(osDiags...)
	base.OsInfo = osInfoObj

	if respDiags.HasError() {
		return false
	}

	return true
}

func mapImageBaseModel(
	ctx context.Context,
	base *imageBaseModel,
	imageResult *image.BcsImageV1ApiGetImageModelImageModel,
	respDiags *diag.Diagnostics,
) {
	base.Id = types.StringValue(imageResult.Id)
	base.Name = ConvertNullableString(imageResult.Name)
	base.Size = ConvertNullableInt64(imageResult.Size)
	base.Status = ConvertNullableString(imageResult.Status)
	base.Owner = ConvertNullableString(imageResult.Owner)
	base.Visibility = ConvertNullableString(imageResult.Visibility)
	base.Description = ConvertNullableString(imageResult.Description)
	base.IsShared = ConvertNullableBool(imageResult.IsShared)
	base.DiskFormat = ConvertNullableString(imageResult.DiskFormat)
	base.ContainerFormat = ConvertNullableString(imageResult.ContainerFormat)
	base.MinDisk = ConvertNullableInt32(imageResult.MinDisk)
	base.MinRam = ConvertNullableInt32(imageResult.MinRam)
	base.VirtualSize = ConvertNullableInt64(imageResult.VirtualSize)
	base.InstanceType = ConvertNullableString(imageResult.InstanceType)
	base.ImageMemberStatus = ConvertNullableString(imageResult.ImageMemberStatus)
	base.ProjectId = ConvertNullableString(imageResult.ProjectId)
	base.CreatedAt = ConvertNullableTime(imageResult.CreatedAt)
	base.UpdatedAt = ConvertNullableTime(imageResult.UpdatedAt)

	osInfoObj, osDiags := ConvertObjectFromModel(ctx, imageResult.OsInfo, osInfoAttrType,
		func(src image.BcsImageV1ApiGetImageModelOsInfoModel) any {
			return osInfoModel{
				Type:         ConvertNullableString(src.Type),
				Distro:       ConvertNullableString(src.Distro),
				Architecture: ConvertNullableString(src.Architecture),
				AdminUser:    ConvertNullableString(src.AdminUser),
				IsHidden:     ConvertNullableBool(src.IsHidden),
			}
		},
	)
	respDiags.Append(osDiags...)
	base.OsInfo = osInfoObj
}

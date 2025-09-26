// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package bcs

import (
	. "terraform-provider-kakaocloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kakaoenterprise/kc-sdk-go/services/bcs"
	"golang.org/x/net/context"
)

func mapInstanceBaseModel(
	ctx context.Context,
	base *instanceBaseModel,
	instanceResult *bcs.BcsInstanceV1ApiGetInstanceModelInstanceModel,
	respDiags *diag.Diagnostics,
) bool {

	if instanceResult.Metadata == nil {
		base.Metadata = types.MapNull(types.StringType)
	} else {
		metaMap, metaDiags := types.MapValueFrom(ctx, types.StringType, instanceResult.Metadata)
		respDiags.Append(metaDiags...)
		base.Metadata = metaMap
	}

	base.Id = types.StringValue(instanceResult.Id)
	base.Name = ConvertNullableString(instanceResult.Name)
	base.Description = ConvertNullableString(instanceResult.Description)

	base.IsHyperThreading = ConvertNullableBool(instanceResult.IsHyperThreading)
	base.IsHadoop = ConvertNullableBool(instanceResult.IsHadoop)
	base.IsK8se = ConvertNullableBool(instanceResult.IsK8se)

	base.VmState = ConvertNullableString(instanceResult.VmState)
	base.TaskState = ConvertNullableString(instanceResult.TaskState)
	base.PowerState = ConvertNullableString(instanceResult.PowerState)
	base.Status = ConvertNullableString(instanceResult.Status)

	base.UserId = ConvertNullableString(instanceResult.UserId)
	base.ProjectId = ConvertNullableString(instanceResult.ProjectId)
	base.KeyName = ConvertNullableString(instanceResult.KeyName)
	base.Hostname = ConvertNullableString(instanceResult.Hostname)
	base.AvailabilityZone = ConvertNullableString(instanceResult.AvailabilityZone)
	base.AttachedVolumeCount = ConvertNullableInt64(instanceResult.AttachedVolumeCount)
	base.SecurityGroupCount = ConvertNullableInt64(instanceResult.SecurityGroupCount)
	base.InstanceType = ConvertNullableString(instanceResult.InstanceType)
	base.CreatedAt = ConvertNullableTime(instanceResult.CreatedAt)
	base.UpdatedAt = ConvertNullableTime(instanceResult.UpdatedAt)

	flavorObj, flavorDiags := ConvertObjectFromModel(ctx, instanceResult.Flavor, instanceFlavorAttrType, func(src bcs.BcsInstanceV1ApiGetInstanceModelInstanceFlavorModel) any {
		return instanceFlavorModel{
			Id:                        types.StringValue(src.Id),
			Name:                      ConvertNullableString(src.Name),
			Group:                     ConvertNullableString(src.Group),
			Vcpus:                     ConvertNullableInt32(src.Vcpus),
			IsBurstable:               ConvertNullableBool(src.IsBurstable),
			Manufacturer:              ConvertNullableString(src.Manufacturer),
			MemoryMb:                  ConvertNullableInt32(src.MemoryMb),
			RootGb:                    ConvertNullableInt32(src.RootGb),
			DiskType:                  ConvertNullableString(src.DiskType),
			InstanceFamily:            ConvertNullableString(src.InstanceFamily),
			OsDistro:                  ConvertNullableStringList(src.OsDistro),
			MaximumNetworkInterfaces:  ConvertNullableInt32(src.MaximumNetworkInterfaces),
			HwType:                    ConvertNullableString(src.HwType),
			HwCount:                   ConvertNullableInt32(src.HwCount),
			IsHyperThreadingSupported: ConvertNullableBool(src.IsHyperThreadingSupported),
			RealVcpus:                 ConvertNullableInt32(src.RealVcpus),
		}
	})
	respDiags.Append(flavorDiags...)
	base.Flavor = flavorObj

	imageObj, imageDiags := ConvertObjectFromModel(ctx, instanceResult.Image, instanceImageAttrType, func(src bcs.BcsInstanceV1ApiGetInstanceModelInstanceImageModel) any {
		return instanceImageModel{
			Id:             types.StringValue(src.Id),
			Name:           ConvertNullableString(src.Name),
			Description:    ConvertNullableString(src.Description),
			Owner:          ConvertNullableString(src.Owner),
			IsWindows:      ConvertNullableBool(src.IsWindows),
			Size:           ConvertNullableInt64(src.Size),
			Status:         ConvertNullableString(src.Status),
			ImageType:      ConvertNullableString(src.ImageType),
			DiskFormat:     ConvertNullableString(src.DiskFormat),
			InstanceType:   ConvertNullableString(src.InstanceType),
			MemberStatus:   ConvertNullableString(src.MemberStatus),
			MinDisk:        ConvertNullableInt32(src.MinDisk),
			MinMemory:      ConvertNullableInt32(src.MinMemory),
			OsAdmin:        ConvertNullableString(src.OsAdmin),
			OsDistro:       ConvertNullableString(src.OsDistro),
			OsType:         ConvertNullableString(src.OsType),
			OsArchitecture: ConvertNullableString(src.OsArchitecture),
			CreatedAt:      ConvertNullableTime(src.CreatedAt),
			UpdatedAt:      ConvertNullableTime(src.UpdatedAt),
		}
	})
	respDiags.Append(imageDiags...)
	base.Image = imageObj

	addressList, addressDiags := ConvertListFromModel(ctx, instanceResult.Addresses, instanceAddressesAttrType, func(src bcs.BcsInstanceV1ApiGetInstanceModelInstanceAddressModel) any {
		return instanceAddressModel{
			PrivateIp:          ConvertNullableString(src.PrivateIp),
			PublicIp:           ConvertNullableString(src.PublicIp),
			NetworkInterfaceId: ConvertNullableString(src.NetworkInterfaceId),
		}
	})
	respDiags.Append(addressDiags...)
	base.Addresses = addressList

	volumeList, volumeDiags := ConvertListFromModel(ctx, instanceResult.AttachedVolumes, instanceAttachedVolumesAttrType, func(src bcs.BcsInstanceV1ApiGetInstanceModelInstanceAttachedVolumeModel) any {
		return instanceAttachedVolumeModel{
			Id:                    types.StringValue(src.Id),
			Name:                  ConvertNullableString(src.Name),
			Status:                ConvertNullableString(src.Status),
			MountPoint:            ConvertNullableString(src.MountPoint),
			Type:                  ConvertNullableString(src.Type),
			Size:                  types.Int32Value(src.Size),
			IsDeleteOnTermination: ConvertNullableBool(src.IsDeleteOnTermination),
			CreatedAt:             ConvertNullableTime(src.CreatedAt),
			IsRoot:                ConvertNullableBool(src.IsRoot),
		}
	})
	respDiags.Append(volumeDiags...)
	base.AttachedVolumes = volumeList

	sgSet, sgDiags := ConvertSetFromModel(ctx, instanceResult.SecurityGroups, instanceSecurityGroupsAttrType, func(src bcs.BcsInstanceV1ApiGetInstanceModelInstanceSecurityGroupModel) any {
		return instanceSecurityGroupModel{
			Id:   types.StringValue(src.Id),
			Name: ConvertNullableString(src.Name),
		}
	})
	respDiags.Append(sgDiags...)
	base.SecurityGroups = sgSet

	if respDiags.HasError() {
		return false
	}

	return true
}

func mapInstanceListModel(
	ctx context.Context,
	base *instanceBaseModel,
	instanceResult *bcs.BcsInstanceV1ApiListInstancesModelInstanceModel,
	respDiags *diag.Diagnostics,
) bool {

	if instanceResult.Metadata == nil {
		base.Metadata = types.MapNull(types.StringType)
	} else {
		metaMap, metaDiags := types.MapValueFrom(ctx, types.StringType, instanceResult.Metadata)
		respDiags.Append(metaDiags...)
		base.Metadata = metaMap
	}

	base.Id = types.StringValue(instanceResult.Id)
	base.Name = ConvertNullableString(instanceResult.Name)
	base.Description = ConvertNullableString(instanceResult.Description)

	base.IsHyperThreading = ConvertNullableBool(instanceResult.IsHyperThreading)
	base.IsHadoop = ConvertNullableBool(instanceResult.IsHadoop)
	base.IsK8se = ConvertNullableBool(instanceResult.IsK8se)

	base.VmState = ConvertNullableString(instanceResult.VmState)
	base.TaskState = ConvertNullableString(instanceResult.TaskState)
	base.PowerState = ConvertNullableString(instanceResult.PowerState)
	base.Status = ConvertNullableString(instanceResult.Status)

	base.UserId = ConvertNullableString(instanceResult.UserId)
	base.ProjectId = ConvertNullableString(instanceResult.ProjectId)
	base.KeyName = ConvertNullableString(instanceResult.KeyName)
	base.Hostname = ConvertNullableString(instanceResult.Hostname)
	base.AvailabilityZone = ConvertNullableString(instanceResult.AvailabilityZone)
	base.AttachedVolumeCount = ConvertNullableInt64(instanceResult.AttachedVolumeCount)
	base.SecurityGroupCount = ConvertNullableInt64(instanceResult.SecurityGroupCount)
	base.InstanceType = ConvertNullableString(instanceResult.InstanceType)
	base.CreatedAt = ConvertNullableTime(instanceResult.CreatedAt)
	base.UpdatedAt = ConvertNullableTime(instanceResult.UpdatedAt)

	flavorObj, flavorDiags := ConvertObjectFromModel(ctx, instanceResult.Flavor, instanceFlavorAttrType, func(src bcs.BcsInstanceV1ApiListInstancesModelInstanceFlavorModel) any {
		return instanceFlavorModel{
			Id:                        types.StringValue(src.Id),
			Name:                      ConvertNullableString(src.Name),
			Group:                     ConvertNullableString(src.Group),
			Vcpus:                     ConvertNullableInt32(src.Vcpus),
			IsBurstable:               ConvertNullableBool(src.IsBurstable),
			Manufacturer:              ConvertNullableString(src.Manufacturer),
			MemoryMb:                  ConvertNullableInt32(src.MemoryMb),
			RootGb:                    ConvertNullableInt32(src.RootGb),
			DiskType:                  ConvertNullableString(src.DiskType),
			InstanceFamily:            ConvertNullableString(src.InstanceFamily),
			OsDistro:                  ConvertNullableStringList(src.OsDistro),
			MaximumNetworkInterfaces:  ConvertNullableInt32(src.MaximumNetworkInterfaces),
			HwType:                    ConvertNullableString(src.HwType),
			HwCount:                   ConvertNullableInt32(src.HwCount),
			IsHyperThreadingSupported: ConvertNullableBool(src.IsHyperThreadingSupported),
			RealVcpus:                 ConvertNullableInt32(src.RealVcpus),
		}
	})
	respDiags.Append(flavorDiags...)
	base.Flavor = flavorObj

	imageObj, imageDiags := ConvertObjectFromModel(ctx, instanceResult.Image, instanceImageAttrType, func(src bcs.BcsInstanceV1ApiListInstancesModelInstanceImageModel) any {
		return instanceImageModel{
			Id:             types.StringValue(src.Id),
			Name:           ConvertNullableString(src.Name),
			Description:    ConvertNullableString(src.Description),
			Owner:          ConvertNullableString(src.Owner),
			IsWindows:      ConvertNullableBool(src.IsWindows),
			Size:           ConvertNullableInt64(src.Size),
			Status:         ConvertNullableString(src.Status),
			ImageType:      ConvertNullableString(src.ImageType),
			DiskFormat:     ConvertNullableString(src.DiskFormat),
			InstanceType:   ConvertNullableString(src.InstanceType),
			MemberStatus:   ConvertNullableString(src.MemberStatus),
			MinDisk:        ConvertNullableInt32(src.MinDisk),
			MinMemory:      ConvertNullableInt32(src.MinMemory),
			OsAdmin:        ConvertNullableString(src.OsAdmin),
			OsDistro:       ConvertNullableString(src.OsDistro),
			OsType:         ConvertNullableString(src.OsType),
			OsArchitecture: ConvertNullableString(src.OsArchitecture),
			CreatedAt:      ConvertNullableTime(src.CreatedAt),
			UpdatedAt:      ConvertNullableTime(src.UpdatedAt),
		}
	})
	respDiags.Append(imageDiags...)
	base.Image = imageObj

	addressList, addressDiags := ConvertListFromModel(ctx, instanceResult.Addresses, instanceAddressesAttrType, func(src bcs.BcsInstanceV1ApiListInstancesModelInstanceAddressModel) any {
		return instanceAddressModel{
			PrivateIp:          ConvertNullableString(src.PrivateIp),
			PublicIp:           ConvertNullableString(src.PublicIp),
			NetworkInterfaceId: ConvertNullableString(src.NetworkInterfaceId),
		}
	})
	respDiags.Append(addressDiags...)
	base.Addresses = addressList

	volumeList, volumeDiags := ConvertListFromModel(ctx, instanceResult.AttachedVolumes, instanceAttachedVolumesAttrType, func(src bcs.BcsInstanceV1ApiListInstancesModelInstanceAttachedVolumeModel) any {
		return instanceAttachedVolumeModel{
			Id:                    types.StringValue(src.Id),
			Name:                  ConvertNullableString(src.Name),
			Status:                ConvertNullableString(src.Status),
			MountPoint:            ConvertNullableString(src.MountPoint),
			Type:                  ConvertNullableString(src.Type),
			Size:                  types.Int32Value(src.Size),
			IsDeleteOnTermination: ConvertNullableBool(src.IsDeleteOnTermination),
			CreatedAt:             ConvertNullableTime(src.CreatedAt),
			IsRoot:                ConvertNullableBool(src.IsRoot),
		}
	})
	respDiags.Append(volumeDiags...)
	base.AttachedVolumes = volumeList

	sgSet, sgDiags := ConvertSetFromModel(ctx, instanceResult.SecurityGroups, instanceSecurityGroupsAttrType, func(src bcs.BcsInstanceV1ApiListInstancesModelInstanceSecurityGroupModel) any {
		return instanceSecurityGroupModel{
			Id:   types.StringValue(src.Id),
			Name: ConvertNullableString(src.Name),
		}
	})
	respDiags.Append(sgDiags...)
	base.SecurityGroups = sgSet

	if respDiags.HasError() {
		return false
	}

	return true
}

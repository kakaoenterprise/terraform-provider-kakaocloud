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

func mapInstanceFlavorBaseModel(
	ctx context.Context,
	origin *instanceFlavorBaseModel,
	result *bcs.BcsInstanceV1ApiGetInstanceTypeModelFlavorModel,
	respDiags *diag.Diagnostics,
) bool {
	origin.Id = types.StringValue(result.Id)
	origin.Name = ConvertNullableString(result.Name)
	origin.Vcpus = ConvertNullableInt32(result.Vcpus)
	origin.Description = ConvertNullableString(result.Description)
	origin.IsBurstable = ConvertNullableBool(result.IsBurstable)
	origin.Architecture = ConvertNullableString(result.Architecture)
	origin.Manufacturer = ConvertNullableString(result.Manufacturer)
	origin.Group = ConvertNullableString(result.Group)
	origin.InstanceType = ConvertNullableString(result.InstanceType)
	origin.Processor = ConvertNullableString(result.Processor)
	origin.MemoryMb = ConvertNullableInt64(result.MemoryMb)
	origin.CreatedAt = ConvertNullableTime(result.CreatedAt)
	origin.UpdatedAt = ConvertNullableTime(result.UpdatedAt)
	origin.AvailabilityZone = ConvertNullableStringList(result.AvailabilityZone)
	origin.InstanceFamily = ConvertNullableString(result.InstanceFamily)
	origin.InstanceSize = ConvertNullableString(result.InstanceSize)
	origin.DiskType = ConvertNullableString(result.DiskType)
	origin.RootGb = ConvertNullableInt32(result.RootGb)
	origin.OsDistro = ConvertNullableString(result.OsDistro)
	origin.HwCount = ConvertNullableInt32(result.HwCount)
	origin.HwType = ConvertNullableString(result.HwType)
	origin.HwName = ConvertNullableString(result.HwName)
	origin.MaximumNetworkInterfaces = ConvertNullableInt32(result.MaximumNetworkInterfaces)
	origin.IsHyperThreadingDisabled = ConvertNullableBool(result.IsHyperThreadingDisabled)
	origin.IsHyperThreadingSupported = ConvertNullableBool(result.IsHyperThreadingSupported)
	origin.IsHyperThreadingDisableSupported = ConvertNullableBool(result.IsHyperThreadingDisableSupported)

	available, availableDiags := types.MapValueFrom(ctx, types.Int32Type, result.Available)
	respDiags.Append(availableDiags...)
	origin.Available = available

	if respDiags.HasError() {
		return false
	}
	return true
}

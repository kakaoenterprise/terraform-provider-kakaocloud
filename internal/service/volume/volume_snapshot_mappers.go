// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package volume

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/kakaoenterprise/kc-sdk-go/v2/services/volume"
	. "terraform-provider-kakaocloud/internal/utils"
)

func mapVolumeSnapshotModel(
	originSnapshot *volumeSnapshotBaseModel,
	snapshotResult *volume.Snapshot,
	respDiags *diag.Diagnostics,
) bool {
	originSnapshot.Id = ConvertNullableString(snapshotResult.Id)
	originSnapshot.Name = ConvertNullableString(snapshotResult.Name)
	originSnapshot.Description = ConvertNullableString(snapshotResult.Description)
	originSnapshot.Size = ConvertNullableInt64(snapshotResult.Size)
	originSnapshot.Status = ConvertNullableString(snapshotResult.Status)
	originSnapshot.VolumeId = ConvertNullableString(snapshotResult.VolumeId)
	originSnapshot.UserId = ConvertNullableString(snapshotResult.UserId)
	originSnapshot.ProjectId = ConvertNullableString(snapshotResult.ProjectId)
	originSnapshot.ParentId = ConvertNullableString(snapshotResult.ParentId)
	originSnapshot.CreatedAt = ConvertNullableTime(snapshotResult.CreatedAt)
	originSnapshot.UpdatedAt = ConvertNullableTime(snapshotResult.UpdatedAt)
	originSnapshot.IsIncremental = ConvertNullableBool(snapshotResult.IsIncremental)
	originSnapshot.IsDependentSnapshot = ConvertNullableBool(snapshotResult.IsDependentSnapshot)
	originSnapshot.RealSize = ConvertNullableInt64(snapshotResult.RealSize)
	originSnapshot.ScheduleId = ConvertNullableString(snapshotResult.ScheduleId)

	if respDiags.HasError() {
		return false
	}

	return true
}

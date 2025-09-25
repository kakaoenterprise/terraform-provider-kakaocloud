// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package volume

import (
	"terraform-provider-kakaocloud/internal/common"

	datasourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	resourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type volumeSnapshotBaseModel struct {
	Id                  types.String `tfsdk:"id"`
	Name                types.String `tfsdk:"name"`
	Description         types.String `tfsdk:"description"`
	Size                types.Int64  `tfsdk:"size"`
	RealSize            types.Int64  `tfsdk:"real_size"`
	Status              types.String `tfsdk:"status"`
	VolumeId            types.String `tfsdk:"volume_id"`
	ProjectId           types.String `tfsdk:"project_id"`
	ParentId            types.String `tfsdk:"parent_id"`
	UserId              types.String `tfsdk:"user_id"`
	IsIncremental       types.Bool   `tfsdk:"is_incremental"`
	IsDependentSnapshot types.Bool   `tfsdk:"is_dependent_snapshot"`
	ScheduleId          types.String `tfsdk:"schedule_id"`
	CreatedAt           types.String `tfsdk:"created_at"`
	UpdatedAt           types.String `tfsdk:"updated_at"`
}

type volumeSnapshotResourceModel struct {
	volumeSnapshotBaseModel
	Timeouts resourceTimeouts.Value `tfsdk:"timeouts"`
}

type volumeSnapshotDataSourceModel struct {
	volumeSnapshotBaseModel
	Timeouts datasourceTimeouts.Value `tfsdk:"timeouts"`
}

type volumeSnapshotsDataSourceModel struct {
	Filter          []common.FilterModel      `tfsdk:"filter"`
	VolumeSnapshots []volumeSnapshotBaseModel `tfsdk:"volume_snapshots"`
	Timeouts        datasourceTimeouts.Value  `tfsdk:"timeouts"`
}

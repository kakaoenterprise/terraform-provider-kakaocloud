// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package mysql

import (
	datasourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type backupExtraInfoModel struct {
	UseCaseSensitiveTableNames types.Bool `tfsdk:"use_case_sensitive_table_names"`
}

type backupModel struct {
	Id                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	CreatedAt         types.String `tfsdk:"created_at"`
	CreatorName       types.String `tfsdk:"creator_name"`
	Description       types.String `tfsdk:"description"`
	DiskSize          types.Int32  `tfsdk:"disk_size"`
	ExpireAt          types.String `tfsdk:"expire_at"`
	ExpiryDuration    types.Int32  `tfsdk:"expiry_duration"`
	ExtraInfo         types.Object `tfsdk:"extra_info"`
	InstanceGroupId   types.String `tfsdk:"instance_group_id"`
	InstanceGroupName types.String `tfsdk:"instance_group_name"`
	ProjectId         types.String `tfsdk:"project_id"`
	Size              types.Int64  `tfsdk:"size"`
	Status            types.String `tfsdk:"status"`
	Type              types.String `tfsdk:"type"`
	StartedAt         types.String `tfsdk:"started_at"`
	UpdatedAt         types.String `tfsdk:"updated_at"`
	EngineVersion     types.String `tfsdk:"engine_version"`
}

type backupDataSourceModel struct {
	backupModel
	Timeouts datasourceTimeouts.Value `tfsdk:"timeouts"`
}

type backupsDataSourceModel struct {
	InstanceGroupId types.String             `tfsdk:"instance_group_id"`
	Backups         types.List               `tfsdk:"backups"`
	Timeouts        datasourceTimeouts.Value `tfsdk:"timeouts"`
}

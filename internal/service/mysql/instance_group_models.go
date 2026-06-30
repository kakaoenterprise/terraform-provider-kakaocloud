// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package mysql

import (
	datasourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type subnetInfoModel struct {
	Replicas         types.Int32  `tfsdk:"replicas"`
	AvailabilityZone types.String `tfsdk:"availability_zone"`
	SubnetId         types.String `tfsdk:"subnet_id"`
}

type networkInfoModel struct {
	PrimarySubnetInfo types.Object `tfsdk:"primary_subnet_info"`
	StandbySubnetInfo types.List   `tfsdk:"standby_subnet_info"`
	SecurityGroupIds  types.Set    `tfsdk:"security_group_ids"`
}

type specContentModel struct {
	DatabaseUserName  types.String `tfsdk:"database_user_name"`
	PrimaryPort       types.Int32  `tfsdk:"primary_port"`
	StandbyPort       types.Int32  `tfsdk:"standby_port"`
	EngineVersion     types.String `tfsdk:"engine_version"`
	FlavorId          types.String `tfsdk:"flavor_id"`
	Vcpu              types.Int32  `tfsdk:"vcpu"`
	Memory            types.Int32  `tfsdk:"memory"`
	LogDiskSize       types.Int32  `tfsdk:"log_disk_size"`
	DataDiskSize      types.Int32  `tfsdk:"data_disk_size"`
	InstanceGroupType types.String `tfsdk:"instance_group_type"`
	NodeSize          types.Int32  `tfsdk:"node_size"`
}

type backupScheduleModel struct {
	Id             types.String `tfsdk:"id"`
	Type           types.String `tfsdk:"type"`
	StartTime      types.String `tfsdk:"start_time"`
	ExpiryDuration types.Int32  `tfsdk:"expiry_duration"`
	Enabled        types.Bool   `tfsdk:"enabled"`
}

type parameterGroupModel struct {
	Id                      types.String `tfsdk:"id"`
	Type                    types.String `tfsdk:"type"`
	ApplyStatus             types.String `tfsdk:"apply_status"`
	EngineVersion           types.String `tfsdk:"engine_version"`
	IsEngineVersionMismatch types.Bool   `tfsdk:"is_engine_version_mismatch"`
}

type extraInfoModel struct {
	UseCaseSensitiveTableNames types.Bool `tfsdk:"use_case_sensitive_table_names"`
}

type instanceNodeModel struct {
	InstanceId       types.String `tfsdk:"instance_id"`
	SubnetId         types.String `tfsdk:"subnet_id"`
	AvailabilityZone types.String `tfsdk:"availability_zone"`
}

type instanceGroupModel struct {
	Id             types.String `tfsdk:"id"`
	CreatedAt      types.String `tfsdk:"created_at"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
	License        types.String `tfsdk:"license"`
	Name           types.String `tfsdk:"name"`
	ProjectId      types.String `tfsdk:"project_id"`
	Description    types.String `tfsdk:"description"`
	Creator        types.String `tfsdk:"creator"`
	SourceBackupId types.String `tfsdk:"source_backup_id"`
	IsMultiAz      types.Bool   `tfsdk:"is_multi_az"`
	Endpoint       types.List   `tfsdk:"endpoint"`
	Status         types.String `tfsdk:"status"`

	NetworkInfo    types.Object `tfsdk:"network_info"`
	SpecContent    types.Object `tfsdk:"spec_content"`
	BackupSchedule types.Object `tfsdk:"backup_schedule"`
	ParameterGroup types.Object `tfsdk:"parameter_group"`
	ExtraInfo      types.Object `tfsdk:"extra_info"`
	Instances      types.Object `tfsdk:"instances"`
}

type instanceGroupListNodeModel struct {
	InstanceId types.String `tfsdk:"instance_id"`
}

type instanceGroupListModel struct {
	Id             types.String `tfsdk:"id"`
	CreatedAt      types.String `tfsdk:"created_at"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
	License        types.String `tfsdk:"license"`
	Name           types.String `tfsdk:"name"`
	ProjectId      types.String `tfsdk:"project_id"`
	Description    types.String `tfsdk:"description"`
	Creator        types.String `tfsdk:"creator"`
	SourceBackupId types.String `tfsdk:"source_backup_id"`
	IsMultiAz      types.Bool   `tfsdk:"is_multi_az"`
	Endpoint       types.List   `tfsdk:"endpoint"`
	Status         types.String `tfsdk:"status"`

	NetworkInfo    types.Object `tfsdk:"network_info"`
	SpecContent    types.Object `tfsdk:"spec_content"`
	BackupSchedule types.Object `tfsdk:"backup_schedule"`
	ParameterGroup types.Object `tfsdk:"parameter_group"`
	ExtraInfo      types.Object `tfsdk:"extra_info"`
	Instances      types.Object `tfsdk:"instances"`
}

type instanceGroupDataSourceModel struct {
	instanceGroupModel
	Timeouts datasourceTimeouts.Value `tfsdk:"timeouts"`
}

type instanceGroupsDataSourceModel struct {
	InstanceGroups types.List               `tfsdk:"instance_groups"`
	Timeouts       datasourceTimeouts.Value `tfsdk:"timeouts"`
}

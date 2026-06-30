// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package mysql

import (
	resourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var instanceGroupResourceSpecContentAttrTypes = map[string]attr.Type{
	"database_user_name":     types.StringType,
	"database_user_password": types.StringType,
	"primary_port":           types.Int32Type,
	"standby_port":           types.Int32Type,
	"engine_version":         types.StringType,
	"flavor_id":              types.StringType,
	"vcpu":                   types.Int32Type,
	"memory":                 types.Int32Type,
	"log_disk_size":          types.Int32Type,
	"data_disk_size":         types.Int32Type,
	"instance_group_type":    types.StringType,
	"node_size":              types.Int32Type,
}

type instanceGroupSpecContentResourceModel struct {
	DatabaseUserName     types.String `tfsdk:"database_user_name"`
	DatabaseUserPassword types.String `tfsdk:"database_user_password"`
	PrimaryPort          types.Int32  `tfsdk:"primary_port"`
	StandbyPort          types.Int32  `tfsdk:"standby_port"`
	EngineVersion        types.String `tfsdk:"engine_version"`
	FlavorId             types.String `tfsdk:"flavor_id"`
	Vcpu                 types.Int32  `tfsdk:"vcpu"`
	Memory               types.Int32  `tfsdk:"memory"`
	LogDiskSize          types.Int32  `tfsdk:"log_disk_size"`
	DataDiskSize         types.Int32  `tfsdk:"data_disk_size"`
	InstanceGroupType    types.String `tfsdk:"instance_group_type"`
	NodeSize             types.Int32  `tfsdk:"node_size"`
}

var instanceGroupResourceDesiredSubnetInfoAttrTypes = map[string]attr.Type{
	"replicas":  types.Int32Type,
	"subnet_id": types.StringType,
}

type instanceGroupResourceDesiredSubnetInfoModel struct {
	Replicas types.Int32  `tfsdk:"replicas"`
	SubnetId types.String `tfsdk:"subnet_id"`
}

var instanceGroupResourceDesiredNetworkInfoAttrTypes = map[string]attr.Type{
	"primary_subnet_info": types.ObjectType{AttrTypes: instanceGroupResourceDesiredSubnetInfoAttrTypes},
	"standby_subnet_info": types.SetType{ElemType: types.ObjectType{AttrTypes: instanceGroupResourceDesiredSubnetInfoAttrTypes}},
	"security_group_ids":  types.SetType{ElemType: types.StringType},
}

type instanceGroupResourceDesiredNetworkInfoModel struct {
	PrimarySubnetInfo types.Object `tfsdk:"primary_subnet_info"`
	StandbySubnetInfo types.Set    `tfsdk:"standby_subnet_info"`
	SecurityGroupIds  types.Set    `tfsdk:"security_group_ids"`
}

var instanceGroupResourceSubnetInfoAttrTypes = map[string]attr.Type{
	"replicas":          types.Int32Type,
	"availability_zone": types.StringType,
	"subnet_id":         types.StringType,
}

type instanceGroupResourceSubnetInfoModel struct {
	Replicas         types.Int32  `tfsdk:"replicas"`
	AvailabilityZone types.String `tfsdk:"availability_zone"`
	SubnetId         types.String `tfsdk:"subnet_id"`
}

var instanceGroupResourceNetworkInfoAttrTypes = map[string]attr.Type{
	"primary_subnet_info": types.ObjectType{AttrTypes: instanceGroupResourceSubnetInfoAttrTypes},
	"standby_subnet_info": types.SetType{ElemType: types.ObjectType{AttrTypes: instanceGroupResourceSubnetInfoAttrTypes}},
	"security_group_ids":  types.SetType{ElemType: types.StringType},
}

type instanceGroupResourceNetworkInfoModel struct {
	PrimarySubnetInfo types.Object `tfsdk:"primary_subnet_info"`
	StandbySubnetInfo types.Set    `tfsdk:"standby_subnet_info"`
	SecurityGroupIds  types.Set    `tfsdk:"security_group_ids"`
}

type instanceGroupRestoreSourceResourceModel struct {
	Type types.String `tfsdk:"type"`
	Id   types.String `tfsdk:"id"`
	Time types.String `tfsdk:"time"`
}

type instanceGroupSecurityGroupsModel struct {
	Id               types.String `tfsdk:"id"`
	InstanceGroupId  types.String `tfsdk:"instance_group_id"`
	SecurityGroupIds types.Set    `tfsdk:"security_group_ids"`
}

type instanceGroupExtendVolumeModel struct {
	Id                  types.String `tfsdk:"id"`
	InstanceGroupId     types.String `tfsdk:"instance_group_id"`
	LogDiskSize         types.Int32  `tfsdk:"log_disk_size"`
	DataDiskSize        types.Int32  `tfsdk:"data_disk_size"`
	InstanceGroupStatus types.String `tfsdk:"instance_group_status"`
}

type instanceGroupBackupScheduleModel struct {
	Id               types.String `tfsdk:"id"`
	InstanceGroupId  types.String `tfsdk:"instance_group_id"`
	BackupScheduleId types.String `tfsdk:"backup_schedule_id"`
	Type             types.String `tfsdk:"type"`
	StartTime        types.String `tfsdk:"start_time"`
	ExpiryDuration   types.Int32  `tfsdk:"expiry_duration"`
	Enabled          types.Bool   `tfsdk:"enabled"`
}

type instanceGroupParameterGroupModel struct {
	Id                      types.String `tfsdk:"id"`
	InstanceGroupId         types.String `tfsdk:"instance_group_id"`
	ParameterGroupId        types.String `tfsdk:"parameter_group_id"`
	ParameterGroupType      types.String `tfsdk:"parameter_group_type"`
	ApplyStatus             types.String `tfsdk:"apply_status"`
	EngineVersion           types.String `tfsdk:"engine_version"`
	IsEngineVersionMismatch types.Bool   `tfsdk:"is_engine_version_mismatch"`
}

type instanceGroupResourceModel struct {
	Id                 types.String           `tfsdk:"id"`
	CreatedAt          types.String           `tfsdk:"created_at"`
	UpdatedAt          types.String           `tfsdk:"updated_at"`
	License            types.String           `tfsdk:"license"`
	Name               types.String           `tfsdk:"name"`
	ProjectId          types.String           `tfsdk:"project_id"`
	Description        types.String           `tfsdk:"description"`
	Creator            types.String           `tfsdk:"creator"`
	SourceBackupId     types.String           `tfsdk:"source_backup_id"`
	IsMultiAz          types.Bool             `tfsdk:"is_multi_az"`
	Endpoint           types.List             `tfsdk:"endpoint"`
	Status             types.String           `tfsdk:"status"`
	NetworkInfo        types.Object           `tfsdk:"network_info"`
	DesiredNetworkInfo types.Object           `tfsdk:"desired_network_info"`
	SpecContent        types.Object           `tfsdk:"spec_content"`
	Source             types.Object           `tfsdk:"source"`
	BackupSchedule     types.Object           `tfsdk:"backup_schedule"`
	ParameterGroup     types.Object           `tfsdk:"parameter_group"`
	ExtraInfo          types.Object           `tfsdk:"extra_info"`
	Instances          types.Object           `tfsdk:"instances"`
	Timeouts           resourceTimeouts.Value `tfsdk:"timeouts"`
}

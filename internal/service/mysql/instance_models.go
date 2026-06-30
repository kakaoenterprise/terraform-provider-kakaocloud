// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package mysql

import (
	datasourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type networkPortModel struct {
	SubnetId         types.String `tfsdk:"subnet_id"`
	SecurityGroupIds types.Set    `tfsdk:"security_group_ids"`
}

type instanceStatusContentModel struct {
	NeedsRestart       types.Bool `tfsdk:"needs_restart"`
	NeedsRestartReason types.List `tfsdk:"needs_restart_reason"`
}

type instanceSpecContentModel struct {
	AvailabilityZone types.String `tfsdk:"availability_zone"`
	FlavorId         types.String `tfsdk:"flavor_id"`
	DataDiskSize     types.Int32  `tfsdk:"data_disk_size"`
	LogDiskSize      types.Int32  `tfsdk:"log_disk_size"`
	EngineVersion    types.String `tfsdk:"engine_version"`
	NetworkPorts     types.List   `tfsdk:"network_ports"`
}

type instanceModel struct {
	Id                 types.String `tfsdk:"id"`
	ProjectId          types.String `tfsdk:"project_id"`
	InstanceGroupId    types.String `tfsdk:"instance_group_id"`
	InstanceGroupName  types.String `tfsdk:"instance_group_name"`
	Name               types.String `tfsdk:"name"`
	Status             types.String `tfsdk:"status"`
	AvailabilityStatus types.String `tfsdk:"availability_status"`
	StatusContent      types.Object `tfsdk:"status_content"`
	Role               types.String `tfsdk:"role"`
	DataDiskUsage      types.Int32  `tfsdk:"data_disk_usage"`
	LogDiskUsage       types.Int32  `tfsdk:"log_disk_usage"`
	SpecContent        types.Object `tfsdk:"spec_content"`
	CreatedAt          types.String `tfsdk:"created_at"`
	UpdatedAt          types.String `tfsdk:"updated_at"`
	StartTime          types.String `tfsdk:"start_time"`
}

type instancesDataSourceModel struct {
	InstanceGroupId types.String             `tfsdk:"instance_group_id"`
	Instances       types.List               `tfsdk:"instances"`
	Timeouts        datasourceTimeouts.Value `tfsdk:"timeouts"`
}

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package kubernetesengine

import (
	datasourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	resourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type scheduledScalingBaseModel struct {
	ClusterName  types.String `tfsdk:"cluster_name"`
	NodePoolName types.String `tfsdk:"node_pool_name"`
	CreatedAt    types.String `tfsdk:"created_at"`
	DesiredNodes types.Int32  `tfsdk:"desired_nodes"`
	Name         types.String `tfsdk:"name"`
	Schedule     types.String `tfsdk:"schedule"`
	ScheduleType types.String `tfsdk:"schedule_type"`
	StartTime    types.String `tfsdk:"start_time"`
	Status       types.Object `tfsdk:"status"`
}

type scheduledScalingStatusModel struct {
	Histories types.List `tfsdk:"histories"`
}

var scheduledScalingHistoryAttrTypes = map[string]attr.Type{
	"description":   types.StringType,
	"occurred_time": types.StringType,
	"state":         types.StringType,
}

var scheduledScalingStatusAttrTypes = map[string]attr.Type{
	"histories": types.ListType{
		ElemType: types.ObjectType{AttrTypes: scheduledScalingHistoryAttrTypes},
	},
}

type scheduledScalingsDataSourceModel struct {
	ScheduledScaling []scheduledScalingBaseModel `tfsdk:"scheduled_scaling"`
	ClusterName      types.String                `tfsdk:"cluster_name"`
	NodePoolName     types.String                `tfsdk:"node_pool_name"`
	Timeouts         datasourceTimeouts.Value    `tfsdk:"timeouts"`
}

type scheduledScalingResourceModel struct {
	scheduledScalingBaseModel
	Timeouts resourceTimeouts.Value `tfsdk:"timeouts"`
}

type scheduledScalingDataSourceModel struct {
	scheduledScalingBaseModel
	Timeouts datasourceTimeouts.Value `tfsdk:"timeouts"`
}

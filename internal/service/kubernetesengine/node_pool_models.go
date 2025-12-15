// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package kubernetesengine

import (
	datasourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	resourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type NodePoolVpcInfoModelSet struct {
	Id      types.String `tfsdk:"id"`
	Subnets types.Set    `tfsdk:"subnets"`
}

type NodePoolVpcInfoModelList struct {
	Id      types.String `tfsdk:"id"`
	Subnets types.Set    `tfsdk:"subnets"`
}

type NodePoolLabelModel struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

type NodePoolTaintModel struct {
	Key    types.String `tfsdk:"key"`
	Value  types.String `tfsdk:"value"`
	Effect types.String `tfsdk:"effect"`
}

type NodePoolBaseModel struct {
	Id          types.String `tfsdk:"id"`
	ClusterName types.String `tfsdk:"cluster_name"`
	Name        types.String `tfsdk:"name"`

	Description      types.String `tfsdk:"description"`
	FlavorId         types.String `tfsdk:"flavor_id"`
	VolumeSize       types.Int32  `tfsdk:"volume_size"`
	NodeCount        types.Int32  `tfsdk:"node_count"`
	SshKeyName       types.String `tfsdk:"ssh_key_name"`
	UserData         types.String `tfsdk:"user_data"`
	VpcInfo          types.Object `tfsdk:"vpc_info"`
	IsHyperThreading types.Bool   `tfsdk:"is_hyper_threading"`

	SecurityGroups types.Set `tfsdk:"security_groups"`
	Labels         types.Set `tfsdk:"labels"`
	Taints         types.Set `tfsdk:"taints"`

	CreatedAt      types.String `tfsdk:"created_at"`
	FailureMessage types.String `tfsdk:"failure_message"`
	IsGpu          types.Bool   `tfsdk:"is_gpu"`
	IsBareMetal    types.Bool   `tfsdk:"is_bare_metal"`
	IsUpgradable   types.Bool   `tfsdk:"is_upgradable"`
	Flavor         types.String `tfsdk:"flavor"`
	Status         types.Object `tfsdk:"status"`
	Image          types.Object `tfsdk:"image"`
	Version        types.String `tfsdk:"version"`
	IsCordon       types.Bool   `tfsdk:"is_cordon"`
	Autoscaling    types.Object `tfsdk:"autoscaling"`
}

type NodePoolResourceModel struct {
	NodePoolBaseModel
	ImageId               types.String           `tfsdk:"image_id"`
	RequestNodeCount      types.Int32            `tfsdk:"request_node_count"`
	MinorVersion          types.String           `tfsdk:"minor_version"`
	RequestSecurityGroups types.Set              `tfsdk:"request_security_groups"`
	Timeouts              resourceTimeouts.Value `tfsdk:"timeouts"`
}

type nodePoolDataSourceModel struct {
	NodePoolBaseModel
	Timeouts datasourceTimeouts.Value `tfsdk:"timeouts"`
}

type nodePoolsDataSourceModel struct {
	ClusterName types.String             `tfsdk:"cluster_name"`
	NodePools   []NodePoolBaseModel      `tfsdk:"node_pools"`
	Timeouts    datasourceTimeouts.Value `tfsdk:"timeouts"`
}

var nodePoolLabelAttrTypes = map[string]attr.Type{
	"key":   types.StringType,
	"value": types.StringType,
}

var nodePoolTaintAttrTypes = map[string]attr.Type{
	"key":    types.StringType,
	"value":  types.StringType,
	"effect": types.StringType,
}

var nodePoolStatusAttrTypes = map[string]attr.Type{
	"phase":             types.StringType,
	"available_nodes":   types.Int32Type,
	"unavailable_nodes": types.Int32Type,
}

type nodePoolStatusModel struct {
	Phase            types.String `tfsdk:"phase"`
	AvailableNodes   types.Int32  `tfsdk:"available_nodes"`
	UnavailableNodes types.Int32  `tfsdk:"unavailable_nodes"`
}

type imageInfoModel struct {
	Id            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Architecture  types.String `tfsdk:"architecture"`
	IsGpuType     types.Bool   `tfsdk:"is_gpu_type"`
	InstanceType  types.String `tfsdk:"instance_type"`
	KernelVersion types.String `tfsdk:"kernel_version"`
	KeyPackage    types.String `tfsdk:"key_package"`
	OsDistro      types.String `tfsdk:"os_distro"`
	OsType        types.String `tfsdk:"os_type"`
	OsVersion     types.String `tfsdk:"os_version"`
}

var imageInfoAttrTypes = map[string]attr.Type{
	"id":             types.StringType,
	"name":           types.StringType,
	"architecture":   types.StringType,
	"is_gpu_type":    types.BoolType,
	"instance_type":  types.StringType,
	"kernel_version": types.StringType,
	"key_package":    types.StringType,
	"os_distro":      types.StringType,
	"os_type":        types.StringType,
	"os_version":     types.StringType,
}

var nodePoolAutoscalingAttrTypes = map[string]attr.Type{
	"is_autoscaler_enable":                types.BoolType,
	"autoscaler_desired_node_count":       types.Int32Type,
	"autoscaler_max_node_count":           types.Int32Type,
	"autoscaler_min_node_count":           types.Int32Type,
	"autoscaler_scale_down_threshold":     types.Float32Type,
	"autoscaler_scale_down_unneeded_time": types.Int32Type,
	"autoscaler_scale_down_unready_time":  types.Int32Type,
}

type NodePoolAutoscalingModel struct {
	IsAutoscalerEnable              types.Bool    `tfsdk:"is_autoscaler_enable"`
	AutoscalerDesiredNodeCount      types.Int32   `tfsdk:"autoscaler_desired_node_count"`
	AutoscalerMaxNodeCount          types.Int32   `tfsdk:"autoscaler_max_node_count"`
	AutoscalerMinNodeCount          types.Int32   `tfsdk:"autoscaler_min_node_count"`
	AutoscalerScaleDownThreshold    types.Float32 `tfsdk:"autoscaler_scale_down_threshold"`
	AutoscalerScaleDownUnneededTime types.Int32   `tfsdk:"autoscaler_scale_down_unneeded_time"`
	AutoscalerScaleDownUnreadyTime  types.Int32   `tfsdk:"autoscaler_scale_down_unready_time"`
}

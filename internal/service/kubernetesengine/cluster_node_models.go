// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package kubernetesengine

import (
	datasourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	resourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type nodeBaseModel struct {
	IsCordon       types.Bool   `tfsdk:"is_cordon"`
	CreatedAt      types.String `tfsdk:"created_at"`
	Flavor         types.String `tfsdk:"flavor"`
	Id             types.String `tfsdk:"id"`
	Ip             types.String `tfsdk:"ip"`
	Name           types.String `tfsdk:"name"`
	NodePoolName   types.String `tfsdk:"node_pool_name"`
	SshKeyName     types.String `tfsdk:"ssh_key_name"`
	FailureMessage types.String `tfsdk:"failure_message"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
	Version        types.String `tfsdk:"version"`
	VolumeSize     types.Int32  `tfsdk:"volume_size"`

	IsHyperThreading types.Bool   `tfsdk:"is_hyper_threading"`
	Image            types.Object `tfsdk:"image"`
	Status           types.Object `tfsdk:"status"`
	VpcInfo          types.Object `tfsdk:"vpc_info"`
}

type nodeImageInfoModel struct {
	Architecture  types.String `tfsdk:"architecture"`
	IsGpuType     types.Bool   `tfsdk:"is_gpu_type"`
	Id            types.String `tfsdk:"id"`
	InstanceType  types.String `tfsdk:"instance_type"`
	KernelVersion types.String `tfsdk:"kernel_version"`
	KeyPackage    types.String `tfsdk:"key_package"`
	Name          types.String `tfsdk:"name"`
	OsDistro      types.String `tfsdk:"os_distro"`
	OsType        types.String `tfsdk:"os_type"`
	OsVersion     types.String `tfsdk:"os_version"`
}

type nodeStatusModel struct {
	Phase types.String `tfsdk:"phase"`
}

type nodeVpcInfoModel struct {
	Id      types.String `tfsdk:"id"`
	Subnets types.Set    `tfsdk:"subnets"`
}

type nodeSubnetModel struct {
	AvailabilityZone types.String `tfsdk:"availability_zone"`
	CidrBlock        types.String `tfsdk:"cidr_block"`
	SubnetId         types.String `tfsdk:"id"`
}

var nodeSubnetAttrTypes = map[string]attr.Type{
	"availability_zone": types.StringType,
	"cidr_block":        types.StringType,
	"id":                types.StringType,
}

var nodeVpcInfoAttrTypes = map[string]attr.Type{
	"id": types.StringType,
	"subnets": types.SetType{
		ElemType: types.ObjectType{AttrTypes: nodeSubnetAttrTypes},
	},
}

var nodeImageInfoAttrTypes = map[string]attr.Type{
	"architecture":   types.StringType,
	"is_gpu_type":    types.BoolType,
	"id":             types.StringType,
	"instance_type":  types.StringType,
	"kernel_version": types.StringType,
	"key_package":    types.StringType,
	"name":           types.StringType,
	"os_distro":      types.StringType,
	"os_type":        types.StringType,
	"os_version":     types.StringType,
}

var nodeStatusAttrTypes = map[string]attr.Type{
	"phase": types.StringType,
}

var nodeBaseAttrTypes = map[string]attr.Type{
	"is_cordon":          types.BoolType,
	"created_at":         types.StringType,
	"flavor":             types.StringType,
	"id":                 types.StringType,
	"ip":                 types.StringType,
	"name":               types.StringType,
	"node_pool_name":     types.StringType,
	"ssh_key_name":       types.StringType,
	"failure_message":    types.StringType,
	"updated_at":         types.StringType,
	"version":            types.StringType,
	"volume_size":        types.Int32Type,
	"is_hyper_threading": types.BoolType,

	"image":    types.ObjectType{AttrTypes: nodeImageInfoAttrTypes},
	"status":   types.ObjectType{AttrTypes: nodeStatusAttrTypes},
	"vpc_info": types.ObjectType{AttrTypes: nodeVpcInfoAttrTypes},
}

type clusterNodeDataSourceModel struct {
	Nodes        []nodeBaseModel          `tfsdk:"nodes"`
	ClusterName  types.String             `tfsdk:"cluster_name"`
	NodePoolName types.String             `tfsdk:"node_pool_name"`
	Timeouts     datasourceTimeouts.Value `tfsdk:"timeouts"`
}

type clusterNodeResourceModel struct {
	Id          types.String           `tfsdk:"id"`
	ClusterName types.String           `tfsdk:"cluster_name"`
	NodeNames   types.Set              `tfsdk:"node_names"`
	IsRemove    types.Bool             `tfsdk:"is_remove"`
	IsCordon    types.Bool             `tfsdk:"is_cordon"`
	Timeouts    resourceTimeouts.Value `tfsdk:"timeouts"`
}

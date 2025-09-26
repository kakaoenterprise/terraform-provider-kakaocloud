// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package kubernetesengine

import (
	datasourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	resourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ClusterBaseModel struct {
	IsAllocateFip        types.Bool   `tfsdk:"is_allocate_fip"`
	ApiVersion           types.String `tfsdk:"api_version"`
	Network              types.Object `tfsdk:"network"`
	ControlPlaneEndpoint types.Object `tfsdk:"control_plane_endpoint"`
	CreatedAt            types.String `tfsdk:"created_at"`
	CreatorInfo          types.Object `tfsdk:"creator_info"`
	Version              types.Object `tfsdk:"version"`
	Description          types.String `tfsdk:"description"`
	Id                   types.String `tfsdk:"id"`
	Name                 types.String `tfsdk:"name"`
	Status               types.Object `tfsdk:"status"`
	FailureMessage       types.String `tfsdk:"failure_message"`
	IsUpgradable         types.Bool   `tfsdk:"is_upgradable"`
	VpcInfo              types.Object `tfsdk:"vpc_info"`
}

type ClusterNetworkModel struct {
	Cni         types.String `tfsdk:"cni"`
	PodCidr     types.String `tfsdk:"pod_cidr"`
	ServiceCidr types.String `tfsdk:"service_cidr"`
}

type ControlPlaneEndpointModel struct {
	Host types.String `tfsdk:"host"`
	Port types.Int32  `tfsdk:"port"`
}

type CreatorInfoModel struct {
	Id   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type OmtInfoModel struct {
	IsDeprecated types.Bool   `tfsdk:"is_deprecated"`
	Eol          types.String `tfsdk:"eol"`
	MinorVersion types.String `tfsdk:"minor_version"`
	NextVersion  types.String `tfsdk:"next_version"`
	PatchVersion types.String `tfsdk:"patch_version"`
}

type StatusModel struct {
	Phase types.String `tfsdk:"phase"`
}

type VpcInfoModel struct {
	Id      types.String `tfsdk:"id"`
	Subnets types.Set    `tfsdk:"subnets"`
}

type SubnetModel struct {
	AvailabilityZone types.String `tfsdk:"availability_zone"`
	CidrBlock        types.String `tfsdk:"cidr_block"`
	Id               types.String `tfsdk:"id"`
}

var clusterNetworkAttrTypes = map[string]attr.Type{
	"cni":          types.StringType,
	"pod_cidr":     types.StringType,
	"service_cidr": types.StringType,
}

var subnetAttrTypes = map[string]attr.Type{
	"availability_zone": types.StringType,
	"cidr_block":        types.StringType,
	"id":                types.StringType,
}

var vpcInfoAttrTypes = map[string]attr.Type{
	"id": types.StringType,
	"subnets": types.SetType{
		ElemType: types.ObjectType{AttrTypes: subnetAttrTypes},
	},
}

var controlPlaneEndpointAttrTypes = map[string]attr.Type{
	"host": types.StringType,
	"port": types.Int32Type,
}

var creatorInfoAttrTypes = map[string]attr.Type{
	"id":   types.StringType,
	"name": types.StringType,
}

var omtInfoAttrTypes = map[string]attr.Type{
	"is_deprecated": types.BoolType,
	"eol":           types.StringType,
	"minor_version": types.StringType,
	"next_version":  types.StringType,
	"patch_version": types.StringType,
}

var statusAttrTypes = map[string]attr.Type{
	"phase": types.StringType,
}

type clusterDataSourceModel struct {
	ClusterBaseModel
	Timeouts datasourceTimeouts.Value `tfsdk:"timeouts"`
}

type clustersDataSourceModel struct {
	Clusters []ClusterBaseModel       `tfsdk:"clusters"`
	Timeouts datasourceTimeouts.Value `tfsdk:"timeouts"`
}

type clusterResourceModel struct {
	ClusterBaseModel
	Timeouts resourceTimeouts.Value `tfsdk:"timeouts"`
}

type targetVpcInfoModel struct {
	Id      types.String `tfsdk:"id"`
	Subnets types.Set    `tfsdk:"subnets"`
}

type targetNetworkModel struct {
	Cni         types.String `tfsdk:"cni"`
	ServiceCidr types.String `tfsdk:"service_cidr"`
	PodCidr     types.String `tfsdk:"pod_cidr"`
}

var targetVpcInfoAttrTypes = map[string]attr.Type{
	"id":      types.StringType,
	"subnets": types.SetType{ElemType: types.StringType},
}

var targetNetworkAttrTypes = map[string]attr.Type{
	"cni":          types.StringType,
	"service_cidr": types.StringType,
	"pod_cidr":     types.StringType,
}

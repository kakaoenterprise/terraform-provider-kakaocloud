// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package bcs

import (
	"terraform-provider-kakaocloud/internal/common"

	datasourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type instanceFlavorBaseModel struct {
	Id                               types.String `tfsdk:"id"`
	Name                             types.String `tfsdk:"name"`
	Vcpus                            types.Int32  `tfsdk:"vcpus"`
	Description                      types.String `tfsdk:"description"`
	IsBurstable                      types.Bool   `tfsdk:"is_burstable"`
	Architecture                     types.String `tfsdk:"architecture"`
	Manufacturer                     types.String `tfsdk:"manufacturer"`
	Group                            types.String `tfsdk:"group"`
	InstanceType                     types.String `tfsdk:"instance_type"`
	Processor                        types.String `tfsdk:"processor"`
	MemoryMb                         types.Int64  `tfsdk:"memory_mb"`
	CreatedAt                        types.String `tfsdk:"created_at"`
	UpdatedAt                        types.String `tfsdk:"updated_at"`
	AvailabilityZone                 types.List   `tfsdk:"availability_zone"`
	Available                        types.Map    `tfsdk:"available"`
	InstanceFamily                   types.String `tfsdk:"instance_family"`
	InstanceSize                     types.String `tfsdk:"instance_size"`
	DiskType                         types.String `tfsdk:"disk_type"`
	RootGb                           types.Int32  `tfsdk:"root_gb"`
	OsDistro                         types.String `tfsdk:"os_distro"`
	HwCount                          types.Int32  `tfsdk:"hw_count"`
	HwType                           types.String `tfsdk:"hw_type"`
	HwName                           types.String `tfsdk:"hw_name"`
	MaximumNetworkInterfaces         types.Int32  `tfsdk:"maximum_network_interfaces"`
	IsHyperThreadingDisabled         types.Bool   `tfsdk:"is_hyper_threading_disabled"`
	IsHyperThreadingSupported        types.Bool   `tfsdk:"is_hyper_threading_supported"`
	IsHyperThreadingDisableSupported types.Bool   `tfsdk:"is_hyper_threading_disable_supported"`
}

type instanceFlavorDataSourceModel struct {
	instanceFlavorBaseModel
	Timeouts datasourceTimeouts.Value `tfsdk:"timeouts"`
}

type instanceFlavorsDataSourceModel struct {
	Filter          []common.FilterModel      `tfsdk:"filter"`
	InstanceFlavors []instanceFlavorBaseModel `tfsdk:"instance_flavors"`
	Timeouts        datasourceTimeouts.Value  `tfsdk:"timeouts"`
}

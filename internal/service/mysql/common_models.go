// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package mysql

import (
	datasourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type flavorModel struct {
	Id                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	Type              types.String `tfsdk:"type"`
	Vcpus             types.Int32  `tfsdk:"vcpus"`
	Memory            types.Int32  `tfsdk:"memory"`
	MemoryMb          types.Int32  `tfsdk:"memory_mb"`
	Group             types.String `tfsdk:"group"`
	Family            types.String `tfsdk:"family"`
	AvailabilityZones types.List   `tfsdk:"availability_zones"`
	Deprecated        types.Bool   `tfsdk:"deprecated"`
}

type flavorsDataSourceModel struct {
	ShowAll  types.Bool               `tfsdk:"show_all"`
	Flavors  []flavorModel            `tfsdk:"flavors"`
	Timeouts datasourceTimeouts.Value `tfsdk:"timeouts"`
}

type engineVersionModel struct {
	EngineVersion types.String `tfsdk:"engine_version"`
	License       types.String `tfsdk:"license"`
}

type engineVersionsDataSourceModel struct {
	EngineVersions []engineVersionModel     `tfsdk:"engine_versions"`
	Timeouts       datasourceTimeouts.Value `tfsdk:"timeouts"`
}

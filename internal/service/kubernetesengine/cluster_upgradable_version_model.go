// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package kubernetesengine

import (
	datasourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var versionAttrTypes = map[string]attr.Type{
	"is_deprecated": types.BoolType,
	"eol":           types.StringType,
	"minor_version": types.StringType,
	"next_version":  types.StringType,
	"patch_version": types.StringType,
}

type upgradableVersionsDataSourceModel struct {
	ClusterName types.String             `tfsdk:"cluster_name"`
	Current     types.Object             `tfsdk:"current"`
	Upgradable  []versionBaseModel       `tfsdk:"upgradable"`
	Timeouts    datasourceTimeouts.Value `tfsdk:"timeouts"`
}

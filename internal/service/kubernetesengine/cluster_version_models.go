// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kubernetesengine

import (
	datasourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type versionBaseModel struct {
	IsDeprecated types.Bool   `tfsdk:"is_deprecated"`
	MinorVersion types.String `tfsdk:"minor_version"`
	PatchVersion types.String `tfsdk:"patch_version"`
	Eol          types.String `tfsdk:"eol"`
	NextVersion  types.String `tfsdk:"next_version"`
}

type versionDataSourceModel struct {
	Versions []versionBaseModel       `tfsdk:"versions"`
	Timeouts datasourceTimeouts.Value `tfsdk:"timeouts"`
}

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package volume

import (
	datasourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type volumeTypeModel struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}

type volumeTypesDataSourceModel struct {
	VolumeTypes []volumeTypeModel        `tfsdk:"volume_types"`
	Timeouts    datasourceTimeouts.Value `tfsdk:"timeouts"`
}

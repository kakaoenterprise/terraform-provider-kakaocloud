// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpc

import (
	datasourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	resourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type subnetShareBaseModel struct {
	Id       types.String `tfsdk:"id"`
	Projects types.List   `tfsdk:"projects"`
}

type subnetShareResourceModel struct {
	subnetShareBaseModel
	ProjectIds types.Set              `tfsdk:"project_ids"`
	Timeouts   resourceTimeouts.Value `tfsdk:"timeouts"`
}

type subnetShareDataSourceModel struct {
	subnetShareBaseModel
	Timeouts datasourceTimeouts.Value `tfsdk:"timeouts"`
}

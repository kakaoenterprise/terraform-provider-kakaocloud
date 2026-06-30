// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package mysql

import (
	datasourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type restorableTimeModel struct {
	FromTime types.String `tfsdk:"from_time"`
	ToTime   types.String `tfsdk:"to_time"`
}

type instanceGroupRestorableTimeDataSourceModel struct {
	InstanceGroupId types.String             `tfsdk:"instance_group_id"`
	RestorableTime  types.Object             `tfsdk:"restorable_time"`
	Timeouts        datasourceTimeouts.Value `tfsdk:"timeouts"`
}

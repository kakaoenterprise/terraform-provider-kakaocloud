// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package loadbalancer

import (
	datasourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type loadBalancerFlavorBaseModel struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	IsEnabled   types.Bool   `tfsdk:"is_enabled"`
}

type loadBalancerFlavorsDataSourceModel struct {
	LoadBalancerFlavors []loadBalancerFlavorBaseModel `tfsdk:"flavors"`
	Timeouts            datasourceTimeouts.Value      `tfsdk:"timeouts"`
}

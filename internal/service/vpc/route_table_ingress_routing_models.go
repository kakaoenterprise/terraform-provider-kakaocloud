// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package vpc

import (
	resourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type routeTableIngressRoutingResourceModel struct {
	Id                 types.String           `tfsdk:"id"`
	RouteTableId       types.String           `tfsdk:"route_table_id"`
	VpcId              types.String           `tfsdk:"vpc_id"`
	InternetGatewayId  types.String           `tfsdk:"internet_gateway_id"`
	ProvisioningStatus types.String           `tfsdk:"provisioning_status"`
	Timeouts           resourceTimeouts.Value `tfsdk:"timeouts"`
}

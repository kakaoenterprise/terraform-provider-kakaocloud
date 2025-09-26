// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package loadbalancer

import (
	datasourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	resourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type healthMonitorTargetGroupModel struct {
	Id types.String `tfsdk:"id"`
}

type loadBalancerHealthMonitorBaseModel struct {
	Id                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	Type               types.String `tfsdk:"type"`
	Delay              types.Int64  `tfsdk:"delay"`
	Timeout            types.Int64  `tfsdk:"timeout"`
	MaxRetries         types.Int64  `tfsdk:"max_retries"`
	MaxRetriesDown     types.Int64  `tfsdk:"max_retries_down"`
	HttpMethod         types.String `tfsdk:"http_method"`
	HttpVersion        types.String `tfsdk:"http_version"`
	UrlPath            types.String `tfsdk:"url_path"`
	ExpectedCodes      types.String `tfsdk:"expected_codes"`
	ProjectId          types.String `tfsdk:"project_id"`
	TargetGroupId      types.String `tfsdk:"target_group_id"`
	TargetGroups       types.List   `tfsdk:"target_groups"`
	ProvisioningStatus types.String `tfsdk:"provisioning_status"`
	OperatingStatus    types.String `tfsdk:"operating_status"`
	CreatedAt          types.String `tfsdk:"created_at"`
	UpdatedAt          types.String `tfsdk:"updated_at"`
}

type loadBalancerHealthMonitorResourceModel struct {
	loadBalancerHealthMonitorBaseModel
	Timeouts resourceTimeouts.Value `tfsdk:"timeouts"`
}

type loadBalancerHealthMonitorDataSourceModel struct {
	loadBalancerHealthMonitorBaseModel
	Timeouts datasourceTimeouts.Value `tfsdk:"timeouts"`
}

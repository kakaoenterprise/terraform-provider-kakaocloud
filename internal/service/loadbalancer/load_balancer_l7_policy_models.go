// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package loadbalancer

import (
	"terraform-provider-kakaocloud/internal/common"

	datasourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	resourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// loadBalancerL7PolicyBaseModel represents the base L7 policy model for standalone L7 policy resources
// This model includes listener_id as it's present in standalone L7 policy API responses
type loadBalancerL7PolicyBaseModel struct {
	Id                    types.String `tfsdk:"id"`
	Name                  types.String `tfsdk:"name"`
	Description           types.String `tfsdk:"description"`
	ListenerId            types.String `tfsdk:"listener_id"`
	Action                types.String `tfsdk:"action"`
	Position              types.Int64  `tfsdk:"position"`
	RedirectTargetGroupId types.String `tfsdk:"redirect_target_group_id"`
	RedirectUrl           types.String `tfsdk:"redirect_url"`
	RedirectPrefix        types.String `tfsdk:"redirect_prefix"`
	RedirectHttpCode      types.Int64  `tfsdk:"redirect_http_code"`
	ProvisioningStatus    types.String `tfsdk:"provisioning_status"`
	OperatingStatus       types.String `tfsdk:"operating_status"`
	ProjectId             types.String `tfsdk:"project_id"`
	Rules                 types.List   `tfsdk:"rules"`
}

type loadBalancerL7PolicyResourceModel struct {
	loadBalancerL7PolicyBaseModel
	Timeouts resourceTimeouts.Value `tfsdk:"timeouts"`
}

type loadBalancerL7PolicyDataSourceModel struct {
	loadBalancerL7PolicyBaseModel
	Timeouts datasourceTimeouts.Value `tfsdk:"timeouts"`
}

type loadBalancerL7PoliciesDataSourceModel struct {
	LoadBalancerId types.String                    `tfsdk:"load_balancer_id"`
	ListenerId     types.String                    `tfsdk:"listener_id"`
	Filter         []common.FilterModel            `tfsdk:"filter"`
	L7Policies     []loadBalancerL7PolicyBaseModel `tfsdk:"l7_policies"`
	Timeouts       datasourceTimeouts.Value        `tfsdk:"timeouts"`
}

// loadBalancerListenerL7PolicyModel represents L7 policy within listener context
// This model does NOT include listener_id as it's not present in the listener API response
type loadBalancerListenerL7PolicyModel struct {
	Id                    types.String `tfsdk:"id"`
	Name                  types.String `tfsdk:"name"`
	Description           types.String `tfsdk:"description"`
	Action                types.String `tfsdk:"action"`
	Position              types.Int64  `tfsdk:"position"`
	RedirectTargetGroupId types.String `tfsdk:"redirect_target_group_id"`
	RedirectUrl           types.String `tfsdk:"redirect_url"`
	RedirectPrefix        types.String `tfsdk:"redirect_prefix"`
	RedirectHttpCode      types.Int64  `tfsdk:"redirect_http_code"`
	ProvisioningStatus    types.String `tfsdk:"provisioning_status"`
	OperatingStatus       types.String `tfsdk:"operating_status"`
	ProjectId             types.String `tfsdk:"project_id"`
	Rules                 types.List   `tfsdk:"rules"`
}

var loadBalancerListenerL7PolicyAttrType = map[string]attr.Type{
	"id":                       types.StringType,
	"name":                     types.StringType,
	"description":              types.StringType,
	"provisioning_status":      types.StringType,
	"operating_status":         types.StringType,
	"project_id":               types.StringType,
	"action":                   types.StringType,
	"position":                 types.Int64Type,
	"rules":                    types.ListType{ElemType: types.ObjectType{AttrTypes: loadBalancerListenerL7PolicyRuleAttrType}},
	"redirect_target_group_id": types.StringType,
	"redirect_url":             types.StringType,
	"redirect_prefix":          types.StringType,
	"redirect_http_code":       types.Int64Type,
}

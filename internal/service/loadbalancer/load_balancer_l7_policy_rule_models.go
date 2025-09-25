// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package loadbalancer

import (
	datasourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	resourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// loadBalancerL7PolicyRuleModel represents the L7 policy rule model used in standalone L7 policy context
type loadBalancerL7PolicyRuleModel struct {
	Id                 types.String `tfsdk:"id"`
	Type               types.String `tfsdk:"type"`
	CompareType        types.String `tfsdk:"compare_type"`
	Key                types.String `tfsdk:"key"`
	Value              types.String `tfsdk:"value"`
	IsInverted         types.Bool   `tfsdk:"is_inverted"`
	ProvisioningStatus types.String `tfsdk:"provisioning_status"`
	OperatingStatus    types.String `tfsdk:"operating_status"`
	ProjectId          types.String `tfsdk:"project_id"`
}

// loadBalancerL7PolicyRuleBaseModel represents the base model for L7 policy rules
// Based on SDK Get Response structure
type loadBalancerL7PolicyRuleBaseModel struct {
	Id                 types.String `tfsdk:"id"`
	L7PolicyId         types.String `tfsdk:"l7_policy_id"`
	Type               types.String `tfsdk:"type"`
	CompareType        types.String `tfsdk:"compare_type"`
	Key                types.String `tfsdk:"key"`
	Value              types.String `tfsdk:"value"`
	IsInverted         types.Bool   `tfsdk:"is_inverted"`
	ProvisioningStatus types.String `tfsdk:"provisioning_status"`
	OperatingStatus    types.String `tfsdk:"operating_status"`
	ProjectId          types.String `tfsdk:"project_id"`
}

// loadBalancerL7PolicyRuleDataSourceModel represents the data source model
type loadBalancerL7PolicyRuleDataSourceModel struct {
	loadBalancerL7PolicyRuleBaseModel
	Timeouts datasourceTimeouts.Value `tfsdk:"timeouts"`
}

// loadBalancerL7PolicyRuleResourceModel represents the resource model
type loadBalancerL7PolicyRuleResourceModel struct {
	loadBalancerL7PolicyRuleBaseModel
	Timeouts resourceTimeouts.Value `tfsdk:"timeouts"`
}

// loadBalancerL7PolicyRulesDataSourceModel represents the list data source model
type loadBalancerL7PolicyRulesDataSourceModel struct {
	Id         types.String                        `tfsdk:"id"`
	L7Rules    []loadBalancerL7PolicyRuleBaseModel `tfsdk:"l7_rules"`
	RulesCount types.Int64                         `tfsdk:"rules_count"`
	Timeouts   datasourceTimeouts.Value            `tfsdk:"timeouts"`
}

// loadBalancerL7PolicyRuleListDataSourceModel represents the list data source model (alternative naming)
type loadBalancerL7PolicyRuleListDataSourceModel struct {
	Id       types.String                        `tfsdk:"id"`
	L7Rules  []loadBalancerL7PolicyRuleBaseModel `tfsdk:"l7_rules"`
	Timeouts datasourceTimeouts.Value            `tfsdk:"timeouts"`
}

// loadBalancerListenerL7PolicyRuleModel represents the L7 policy rule model used within listener context
type loadBalancerListenerL7PolicyRuleModel struct {
	Id                 types.String `tfsdk:"id"`
	CompareType        types.String `tfsdk:"compare_type"`
	IsInverted         types.Bool   `tfsdk:"is_inverted"`
	Key                types.String `tfsdk:"key"`
	Value              types.String `tfsdk:"value"`
	ProvisioningStatus types.String `tfsdk:"provisioning_status"`
	OperatingStatus    types.String `tfsdk:"operating_status"`
	ProjectId          types.String `tfsdk:"project_id"`
	Type               types.String `tfsdk:"type"`
}

var loadBalancerL7PolicyRuleAttrType = map[string]attr.Type{
	"id":                  types.StringType,
	"type":                types.StringType,
	"compare_type":        types.StringType,
	"key":                 types.StringType,
	"value":               types.StringType,
	"is_inverted":         types.BoolType,
	"provisioning_status": types.StringType,
	"operating_status":    types.StringType,
	"project_id":          types.StringType,
}

var loadBalancerListenerL7PolicyRuleAttrType = map[string]attr.Type{
	"id":                  types.StringType,
	"compare_type":        types.StringType,
	"is_inverted":         types.BoolType,
	"key":                 types.StringType,
	"value":               types.StringType,
	"provisioning_status": types.StringType,
	"operating_status":    types.StringType,
	"project_id":          types.StringType,
	"type":                types.StringType,
}

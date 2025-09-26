// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package loadbalancer

import (
	datasourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	resourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

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

type loadBalancerL7PolicyRuleDataSourceModel struct {
	loadBalancerL7PolicyRuleBaseModel
	Timeouts datasourceTimeouts.Value `tfsdk:"timeouts"`
}

type loadBalancerL7PolicyRuleResourceModel struct {
	loadBalancerL7PolicyRuleBaseModel
	Timeouts resourceTimeouts.Value `tfsdk:"timeouts"`
}

type loadBalancerL7PolicyRulesDataSourceModel struct {
	Id         types.String                        `tfsdk:"id"`
	L7Rules    []loadBalancerL7PolicyRuleBaseModel `tfsdk:"l7_rules"`
	RulesCount types.Int64                         `tfsdk:"rules_count"`
	Timeouts   datasourceTimeouts.Value            `tfsdk:"timeouts"`
}

type loadBalancerL7PolicyRuleListDataSourceModel struct {
	Id       types.String                        `tfsdk:"id"`
	L7Rules  []loadBalancerL7PolicyRuleBaseModel `tfsdk:"l7_rules"`
	Timeouts datasourceTimeouts.Value            `tfsdk:"timeouts"`
}

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

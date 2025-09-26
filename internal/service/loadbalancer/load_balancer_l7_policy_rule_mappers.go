// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package loadbalancer

import (
	. "terraform-provider-kakaocloud/internal/utils"

	datasourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	resourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kakaoenterprise/kc-sdk-go/services/loadbalancer"
	"golang.org/x/net/context"
)

func mapLoadBalancerL7PolicyRuleBaseModel(
	ctx context.Context,
	base *loadBalancerL7PolicyRuleBaseModel,
	src *loadbalancer.BnsLoadBalancerV1ApiGetL7PolicyRuleModelL7PolicyRuleModel,
	l7PolicyId string,
	diags *diag.Diagnostics,
) bool {
	base.Id = types.StringValue(src.Id)
	base.L7PolicyId = types.StringValue(l7PolicyId)
	base.Type = ConvertNullableString(src.Type)
	base.CompareType = ConvertNullableString(src.CompareType)
	base.Key = ConvertNullableString(src.Key)
	base.Value = ConvertNullableString(src.Value)
	base.IsInverted = types.BoolValue(src.IsInverted)
	base.ProvisioningStatus = ConvertNullableString(src.ProvisioningStatus)
	base.OperatingStatus = ConvertNullableString(src.OperatingStatus)
	base.ProjectId = types.StringValue(src.ProjectId)

	return !diags.HasError()
}

func mapLoadBalancerL7PolicyRuleDataSourceFromGetRuleResponse(
	ctx context.Context,
	config *loadBalancerL7PolicyRuleDataSourceModel,
	src *loadbalancer.BnsLoadBalancerV1ApiGetL7PolicyRuleModelL7PolicyRuleModel,
	l7PolicyId string,
	diags *diag.Diagnostics,
) bool {
	ok := mapLoadBalancerL7PolicyRuleBaseModel(ctx, &config.loadBalancerL7PolicyRuleBaseModel, src, l7PolicyId, diags)
	if !ok {
		return false
	}

	return !diags.HasError()
}

func mapLoadBalancerL7PolicyRuleToCreateRequest(plan loadBalancerL7PolicyRuleResourceModel) loadbalancer.CreateL7PolicyRuleModel {
	var key loadbalancer.NullableString
	if !plan.Key.IsNull() && !plan.Key.IsUnknown() {
		key = *loadbalancer.NewNullableString(plan.Key.ValueStringPointer())
	}

	var isInverted loadbalancer.NullableBool
	if !plan.IsInverted.IsNull() && !plan.IsInverted.IsUnknown() {
		boolVal := plan.IsInverted.ValueBool()
		isInverted = *loadbalancer.NewNullableBool(&boolVal)
	}

	return loadbalancer.CreateL7PolicyRuleModel{
		Type:        loadbalancer.L7RuleType(plan.Type.ValueString()),
		CompareType: loadbalancer.L7RuleCompareType(plan.CompareType.ValueString()),
		Key:         key,
		Value:       plan.Value.ValueString(),
		IsInverted:  isInverted,
	}
}

func mapLoadBalancerL7PolicyRuleToUpdateRequest(plan loadBalancerL7PolicyRuleResourceModel) loadbalancer.EditL7PolicyRuleModel {
	var key loadbalancer.NullableString
	if !plan.Key.IsNull() && !plan.Key.IsUnknown() {
		key = *loadbalancer.NewNullableString(plan.Key.ValueStringPointer())
	}

	var isInverted loadbalancer.NullableBool
	if !plan.IsInverted.IsNull() && !plan.IsInverted.IsUnknown() {
		boolVal := plan.IsInverted.ValueBool()
		isInverted = *loadbalancer.NewNullableBool(&boolVal)
	}

	return loadbalancer.EditL7PolicyRuleModel{
		Type:        loadbalancer.L7RuleType(plan.Type.ValueString()),
		CompareType: loadbalancer.L7RuleCompareType(plan.CompareType.ValueString()),
		Key:         key,
		Value:       plan.Value.ValueString(),
		IsInverted:  isInverted,
	}
}

func mapLoadBalancerL7PolicyRuleFromGetResponse(src loadbalancer.BnsLoadBalancerV1ApiGetL7PolicyRuleModelL7PolicyRuleModel, l7PolicyId string, timeouts resourceTimeouts.Value) loadBalancerL7PolicyRuleResourceModel {
	return loadBalancerL7PolicyRuleResourceModel{
		loadBalancerL7PolicyRuleBaseModel: loadBalancerL7PolicyRuleBaseModel{
			Id:                 types.StringValue(src.Id),
			L7PolicyId:         types.StringValue(l7PolicyId),
			Type:               ConvertNullableString(src.Type),
			CompareType:        ConvertNullableString(src.CompareType),
			Key:                ConvertNullableString(src.Key),
			Value:              ConvertNullableString(src.Value),
			IsInverted:         types.BoolValue(src.IsInverted),
			ProvisioningStatus: ConvertNullableString(src.ProvisioningStatus),
			OperatingStatus:    ConvertNullableString(src.OperatingStatus),
			ProjectId:          types.StringValue(src.ProjectId),
		},
		Timeouts: timeouts,
	}
}

func mapLoadBalancerL7PolicyRuleListFromGetPolicyResponse(src loadbalancer.BnsLoadBalancerV1ApiGetL7PolicyModelResponseL7PolicyModel, l7PolicyId string, timeouts datasourceTimeouts.Value) loadBalancerL7PolicyRuleListDataSourceModel {
	var l7Rules []loadBalancerL7PolicyRuleBaseModel

	for _, rule := range src.L7Policy.Rules {
		l7Rules = append(l7Rules, loadBalancerL7PolicyRuleBaseModel{
			Id:                 types.StringValue(rule.Id),
			L7PolicyId:         types.StringValue(l7PolicyId),
			Type:               ConvertNullableString(rule.Type),
			CompareType:        ConvertNullableString(rule.CompareType),
			Key:                ConvertNullableString(rule.Key),
			Value:              ConvertNullableString(rule.Value),
			IsInverted:         types.BoolValue(rule.IsInverted),
			ProvisioningStatus: ConvertNullableString(rule.ProvisioningStatus),
			OperatingStatus:    ConvertNullableString(rule.OperatingStatus),
			ProjectId:          types.StringValue(rule.ProjectId),
		})
	}

	return loadBalancerL7PolicyRuleListDataSourceModel{
		Id:       types.StringValue(l7PolicyId),
		L7Rules:  l7Rules,
		Timeouts: timeouts,
	}
}

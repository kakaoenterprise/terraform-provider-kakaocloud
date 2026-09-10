// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package loadbalancer

import (
	. "terraform-provider-kakaocloud/internal/utils"

	datasourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kakaoenterprise/kc-sdk-go/v2/services/loadbalancer"
)

func mapLoadBalancerL7PolicyRuleModel(
	base *loadBalancerL7PolicyRuleBaseModel,
	src *loadbalancer.Rule,
	l7PolicyId string,
	diags *diag.Diagnostics,
) bool {
	base.Id = ConvertNullableString(src.Id)
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

func mapLoadBalancerL7PolicyRuleListFromGetPolicyResponse(src loadbalancer.GetL7PolicyResponse, l7PolicyId string, timeouts datasourceTimeouts.Value) loadBalancerL7PolicyRulesDataSourceModel {
	l7Rules := make([]loadBalancerL7PolicyRuleBaseModel, 0, len(src.L7Policy.Rules))

	for _, rule := range src.L7Policy.Rules {
		l7Rules = append(l7Rules, loadBalancerL7PolicyRuleBaseModel{
			Id:                 ConvertNullableString(rule.Id),
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

	return loadBalancerL7PolicyRulesDataSourceModel{
		Id:       types.StringValue(l7PolicyId),
		L7Rules:  l7Rules,
		Timeouts: timeouts,
	}
}

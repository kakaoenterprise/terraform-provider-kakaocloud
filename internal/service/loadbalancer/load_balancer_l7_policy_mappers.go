// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package loadbalancer

import (
	"context"
	"terraform-provider-kakaocloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kakaoenterprise/kc-sdk-go/services/loadbalancer"
)

func mapLoadBalancerL7PolicyFromGetResponse(
	ctx context.Context,
	model *loadBalancerL7PolicyBaseModel,
	src *loadbalancer.BnsLoadBalancerV1ApiGetL7PolicyModelL7PolicyModel,
	diags *diag.Diagnostics,
) bool {
	model.Id = types.StringValue(src.Id)
	model.Action = utils.ConvertNullableString(src.Action)

	if src.Name.IsSet() {
		model.Name = utils.ConvertNullableString(src.Name)
	}
	if src.Description.IsSet() {
		model.Description = utils.ConvertNullableString(src.Description)
	}
	if src.Position.IsSet() {
		model.Position = utils.ConvertNullableInt32ToInt64(src.Position)
	}
	if src.RedirectTargetGroupId.IsSet() {
		model.RedirectTargetGroupId = utils.ConvertNullableString(src.RedirectTargetGroupId)
	}
	if src.RedirectUrl.IsSet() {
		model.RedirectUrl = utils.ConvertNullableString(src.RedirectUrl)
	}
	if src.RedirectPrefix.IsSet() {
		model.RedirectPrefix = utils.ConvertNullableString(src.RedirectPrefix)
	}
	if src.RedirectHttpCode.IsSet() {
		model.RedirectHttpCode = utils.ConvertNullableInt32ToInt64(src.RedirectHttpCode)
	}

	model.ProvisioningStatus = utils.ConvertNullableString(src.ProvisioningStatus)
	model.OperatingStatus = utils.ConvertNullableString(src.OperatingStatus)
	model.ProjectId = utils.ConvertNullableString(src.ProjectId)

	rules, ruleDiags := utils.ConvertListFromModel(ctx, src.Rules, loadBalancerL7PolicyRuleAttrType, func(rule loadbalancer.BnsLoadBalancerV1ApiGetL7PolicyModelRuleModel) any {
		return loadBalancerL7PolicyRuleModel{
			Id:                 types.StringValue(rule.Id),
			Type:               utils.ConvertNullableString(rule.Type),
			CompareType:        utils.ConvertNullableString(rule.CompareType),
			Key:                utils.ConvertNullableString(rule.Key),
			Value:              utils.ConvertNullableString(rule.Value),
			IsInverted:         types.BoolValue(rule.IsInverted),
			ProvisioningStatus: utils.ConvertNullableString(rule.ProvisioningStatus),
			OperatingStatus:    utils.ConvertNullableString(rule.OperatingStatus),
			ProjectId:          types.StringValue(rule.ProjectId),
		}
	})
	diags.Append(ruleDiags...)
	model.Rules = rules

	return !diags.HasError()
}

func mapLoadBalancerL7PolicyDataSourceFromGetResponse(
	ctx context.Context,
	model *loadBalancerL7PolicyBaseModel,
	src *loadbalancer.BnsLoadBalancerV1ApiGetL7PolicyModelL7PolicyModel,
	diags *diag.Diagnostics,
) bool {
	model.Id = types.StringValue(src.Id)
	model.Name = utils.ConvertNullableString(src.Name)
	model.Description = utils.ConvertNullableString(src.Description)
	model.Action = utils.ConvertNullableString(src.Action)
	model.Position = utils.ConvertNullableInt32ToInt64(src.Position)
	model.RedirectTargetGroupId = utils.ConvertNullableString(src.RedirectTargetGroupId)
	model.RedirectUrl = utils.ConvertNullableString(src.RedirectUrl)
	model.RedirectPrefix = utils.ConvertNullableString(src.RedirectPrefix)
	model.RedirectHttpCode = utils.ConvertNullableInt32ToInt64(src.RedirectHttpCode)
	model.ProvisioningStatus = utils.ConvertNullableString(src.ProvisioningStatus)
	model.OperatingStatus = utils.ConvertNullableString(src.OperatingStatus)
	model.ProjectId = utils.ConvertNullableString(src.ProjectId)

	rules, ruleDiags := utils.ConvertListFromModel(ctx, src.Rules, loadBalancerL7PolicyRuleAttrType, func(rule loadbalancer.BnsLoadBalancerV1ApiGetL7PolicyModelRuleModel) any {
		return loadBalancerL7PolicyRuleModel{
			Id:                 types.StringValue(rule.Id),
			Type:               utils.ConvertNullableString(rule.Type),
			CompareType:        utils.ConvertNullableString(rule.CompareType),
			Key:                utils.ConvertNullableString(rule.Key),
			Value:              utils.ConvertNullableString(rule.Value),
			IsInverted:         types.BoolValue(rule.IsInverted),
			ProvisioningStatus: utils.ConvertNullableString(rule.ProvisioningStatus),
			OperatingStatus:    utils.ConvertNullableString(rule.OperatingStatus),
			ProjectId:          types.StringValue(rule.ProjectId),
		}
	})
	diags.Append(ruleDiags...)
	model.Rules = rules

	return !diags.HasError()
}

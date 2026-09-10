// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package loadbalancer

import (
	"context"
	"terraform-provider-kakaocloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kakaoenterprise/kc-sdk-go/v2/services/loadbalancer"
)

func mapLoadBalancerL7PolicyFromGetResponse(
	ctx context.Context,
	model *loadBalancerL7PolicyBaseModel,
	src *loadbalancer.L7Policy,
	diags *diag.Diagnostics,
) bool {
	model.Id = utils.ConvertNullableString(src.Id)
	model.Action = utils.ConvertNullableString(src.Action)
	model.Name = utils.ConvertNullableString(src.Name)
	model.Description = utils.ConvertNullableString(src.Description)
	model.Position = utils.ConvertNullableInt32(src.Position)
	model.RedirectTargetGroupId = utils.ConvertNullableString(src.RedirectTargetGroupId)
	model.RedirectUrl = utils.ConvertNullableString(src.RedirectUrl)
	model.RedirectPrefix = utils.ConvertNullableString(src.RedirectPrefix)
	model.RedirectHttpCode = utils.ConvertNullableInt32(src.RedirectHttpCode)
	model.ProvisioningStatus = utils.ConvertNullableString(src.ProvisioningStatus)
	model.OperatingStatus = utils.ConvertNullableString(src.OperatingStatus)
	model.ProjectId = utils.ConvertNullableString(src.ProjectId)

	rules, ruleDiags := utils.ConvertListFromModel(ctx, src.Rules, loadBalancerL7PolicyRuleAttrType, func(rule loadbalancer.Rule) any {
		return loadBalancerL7PolicyRuleModel{
			Id:                 utils.ConvertNullableString(rule.Id),
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

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package loadbalancer

import (
	"terraform-provider-kakaocloud/internal/utils"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kakaoenterprise/kc-sdk-go/services/loadbalancer"
)

func mapHealthMonitorToCreateRequest(model *loadBalancerHealthMonitorResourceModel) *loadbalancer.CreateHealthMonitor {

	healthMonitor := &loadbalancer.CreateHealthMonitor{
		Delay:          model.Delay.ValueInt32(),
		MaxRetries:     model.MaxRetries.ValueInt32(),
		MaxRetriesDown: model.MaxRetriesDown.ValueInt32(),
		TargetGroupId:  model.TargetGroupId.ValueString(),
		Timeout:        model.Timeout.ValueInt32(),
		Type:           loadbalancer.HealthMonitorType(model.Type.ValueString()),
	}

	if !model.HttpMethod.IsNull() {
		httpMethod := loadbalancer.HealthMonitorMethod(model.HttpMethod.ValueString())
		healthMonitor.HttpMethod = *loadbalancer.NewNullableHealthMonitorMethod(&httpMethod)
	}

	if !model.HttpVersion.IsNull() {
		httpVersion := loadbalancer.HealthMonitorHttpVersion(model.HttpVersion.ValueString())
		healthMonitor.HttpVersion = *loadbalancer.NewNullableHealthMonitorHttpVersion(&httpVersion)
	}

	if !model.UrlPath.IsNull() {
		urlPath := model.UrlPath.ValueString()
		healthMonitor.UrlPath = *loadbalancer.NewNullableString(&urlPath)
	}

	if !model.ExpectedCodes.IsNull() {
		expectedCodes := model.ExpectedCodes.ValueString()
		healthMonitor.ExpectedCodes = *loadbalancer.NewNullableString(&expectedCodes)
	}

	return healthMonitor
}

func mapHealthMonitorFromGetResponse(model *loadBalancerHealthMonitorBaseModel, apiModel *loadbalancer.BnsLoadBalancerV1ApiGetTargetGroupHealthMonitorModelHealthMonitorModel) {
	model.Name = types.StringValue(apiModel.Name)
	model.Type = types.StringValue(string(apiModel.Type))
	model.Delay = types.Int32Value(apiModel.Delay)
	model.Timeout = types.Int32Value(apiModel.Timeout)
	model.MaxRetries = types.Int32Value(apiModel.MaxRetries)
	model.MaxRetriesDown = types.Int32Value(apiModel.MaxRetriesDown)
	model.ProjectId = types.StringValue(apiModel.ProjectId)
	model.ProvisioningStatus = types.StringValue(string(apiModel.ProvisioningStatus))
	model.OperatingStatus = types.StringValue(string(apiModel.OperatingStatus))
	model.CreatedAt = types.StringValue(apiModel.CreatedAt.Format(time.RFC3339))
	model.UpdatedAt = utils.ConvertNullableTime(apiModel.UpdatedAt)

	model.HttpMethod = utils.ConvertNullableString(apiModel.HttpMethod)
	model.HttpVersion = utils.ConvertNullableString(apiModel.HttpVersion)
	model.UrlPath = utils.ConvertNullableString(apiModel.UrlPath)
	model.ExpectedCodes = utils.ConvertNullableString(apiModel.ExpectedCodes)

	if len(apiModel.TargetGroups) > 0 {

		model.TargetGroupId = types.StringValue(apiModel.TargetGroups[0].Id)

		targetGroups := make([]attr.Value, 0, len(apiModel.TargetGroups))
		for _, targetGroup := range apiModel.TargetGroups {
			targetGroupObj := map[string]attr.Value{
				"id": types.StringValue(targetGroup.Id),
			}
			targetGroups = append(targetGroups, types.ObjectValueMust(
				map[string]attr.Type{"id": types.StringType},
				targetGroupObj,
			))
		}
		model.TargetGroups = types.ListValueMust(
			types.ObjectType{AttrTypes: map[string]attr.Type{"id": types.StringType}},
			targetGroups,
		)
	} else {
		model.TargetGroupId = types.StringNull()
		model.TargetGroups = types.ListNull(types.ObjectType{AttrTypes: map[string]attr.Type{"id": types.StringType}})
	}
}

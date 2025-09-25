// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package loadbalancer

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kakaoenterprise/kc-sdk-go/services/loadbalancer"
)

// Map health monitor from Terraform model to SDK create request
func mapHealthMonitorToCreateRequest(ctx context.Context, model *loadBalancerHealthMonitorResourceModel) (*loadbalancer.CreateHealthMonitor, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Create the health monitor
	healthMonitor := &loadbalancer.CreateHealthMonitor{
		Delay:          int32(model.Delay.ValueInt64()),
		MaxRetries:     int32(model.MaxRetries.ValueInt64()),
		MaxRetriesDown: int32(model.MaxRetriesDown.ValueInt64()),
		TargetGroupId:  model.TargetGroupId.ValueString(),
		Timeout:        int32(model.Timeout.ValueInt64()),
		Type:           loadbalancer.HealthMonitorType(model.Type.ValueString()),
	}

	// Set optional HTTP-specific fields
	if !model.HttpMethod.IsNull() && !model.HttpMethod.IsUnknown() {
		httpMethod := loadbalancer.HealthMonitorMethod(model.HttpMethod.ValueString())
		healthMonitor.HttpMethod = *loadbalancer.NewNullableHealthMonitorMethod(&httpMethod)
	}

	if !model.HttpVersion.IsNull() && !model.HttpVersion.IsUnknown() {
		httpVersion := loadbalancer.HealthMonitorHttpVersion(model.HttpVersion.ValueString())
		healthMonitor.HttpVersion = *loadbalancer.NewNullableHealthMonitorHttpVersion(&httpVersion)
	}

	if !model.UrlPath.IsNull() && !model.UrlPath.IsUnknown() {
		urlPath := model.UrlPath.ValueString()
		healthMonitor.UrlPath = *loadbalancer.NewNullableString(&urlPath)
	}

	if !model.ExpectedCodes.IsNull() && !model.ExpectedCodes.IsUnknown() {
		expectedCodes := model.ExpectedCodes.ValueString()
		healthMonitor.ExpectedCodes = *loadbalancer.NewNullableString(&expectedCodes)
	}

	return healthMonitor, diags
}

// Map health monitor from Terraform model to SDK update request
func mapHealthMonitorToUpdateRequest(ctx context.Context, model *loadBalancerHealthMonitorResourceModel) (*loadbalancer.EditHealthMonitor, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Create the health monitor
	delay := int32(model.Delay.ValueInt64())
	maxRetries := int32(model.MaxRetries.ValueInt64())
	maxRetriesDown := int32(model.MaxRetriesDown.ValueInt64())
	timeout := int32(model.Timeout.ValueInt64())

	healthMonitor := &loadbalancer.EditHealthMonitor{
		Delay:          *loadbalancer.NewNullableInt32(&delay),
		MaxRetries:     *loadbalancer.NewNullableInt32(&maxRetries),
		MaxRetriesDown: *loadbalancer.NewNullableInt32(&maxRetriesDown),
		Timeout:        *loadbalancer.NewNullableInt32(&timeout),
	}

	// Set optional HTTP-specific fields
	if !model.HttpMethod.IsNull() && !model.HttpMethod.IsUnknown() {
		httpMethod := loadbalancer.HealthMonitorMethod(model.HttpMethod.ValueString())
		healthMonitor.HttpMethod = &httpMethod
	}

	if !model.HttpVersion.IsNull() && !model.HttpVersion.IsUnknown() {
		httpVersion := loadbalancer.HealthMonitorHttpVersion(model.HttpVersion.ValueString())
		healthMonitor.HttpVersion = &httpVersion
	}

	if !model.UrlPath.IsNull() && !model.UrlPath.IsUnknown() {
		urlPath := model.UrlPath.ValueString()
		healthMonitor.UrlPath = *loadbalancer.NewNullableString(&urlPath)
	}

	if !model.ExpectedCodes.IsNull() && !model.ExpectedCodes.IsUnknown() {
		expectedCodes := model.ExpectedCodes.ValueString()
		healthMonitor.ExpectedCodes = &expectedCodes
	}

	return healthMonitor, diags
}

// Map health monitor from SDK response to Terraform model
func mapHealthMonitorFromGetResponse(ctx context.Context, model *loadBalancerHealthMonitorBaseModel, apiModel *loadbalancer.BnsLoadBalancerV1ApiGetTargetGroupHealthMonitorModelHealthMonitorModel, diags *diag.Diagnostics) bool {
	model.Id = types.StringValue(apiModel.Id)
	// Handle empty name field due to JSON field mismatch
	if apiModel.Name == "" {
		// Try to preserve the name from the current model if available
		if !model.Name.IsNull() && !model.Name.IsUnknown() {
			// Keep the existing name
		} else {
			model.Name = types.StringValue("") // Set empty string explicitly
		}
	} else {
		model.Name = types.StringValue(apiModel.Name)
	}
	model.Type = types.StringValue(string(apiModel.Type))
	model.Delay = types.Int64Value(int64(apiModel.Delay))
	model.Timeout = types.Int64Value(int64(apiModel.Timeout))
	model.MaxRetries = types.Int64Value(int64(apiModel.MaxRetries))
	model.MaxRetriesDown = types.Int64Value(int64(apiModel.MaxRetriesDown))
	model.ProjectId = types.StringValue(apiModel.ProjectId)
	model.ProvisioningStatus = types.StringValue(string(apiModel.ProvisioningStatus))
	model.OperatingStatus = types.StringValue(string(apiModel.OperatingStatus))
	model.CreatedAt = types.StringValue(apiModel.CreatedAt.Format(time.RFC3339))
	if apiModel.UpdatedAt.IsSet() && apiModel.UpdatedAt.Get() != nil {
		model.UpdatedAt = types.StringValue(apiModel.UpdatedAt.Get().Format(time.RFC3339))
	} else {
		model.UpdatedAt = types.StringNull()
	}

	// Handle optional HTTP-specific fields
	if apiModel.HttpMethod.IsSet() && apiModel.HttpMethod.Get() != nil {
		model.HttpMethod = types.StringValue(string(*apiModel.HttpMethod.Get()))
	} else {
		model.HttpMethod = types.StringNull()
	}

	if apiModel.HttpVersion.IsSet() && apiModel.HttpVersion.Get() != nil {
		model.HttpVersion = types.StringValue(string(*apiModel.HttpVersion.Get()))
	} else {
		model.HttpVersion = types.StringNull()
	}

	if apiModel.UrlPath.IsSet() && apiModel.UrlPath.Get() != nil {
		model.UrlPath = types.StringValue(*apiModel.UrlPath.Get())
	} else {
		model.UrlPath = types.StringNull()
	}

	if apiModel.ExpectedCodes.IsSet() && apiModel.ExpectedCodes.Get() != nil {
		model.ExpectedCodes = types.StringValue(*apiModel.ExpectedCodes.Get())
	} else {
		model.ExpectedCodes = types.StringNull()
	}

	// Handle target groups
	if len(apiModel.TargetGroups) > 0 {
		// Set the first target group ID (for backward compatibility)
		model.TargetGroupId = types.StringValue(apiModel.TargetGroups[0].Id)

		// Convert target groups to list of target group objects
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

	return true
}

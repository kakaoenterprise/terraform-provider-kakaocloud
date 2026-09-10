// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package loadbalancer

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-kakaocloud/internal/common"
	"terraform-provider-kakaocloud/internal/utils"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kakaoenterprise/kc-sdk-go/v2/services/loadbalancer"
)

var (
	_ resource.Resource                = &loadBalancerHealthMonitorResource{}
	_ resource.ResourceWithConfigure   = &loadBalancerHealthMonitorResource{}
	_ resource.ResourceWithImportState = &loadBalancerHealthMonitorResource{}
)

func NewLoadBalancerHealthMonitorResource() resource.Resource {
	return &loadBalancerHealthMonitorResource{}
}

type loadBalancerHealthMonitorResource struct {
	kc *common.KakaoCloudClient
}

type healthMonitorExpectedState struct {
	Delay          *types.Int32
	MaxRetries     *types.Int32
	MaxRetriesDown *types.Int32
	Timeout        *types.Int32
	HttpMethod     *types.String
	HttpVersion    *types.String
	UrlPath        *types.String
	ExpectedCodes  *types.String
}

const healthMonitorSynchronizedStatus = "SYNCHRONIZED"

func (r *loadBalancerHealthMonitorResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*common.KakaoCloudClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *common.KakaoCloudClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.kc = client
}

func (r *loadBalancerHealthMonitorResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_load_balancer_health_monitor"
}

func (r *loadBalancerHealthMonitorResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: utils.MergeResourceSchemaAttributes(
			loadBalancerHealthMonitorResourceSchema,
			map[string]schema.Attribute{
				"timeouts": timeouts.AttributesAll(ctx),
			},
		),
	}
}

func (r *loadBalancerHealthMonitorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan loadBalancerHealthMonitorResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	timeout, diags := plan.Timeouts.Create(ctx, common.DefaultCreateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	loadBalancerId, ok := r.getLoadBalancerIdByTargetGroupId(ctx, plan.TargetGroupId.ValueString(), &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}
	unlock, ok := common.AcquireLockForID(ctx, *loadBalancerId, &resp.Diagnostics)
	if !ok {
		return
	}
	defer unlock()

	_, ok = pollLoadBalancerStatus(ctx, *loadBalancerId, true, r, r.kc, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	createReq := mapHealthMonitorToCreateRequest(&plan)

	body := *loadbalancer.NewCreateHealthMonitorRequest(*createReq)

	respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*loadbalancer.CreateHealthMonitorResponse, *http.Response, error) {
			return r.kc.ApiClient.LoadBalancerHealthMonitorAPI.CreateHealthMonitor(ctx).XAuthToken(r.kc.XAuthToken).CreateHealthMonitorRequest(body).Execute()
		},
	)

	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "CreateHealthMonitor", err, &resp.Diagnostics)
		return
	}

	plan.Id = types.StringValue(respModel.HealthMonitor.Id)

	result, ok := r.pollHealthMonitorUntilStatus(
		ctx,
		plan.Id.ValueString(),
		loadBalancerAvailableAndErrorStatuses,
		nil,
		&resp.Diagnostics,
	)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	_, ok = pollLoadBalancerStatus(ctx, *loadBalancerId, true, r, r.kc, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	common.CheckResourceAvailableStatus(ctx, r, (*string)(&result.ProvisioningStatus), []string{common.LoadBalancerProvisioningStatusActive}, &resp.Diagnostics)

	mapHealthMonitorFromGetResponse(&plan.loadBalancerHealthMonitorBaseModel, result)

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *loadBalancerHealthMonitorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state loadBalancerHealthMonitorResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	timeout, diags := state.Timeouts.Read(ctx, common.DefaultReadTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	healthMonitor, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*loadbalancer.GetTargetGroupHealthMonitorResponse, *http.Response, error) {
			return r.kc.ApiClient.LoadBalancerHealthMonitorAPI.
				GetTargetGroupHealthMonitor(ctx, state.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				Execute()
		},
	)

	if httpResp != nil && httpResp.StatusCode == 404 {
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "GetHealthMonitor", err, &resp.Diagnostics)
		return
	}

	mapHealthMonitorFromGetResponse(&state.loadBalancerHealthMonitorBaseModel, &healthMonitor.HealthMonitor)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *loadBalancerHealthMonitorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state loadBalancerHealthMonitorResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	timeout, diags := plan.Timeouts.Update(ctx, common.DefaultUpdateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	loadBalancerId, ok := r.getLoadBalancerIdByTargetGroupId(ctx, state.TargetGroupId.ValueString(), &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}
	unlock, ok := common.AcquireLockForID(ctx, *loadBalancerId, &resp.Diagnostics)
	if !ok {
		return
	}
	defer unlock()

	_, ok = pollLoadBalancerStatus(ctx, *loadBalancerId, true, r, r.kc, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	expected := buildHealthMonitorExpectedState(&plan, &state)
	updateReq := loadbalancer.NewUpdateHealthMonitor()

	if expected.Delay != nil {
		updateReq.SetDelay(expected.Delay.ValueInt32())
	}
	if expected.MaxRetries != nil {
		updateReq.SetMaxRetries(expected.MaxRetries.ValueInt32())
	}
	if expected.MaxRetriesDown != nil {
		updateReq.SetMaxRetriesDown(expected.MaxRetriesDown.ValueInt32())
	}
	if expected.Timeout != nil {
		updateReq.SetTimeout(expected.Timeout.ValueInt32())
	}

	if expected.HttpMethod != nil {
		httpMethod := loadbalancer.HealthMonitorMethod(expected.HttpMethod.ValueString())
		updateReq.SetHttpMethod(httpMethod)
	}
	if expected.HttpVersion != nil {
		httpVersion := loadbalancer.HealthMonitorHttpVersion(expected.HttpVersion.ValueString())
		updateReq.SetHttpVersion(httpVersion)
	}
	if expected.UrlPath != nil {
		updateReq.SetUrlPath(expected.UrlPath.ValueString())
	}
	if expected.ExpectedCodes != nil {
		updateReq.SetExpectedCodes(expected.ExpectedCodes.ValueString())
	}

	body := loadbalancer.NewUpdateHealthMonitorRequest(*updateReq)

	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*loadbalancer.UpdateHealthMonitorResponse, *http.Response, error) {
			return r.kc.ApiClient.LoadBalancerHealthMonitorAPI.
				UpdateHealthMonitor(ctx, state.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				UpdateHealthMonitorRequest(*body).
				Execute()
		},
	)

	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "UpdateHealthMonitor", err, &resp.Diagnostics)
		return
	}

	result, ok := r.pollHealthMonitorUntilStatus(
		ctx,
		state.Id.ValueString(),
		loadBalancerAvailableAndErrorStatuses,
		&expected,
		&resp.Diagnostics,
	)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	_, ok = pollLoadBalancerStatus(ctx, *loadBalancerId, true, r, r.kc, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	common.CheckResourceAvailableStatus(ctx, r, (*string)(&result.ProvisioningStatus), []string{common.LoadBalancerProvisioningStatusActive}, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	mapHealthMonitorFromGetResponse(&plan.loadBalancerHealthMonitorBaseModel, result)

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *loadBalancerHealthMonitorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state loadBalancerHealthMonitorResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	timeout, diags := state.Timeouts.Delete(ctx, common.DefaultDeleteTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	loadBalancerId, ok := r.getLoadBalancerIdByTargetGroupId(ctx, state.TargetGroupId.ValueString(), &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}
	unlock, ok := common.AcquireLockForID(ctx, *loadBalancerId, &resp.Diagnostics)
	if !ok {
		return
	}
	defer unlock()

	_, ok = pollLoadBalancerStatus(ctx, *loadBalancerId, true, r, r.kc, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (struct{}, *http.Response, error) {
			httpResp, err := r.kc.ApiClient.LoadBalancerHealthMonitorAPI.
				DeleteHealthMonitor(ctx, state.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				Execute()
			return struct{}{}, httpResp, err
		},
	)

	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "DeleteHealthMonitor", err, &resp.Diagnostics)
		return
	}

	common.PollUntilDeletion(ctx, r, 5*time.Second, &resp.Diagnostics, func(ctx context.Context) (bool, *http.Response, error) {
		_, httpResp, err := r.kc.ApiClient.LoadBalancerHealthMonitorAPI.
			GetTargetGroupHealthMonitor(ctx, state.Id.ValueString()).
			XAuthToken(r.kc.XAuthToken).
			Execute()
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			return true, httpResp, nil
		}
		return false, httpResp, err
	})
}

func (r *loadBalancerHealthMonitorResource) pollHealthMonitorUntilStatus(
	ctx context.Context,
	healthMonitorId string,
	targetStatuses []string,
	expected *healthMonitorExpectedState,
	resp *diag.Diagnostics,
) (*loadbalancer.HealthMonitor, bool) {
	pollingTargetStatuses := targetStatuses
	if expected != nil {
		pollingTargetStatuses = []string{
			healthMonitorSynchronizedStatus,
			common.LoadBalancerProvisioningStatusError,
		}
	}

	return common.PollUntilResult(
		ctx,
		r,
		2*time.Second,
		"target group health monitor",
		healthMonitorId,
		pollingTargetStatuses,
		resp,
		func(ctx context.Context) (*loadbalancer.HealthMonitor, *http.Response, error) {
			respModel, httpResp, err := r.kc.ApiClient.LoadBalancerHealthMonitorAPI.
				GetTargetGroupHealthMonitor(ctx, healthMonitorId).
				XAuthToken(r.kc.XAuthToken).
				Execute()
			if err != nil {
				return nil, httpResp, err
			}
			return &respModel.HealthMonitor, httpResp, nil
		},
		func(hm *loadbalancer.HealthMonitor) string {
			status := string(hm.ProvisioningStatus)
			if status != common.LoadBalancerProvisioningStatusActive || expected == nil {
				return status
			}

			if healthMonitorMatchesExpectedState(hm, expected) {
				return healthMonitorSynchronizedStatus
			}

			return status
		},
	)
}

func buildHealthMonitorExpectedState(
	plan *loadBalancerHealthMonitorResourceModel,
	state *loadBalancerHealthMonitorResourceModel,
) healthMonitorExpectedState {
	expected := healthMonitorExpectedState{}

	if !plan.Delay.Equal(state.Delay) {
		expected.Delay = new(plan.Delay)
	}
	if !plan.MaxRetries.Equal(state.MaxRetries) {
		expected.MaxRetries = new(plan.MaxRetries)
	}
	if !plan.MaxRetriesDown.Equal(state.MaxRetriesDown) {
		expected.MaxRetriesDown = new(plan.MaxRetriesDown)
	}
	if !plan.Timeout.Equal(state.Timeout) {
		expected.Timeout = new(plan.Timeout)
	}
	if !plan.HttpMethod.IsNull() && !plan.HttpMethod.IsUnknown() && !plan.HttpMethod.Equal(state.HttpMethod) {
		expected.HttpMethod = new(plan.HttpMethod)
	}
	if !plan.HttpVersion.IsNull() && !plan.HttpVersion.IsUnknown() && !plan.HttpVersion.Equal(state.HttpVersion) {
		expected.HttpVersion = new(plan.HttpVersion)
	}
	if !plan.UrlPath.IsNull() && !plan.UrlPath.IsUnknown() && !plan.UrlPath.Equal(state.UrlPath) {
		expected.UrlPath = new(plan.UrlPath)
	}
	if !plan.ExpectedCodes.IsNull() && !plan.ExpectedCodes.IsUnknown() && !plan.ExpectedCodes.Equal(state.ExpectedCodes) {
		expected.ExpectedCodes = new(plan.ExpectedCodes)
	}

	return expected
}

func healthMonitorMatchesExpectedState(
	healthMonitor *loadbalancer.HealthMonitor,
	expected *healthMonitorExpectedState,
) bool {
	if expected.Delay != nil && !expected.Delay.Equal(types.Int32Value(healthMonitor.Delay)) {
		return false
	}
	if expected.MaxRetries != nil && !expected.MaxRetries.Equal(types.Int32Value(healthMonitor.MaxRetries)) {
		return false
	}
	if expected.MaxRetriesDown != nil && !expected.MaxRetriesDown.Equal(types.Int32Value(healthMonitor.MaxRetriesDown)) {
		return false
	}
	if expected.Timeout != nil && !expected.Timeout.Equal(types.Int32Value(healthMonitor.Timeout)) {
		return false
	}
	if expected.HttpMethod != nil && !expected.HttpMethod.Equal(utils.ConvertNullableString(healthMonitor.HttpMethod)) {
		return false
	}
	if expected.HttpVersion != nil && !expected.HttpVersion.Equal(utils.ConvertNullableString(healthMonitor.HttpVersion)) {
		return false
	}
	if expected.UrlPath != nil && !expected.UrlPath.Equal(utils.ConvertNullableString(healthMonitor.UrlPath)) {
		return false
	}
	if expected.ExpectedCodes != nil && !expected.ExpectedCodes.Equal(utils.ConvertNullableString(healthMonitor.ExpectedCodes)) {
		return false
	}

	return true
}

func (r *loadBalancerHealthMonitorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *loadBalancerHealthMonitorResource) getLoadBalancerIdByTargetGroupId(ctx context.Context, targetGroupId string, respDiags *diag.Diagnostics) (*string, bool) {
	respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, respDiags,
		func() (*loadbalancer.GetTargetGroupResponse, *http.Response, error) {
			return r.kc.ApiClient.LoadBalancerTargetGroupAPI.GetTargetGroup(ctx, targetGroupId).
				XAuthToken(r.kc.XAuthToken).Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "GetTargetGroup", err, respDiags)
		return nil, false
	}

	return respModel.TargetGroup.LoadBalancerId.Get(), true
}

func (r *loadBalancerHealthMonitorResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config loadBalancerHealthMonitorResourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !config.Delay.IsUnknown() && !config.Timeout.IsUnknown() &&
		config.Delay.ValueInt32() <= config.Timeout.ValueInt32() {
		common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
			fmt.Sprintf("'delay' must be greater than 'timeout'.: %d <= %d", config.Delay.ValueInt32(), config.Timeout.ValueInt32()),
		)
	}

	if config.Type.ValueString() == string(loadbalancer.HEALTHMONITORTYPE_HTTP) ||
		config.Type.ValueString() == string(loadbalancer.HEALTHMONITORTYPE_HTTPS) {

		var missingFields []string

		if config.HttpMethod.IsNull() {
			missingFields = append(missingFields, "'http_method'")
		}
		if config.HttpVersion.IsNull() {
			missingFields = append(missingFields, "'http_version'")
		}
		if config.UrlPath.IsNull() {
			missingFields = append(missingFields, "'url_path'")
		}
		if config.ExpectedCodes.IsNull() {
			missingFields = append(missingFields, "'expected_codes'")
		}

		if len(missingFields) > 0 {
			common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
				fmt.Sprintf("Missing required attributes for type '%s': %v",
					config.Type.ValueString(), missingFields),
			)
		}
	} else {
		var unnecessaryFields []string

		if !config.HttpMethod.IsNull() {
			unnecessaryFields = append(unnecessaryFields, "'http_method'")
		}
		if !config.HttpVersion.IsNull() {
			unnecessaryFields = append(unnecessaryFields, "'http_version'")
		}
		if !config.UrlPath.IsNull() {
			unnecessaryFields = append(unnecessaryFields, "'url_path'")
		}
		if !config.ExpectedCodes.IsNull() {
			unnecessaryFields = append(unnecessaryFields, "'expected_codes'")
		}

		if len(unnecessaryFields) > 0 {
			common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
				fmt.Sprintf("The following attributes must not be set for type '%s': %v",
					config.Type.ValueString(), unnecessaryFields),
			)
		}
	}
}

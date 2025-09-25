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
	"github.com/kakaoenterprise/kc-sdk-go/services/loadbalancer"
)

// Ensure the implementation satisfies the expected interfaces
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
		Description: "Manages a KakaoCloud Load Balancer Health Monitor.",
		Attributes: utils.MergeResourceSchemaAttributes(
			loadBalancerHealthMonitorResourceSchema,
			map[string]schema.Attribute{
				"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
					Create: true,
					Read:   true,
					Update: true,
					Delete: true,
				}),
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

	// Map Terraform model to SDK create request
	createReq, diags := mapHealthMonitorToCreateRequest(ctx, &plan)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Create the health monitor
	body := *loadbalancer.NewBodyCreateHealthMonitor(*createReq)

	// First try with normal auth retry, then with conflict retry if needed
	respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*loadbalancer.BnsLoadBalancerV1ApiCreateHealthMonitorModelResponseHealthMonitorModel, *http.Response, error) {
			return r.kc.ApiClient.LoadBalancerTargetGroupAPI.CreateHealthMonitor(ctx).XAuthToken(r.kc.XAuthToken).BodyCreateHealthMonitor(body).Execute()
		},
	)

	// If we get a 409 conflict, retry with loadbalancer-specific conflict logic
	if httpResp != nil && httpResp.StatusCode == http.StatusConflict {
		respModel, httpResp, err = ExecuteWithLoadBalancerConflictRetry(ctx, r.kc, &resp.Diagnostics,
			func() (*loadbalancer.BnsLoadBalancerV1ApiCreateHealthMonitorModelResponseHealthMonitorModel, *http.Response, error) {
				return r.kc.ApiClient.LoadBalancerTargetGroupAPI.CreateHealthMonitor(ctx).XAuthToken(r.kc.XAuthToken).BodyCreateHealthMonitor(body).Execute()
			},
		)
	}

	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "CreateHealthMonitor", err, &resp.Diagnostics)
		return
	}

	plan.Id = types.StringValue(respModel.HealthMonitor.Id)

	result, ok := r.pollHealthMonitorUntilStatus(
		ctx,
		plan.Id.ValueString(),
		[]string{ProvisioningStatusActive, ProvisioningStatusError},
		&resp.Diagnostics,
	)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	common.CheckResourceAvailableStatus(ctx, r, (*string)(&result.ProvisioningStatus), []string{ProvisioningStatusActive}, &resp.Diagnostics)

	ok = mapHealthMonitorFromGetResponse(ctx, &plan.loadBalancerHealthMonitorBaseModel, result, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	// Set the state
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

	// Get the health monitor
	healthMonitor, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*loadbalancer.BnsLoadBalancerV1ApiGetTargetGroupHealthMonitorModelResponseHealthMonitorModel, *http.Response, error) {
			return r.kc.ApiClient.LoadBalancerTargetGroupAPI.
				GetTargetGroupHealthMonitor(ctx, state.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				Execute()
		},
	)

	// 404 → Remove from Terraform state
	if httpResp != nil && httpResp.StatusCode == 404 {
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "GetHealthMonitor", err, &resp.Diagnostics)
		return
	}

	// Map response back to Terraform model
	ok := mapHealthMonitorFromGetResponse(ctx, &state.loadBalancerHealthMonitorBaseModel, &healthMonitor.HealthMonitor, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	// Set the state
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

	// Map Terraform model to SDK update request
	updateReq, diags := mapHealthMonitorToUpdateRequest(ctx, &plan)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Update the health monitor
	body := loadbalancer.BodyUpdateHealthMonitor{HealthMonitor: *updateReq}

	// First try with normal auth retry, then with conflict retry if needed
	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*loadbalancer.BnsLoadBalancerV1ApiUpdateHealthMonitorModelResponseHealthMonitorModel, *http.Response, error) {
			return r.kc.ApiClient.LoadBalancerTargetGroupAPI.
				UpdateHealthMonitor(ctx, state.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				BodyUpdateHealthMonitor(body).
				Execute()
		},
	)

	// If we get a 409 conflict, retry with loadbalancer-specific conflict logic
	if httpResp != nil && httpResp.StatusCode == http.StatusConflict {
		_, httpResp, err = ExecuteWithLoadBalancerConflictRetry(ctx, r.kc, &resp.Diagnostics,
			func() (*loadbalancer.BnsLoadBalancerV1ApiUpdateHealthMonitorModelResponseHealthMonitorModel, *http.Response, error) {
				return r.kc.ApiClient.LoadBalancerTargetGroupAPI.
					UpdateHealthMonitor(ctx, state.Id.ValueString()).
					XAuthToken(r.kc.XAuthToken).
					BodyUpdateHealthMonitor(body).
					Execute()
			},
		)
	}

	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "UpdateHealthMonitor", err, &resp.Diagnostics)
		return
	}

	// Wait for the health monitor to become active again
	result, ok := common.PollUntilResult(
		ctx,
		r,
		5*time.Second,
		[]string{ProvisioningStatusActive, ProvisioningStatusError},
		&resp.Diagnostics,
		func(ctx context.Context) (*loadbalancer.BnsLoadBalancerV1ApiGetTargetGroupHealthMonitorModelHealthMonitorModel, *http.Response, error) {
			getResp, httpResp, err := r.kc.ApiClient.LoadBalancerTargetGroupAPI.
				GetTargetGroupHealthMonitor(ctx, state.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				Execute()
			return &getResp.HealthMonitor, httpResp, err
		},
		func(hm *loadbalancer.BnsLoadBalancerV1ApiGetTargetGroupHealthMonitorModelHealthMonitorModel) string {
			return string(hm.ProvisioningStatus)
		},
	)

	if !ok || resp.Diagnostics.HasError() {
		return
	}

	// Update the model with the final state
	ok = mapHealthMonitorFromGetResponse(ctx, &plan.loadBalancerHealthMonitorBaseModel, result, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	// Set the state
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

	// Delete the health monitor
	// First try with normal auth retry, then with conflict retry if needed
	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (struct{}, *http.Response, error) {
			httpResp, err := r.kc.ApiClient.LoadBalancerTargetGroupAPI.
				DeleteHealthMonitor(ctx, state.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				Execute()
			return struct{}{}, httpResp, err
		},
	)

	// If we get a 409 conflict, retry with loadbalancer-specific conflict logic
	if httpResp != nil && httpResp.StatusCode == http.StatusConflict {
		_, httpResp, err = ExecuteWithLoadBalancerConflictRetry(ctx, r.kc, &resp.Diagnostics,
			func() (struct{}, *http.Response, error) {
				httpResp, err := r.kc.ApiClient.LoadBalancerTargetGroupAPI.
					DeleteHealthMonitor(ctx, state.Id.ValueString()).
					XAuthToken(r.kc.XAuthToken).
					Execute()
				return struct{}{}, httpResp, err
			},
		)
	}

	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "DeleteHealthMonitor", err, &resp.Diagnostics)
		return
	}

	// Wait for deletion to complete
	common.PollUntilDeletion(ctx, r, 5*time.Second, &resp.Diagnostics, func(ctx context.Context) (bool, *http.Response, error) {
		_, httpResp, err := r.kc.ApiClient.LoadBalancerTargetGroupAPI.
			GetTargetGroupHealthMonitor(ctx, state.Id.ValueString()).
			XAuthToken(r.kc.XAuthToken).
			Execute()
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			return true, httpResp, nil
		}
		return false, httpResp, err
	})
}

// pollHealthMonitorUntilStatus polls the health monitor until it reaches one of the target statuses
func (r *loadBalancerHealthMonitorResource) pollHealthMonitorUntilStatus(
	ctx context.Context,
	healthMonitorId string,
	targetStatuses []string,
	resp *diag.Diagnostics,
) (*loadbalancer.BnsLoadBalancerV1ApiGetTargetGroupHealthMonitorModelHealthMonitorModel, bool) {
	return common.PollUntilResult(
		ctx,
		r,
		2*time.Second,
		targetStatuses,
		resp,
		func(ctx context.Context) (*loadbalancer.BnsLoadBalancerV1ApiGetTargetGroupHealthMonitorModelHealthMonitorModel, *http.Response, error) {
			respModel, httpResp, err := r.kc.ApiClient.LoadBalancerTargetGroupAPI.
				GetTargetGroupHealthMonitor(ctx, healthMonitorId).
				XAuthToken(r.kc.XAuthToken).
				Execute()
			if err != nil {
				return nil, httpResp, err
			}
			return &respModel.HealthMonitor, httpResp, nil
		},
		func(hm *loadbalancer.BnsLoadBalancerV1ApiGetTargetGroupHealthMonitorModelHealthMonitorModel) string {
			return string(hm.ProvisioningStatus)
		},
	)
}

func (r *loadBalancerHealthMonitorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

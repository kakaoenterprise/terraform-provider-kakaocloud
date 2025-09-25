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

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.ResourceWithConfigure   = &loadBalancerTargetGroupResource{}
	_ resource.ResourceWithImportState = &loadBalancerTargetGroupResource{}
)

func NewLoadBalancerTargetGroupResource() resource.Resource {
	return &loadBalancerTargetGroupResource{}
}

type loadBalancerTargetGroupResource struct {
	kc *common.KakaoCloudClient
}

// Metadata returns the resource type name.
func (r *loadBalancerTargetGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_load_balancer_target_group"
}

// Schema defines the schema for the resource.
func (r *loadBalancerTargetGroupResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Represents a load balancer target group resource.",
		Attributes: utils.MergeResourceSchemaAttributes(
			loadBalancerTargetGroupResourceSchema,
			map[string]schema.Attribute{
				"timeouts": timeouts.AttributesAll(ctx),
			},
		),
	}
}

// Configure adds the provider configured client to the resource.
func (r *loadBalancerTargetGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates the resource and sets the initial Terraform state.
func (r *loadBalancerTargetGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan loadBalancerTargetGroupResourceModel
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

	// Map plan to create request
	createReq := mapLoadBalancerTargetGroupToCreateRequest(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create target group
	body := loadbalancer.BodyCreateTargetGroup{TargetGroup: *createReq}
	// First try with normal auth retry, then with conflict retry if needed
	respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*loadbalancer.BnsLoadBalancerV1ApiCreateTargetGroupModelResponseTargetGroupModel, *http.Response, error) {
			return r.kc.ApiClient.LoadBalancerTargetGroupAPI.
				CreateTargetGroup(ctx).
				XAuthToken(r.kc.XAuthToken).
				BodyCreateTargetGroup(body).
				Execute()
		},
	)

	// If we get a 409 conflict, retry with loadbalancer-specific conflict logic
	if httpResp != nil && httpResp.StatusCode == http.StatusConflict {
		respModel, httpResp, err = ExecuteWithLoadBalancerConflictRetry(ctx, r.kc, &resp.Diagnostics,
			func() (*loadbalancer.BnsLoadBalancerV1ApiCreateTargetGroupModelResponseTargetGroupModel, *http.Response, error) {
				return r.kc.ApiClient.LoadBalancerTargetGroupAPI.
					CreateTargetGroup(ctx).
					XAuthToken(r.kc.XAuthToken).
					BodyCreateTargetGroup(body).
					Execute()
			},
		)
	}

	if err != nil {
		common.AddApiActionError(ctx, createReq, httpResp, "create", err, &resp.Diagnostics)
		return
	}

	// Set the ID from create response
	plan.Id = types.StringValue(respModel.TargetGroup.Id)

	ok := mapLoadBalancerTargetGroupFromCreateResponse(ctx, &plan, &respModel.TargetGroup, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	// Poll until target group is active
	result, ok := r.pollTargetGroupUntilStatus(
		ctx,
		plan.Id.ValueString(),
		[]string{ProvisioningStatusActive, ProvisioningStatusError},
		&resp.Diagnostics,
	)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	// Check if target group is in error state
	if result.ProvisioningStatus.IsSet() && result.ProvisioningStatus.Get() != nil && string(*result.ProvisioningStatus.Get()) == "ERROR" {
		resp.Diagnostics.AddError(
			"Target Group Creation Failed",
			fmt.Sprintf("Target group %s failed to create and is in ERROR state", plan.Id.ValueString()),
		)
		return
	}

	common.CheckResourceAvailableStatus(ctx, r, (*string)(result.ProvisioningStatus.Get()), []string{ProvisioningStatusActive}, &resp.Diagnostics)

	// Preserve original health monitor configuration before mapping
	originalHealthMonitor := plan.HealthMonitor
	originalSessionPersistence := plan.SessionPersistence

	// Map the final Get response after polling to get all computed fields
	ok = mapLoadBalancerTargetGroupFromGetResponse(ctx, &plan, result, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	// Only restore original health monitor if API didn't return it at all
	// If API returns adjusted values, we should accept them as the API has business logic constraints
	if plan.HealthMonitor.IsNull() && !originalHealthMonitor.IsNull() && !originalHealthMonitor.IsUnknown() {
		plan.HealthMonitor = preserveHealthMonitorUserConfig(ctx, originalHealthMonitor)
	}
	// FIXED: For session persistence, only restore if the original plan had it (not null)
	// This prevents issues when creating without session persistence
	if plan.SessionPersistence.IsNull() && !originalSessionPersistence.IsNull() && !originalSessionPersistence.IsUnknown() {
		plan.SessionPersistence = originalSessionPersistence
	}

	// Set state
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *loadBalancerTargetGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state loadBalancerTargetGroupResourceModel
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

	// Get target group - now returns single object (SDK updated)
	respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*loadbalancer.TargetGroupResponseModel, *http.Response, error) {
			return r.kc.ApiClient.LoadBalancerTargetGroupAPI.
				GetTargetGroup(ctx, state.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				Execute()
		},
	)

	if httpResp != nil && httpResp.StatusCode == 404 {
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		common.AddApiActionError(ctx, nil, httpResp, "read", err, &resp.Diagnostics)
		return
	}

	// Map response to state - now using single object response
	ok := mapLoadBalancerTargetGroupFromGetResponse(ctx, &state, &respModel.TargetGroup, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *loadBalancerTargetGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state loadBalancerTargetGroupResourceModel
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

	// Map plan to update request
	updateReq := mapLoadBalancerTargetGroupToUpdateRequest(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update target group
	body := loadbalancer.NewBodyUpdateTargetGroup(*updateReq)
	// First try with normal auth retry, then with conflict retry if needed
	respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*loadbalancer.BnsLoadBalancerV1ApiUpdateTargetGroupModelResponseTargetGroupModel, *http.Response, error) {
			return r.kc.ApiClient.LoadBalancerTargetGroupAPI.
				UpdateTargetGroup(ctx, state.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				BodyUpdateTargetGroup(*body).
				Execute()
		},
	)

	// If we get a 409 conflict, retry with loadbalancer-specific conflict logic
	if httpResp != nil && httpResp.StatusCode == http.StatusConflict {
		respModel, httpResp, err = ExecuteWithLoadBalancerConflictRetry(ctx, r.kc, &resp.Diagnostics,
			func() (*loadbalancer.BnsLoadBalancerV1ApiUpdateTargetGroupModelResponseTargetGroupModel, *http.Response, error) {
				return r.kc.ApiClient.LoadBalancerTargetGroupAPI.
					UpdateTargetGroup(ctx, state.Id.ValueString()).
					XAuthToken(r.kc.XAuthToken).
					BodyUpdateTargetGroup(*body).
					Execute()
			},
		)
	}

	if err != nil {
		common.AddApiActionError(ctx, updateReq, httpResp, "update", err, &resp.Diagnostics)
		return
	}

	// Map response to state
	ok := mapLoadBalancerTargetGroupFromUpdateResponse(ctx, &plan, &respModel.TargetGroup, &state, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	// Set the ID from create response
	plan.Id = types.StringValue(respModel.TargetGroup.Id)

	result, ok := r.pollTargetGroupUntilStatus(
		ctx,
		plan.Id.ValueString(),
		[]string{ProvisioningStatusActive, ProvisioningStatusError},
		&resp.Diagnostics,
	)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	// Check if target group is in error state
	if result.ProvisioningStatus.IsSet() && result.ProvisioningStatus.Get() != nil && string(*result.ProvisioningStatus.Get()) == "ERROR" {
		resp.Diagnostics.AddError(
			"Target Group Update Failed",
			fmt.Sprintf("Target group %s failed to update and is in ERROR state", plan.Id.ValueString()),
		)
		return
	}

	common.CheckResourceAvailableStatus(ctx, r, (*string)(result.ProvisioningStatus.Get()), []string{ProvisioningStatusActive}, &resp.Diagnostics)

	// FIXED: Preserve original configurations BEFORE mapping from Get response
	originalHealthMonitor := plan.HealthMonitor
	originalSessionPersistence := plan.SessionPersistence

	// Map the final Get response after polling to get all computed fields
	ok = mapLoadBalancerTargetGroupFromGetResponse(ctx, &plan, result, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	// Only restore original health monitor if API didn't return it at all
	// If API returns adjusted values, we should accept them as the API has business logic constraints
	if plan.HealthMonitor.IsNull() && !originalHealthMonitor.IsNull() && !originalHealthMonitor.IsUnknown() {
		plan.HealthMonitor = preserveHealthMonitorUserConfig(ctx, originalHealthMonitor)
	}

	// FIXED: For session persistence, respect user's intent
	// If the original plan had session persistence (not null), preserve it
	// If the original plan was null, that means user wants to remove it - keep it null regardless of what API returns
	if !originalSessionPersistence.IsNull() && !originalSessionPersistence.IsUnknown() {
		// User had session persistence in original plan - preserve it
		plan.SessionPersistence = originalSessionPersistence
	} else {
		// User wants to remove session persistence - keep it null even if API still returns it
		plan.SessionPersistence = types.ObjectNull(loadBalancerTargetGroupSessionPersistenceAttrType)
	}

	// Set state
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *loadBalancerTargetGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state loadBalancerTargetGroupResourceModel
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

	// Delete target group
	// First try with normal auth retry, then with conflict retry if needed
	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (interface{}, *http.Response, error) {
			httpResp, err := r.kc.ApiClient.LoadBalancerTargetGroupAPI.
				DeleteTargetGroup(ctx, state.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				Execute()
			return nil, httpResp, err
		},
	)

	// If we get a 409 conflict, retry with loadbalancer-specific conflict logic
	if httpResp != nil && httpResp.StatusCode == http.StatusConflict {
		_, httpResp, err = ExecuteWithLoadBalancerConflictRetry(ctx, r.kc, &resp.Diagnostics,
			func() (interface{}, *http.Response, error) {
				httpResp, err := r.kc.ApiClient.LoadBalancerTargetGroupAPI.
					DeleteTargetGroup(ctx, state.Id.ValueString()).
					XAuthToken(r.kc.XAuthToken).
					Execute()
				return nil, httpResp, err
			},
		)
	}

	if httpResp != nil && httpResp.StatusCode == 404 {
		// Target group already deleted, nothing to do
		return
	}

	if err != nil {
		common.AddApiActionError(ctx, nil, httpResp, "delete", err, &resp.Diagnostics)
		return
	}

	// Poll until deletion is confirmed
	common.PollUntilDeletion(ctx, r, 2*time.Second, &resp.Diagnostics,
		func(ctx context.Context) (bool, *http.Response, error) {
			_, httpResp, err := r.kc.ApiClient.LoadBalancerTargetGroupAPI.
				GetTargetGroup(ctx, state.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				Execute()
			if httpResp != nil && httpResp.StatusCode == 404 {
				return true, httpResp, nil // Deleted successfully
			}
			return false, httpResp, err
		},
	)
}

// ImportState imports the resource from the existing infrastructure.
func (r *loadBalancerTargetGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by target group ID
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// pollTargetGroupUntilStatus polls the target group until it reaches one of the target statuses
func (r *loadBalancerTargetGroupResource) pollTargetGroupUntilStatus(
	ctx context.Context,
	targetGroupId string,
	targetStatuses []string,
	diags *diag.Diagnostics,
) (*loadbalancer.BnsLoadBalancerV1ApiGetTargetGroupModelTargetGroupModel, bool) {
	return common.PollUntilResult(
		ctx,
		r,
		2*time.Second,
		targetStatuses,
		diags,
		func(ctx context.Context) (*loadbalancer.BnsLoadBalancerV1ApiGetTargetGroupModelTargetGroupModel, *http.Response, error) {
			respModel, httpResp, err := r.kc.ApiClient.LoadBalancerTargetGroupAPI.
				GetTargetGroup(ctx, targetGroupId).
				XAuthToken(r.kc.XAuthToken).
				Execute()
			if err != nil {
				return nil, httpResp, err
			}
			return &respModel.TargetGroup, httpResp, nil
		},
		func(model *loadbalancer.BnsLoadBalancerV1ApiGetTargetGroupModelTargetGroupModel) string {
			if model.ProvisioningStatus.IsSet() && model.ProvisioningStatus.Get() != nil {
				return string(*model.ProvisioningStatus.Get())
			}
			return ""
		},
	)
}

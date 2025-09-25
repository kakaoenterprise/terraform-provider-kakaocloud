// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package loadbalancer

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"terraform-provider-kakaocloud/internal/common"
	"terraform-provider-kakaocloud/internal/utils"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/kakaoenterprise/kc-sdk-go/services/loadbalancer"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.ResourceWithConfigure   = &loadBalancerTargetGroupMemberResource{}
	_ resource.ResourceWithImportState = &loadBalancerTargetGroupMemberResource{}
)

// NewLoadBalancerTargetGroupMemberResource is a helper function to simplify the provider implementation.
func NewLoadBalancerTargetGroupMemberResource() resource.Resource {
	return &loadBalancerTargetGroupMemberResource{}
}

// loadBalancerTargetGroupMemberResource is the resource implementation.
type loadBalancerTargetGroupMemberResource struct {
	kc *common.KakaoCloudClient
}

// Metadata returns the resource type name.
func (r *loadBalancerTargetGroupMemberResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_load_balancer_target_group_member"
}

// Schema defines the schema for the resource.
func (r *loadBalancerTargetGroupMemberResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Load Balancer Target Group Member (target instance) in KakaoCloud.",
		Attributes: utils.MergeResourceSchemaAttributes(
			loadBalancerTargetGroupMemberResourceSchema,
			map[string]schema.Attribute{
				"timeouts": timeouts.AttributesAll(ctx),
			},
		),
	}
}

// Configure adds the provider configured client to the resource.
func (r *loadBalancerTargetGroupMemberResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	kc, ok := req.ProviderData.(*common.KakaoCloudClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *common.KakaoCloudClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.kc = kc
}

// Create creates the resource and sets the initial Terraform state.
func (r *loadBalancerTargetGroupMemberResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan loadBalancerTargetGroupMemberResourceModel
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
	createReq := mapLoadBalancerTargetGroupMemberToCreateRequest(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create target group member
	body := loadbalancer.NewBodyAddTarget(*createReq)
	// First try with normal auth retry, then with conflict retry if needed
	respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*loadbalancer.BnsLoadBalancerV1ApiAddTargetModelResponseTargetGroupMemberModel, *http.Response, error) {
			return r.kc.ApiClient.LoadBalancerTargetGroupAPI.
				AddTarget(ctx, plan.TargetGroupId.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				BodyAddTarget(*body).
				Execute()
		},
	)

	// If we get a 409 conflict, retry with loadbalancer-specific conflict logic
	if httpResp != nil && httpResp.StatusCode == http.StatusConflict {
		respModel, httpResp, err = ExecuteWithLoadBalancerConflictRetry(ctx, r.kc, &resp.Diagnostics,
			func() (*loadbalancer.BnsLoadBalancerV1ApiAddTargetModelResponseTargetGroupMemberModel, *http.Response, error) {
				return r.kc.ApiClient.LoadBalancerTargetGroupAPI.
					AddTarget(ctx, plan.TargetGroupId.ValueString()).
					XAuthToken(r.kc.XAuthToken).
					BodyAddTarget(*body).
					Execute()
			},
		)
	}

	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "AddTarget", err, &resp.Diagnostics)
		return
	}

	// After create, we need to poll until the creation is complete
	// This follows the same pattern as other load balancer resources
	result, ok := r.pollTargetGroupMemberUntilStatus(
		ctx,
		plan.TargetGroupId.ValueString(),
		respModel.Member.Id,
		[]string{ProvisioningStatusActive, ProvisioningStatusError},
		&resp.Diagnostics,
	)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	// Check if target group member is in error state
	if result.ProvisioningStatus.IsSet() && result.ProvisioningStatus.Get() != nil && string(*result.ProvisioningStatus.Get()) == "ERROR" {
		resp.Diagnostics.AddError(
			"Target Group Member Creation Failed",
			"The target group member creation failed and is in ERROR state",
		)
		return
	}

	common.CheckResourceAvailableStatus(ctx, r, (*string)(result.ProvisioningStatus.Get()), []string{ProvisioningStatusActive}, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Map the current state from the API response
	ok = mapLoadBalancerTargetGroupMemberFromGetResponse(ctx, &plan, result, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	// Set state with the current values from the API
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

// Read refreshes the Terraform state with the latest data.
func (r *loadBalancerTargetGroupMemberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state loadBalancerTargetGroupMemberResourceModel
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

	// List targets in target group to find the specific member
	respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*loadbalancer.TargetGroupMemberListModel, *http.Response, error) {
			return r.kc.ApiClient.LoadBalancerTargetGroupAPI.
				ListTargetsInTargetGroup(ctx, state.TargetGroupId.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				Execute()
		},
	)

	if httpResp != nil && httpResp.StatusCode == 404 {
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "ListTargetsInTargetGroup", err, &resp.Diagnostics)
		return
	}

	// Find the specific member by ID
	var foundMember *loadbalancer.BnsLoadBalancerV1ApiListTargetsInTargetGroupModelTargetGroupMemberModel
	for _, member := range respModel.Members {
		if member.Id == state.Id.ValueString() {
			foundMember = &member
			break
		}
	}

	if foundMember == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Map response to state
	ok := mapLoadBalancerTargetGroupMemberFromGetResponse(ctx, &state, foundMember, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *loadBalancerTargetGroupMemberResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state loadBalancerTargetGroupMemberResourceModel
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
	updateReq := mapLoadBalancerTargetGroupMemberToUpdateRequest(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update target group member
	body := loadbalancer.NewBodyUpdateTarget(*updateReq)
	// First try with normal auth retry, then with conflict retry if needed
	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*loadbalancer.BnsLoadBalancerV1ApiUpdateTargetModelResponseTargetGroupMemberModel, *http.Response, error) {
			return r.kc.ApiClient.LoadBalancerTargetGroupAPI.
				UpdateTarget(ctx, state.TargetGroupId.ValueString(), state.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				BodyUpdateTarget(*body).
				Execute()
		},
	)

	// If we get a 409 conflict, retry with loadbalancer-specific conflict logic
	if httpResp != nil && httpResp.StatusCode == http.StatusConflict {
		_, httpResp, err = ExecuteWithLoadBalancerConflictRetry(ctx, r.kc, &resp.Diagnostics,
			func() (*loadbalancer.BnsLoadBalancerV1ApiUpdateTargetModelResponseTargetGroupMemberModel, *http.Response, error) {
				return r.kc.ApiClient.LoadBalancerTargetGroupAPI.
					UpdateTarget(ctx, state.TargetGroupId.ValueString(), state.Id.ValueString()).
					XAuthToken(r.kc.XAuthToken).
					BodyUpdateTarget(*body).
					Execute()
			},
		)
	}

	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "UpdateTarget", err, &resp.Diagnostics)
		return
	}

	// After update, we need to poll until the update is complete
	// This follows the same pattern as other load balancer resources
	result, ok := r.pollTargetGroupMemberUntilStatus(
		ctx,
		state.TargetGroupId.ValueString(),
		state.Id.ValueString(),
		[]string{ProvisioningStatusActive, ProvisioningStatusError},
		&resp.Diagnostics,
	)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	// Check if target group member is in error state
	if result.ProvisioningStatus.IsSet() && result.ProvisioningStatus.Get() != nil && string(*result.ProvisioningStatus.Get()) == "ERROR" {
		resp.Diagnostics.AddError(
			"Target Group Member Update Failed",
			"The target group member update failed and is in ERROR state",
		)
		return
	}

	common.CheckResourceAvailableStatus(ctx, r, (*string)(result.ProvisioningStatus.Get()), []string{ProvisioningStatusActive}, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Map the current state from the API response
	ok = mapLoadBalancerTargetGroupMemberFromGetResponse(ctx, &plan, result, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	// Set state with the current values from the API
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *loadBalancerTargetGroupMemberResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state loadBalancerTargetGroupMemberResourceModel
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

	// Delete target group member
	// First try with normal auth retry, then with conflict retry if needed
	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (interface{}, *http.Response, error) {
			httpResp, err := r.kc.ApiClient.LoadBalancerTargetGroupAPI.
				RemoveTarget(ctx, state.TargetGroupId.ValueString(), state.Id.ValueString()).
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
					RemoveTarget(ctx, state.TargetGroupId.ValueString(), state.Id.ValueString()).
					XAuthToken(r.kc.XAuthToken).
					Execute()
				return nil, httpResp, err
			},
		)
	}

	if httpResp != nil && httpResp.StatusCode == 404 {
		// Target group member already deleted, nothing to do
		return
	}

	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "RemoveTarget", err, &resp.Diagnostics)
		return
	}

	// Poll until deletion is confirmed
	common.PollUntilDeletion(ctx, r, 2*time.Second, &resp.Diagnostics,
		func(ctx context.Context) (bool, *http.Response, error) {
			respModel, httpResp, err := r.kc.ApiClient.LoadBalancerTargetGroupAPI.
				ListTargetsInTargetGroup(ctx, state.TargetGroupId.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				Execute()
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					return true, httpResp, nil // Target group or member deleted
				}
				return false, httpResp, err
			}
			// Check if the specific member still exists
			for _, member := range respModel.Members {
				if member.Id == state.Id.ValueString() {
					return false, httpResp, nil // Member still exists
				}
			}
			return true, httpResp, nil // Member deleted
		},
	)
}

// ImportState imports the resource.
func (r *loadBalancerTargetGroupMemberResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Parse import ID format: target_group_id:member_id
	parts := strings.Split(req.ID, ":")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Import ID must be in the format: target_group_id:member_id",
		)
		return
	}

	targetGroupId := parts[0]
	memberId := parts[1]

	// Set the resource attributes
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("target_group_id"), targetGroupId)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), memberId)...)
}

// pollTargetGroupMemberUntilStatus polls until the target group member reaches one of the target statuses
func (r *loadBalancerTargetGroupMemberResource) pollTargetGroupMemberUntilStatus(
	ctx context.Context,
	targetGroupId string,
	memberId string,
	targetStatuses []string,
	resp *diag.Diagnostics,
) (*loadbalancer.BnsLoadBalancerV1ApiListTargetsInTargetGroupModelTargetGroupMemberModel, bool) {
	return common.PollUntilResult(
		ctx,
		r,
		2*time.Second,
		targetStatuses,
		resp,
		func(ctx context.Context) (*loadbalancer.BnsLoadBalancerV1ApiListTargetsInTargetGroupModelTargetGroupMemberModel, *http.Response, error) {
			respModel, httpResp, err := r.kc.ApiClient.LoadBalancerTargetGroupAPI.
				ListTargetsInTargetGroup(ctx, targetGroupId).
				XAuthToken(r.kc.XAuthToken).
				Execute()
			if err != nil {
				return nil, httpResp, err
			}

			// Find the specific member by ID
			for _, member := range respModel.Members {
				if member.Id == memberId {
					return &member, httpResp, nil
				}
			}

			return nil, httpResp, fmt.Errorf("member %s not found in target group %s", memberId, targetGroupId)
		},
		func(v *loadbalancer.BnsLoadBalancerV1ApiListTargetsInTargetGroupModelTargetGroupMemberModel) string {
			return string(*v.ProvisioningStatus.Get())
		},
	)
}

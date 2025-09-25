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
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kakaoenterprise/kc-sdk-go/services/loadbalancer"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.ResourceWithConfigure      = &loadBalancerL7PolicyRuleResource{}
	_ resource.ResourceWithImportState    = &loadBalancerL7PolicyRuleResource{}
	_ resource.ResourceWithValidateConfig = &loadBalancerL7PolicyRuleResource{}
)

func NewLoadBalancerL7PolicyRuleResource() resource.Resource {
	return &loadBalancerL7PolicyRuleResource{}
}

type loadBalancerL7PolicyRuleResource struct {
	kc *common.KakaoCloudClient
}

// Metadata returns the resource type name.
func (r *loadBalancerL7PolicyRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_load_balancer_l7_policy_rule"
}

// Schema defines the schema for the resource.
func (r *loadBalancerL7PolicyRuleResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Represents a KakaoCloud Load Balancer L7 Policy Rule resource.",
		Attributes: utils.MergeResourceSchemaAttributes(
			loadBalancerL7PolicyRuleResourceSchema,
			map[string]schema.Attribute{
				"timeouts": timeouts.AttributesAll(ctx),
			},
		),
	}
}

// Configure adds the provider configured client to the resource.
func (r *loadBalancerL7PolicyRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// ValidateConfig validates the resource configuration.
func (r *loadBalancerL7PolicyRuleResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	ValidateL7PolicyRuleConfig(ctx, req, resp)
}

// Create creates the resource and sets the initial Terraform state.
func (r *loadBalancerL7PolicyRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan loadBalancerL7PolicyRuleResourceModel

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

	// Map Terraform plan to API request
	createReq := mapLoadBalancerL7PolicyRuleToCreateRequest(plan)

	// Create the L7 policy rule
	// First try with normal auth retry, then with conflict retry if needed
	createResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*loadbalancer.BnsLoadBalancerV1ApiAddL7PolicyRuleModelResponseL7PolicyRuleModel, *http.Response, error) {
			return r.kc.ApiClient.LoadBalancerL7PoliciesAPI.AddL7PolicyRule(ctx, plan.L7PolicyId.ValueString()).BodyAddL7PolicyRule(loadbalancer.BodyAddL7PolicyRule{L7Rule: createReq}).XAuthToken(r.kc.XAuthToken).Execute()
		},
	)

	// If we get a 409 conflict, retry with loadbalancer-specific conflict logic
	if httpResp != nil && httpResp.StatusCode == http.StatusConflict {
		createResp, httpResp, err = ExecuteWithLoadBalancerConflictRetry(ctx, r.kc, &resp.Diagnostics,
			func() (*loadbalancer.BnsLoadBalancerV1ApiAddL7PolicyRuleModelResponseL7PolicyRuleModel, *http.Response, error) {
				return r.kc.ApiClient.LoadBalancerL7PoliciesAPI.AddL7PolicyRule(ctx, plan.L7PolicyId.ValueString()).BodyAddL7PolicyRule(loadbalancer.BodyAddL7PolicyRule{L7Rule: createReq}).XAuthToken(r.kc.XAuthToken).Execute()
			},
		)
	}

	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "AddL7PolicyRule", err, &resp.Diagnostics)
		return
	}

	// Set the ID from create response
	plan.Id = types.StringValue(createResp.L7Rule.Id)

	// Wait for the L7 policy rule to become active
	result, ok := r.pollL7PolicyRuleUntilStatus(
		ctx,
		plan.L7PolicyId.ValueString(),
		createResp.L7Rule.Id,
		[]string{ProvisioningStatusActive, ProvisioningStatusError},
		&resp.Diagnostics,
	)

	if !ok || resp.Diagnostics.HasError() {
		return
	}

	common.CheckResourceAvailableStatus(ctx, r, (*string)(result.L7Rule.ProvisioningStatus.Get()), []string{ProvisioningStatusActive}, &resp.Diagnostics)

	ok = mapLoadBalancerL7PolicyRuleBaseModel(ctx, &plan.loadBalancerL7PolicyRuleBaseModel, &result.L7Rule, plan.L7PolicyId.ValueString(), &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

// Read refreshes the Terraform state with the latest data.
func (r *loadBalancerL7PolicyRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state loadBalancerL7PolicyRuleResourceModel

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

	// Get the L7 policy rule
	getResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*loadbalancer.Responsel7PolicyRuleModel, *http.Response, error) {
			return r.kc.ApiClient.LoadBalancerL7PoliciesAPI.GetL7PolicyRule(ctx, state.L7PolicyId.ValueString(), state.Id.ValueString()).XAuthToken(r.kc.XAuthToken).Execute()
		},
	)

	// 404 → Remove from Terraform state
	if httpResp != nil && httpResp.StatusCode == 404 {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "GetL7PolicyRule", err, &resp.Diagnostics)
		return
	}

	// Map API response to Terraform state
	state = mapLoadBalancerL7PolicyRuleFromGetResponse(getResp.L7Rule, state.L7PolicyId.ValueString(), state.Timeouts)

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *loadBalancerL7PolicyRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state loadBalancerL7PolicyRuleResourceModel

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

	// Map Terraform plan to API request
	updateReq := mapLoadBalancerL7PolicyRuleToUpdateRequest(plan)

	// Update the L7 policy rule
	// First try with normal auth retry, then with conflict retry if needed
	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*loadbalancer.BnsLoadBalancerV1ApiUpdateL7PolicyRuleModelResponseL7PolicyRuleModel, *http.Response, error) {
			return r.kc.ApiClient.LoadBalancerL7PoliciesAPI.UpdateL7PolicyRule(ctx, state.L7PolicyId.ValueString(), state.Id.ValueString()).BodyUpdateL7PolicyRule(loadbalancer.BodyUpdateL7PolicyRule{L7Rule: updateReq}).XAuthToken(r.kc.XAuthToken).Execute()
		},
	)

	// If we get a 409 conflict, retry with loadbalancer-specific conflict logic
	if httpResp != nil && httpResp.StatusCode == http.StatusConflict {
		_, httpResp, err = ExecuteWithLoadBalancerConflictRetry(ctx, r.kc, &resp.Diagnostics,
			func() (*loadbalancer.BnsLoadBalancerV1ApiUpdateL7PolicyRuleModelResponseL7PolicyRuleModel, *http.Response, error) {
				return r.kc.ApiClient.LoadBalancerL7PoliciesAPI.UpdateL7PolicyRule(ctx, state.L7PolicyId.ValueString(), state.Id.ValueString()).BodyUpdateL7PolicyRule(loadbalancer.BodyUpdateL7PolicyRule{L7Rule: updateReq}).XAuthToken(r.kc.XAuthToken).Execute()
			},
		)
	}

	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "UpdateL7PolicyRule", err, &resp.Diagnostics)
		return
	}

	// Wait for the L7 policy rule to become active again
	result, ok := r.pollL7PolicyRuleUntilStatus(
		ctx,
		state.L7PolicyId.ValueString(),
		state.Id.ValueString(),
		[]string{ProvisioningStatusActive, ProvisioningStatusError},
		&resp.Diagnostics,
	)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	common.CheckResourceAvailableStatus(ctx, r, (*string)(result.L7Rule.ProvisioningStatus.Get()), []string{ProvisioningStatusActive}, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Map the final GET response
	ok = mapLoadBalancerL7PolicyRuleBaseModel(ctx, &plan.loadBalancerL7PolicyRuleBaseModel, &result.L7Rule, state.L7PolicyId.ValueString(), &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	plan.Timeouts = state.Timeouts
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *loadBalancerL7PolicyRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state loadBalancerL7PolicyRuleResourceModel

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

	// First try with normal auth retry, then with conflict retry if needed
	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (interface{}, *http.Response, error) {
			httpResp, err := r.kc.ApiClient.LoadBalancerL7PoliciesAPI.DeleteL7PolicyRule(ctx, state.L7PolicyId.ValueString(), state.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				Execute()
			return nil, httpResp, err
		},
	)

	// If we get a 409 conflict, retry with loadbalancer-specific conflict logic
	if httpResp != nil && httpResp.StatusCode == http.StatusConflict {
		_, httpResp, err = ExecuteWithLoadBalancerConflictRetry(ctx, r.kc, &resp.Diagnostics,
			func() (interface{}, *http.Response, error) {
				httpResp, err := r.kc.ApiClient.LoadBalancerL7PoliciesAPI.DeleteL7PolicyRule(ctx, state.L7PolicyId.ValueString(), state.Id.ValueString()).
					XAuthToken(r.kc.XAuthToken).
					Execute()
				return nil, httpResp, err
			},
		)
	}

	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			return
		}
		common.AddApiActionError(ctx, r, httpResp, "DeleteL7PolicyRule", err, &resp.Diagnostics)
		return
	}

	// Poll until resource disappears
	common.PollUntilDeletion(ctx, r, 2*time.Second, &resp.Diagnostics, func(ctx context.Context) (bool, *http.Response, error) {
		_, httpResp, err := r.kc.ApiClient.LoadBalancerL7PoliciesAPI.
			GetL7PolicyRule(ctx, state.L7PolicyId.ValueString(), state.Id.ValueString()).
			XAuthToken(r.kc.XAuthToken).
			Execute()

		if httpResp != nil && httpResp.StatusCode == 404 {
			return true, httpResp, nil
		}
		return false, httpResp, err
	})
}

// ImportState imports the resource from the existing infrastructure.
func (r *loadBalancerL7PolicyRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: l7_policy_id,rule_id
	parts := strings.Split(req.ID, ",")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			"Import ID must be in the format: l7_policy_id,rule_id",
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("l7_policy_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

// pollL7PolicyRuleUntilStatus polls the L7 policy rule until it reaches one of the target statuses
func (r *loadBalancerL7PolicyRuleResource) pollL7PolicyRuleUntilStatus(
	ctx context.Context,
	l7PolicyId, ruleId string,
	targetStatuses []string,
	diags *diag.Diagnostics,
) (*loadbalancer.Responsel7PolicyRuleModel, bool) {
	return common.PollUntilResult(
		ctx,
		r,
		2*time.Second,
		targetStatuses,
		diags,
		func(ctx context.Context) (*loadbalancer.Responsel7PolicyRuleModel, *http.Response, error) {
			resp, httpResp, err := r.kc.ApiClient.LoadBalancerL7PoliciesAPI.
				GetL7PolicyRule(ctx, l7PolicyId, ruleId).
				XAuthToken(r.kc.XAuthToken).
				Execute()
			if err != nil {
				return nil, httpResp, err
			}
			return resp, httpResp, nil
		},
		func(v *loadbalancer.Responsel7PolicyRuleModel) string {
			return string(*v.L7Rule.ProvisioningStatus.Get())
		},
	)
}

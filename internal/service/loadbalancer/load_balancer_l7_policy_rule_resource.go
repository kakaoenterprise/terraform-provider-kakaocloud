// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package loadbalancer

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"terraform-provider-kakaocloud/internal/common"
	"terraform-provider-kakaocloud/internal/docs"
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

func (r *loadBalancerL7PolicyRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_load_balancer_l7_policy_rule"
}

func (r *loadBalancerL7PolicyRuleResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: docs.GetResourceDescription("LoadBalancerL7PolicyRule"),
		Attributes: utils.MergeResourceSchemaAttributes(
			loadBalancerL7PolicyRuleResourceSchema,
			map[string]schema.Attribute{
				"timeouts": timeouts.AttributesAll(ctx),
			},
		),
	}
}

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

func (r *loadBalancerL7PolicyRuleResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	ValidateL7PolicyRuleConfig(ctx, req, resp)
}

func (r *loadBalancerL7PolicyRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan loadBalancerL7PolicyRuleResourceModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	loadBalancerId, ok := r.getLoadBalancerIdByL7PolicyId(ctx, plan.L7PolicyId.ValueString(), &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}
	mutex := common.LockForID(*loadBalancerId)
	mutex.Lock()
	defer mutex.Unlock()

	timeout, diags := plan.Timeouts.Create(ctx, common.DefaultCreateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	createReq := mapLoadBalancerL7PolicyRuleToCreateRequest(plan)

	createResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*loadbalancer.BnsLoadBalancerV1ApiAddL7PolicyRuleModelResponseL7PolicyRuleModel, *http.Response, error) {
			return r.kc.ApiClient.LoadBalancerL7PoliciesAPI.AddL7PolicyRule(ctx, plan.L7PolicyId.ValueString()).BodyAddL7PolicyRule(loadbalancer.BodyAddL7PolicyRule{L7Rule: createReq}).XAuthToken(r.kc.XAuthToken).Execute()
		},
	)

	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "AddL7PolicyRule", err, &resp.Diagnostics)
		return
	}

	plan.Id = types.StringValue(createResp.L7Rule.Id)

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

	getResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*loadbalancer.Responsel7PolicyRuleModel, *http.Response, error) {
			return r.kc.ApiClient.LoadBalancerL7PoliciesAPI.GetL7PolicyRule(ctx, state.L7PolicyId.ValueString(), state.Id.ValueString()).XAuthToken(r.kc.XAuthToken).Execute()
		},
	)

	if httpResp != nil && httpResp.StatusCode == 404 {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "GetL7PolicyRule", err, &resp.Diagnostics)
		return
	}

	state = mapLoadBalancerL7PolicyRuleFromGetResponse(getResp.L7Rule, state.L7PolicyId.ValueString(), state.Timeouts)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

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

	loadBalancerId, ok := r.getLoadBalancerIdByL7PolicyId(ctx, state.L7PolicyId.ValueString(), &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}
	mutex := common.LockForID(*loadBalancerId)
	mutex.Lock()
	defer mutex.Unlock()

	timeout, diags := plan.Timeouts.Update(ctx, common.DefaultUpdateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	updateReq := mapLoadBalancerL7PolicyRuleToUpdateRequest(plan)

	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*loadbalancer.BnsLoadBalancerV1ApiUpdateL7PolicyRuleModelResponseL7PolicyRuleModel, *http.Response, error) {
			return r.kc.ApiClient.LoadBalancerL7PoliciesAPI.UpdateL7PolicyRule(ctx, state.L7PolicyId.ValueString(), state.Id.ValueString()).BodyUpdateL7PolicyRule(loadbalancer.BodyUpdateL7PolicyRule{L7Rule: updateReq}).XAuthToken(r.kc.XAuthToken).Execute()
		},
	)

	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "UpdateL7PolicyRule", err, &resp.Diagnostics)
		return
	}

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

	ok = mapLoadBalancerL7PolicyRuleBaseModel(ctx, &plan.loadBalancerL7PolicyRuleBaseModel, &result.L7Rule, state.L7PolicyId.ValueString(), &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	plan.Timeouts = state.Timeouts
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *loadBalancerL7PolicyRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state loadBalancerL7PolicyRuleResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	loadBalancerId, ok := r.getLoadBalancerIdByL7PolicyId(ctx, state.L7PolicyId.ValueString(), &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}
	mutex := common.LockForID(*loadBalancerId)
	mutex.Lock()
	defer mutex.Unlock()

	timeout, diags := state.Timeouts.Delete(ctx, common.DefaultDeleteTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (interface{}, *http.Response, error) {
			httpResp, err := r.kc.ApiClient.LoadBalancerL7PoliciesAPI.DeleteL7PolicyRule(ctx, state.L7PolicyId.ValueString(), state.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				Execute()
			return nil, httpResp, err
		},
	)

	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			return
		}
		common.AddApiActionError(ctx, r, httpResp, "DeleteL7PolicyRule", err, &resp.Diagnostics)
		return
	}

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

func (r *loadBalancerL7PolicyRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {

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

func (r *loadBalancerL7PolicyRuleResource) getLoadBalancerIdByL7PolicyId(ctx context.Context, l7PolicyId string, respDiags *diag.Diagnostics) (*string, bool) {

	listenerId, err := r.findListenerIdByL7PolicyId(ctx, l7PolicyId)
	if err != nil {
		respDiags.AddError("Could not find listener for L7 policy", fmt.Sprintf("Failed to find listener for L7 policy %s: %v", l7PolicyId, err))
		return nil, false
	}

	listenerResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, respDiags,
		func() (*loadbalancer.BnsLoadBalancerV1ApiGetListenerModelResponseListenerModel, *http.Response, error) {
			return r.kc.ApiClient.LoadBalancerListenerAPI.GetListener(ctx, listenerId).
				XAuthToken(r.kc.XAuthToken).Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "GetListener", err, respDiags)
		return nil, false
	}

	return listenerResp.Listener.LoadBalancerId.Get(), true
}

func (r *loadBalancerL7PolicyRuleResource) findListenerIdByL7PolicyId(ctx context.Context, l7PolicyId string) (string, error) {

	listenersResp, _, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &diag.Diagnostics{},
		func() (*loadbalancer.ListenerListModel, *http.Response, error) {
			return r.kc.ApiClient.LoadBalancerListenerAPI.ListListeners(ctx).Limit(1000).XAuthToken(r.kc.XAuthToken).Execute()
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to list listeners: %w", err)
	}

	for _, listener := range listenersResp.Listeners {
		if listener.L7Policies != nil {
			for _, policy := range listener.L7Policies {
				if policy.Id == l7PolicyId {
					return listener.Id, nil
				}
			}
		}
	}

	return "", fmt.Errorf("L7 policy %s not found in any listener", l7PolicyId)
}

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

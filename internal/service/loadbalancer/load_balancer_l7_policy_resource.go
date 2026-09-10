// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package loadbalancer

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
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
	_ resource.Resource                = &loadBalancerL7PolicyResource{}
	_ resource.ResourceWithConfigure   = &loadBalancerL7PolicyResource{}
	_ resource.ResourceWithImportState = &loadBalancerL7PolicyResource{}
)

func NewLoadBalancerL7PolicyResource() resource.Resource {
	return &loadBalancerL7PolicyResource{}
}

type loadBalancerL7PolicyResource struct {
	kc *common.KakaoCloudClient
}

func (r *loadBalancerL7PolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *loadBalancerL7PolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_load_balancer_l7_policy"
}

func (r *loadBalancerL7PolicyResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: utils.MergeResourceSchemaAttributes(
			loadBalancerL7PolicyResourceSchemaAttributes,
			map[string]schema.Attribute{
				"timeouts": timeouts.AttributesAll(ctx),
			},
		),
	}
}

func (r *loadBalancerL7PolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan loadBalancerL7PolicyResourceModel
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

	loadBalancerId, ok := r.getLoadBalancerIdByListenerId(ctx, plan.ListenerId.ValueString(), &resp.Diagnostics)
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

	createReq := loadbalancer.NewCreateL7Policy(
		loadbalancer.L7PolicyAction(plan.Action.ValueString()),
		plan.ListenerId.ValueString(),
	)

	if !plan.Name.IsNull() {
		createReq.SetName(plan.Name.ValueString())
	}

	if !plan.Description.IsNull() {
		createReq.SetDescription(plan.Description.ValueString())
	}

	if !plan.Position.IsNull() && !plan.Position.IsUnknown() {
		ok := r.validatePosition(ctx, *loadBalancerId, plan.ListenerId.ValueString(), common.ActionC, plan.Position.ValueInt32(), &resp.Diagnostics)
		if !ok || resp.Diagnostics.HasError() {
			return
		}

		createReq.SetPosition(plan.Position.ValueInt32())
	}

	if !plan.RedirectTargetGroupId.IsNull() {
		createReq.SetRedirectTargetGroupId(plan.RedirectTargetGroupId.ValueString())
	}

	if !plan.RedirectUrl.IsNull() {
		createReq.SetRedirectUrl(plan.RedirectUrl.ValueString())
	}

	if !plan.RedirectPrefix.IsNull() {
		createReq.SetRedirectPrefix(plan.RedirectPrefix.ValueString())
	}

	body := *loadbalancer.NewCreateL7PolicyRequest(*createReq)

	createResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*loadbalancer.CreateL7PolicyResponse, *http.Response, error) {
			return r.kc.ApiClient.LoadBalancerL7PolicyAPI.CreateL7Policy(ctx).XAuthToken(r.kc.XAuthToken).CreateL7PolicyRequest(body).Execute()
		},
	)

	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "CreateL7Policy", err, &resp.Diagnostics)
		return
	}

	plan.Id = utils.ConvertNullableString(createResp.L7Policy.Id)

	result, ok := r.pollL7PolicyUntilStatus(
		ctx,
		plan.Id.ValueString(),
		loadBalancerAvailableAndErrorStatuses,
		&resp.Diagnostics,
	)

	if !ok || resp.Diagnostics.HasError() {
		return
	}

	common.CheckResourceAvailableStatus(ctx, r, (*string)(result.ProvisioningStatus.Get()), []string{common.LoadBalancerProvisioningStatusActive}, &resp.Diagnostics)

	ok = mapLoadBalancerL7PolicyFromGetResponse(ctx, &plan.loadBalancerL7PolicyBaseModel, result, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *loadBalancerL7PolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state loadBalancerL7PolicyResourceModel
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

	respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*loadbalancer.GetL7PolicyResponse, *http.Response, error) {
			return r.kc.ApiClient.LoadBalancerL7PolicyAPI.GetL7Policy(ctx, state.Id.ValueString()).XAuthToken(r.kc.XAuthToken).Execute()
		},
	)

	if httpResp != nil && httpResp.StatusCode == 404 {
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "GetL7Policy", err, &resp.Diagnostics)
		return
	}

	ok := mapLoadBalancerL7PolicyFromGetResponse(ctx, &state.loadBalancerL7PolicyBaseModel, new(respModel.L7Policy), &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	if state.ListenerId.IsNull() {
		_, foundListenerId, ok := findLbAndListenerByL7PolicyId(ctx, r, r.kc, state.Id.ValueString(), &resp.Diagnostics)
		if !ok {
			return
		}

		state.ListenerId = types.StringValue(*foundListenerId)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *loadBalancerL7PolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state loadBalancerL7PolicyResourceModel
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

	loadBalancerId, ok := r.getLoadBalancerIdByListenerId(ctx, state.ListenerId.ValueString(), &resp.Diagnostics)
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

	action := plan.Action.ValueString()

	editReq := loadbalancer.NewUpdateL7Policy(loadbalancer.L7PolicyAction(action))

	if !plan.Name.IsNull() {
		editReq.SetName(plan.Name.ValueString())
	}

	if !plan.Description.IsNull() {
		editReq.SetDescription(plan.Description.ValueString())
	}

	if !plan.Position.IsNull() && !plan.Position.IsUnknown() {
		ok := r.validatePosition(ctx, *loadBalancerId, plan.ListenerId.ValueString(), common.ActionU, plan.Position.ValueInt32(), &resp.Diagnostics)
		if !ok || resp.Diagnostics.HasError() {
			return
		}

		editReq.SetPosition(plan.Position.ValueInt32())
	}

	if !plan.RedirectUrl.IsNull() {
		editReq.SetRedirectUrl(plan.RedirectUrl.ValueString())
	}
	if !plan.RedirectTargetGroupId.IsNull() {
		editReq.SetRedirectTargetGroupId(plan.RedirectTargetGroupId.ValueString())
	}
	if !plan.RedirectPrefix.IsNull() {
		editReq.SetRedirectPrefix(plan.RedirectPrefix.ValueString())
	}

	body := *loadbalancer.NewUpdateL7PolicyRequest(*editReq)

	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*loadbalancer.UpdateL7PolicyResponse, *http.Response, error) {
			return r.kc.ApiClient.LoadBalancerL7PolicyAPI.UpdateL7Policy(ctx, state.Id.ValueString()).XAuthToken(r.kc.XAuthToken).UpdateL7PolicyRequest(body).Execute()
		},
	)

	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "UpdateL7Policy", err, &resp.Diagnostics)
		return
	}

	time.Sleep(5 * time.Second)

	result, ok := r.pollL7PolicyUntilStatus(
		ctx,
		plan.Id.ValueString(),
		loadBalancerAvailableAndErrorStatuses,
		&resp.Diagnostics,
	)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	common.CheckResourceAvailableStatus(ctx, r, (*string)(result.ProvisioningStatus.Get()), []string{common.LoadBalancerProvisioningStatusActive}, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	ok = mapLoadBalancerL7PolicyFromGetResponse(ctx, &plan.loadBalancerL7PolicyBaseModel, result, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *loadBalancerL7PolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state loadBalancerL7PolicyResourceModel
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

	loadBalancerId, ok := r.getLoadBalancerIdByListenerId(ctx, state.ListenerId.ValueString(), &resp.Diagnostics)
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
		func() (interface{}, *http.Response, error) {
			httpResp, err := r.kc.ApiClient.LoadBalancerL7PolicyAPI.DeleteL7Policy(ctx, state.Id.ValueString()).XAuthToken(r.kc.XAuthToken).Execute()
			return nil, httpResp, err
		},
	)

	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "DeleteL7Policy", err, &resp.Diagnostics)
		return
	}

	common.PollUntilDeletion(ctx, r, 5*time.Second, &resp.Diagnostics, func(ctx context.Context) (bool, *http.Response, error) {
		_, httpResp, err := r.kc.ApiClient.LoadBalancerL7PolicyAPI.
			GetL7Policy(ctx, state.Id.ValueString()).
			XAuthToken(r.kc.XAuthToken).
			Execute()
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			return true, httpResp, nil
		}
		return false, httpResp, err
	})
}

func (r *loadBalancerL7PolicyResource) pollL7PolicyUntilStatus(
	ctx context.Context,
	l7PolicyId string,
	targetStatuses []string,
	resp *diag.Diagnostics,
) (*loadbalancer.L7Policy, bool) {
	return common.PollUntilResult(
		ctx,
		r,
		5*time.Second,
		"l7 policy",
		l7PolicyId,
		targetStatuses,
		resp,
		func(ctx context.Context) (*loadbalancer.L7Policy, *http.Response, error) {
			respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, resp,
				func() (*loadbalancer.GetL7PolicyResponse, *http.Response, error) {
					return r.kc.ApiClient.LoadBalancerL7PolicyAPI.
						GetL7Policy(ctx, l7PolicyId).
						XAuthToken(r.kc.XAuthToken).
						Execute()
				},
			)
			if err != nil {
				return nil, httpResp, err
			}
			return &respModel.L7Policy, httpResp, nil
		},
		func(policy *loadbalancer.L7Policy) string {
			return string(*policy.ProvisioningStatus.Get())
		},
	)
}

func (r *loadBalancerL7PolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *loadBalancerL7PolicyResource) getLoadBalancerIdByListenerId(ctx context.Context, listenerId string, respDiags *diag.Diagnostics) (*string, bool) {
	listenerResult, ok := common.PollUntilResult(
		ctx,
		r,
		3*time.Second,
		"listener",
		listenerId,
		[]string{"ok"},
		respDiags,
		func(ctx context.Context) (*loadbalancer.LoadBalancerListener, *http.Response, error) {
			respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, respDiags,
				func() (*loadbalancer.GetListenerResponse, *http.Response, error) {
					return r.kc.ApiClient.LoadBalancerListenerAPI.GetListener(ctx, listenerId).
						XAuthToken(r.kc.XAuthToken).Execute()
				},
			)
			if err != nil {
				return nil, httpResp, err
			}
			return &respModel.Listener, httpResp, nil
		},
		func(v *loadbalancer.LoadBalancerListener) string {
			return "ok"
		},
	)
	if !ok {
		return nil, false
	}

	return listenerResult.LoadBalancerId.Get(), ok
}

func (r *loadBalancerL7PolicyResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config loadBalancerL7PolicyResourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	action := config.Action.ValueString()

	switch action {
	case string(loadbalancer.L7POLICYACTION_REDIRECT_TO_URL):
		if config.RedirectUrl.IsNull() {
			common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
				fmt.Sprintf("'redirect_url' is required for action type '%v'.", action),
			)
		}
		if !config.RedirectTargetGroupId.IsNull() || !config.RedirectPrefix.IsNull() {
			common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
				fmt.Sprintf("'redirect_target_group_id' and 'redirect_prefix' must not be set for action type '%v'.", action),
			)
		}
	case string(loadbalancer.L7POLICYACTION_REDIRECT_TO_POOL):
		if config.RedirectTargetGroupId.IsNull() {
			common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
				fmt.Sprintf("'redirect_target_group_id' is required for action type '%v'.", action),
			)
		}
		if !config.RedirectUrl.IsNull() || !config.RedirectPrefix.IsNull() {
			common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
				fmt.Sprintf("'redirect_url' and 'redirect_prefix' must not be set for action type '%v'.", action),
			)
		}
	case string(loadbalancer.L7POLICYACTION_REDIRECT_PREFIX):
		if config.RedirectPrefix.IsNull() {
			common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
				fmt.Sprintf("'redirect_prefix' is required for action type '%v'.", action),
			)
		}
		if !config.RedirectUrl.IsNull() || !config.RedirectTargetGroupId.IsNull() {
			common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
				fmt.Sprintf("'redirect_url' and 'redirect_target_group_id' must not be set for action type '%v'.", action),
			)
		}
	}
}

func (r *loadBalancerL7PolicyResource) validatePosition(ctx context.Context, lbId, listenerId, action string, position int32, respDiags *diag.Diagnostics) bool {
	lbL7PoliciesResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, respDiags,
		func() (*loadbalancer.ListL7PoliciesResponse, *http.Response, error) {
			return r.kc.ApiClient.LoadBalancerL7PolicyAPI.ListL7Policies(
				ctx,
				lbId,
				listenerId,
			).Limit(1000).XAuthToken(r.kc.XAuthToken).Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "ListL7Policies", err, respDiags)
		return false
	}

	availablePosition := int32(len(lbL7PoliciesResp.L7Policies))

	if action == common.ActionC {
		availablePosition = availablePosition + 1
	}

	if position > availablePosition {
		common.AddValidationConfigError(ctx, r, respDiags,
			"Invalid position. Position must be less than or equal to "+strconv.Itoa(int(availablePosition))+".")
		return false
	}

	return true
}

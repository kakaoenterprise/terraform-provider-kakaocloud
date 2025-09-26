// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package loadbalancer

import (
	"context"
	"fmt"
	"net/http"
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
		Description: docs.GetResourceDescription("LoadBalancerL7Policy"),
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

	createReq := loadbalancer.CreateL7PolicyModel{
		ListenerId: plan.ListenerId.ValueString(),
		Action:     loadbalancer.L7PolicyAction(plan.Action.ValueString()),
	}

	if !plan.Name.IsNull() && !plan.Name.IsUnknown() {
		createReq.SetName(plan.Name.ValueString())
	}

	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		createReq.SetDescription(plan.Description.ValueString())
	}

	if !plan.Position.IsNull() && !plan.Position.IsUnknown() {
		createReq.SetPosition(int32(plan.Position.ValueInt64()))
	}

	if !plan.RedirectTargetGroupId.IsNull() && !plan.RedirectTargetGroupId.IsUnknown() {
		createReq.SetRedirectTargetGroupId(plan.RedirectTargetGroupId.ValueString())
	}

	if !plan.RedirectUrl.IsNull() && !plan.RedirectUrl.IsUnknown() {
		createReq.SetRedirectUrl(plan.RedirectUrl.ValueString())
	}

	if !plan.RedirectPrefix.IsNull() && !plan.RedirectPrefix.IsUnknown() {
		createReq.SetRedirectPrefix(plan.RedirectPrefix.ValueString())
	}

	body := loadbalancer.BodyCreateL7Policy{
		L7Policy: createReq,
	}

	createResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*loadbalancer.BnsLoadBalancerV1ApiCreateL7PolicyModelResponseL7PolicyModel, *http.Response, error) {
			return r.kc.ApiClient.LoadBalancerL7PoliciesAPI.CreateL7Policy(ctx).XAuthToken(r.kc.XAuthToken).BodyCreateL7Policy(body).Execute()
		},
	)

	if httpResp != nil && httpResp.StatusCode == http.StatusConflict {
		createResp, httpResp, err = ExecuteWithLoadBalancerConflictRetry(ctx, r.kc, &resp.Diagnostics,
			func() (*loadbalancer.BnsLoadBalancerV1ApiCreateL7PolicyModelResponseL7PolicyModel, *http.Response, error) {
				return r.kc.ApiClient.LoadBalancerL7PoliciesAPI.CreateL7Policy(ctx).XAuthToken(r.kc.XAuthToken).BodyCreateL7Policy(body).Execute()
			},
		)
	}

	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "CreateL7Policy", err, &resp.Diagnostics)
		return
	}

	plan.Id = types.StringValue(createResp.L7Policy.Id)

	result, ok := r.pollL7PolicyUntilStatus(
		ctx,
		plan.Id.ValueString(),
		[]string{ProvisioningStatusActive, ProvisioningStatusError},
		&resp.Diagnostics,
	)

	if !ok || resp.Diagnostics.HasError() {
		return
	}

	common.CheckResourceAvailableStatus(ctx, r, (*string)(result.ProvisioningStatus.Get()), []string{ProvisioningStatusActive}, &resp.Diagnostics)

	listenerIdFromPlan := plan.ListenerId
	ok = mapLoadBalancerL7PolicyFromGetResponse(ctx, &plan.loadBalancerL7PolicyBaseModel, result, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}
	plan.ListenerId = listenerIdFromPlan

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
		func() (*loadbalancer.BnsLoadBalancerV1ApiGetL7PolicyModelResponseL7PolicyModel, *http.Response, error) {
			return r.kc.ApiClient.LoadBalancerL7PoliciesAPI.GetL7Policy(ctx, state.Id.ValueString()).XAuthToken(r.kc.XAuthToken).Execute()
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

	l7PolicyResult := respModel.L7Policy

	listenerIdFromState := state.ListenerId
	ok := mapLoadBalancerL7PolicyFromGetResponse(ctx, &state.loadBalancerL7PolicyBaseModel, &l7PolicyResult, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}
	state.ListenerId = listenerIdFromState

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

	editReq := loadbalancer.NewEditL7PolicyModel(loadbalancer.L7PolicyAction(plan.Action.ValueString()))

	if !plan.Name.Equal(state.Name) {
		editReq.SetName(plan.Name.ValueString())
	}

	if !plan.Description.Equal(state.Description) {
		editReq.SetDescription(plan.Description.ValueString())
	}

	if !plan.Position.IsNull() && !plan.Position.IsUnknown() {

		editReq.SetPosition(int32(plan.Position.ValueInt64()))
	} else {

		if !state.Position.IsNull() && !state.Position.IsUnknown() {
			editReq.SetPosition(int32(state.Position.ValueInt64()))
		}
	}

	action := plan.Action.ValueString()

	switch action {
	case "REDIRECT_TO_URL":

		editReq.SetRedirectUrl(plan.RedirectUrl.ValueString())
		editReq.SetRedirectTargetGroupIdNil()
		editReq.SetRedirectPrefixNil()

	case "REDIRECT_TO_POOL":

		editReq.SetRedirectTargetGroupId(plan.RedirectTargetGroupId.ValueString())
		editReq.SetRedirectUrlNil()
		editReq.SetRedirectPrefixNil()

	case "REDIRECT_PREFIX":

		editReq.SetRedirectPrefix(plan.RedirectPrefix.ValueString())
		editReq.SetRedirectUrlNil()
		editReq.SetRedirectTargetGroupIdNil()
	}

	body := *loadbalancer.NewBodyUpdateL7Policy(*editReq)

	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*loadbalancer.BnsLoadBalancerV1ApiUpdateL7PolicyModelResponseL7PolicyModel, *http.Response, error) {
			return r.kc.ApiClient.LoadBalancerL7PoliciesAPI.UpdateL7Policy(ctx, state.Id.ValueString()).XAuthToken(r.kc.XAuthToken).BodyUpdateL7Policy(body).Execute()
		},
	)

	if httpResp != nil && httpResp.StatusCode == http.StatusConflict {
		_, httpResp, err = ExecuteWithLoadBalancerConflictRetry(ctx, r.kc, &resp.Diagnostics,
			func() (*loadbalancer.BnsLoadBalancerV1ApiUpdateL7PolicyModelResponseL7PolicyModel, *http.Response, error) {
				return r.kc.ApiClient.LoadBalancerL7PoliciesAPI.UpdateL7Policy(ctx, state.Id.ValueString()).XAuthToken(r.kc.XAuthToken).BodyUpdateL7Policy(body).Execute()
			},
		)
	}

	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "UpdateL7Policy", err, &resp.Diagnostics)
		return
	}

	result, ok := r.pollL7PolicyUntilStatus(
		ctx,
		plan.Id.ValueString(),
		[]string{ProvisioningStatusActive, ProvisioningStatusError},
		&resp.Diagnostics,
	)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	common.CheckResourceAvailableStatus(ctx, r, (*string)(result.ProvisioningStatus.Get()), []string{ProvisioningStatusActive}, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	listenerIdFromPlan := plan.ListenerId
	rulesFromState := state.Rules
	ok = mapLoadBalancerL7PolicyFromGetResponse(ctx, &plan.loadBalancerL7PolicyBaseModel, result, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}
	plan.ListenerId = listenerIdFromPlan
	plan.Rules = rulesFromState

	plan.Timeouts = state.Timeouts
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

	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (interface{}, *http.Response, error) {
			httpResp, err := r.kc.ApiClient.LoadBalancerL7PoliciesAPI.DeleteL7Policy(ctx, state.Id.ValueString()).XAuthToken(r.kc.XAuthToken).Execute()
			return nil, httpResp, err
		},
	)

	if httpResp != nil && httpResp.StatusCode == http.StatusConflict {
		_, httpResp, err = ExecuteWithLoadBalancerConflictRetry(ctx, r.kc, &resp.Diagnostics,
			func() (interface{}, *http.Response, error) {
				httpResp, err := r.kc.ApiClient.LoadBalancerL7PoliciesAPI.DeleteL7Policy(ctx, state.Id.ValueString()).XAuthToken(r.kc.XAuthToken).Execute()
				return nil, httpResp, err
			},
		)
	}

	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "DeleteL7Policy", err, &resp.Diagnostics)
		return
	}

	common.PollUntilDeletion(ctx, r, 5*time.Second, &resp.Diagnostics, func(ctx context.Context) (bool, *http.Response, error) {
		_, httpResp, err := r.kc.ApiClient.LoadBalancerL7PoliciesAPI.
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
) (*loadbalancer.BnsLoadBalancerV1ApiGetL7PolicyModelL7PolicyModel, bool) {
	return common.PollUntilResult(
		ctx,
		r,
		5*time.Second,
		targetStatuses,
		resp,
		func(ctx context.Context) (*loadbalancer.BnsLoadBalancerV1ApiGetL7PolicyModelL7PolicyModel, *http.Response, error) {
			respModel, httpResp, err := r.kc.ApiClient.LoadBalancerL7PoliciesAPI.
				GetL7Policy(ctx, l7PolicyId).
				XAuthToken(r.kc.XAuthToken).
				Execute()
			if err != nil {
				return nil, httpResp, err
			}
			return &respModel.L7Policy, httpResp, nil
		},
		func(policy *loadbalancer.BnsLoadBalancerV1ApiGetL7PolicyModelL7PolicyModel) string {
			return string(*policy.ProvisioningStatus.Get())
		},
	)
}

func (r *loadBalancerL7PolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

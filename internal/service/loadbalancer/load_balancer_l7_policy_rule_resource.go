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
	"github.com/kakaoenterprise/kc-sdk-go/v2/services/loadbalancer"
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

const l7PolicyRuleSynchronizedStatus = "SYNCHRONIZED"

type l7PolicyRuleExpectedState struct {
	compareType string
	ruleType    string
	key         *types.String
	value       string
	isInverted  *bool
}

func (r *loadBalancerL7PolicyRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_load_balancer_l7_policy_rule"
}

func (r *loadBalancerL7PolicyRuleResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
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
	r.validateL7PolicyRuleConfig(ctx, req, resp)
}

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

	loadBalancerId, _, ok := findLbAndListenerByL7PolicyId(ctx, r, r.kc, plan.L7PolicyId.ValueString(), &resp.Diagnostics)
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

	createReq := loadbalancer.NewAddL7PolicyRule(
		loadbalancer.L7RuleCompareType(plan.CompareType.ValueString()),
		loadbalancer.L7RuleType(plan.Type.ValueString()),
		plan.Value.ValueString(),
	)

	if !plan.Key.IsNull() {
		createReq.SetKey(plan.Key.ValueString())
	}
	if !plan.IsInverted.IsNull() {
		createReq.SetIsInverted(plan.IsInverted.ValueBool())
	}

	createResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*loadbalancer.AddL7PolicyRuleResponse, *http.Response, error) {
			return r.kc.ApiClient.LoadBalancerL7RuleAPI.AddL7PolicyRule(ctx, plan.L7PolicyId.ValueString()).AddL7PolicyRuleRequest(*loadbalancer.NewAddL7PolicyRuleRequest(*createReq)).XAuthToken(r.kc.XAuthToken).Execute()
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
		plan.Id.ValueString(),
		loadBalancerAvailableAndErrorStatuses,
		nil,
		&resp.Diagnostics,
	)

	if !ok || resp.Diagnostics.HasError() {
		return
	}

	common.CheckResourceAvailableStatus(ctx, r, (*string)(result.ProvisioningStatus.Get()), []string{common.LoadBalancerProvisioningStatusActive}, &resp.Diagnostics)

	ok = mapLoadBalancerL7PolicyRuleModel(&plan.loadBalancerL7PolicyRuleBaseModel, result, plan.L7PolicyId.ValueString(), &resp.Diagnostics)
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
		func() (*loadbalancer.GetL7PolicyRuleResponse, *http.Response, error) {
			return r.kc.ApiClient.LoadBalancerL7RuleAPI.GetL7PolicyRule(ctx, state.L7PolicyId.ValueString(), state.Id.ValueString()).XAuthToken(r.kc.XAuthToken).Execute()
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

	ok := mapLoadBalancerL7PolicyRuleModel(&state.loadBalancerL7PolicyRuleBaseModel, &getResp.L7Rule, state.L7PolicyId.ValueString(), &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

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

	timeout, diags := plan.Timeouts.Update(ctx, common.DefaultUpdateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	loadBalancerId, _, ok := findLbAndListenerByL7PolicyId(ctx, r, r.kc, state.L7PolicyId.ValueString(), &resp.Diagnostics)
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

	updateReq := loadbalancer.NewUpdateL7PolicyRule(
		loadbalancer.L7RuleCompareType(plan.CompareType.ValueString()),
		loadbalancer.L7RuleType(plan.Type.ValueString()),
		plan.Value.ValueString(),
	)

	if plan.Key.IsNull() {
		updateReq.SetKeyNil()
	} else if !plan.Key.IsUnknown() {
		updateReq.SetKey(plan.Key.ValueString())
	}
	if !plan.IsInverted.IsNull() && !plan.IsInverted.IsUnknown() {
		updateReq.SetIsInverted(plan.IsInverted.ValueBool())
	}

	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*loadbalancer.UpdateL7PolicyRuleResponse, *http.Response, error) {
			return r.kc.ApiClient.LoadBalancerL7RuleAPI.UpdateL7PolicyRule(ctx, state.L7PolicyId.ValueString(), state.Id.ValueString()).UpdateL7PolicyRuleRequest(*loadbalancer.NewUpdateL7PolicyRuleRequest(*updateReq)).XAuthToken(r.kc.XAuthToken).Execute()
		},
	)

	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "UpdateL7PolicyRule", err, &resp.Diagnostics)
		return
	}

	expected := buildL7PolicyRuleExpectedState(&plan)

	result, ok := r.pollL7PolicyRuleUntilStatus(
		ctx,
		state.L7PolicyId.ValueString(),
		state.Id.ValueString(),
		loadBalancerAvailableAndErrorStatuses,
		&expected,
		&resp.Diagnostics,
	)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	common.CheckResourceAvailableStatus(ctx, r, (*string)(result.ProvisioningStatus.Get()), []string{common.LoadBalancerProvisioningStatusActive}, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	ok = mapLoadBalancerL7PolicyRuleModel(&plan.loadBalancerL7PolicyRuleBaseModel, result, state.L7PolicyId.ValueString(), &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

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

	timeout, diags := state.Timeouts.Delete(ctx, common.DefaultDeleteTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	loadBalancerId, _, ok := findLbAndListenerByL7PolicyId(ctx, r, r.kc, state.L7PolicyId.ValueString(), &resp.Diagnostics)
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
			httpResp, err := r.kc.ApiClient.LoadBalancerL7RuleAPI.DeleteL7PolicyRule(ctx, state.L7PolicyId.ValueString(), state.Id.ValueString()).
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
		_, httpResp, err := r.kc.ApiClient.LoadBalancerL7RuleAPI.
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

	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		common.AddImportFormatError(ctx, r, &resp.Diagnostics,
			"Expected import ID in the format: l7_policy_id/rule_id")
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("l7_policy_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

func (r *loadBalancerL7PolicyRuleResource) pollL7PolicyRuleUntilStatus(
	ctx context.Context,
	l7PolicyId, ruleId string,
	targetStatuses []string,
	expected *l7PolicyRuleExpectedState,
	diags *diag.Diagnostics,
) (*loadbalancer.Rule, bool) {
	pollingTargetStatuses := targetStatuses
	if expected != nil {
		pollingTargetStatuses = []string{
			l7PolicyRuleSynchronizedStatus,
			common.LoadBalancerProvisioningStatusError,
		}
	}

	return common.PollUntilResult(
		ctx,
		r,
		2*time.Second,
		"l7 policy rule",
		ruleId,
		pollingTargetStatuses,
		diags,
		func(ctx context.Context) (*loadbalancer.Rule, *http.Response, error) {
			resp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, diags,
				func() (*loadbalancer.GetL7PolicyRuleResponse, *http.Response, error) {
					return r.kc.ApiClient.LoadBalancerL7RuleAPI.
						GetL7PolicyRule(ctx, l7PolicyId, ruleId).
						XAuthToken(r.kc.XAuthToken).
						Execute()
				},
			)
			if err != nil {
				return nil, httpResp, err
			}
			return &resp.L7Rule, httpResp, nil
		},
		func(v *loadbalancer.Rule) string {
			if v == nil || v.ProvisioningStatus.Get() == nil {
				return ""
			}

			status := string(*v.ProvisioningStatus.Get())
			if status != common.LoadBalancerProvisioningStatusActive || expected == nil {
				return status
			}
			if l7PolicyRuleMatchesExpectedState(v, expected) {
				return l7PolicyRuleSynchronizedStatus
			}
			return status
		},
	)
}

func buildL7PolicyRuleExpectedState(plan *loadBalancerL7PolicyRuleResourceModel) l7PolicyRuleExpectedState {
	expected := l7PolicyRuleExpectedState{
		compareType: plan.CompareType.ValueString(),
		ruleType:    plan.Type.ValueString(),
		value:       plan.Value.ValueString(),
	}
	if !plan.Key.IsUnknown() {
		expected.key = &plan.Key
	}
	if !plan.IsInverted.IsNull() && !plan.IsInverted.IsUnknown() {
		isInverted := plan.IsInverted.ValueBool()
		expected.isInverted = &isInverted
	}
	return expected
}

func l7PolicyRuleMatchesExpectedState(rule *loadbalancer.Rule, expected *l7PolicyRuleExpectedState) bool {
	if rule.CompareType.Get() == nil || string(*rule.CompareType.Get()) != expected.compareType {
		return false
	}
	if rule.Type.Get() == nil || string(*rule.Type.Get()) != expected.ruleType {
		return false
	}
	if rule.Value.Get() == nil || *rule.Value.Get() != expected.value {
		return false
	}
	if expected.key != nil && !utils.ConvertNullableString(rule.Key).Equal(*expected.key) {
		return false
	}
	if expected.isInverted != nil && rule.IsInverted != *expected.isInverted {
		return false
	}
	return true
}

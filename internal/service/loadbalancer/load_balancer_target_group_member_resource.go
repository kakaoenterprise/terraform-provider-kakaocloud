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
	_ resource.ResourceWithConfigure   = &loadBalancerTargetGroupMemberResource{}
	_ resource.ResourceWithImportState = &loadBalancerTargetGroupMemberResource{}
)

func NewLoadBalancerTargetGroupMemberResource() resource.Resource {
	return &loadBalancerTargetGroupMemberResource{}
}

type loadBalancerTargetGroupMemberResource struct {
	kc *common.KakaoCloudClient
}

func (r *loadBalancerTargetGroupMemberResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_load_balancer_target_group_member"
}

func (r *loadBalancerTargetGroupMemberResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: utils.MergeResourceSchemaAttributes(
			loadBalancerTargetGroupMemberResourceSchema,
			map[string]schema.Attribute{
				"timeouts": timeouts.AttributesAll(ctx),
			},
		),
	}
}

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

	createReq := mapLoadBalancerTargetGroupMemberToCreateRequest(&plan)

	body := loadbalancer.NewAddTargetRequest(*createReq)

	respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*loadbalancer.AddTargetResponse, *http.Response, error) {
			return r.kc.ApiClient.LoadBalancerTargetGroupAPI.
				AddTarget(ctx, plan.TargetGroupId.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				AddTargetRequest(*body).
				Execute()
		},
	)

	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "AddTarget", err, &resp.Diagnostics)
		return
	}

	expected := buildTargetGroupMemberExpectedState(&plan)

	result, ok := r.pollTargetGroupMemberUntilStatus(
		ctx,
		plan.TargetGroupId.ValueString(),
		respModel.Member.Id,
		loadBalancerAvailableAndErrorStatuses,
		expected,
		&resp.Diagnostics,
	)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	common.CheckResourceAvailableStatus(ctx, r, (*string)(result.ProvisioningStatus.Get()), []string{common.LoadBalancerProvisioningStatusActive}, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	ok = mapLoadBalancerTargetGroupMemberFromGetResponse(&plan, result)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

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

	respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*loadbalancer.GetTargetResponse, *http.Response, error) {
			return r.kc.ApiClient.LoadBalancerTargetGroupAPI.
				GetTarget(ctx, state.TargetGroupId.ValueString(), state.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				Execute()
		},
	)

	if httpResp != nil && httpResp.StatusCode == 404 {
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "GetTarget", err, &resp.Diagnostics)
		return
	}

	ok := mapLoadBalancerTargetGroupMemberFromGetResponse(&state, &respModel.Member)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

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

	updateReq := loadbalancer.NewUpdateTarget()

	if !plan.Name.Equal(state.Name) {
		if !plan.Name.IsNull() && !plan.Name.IsUnknown() {
			updateReq.SetName(plan.Name.ValueString())
		} else {
			updateReq.SetNameNil()
		}
	}

	if !plan.Weight.Equal(state.Weight) && !plan.Weight.IsNull() && !plan.Weight.IsUnknown() {
		updateReq.SetWeight(plan.Weight.ValueInt32())
	}

	if !plan.MonitorPort.Equal(state.MonitorPort) && !plan.MonitorPort.IsNull() && !plan.MonitorPort.IsUnknown() {
		updateReq.SetMonitorPort(plan.MonitorPort.ValueInt32())
	}

	body := loadbalancer.NewUpdateTargetRequest(*updateReq)

	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*loadbalancer.UpdateTargetResponse, *http.Response, error) {
			return r.kc.ApiClient.LoadBalancerTargetGroupAPI.
				UpdateTarget(ctx, state.TargetGroupId.ValueString(), state.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				UpdateTargetRequest(*body).
				Execute()
		},
	)

	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "UpdateTarget", err, &resp.Diagnostics)
		return
	}

	expected := buildTargetGroupMemberExpectedState(&plan)

	result, ok := r.pollTargetGroupMemberUntilStatus(
		ctx,
		state.TargetGroupId.ValueString(),
		state.Id.ValueString(),
		loadBalancerAvailableAndErrorStatuses,
		expected,
		&resp.Diagnostics,
	)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	common.CheckResourceAvailableStatus(ctx, r, (*string)(result.ProvisioningStatus.Get()), []string{common.LoadBalancerProvisioningStatusActive}, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	ok = mapLoadBalancerTargetGroupMemberFromGetResponse(&plan, result)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

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
		func() (interface{}, *http.Response, error) {
			httpResp, err := r.kc.ApiClient.LoadBalancerTargetGroupAPI.
				RemoveTarget(ctx, state.TargetGroupId.ValueString(), state.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				Execute()
			return nil, httpResp, err
		},
	)

	if httpResp != nil && httpResp.StatusCode == 404 {

		return
	}

	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "RemoveTarget", err, &resp.Diagnostics)
		return
	}

	common.PollUntilDeletion(ctx, r, 2*time.Second, &resp.Diagnostics,
		func(ctx context.Context) (bool, *http.Response, error) {
			respModel, httpResp, err := r.kc.ApiClient.LoadBalancerTargetGroupAPI.
				ListTargetsInTargetGroup(ctx, state.TargetGroupId.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				Limit(1000).
				Execute()
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					return true, httpResp, nil
				}
				return false, httpResp, err
			}

			for _, member := range respModel.Members {
				if member.Id.Get() != nil && *member.Id.Get() == state.Id.ValueString() {
					return false, httpResp, nil
				}
			}
			return true, httpResp, nil
		},
	)
}

func (r *loadBalancerTargetGroupMemberResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {

	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		common.AddImportFormatError(ctx, r, &resp.Diagnostics,
			"Expected import ID in the format: target_group_id/member_id")
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("target_group_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

func (r *loadBalancerTargetGroupMemberResource) getLoadBalancerIdByTargetGroupId(ctx context.Context, targetGroupId string, respDiags *diag.Diagnostics) (*string, bool) {
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

func (r *loadBalancerTargetGroupMemberResource) pollTargetGroupMemberUntilStatus(
	ctx context.Context,
	targetGroupID string,
	memberID string,
	targetStatuses []string,
	expected targetGroupMemberExpectedState,
	diags *diag.Diagnostics,
) (*loadbalancer.PoolMember, bool) {
	const (
		pollInterval = 2 * time.Second
	)

	targetStatusSet := make(map[string]struct{}, len(targetStatuses))
	for _, status := range targetStatuses {
		targetStatusSet[status] = struct{}{}
	}

	var lastStatus string
	var lastErr error

	for {
		respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, diags,
			func() (*loadbalancer.GetTargetResponse, *http.Response, error) {
				return r.kc.ApiClient.
					LoadBalancerTargetGroupAPI.
					GetTarget(ctx, targetGroupID, memberID).
					XAuthToken(r.kc.XAuthToken).
					Execute()
			},
		)
		if err != nil {
			lastErr = err

			if httpResp == nil ||
				httpResp.StatusCode != http.StatusNotFound {
				common.AddApiActionError(ctx, r, httpResp, "GetTarget", err, diags,
					fmt.Sprintf("Failed to get target group member %q from target group %q.", memberID, targetGroupID),
				)
				return nil, false
			}
		} else {

			lastErr = nil

			if respModel.Member.ProvisioningStatus.IsSet() &&
				respModel.Member.ProvisioningStatus.Get() != nil {
				lastStatus = string(*respModel.Member.ProvisioningStatus.Get())

				if _, reached := targetStatusSet[lastStatus]; reached && targetGroupMemberMatchesExpectedState(
					&respModel.Member,
					expected,
				) {
					return &respModel.Member, true
				}
			}
		}

		timer := time.NewTimer(pollInterval)

		select {
		case <-timer.C:

		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}

			detail := fmt.Sprintf("Polling target group member %q in target group %q was canceled: %v", memberID, targetGroupID, ctx.Err())

			if lastStatus != "" {
				detail += fmt.Sprintf(". Last observed status: %q.", lastStatus)
			} else if lastErr != nil {
				detail += fmt.Sprintf(". Last polling error: %v.", lastErr)
			}

			diags.AddError("Target group member polling canceled", detail)
			return nil, false
		}
	}

	return nil, false
}

type targetGroupMemberExpectedState struct {
	Name        *types.String
	Weight      *types.Int32
	MonitorPort *types.Int32
}

func buildTargetGroupMemberExpectedState(
	plan *loadBalancerTargetGroupMemberResourceModel,
) targetGroupMemberExpectedState {
	expected := targetGroupMemberExpectedState{}

	if !plan.Name.IsNull() && !plan.Name.IsUnknown() {
		expected.Name = new(plan.Name)
	}

	if !plan.Weight.IsNull() && !plan.Weight.IsUnknown() {
		expected.Weight = new(plan.Weight)
	}

	if !plan.MonitorPort.IsNull() && !plan.MonitorPort.IsUnknown() {
		expected.MonitorPort = new(plan.MonitorPort)
	}

	return expected
}

func targetGroupMemberMatchesExpectedState(
	member *loadbalancer.PoolMember,
	expected targetGroupMemberExpectedState,
) bool {
	if expected.Name != nil {
		actual := utils.ConvertNullableString(member.Name)

		if !expected.Name.Equal(actual) {
			return false
		}
	}

	if expected.Weight != nil {
		actual := utils.ConvertNullableInt32(member.Weight)

		if !expected.Weight.Equal(actual) {
			return false
		}
	}

	if expected.MonitorPort != nil {
		actual := utils.ConvertNullableInt32(member.MonitorPort)

		if !expected.MonitorPort.Equal(actual) {
			return false
		}
	}

	return true
}

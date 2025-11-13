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
	"github.com/kakaoenterprise/kc-sdk-go/services/loadbalancer"
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
		Description: docs.GetResourceDescription("LoadBalancerTargetGroupMember"),
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

	loadBalancerId, ok := r.getLoadBalancerIdByTargetGroupId(ctx, plan.TargetGroupId.ValueString(), &resp.Diagnostics)
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

	createReq := mapLoadBalancerTargetGroupMemberToCreateRequest(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	body := loadbalancer.NewBodyAddTarget(*createReq)

	respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*loadbalancer.BnsLoadBalancerV1ApiAddTargetModelResponseTargetGroupMemberModel, *http.Response, error) {
			return r.kc.ApiClient.LoadBalancerTargetGroupAPI.
				AddTarget(ctx, plan.TargetGroupId.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				BodyAddTarget(*body).
				Execute()
		},
	)

	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "AddTarget", err, &resp.Diagnostics)
		return
	}

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

	ok = mapLoadBalancerTargetGroupMemberFromGetResponse(ctx, &plan, result, &resp.Diagnostics)
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
		func() (*loadbalancer.TargetGroupMemberListModel, *http.Response, error) {
			return r.kc.ApiClient.LoadBalancerTargetGroupAPI.
				ListTargetsInTargetGroup(ctx, state.TargetGroupId.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				Limit(1000).
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

	ok := mapLoadBalancerTargetGroupMemberFromGetResponse(ctx, &state, foundMember, &resp.Diagnostics)
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

	loadBalancerId, ok := r.getLoadBalancerIdByTargetGroupId(ctx, state.TargetGroupId.ValueString(), &resp.Diagnostics)
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

	updateReq := mapLoadBalancerTargetGroupMemberToUpdateRequest(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	body := loadbalancer.NewBodyUpdateTarget(*updateReq)

	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*loadbalancer.BnsLoadBalancerV1ApiUpdateTargetModelResponseTargetGroupMemberModel, *http.Response, error) {
			return r.kc.ApiClient.LoadBalancerTargetGroupAPI.
				UpdateTarget(ctx, state.TargetGroupId.ValueString(), state.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				BodyUpdateTarget(*body).
				Execute()
		},
	)

	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "UpdateTarget", err, &resp.Diagnostics)
		return
	}

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

	ok = mapLoadBalancerTargetGroupMemberFromGetResponse(ctx, &plan, result, &resp.Diagnostics)
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

	loadBalancerId, ok := r.getLoadBalancerIdByTargetGroupId(ctx, state.TargetGroupId.ValueString(), &resp.Diagnostics)
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
				if member.Id == state.Id.ValueString() {
					return false, httpResp, nil
				}
			}
			return true, httpResp, nil
		},
	)
}

func (r *loadBalancerTargetGroupMemberResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {

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

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("target_group_id"), targetGroupId)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), memberId)...)
}

func (r *loadBalancerTargetGroupMemberResource) getLoadBalancerIdByTargetGroupId(ctx context.Context, targetGroupId string, respDiags *diag.Diagnostics) (*string, bool) {
	respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, respDiags,
		func() (*loadbalancer.TargetGroupResponseModel, *http.Response, error) {
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
	targetGroupId string,
	memberId string,
	targetStatuses []string,
	resp *diag.Diagnostics,
) (*loadbalancer.BnsLoadBalancerV1ApiListTargetsInTargetGroupModelTargetGroupMemberModel, bool) {

	var lastErr error
	attemptCount := 0
	maxAttempts := 30

	for {
		select {
		case <-ctx.Done():
			if lastErr != nil {
				resp.AddError("Polling Timeout", fmt.Sprintf("Failed to find target group member after polling: %v", lastErr))
			} else {
				resp.AddError("Polling Timeout", "Context cancelled while polling for target group member status")
			}
			return nil, false
		default:
		}

		attemptCount++

		respModel, _, err := r.kc.ApiClient.LoadBalancerTargetGroupAPI.
			ListTargetsInTargetGroup(ctx, targetGroupId).
			XAuthToken(r.kc.XAuthToken).
			Limit(1000).
			Execute()
		if err != nil {
			lastErr = fmt.Errorf("failed to list targets in target group: %w", err)
			if attemptCount >= maxAttempts {
				resp.AddError("Polling Failed", fmt.Sprintf("Failed to list targets after %d attempts: %v", maxAttempts, lastErr))
				return nil, false
			}
			time.Sleep(2 * time.Second)
			continue
		}

		for _, member := range respModel.Members {
			if member.Id == memberId {

				if member.ProvisioningStatus.IsSet() && member.ProvisioningStatus.Get() != nil {
					status := string(*member.ProvisioningStatus.Get())
					for _, targetStatus := range targetStatuses {
						if status == targetStatus {
							return &member, true
						}
					}
				}

				if attemptCount >= maxAttempts {
					resp.AddError("Polling Timeout", fmt.Sprintf("Target group member found but did not reach target status within timeout"))
					return nil, false
				}
				time.Sleep(2 * time.Second)
				continue
			}
		}

		lastErr = fmt.Errorf("member %s not found in target group %s", memberId, targetGroupId)
		if attemptCount >= maxAttempts {
			resp.AddError("Member Not Found", fmt.Sprintf("Target group member %s not found in target group %s after %d attempts. This may indicate the member creation failed or there is an API consistency issue.", memberId, targetGroupId, maxAttempts))
			return nil, false
		}

		time.Sleep(2 * time.Second)
	}
}

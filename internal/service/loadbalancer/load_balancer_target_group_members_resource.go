// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package loadbalancer

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"sort"
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

var (
	_ resource.ResourceWithConfigure      = &loadBalancerTargetGroupMembersResource{}
	_ resource.ResourceWithImportState    = &loadBalancerTargetGroupMembersResource{}
	_ resource.ResourceWithValidateConfig = &loadBalancerTargetGroupMembersResource{}
)

func NewLoadBalancerTargetGroupMembersResource() resource.Resource {
	return &loadBalancerTargetGroupMembersResource{}
}

type loadBalancerTargetGroupMembersResource struct {
	kc *common.KakaoCloudClient
}

func (r *loadBalancerTargetGroupMembersResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_load_balancer_target_group_members"
}

func (r *loadBalancerTargetGroupMembersResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: utils.MergeResourceSchemaAttributes(
			loadBalancerTargetGroupMemberListResourceSchema,
			map[string]schema.Attribute{
				"timeouts": timeouts.AttributesAll(ctx),
			},
		),
	}
}

func (r *loadBalancerTargetGroupMembersResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *loadBalancerTargetGroupMembersResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan loadBalancerTargetGroupMemberListResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.batchRequest(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *loadBalancerTargetGroupMembersResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state, newState loadBalancerTargetGroupMemberListResourceModel
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

	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "ListTargetsInTargetGroup", err, &resp.Diagnostics)
		return
	}

	newState.TargetGroupId = state.TargetGroupId
	newState.Timeouts = state.Timeouts

	if state.Members != nil {
		for _, member := range state.Members {
			for _, respMember := range respModel.Members {
				if member.Address == utils.ConvertNullableString(respMember.IpAddress) &&
					member.ProtocolPort == utils.ConvertNullableInt32(respMember.ProtocolPort) &&
					member.SubnetId.ValueString() == respMember.Subnet.Id &&
					(member.Name.IsNull() || member.Name == utils.ConvertNullableString(respMember.Name)) &&
					(member.Weight.IsNull() || member.Weight == utils.ConvertNullableInt32(respMember.Weight)) &&
					(member.MonitorPort.IsNull() || member.MonitorPort == utils.ConvertNullableInt32(respMember.MonitorPort)) {
					newState.Members = append(newState.Members, member)
					break
				}
			}
		}
	} else {
		sort.Slice(respModel.Members, func(i, j int) bool {
			ipI := net.ParseIP(*respModel.Members[i].IpAddress.Get())
			ipJ := net.ParseIP(*respModel.Members[j].IpAddress.Get())

			if ipI == nil || ipJ == nil {
				return false
			}
			return bytes.Compare(ipI, ipJ) < 0
		})

		for _, respMember := range respModel.Members {
			member := loadBalancerTargetGroupMemberBatchModel{
				Address:      utils.ConvertNullableString(respMember.IpAddress),
				ProtocolPort: utils.ConvertNullableInt32(respMember.ProtocolPort),
				SubnetId:     types.StringValue(respMember.Subnet.Id),
			}

			name := respMember.Name.Get()
			if name != nil {
				member.Name = types.StringValue(*name)
			}
			weight := respMember.Weight.Get()
			if weight != nil {
				member.Weight = types.Int32Value(*weight)
			}
			monitorPort := respMember.MonitorPort.Get()
			if monitorPort != nil {
				member.MonitorPort = types.Int32Value(*monitorPort)
			}
			newState.Members = append(newState.Members, member)
		}
	}

	diags = resp.State.Set(ctx, &newState)
	resp.Diagnostics.Append(diags...)
}

func (r *loadBalancerTargetGroupMembersResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan loadBalancerTargetGroupMemberListResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.batchRequest(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *loadBalancerTargetGroupMembersResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state loadBalancerTargetGroupMemberListResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan := loadBalancerTargetGroupMemberListResourceModel{
		state.TargetGroupId,
		[]loadBalancerTargetGroupMemberBatchModel{},
		state.Timeouts,
	}

	r.batchRequest(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
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
				for _, deleteMember := range state.Members {
					if deleteMember.Address == utils.ConvertNullableString(member.IpAddress) {
						return false, httpResp, nil
					}
				}
			}
			return true, httpResp, nil
		},
	)
}

func (r *loadBalancerTargetGroupMembersResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {

	resource.ImportStatePassthroughID(ctx, path.Root("target_group_id"), req, resp)
}

func (r *loadBalancerTargetGroupMembersResource) getLoadBalancerIdByTargetGroupId(ctx context.Context, targetGroupId string, respDiags *diag.Diagnostics) (*string, bool) {
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

func (r *loadBalancerTargetGroupMembersResource) batchRequest(ctx context.Context, plan *loadBalancerTargetGroupMemberListResourceModel, resp *diag.Diagnostics) {

	loadBalancerId, ok := r.getLoadBalancerIdByTargetGroupId(ctx, plan.TargetGroupId.ValueString(), resp)
	if !ok || resp.HasError() {
		return
	}
	mutex := common.LockForID(*loadBalancerId)
	mutex.Lock()
	defer mutex.Unlock()

	timeout, diags := plan.Timeouts.Create(ctx, common.DefaultCreateTimeout)
	resp.Append(diags...)
	if resp.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ok = CheckLoadBalancerStatus(ctx, *loadBalancerId, true, r, r.kc, resp)
	if !ok || resp.HasError() {
		return
	}

	batchReqBody := mapLoadBalancerTargetGroupMembersToBatchRequest(plan)

	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, resp,
		func() (interface{}, *http.Response, error) {
			return r.kc.ApiClient.LoadBalancerTargetGroupAPI.
				UpdateTargets(ctx, plan.TargetGroupId.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				BodyUpdateTargets(*batchReqBody).
				Execute()
		},
	)

	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "UpdateTargets", err, resp)
		return
	}

	time.Sleep(5 * time.Second)

	result, ok := r.pollTargetGroupMembersUntilStatus(
		ctx,
		plan.TargetGroupId.ValueString(),
		plan.Members,
		[]string{common.LoadBalancerProvisioningStatusActive, common.LoadBalancerProvisioningStatusError},
		resp,
	)
	if !ok || resp.HasError() {
		return
	}

	for _, memberResult := range result {
		common.CheckResourceAvailableStatus(ctx, r, (*string)(memberResult.ProvisioningStatus.Get()), []string{common.LoadBalancerProvisioningStatusActive}, resp)
		if resp.HasError() {
			return
		}
	}

	for i := range plan.Members {
		for _, resultMember := range result {
			if plan.Members[i].Address == utils.ConvertNullableString(resultMember.IpAddress) {
				plan.Members[i].Name = utils.ConvertNullableString(resultMember.Name)
				plan.Members[i].Weight = utils.ConvertNullableInt32(resultMember.Weight)
				plan.Members[i].MonitorPort = utils.ConvertNullableInt32(resultMember.MonitorPort)
			}
		}
	}
}

func (r *loadBalancerTargetGroupMembersResource) pollTargetGroupMembersUntilStatus(
	ctx context.Context,
	targetGroupId string,
	members []loadBalancerTargetGroupMemberBatchModel,
	targetStatuses []string,
	resp *diag.Diagnostics,
) ([]loadbalancer.BnsLoadBalancerV1ApiListTargetsInTargetGroupModelTargetGroupMemberModel, bool) {

	var lastErr error
	attemptCount := 0
	maxAttempts := 900

	expected := make(map[string]struct{})
	for _, m := range members {
		key := fmt.Sprintf(
			"%s:%d:%s",
			m.Address,
			m.ProtocolPort,
			m.SubnetId.ValueString(),
		)
		expected[key] = struct{}{}
	}

	if len(expected) == 0 {
		return nil, true
	}

	for {
		select {
		case <-ctx.Done():
			if lastErr != nil {
				resp.AddError(
					"Polling Timeout",
					fmt.Sprintf("Failed while polling target group members: %v", lastErr),
				)
			} else {
				resp.AddError(
					"Polling Timeout",
					"Context cancelled while polling for target group members status",
				)
			}
			return nil, false
		default:
		}

		attemptCount++

		respModel, _, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, resp,
			func() (*loadbalancer.TargetGroupMemberListModel, *http.Response, error) {
				return r.kc.ApiClient.LoadBalancerTargetGroupAPI.
					ListTargetsInTargetGroup(ctx, targetGroupId).
					XAuthToken(r.kc.XAuthToken).
					Limit(1000).
					Execute()
			},
		)
		if err != nil {
			lastErr = fmt.Errorf("failed to list targets in target group: %w", err)
			if attemptCount >= maxAttempts {
				resp.AddError(
					"Polling Failed",
					fmt.Sprintf("Failed to list targets after %d attempts: %v", maxAttempts, lastErr),
				)
				return nil, false
			}
			time.Sleep(2 * time.Second)
			continue
		}

		matched := 0

		var matchedResp []loadbalancer.BnsLoadBalancerV1ApiListTargetsInTargetGroupModelTargetGroupMemberModel

		for _, apiMember := range respModel.Members {
			key := fmt.Sprintf(
				"%s:%d:%s",
				utils.ConvertNullableString(apiMember.IpAddress),
				utils.ConvertNullableInt32(apiMember.ProtocolPort),
				apiMember.Subnet.Id,
			)

			if _, ok := expected[key]; !ok {
				continue
			}

			if apiMember.ProvisioningStatus.IsSet() && apiMember.ProvisioningStatus.Get() != nil {
				status := string(*apiMember.ProvisioningStatus.Get())
				for _, targetStatus := range targetStatuses {
					if status == targetStatus {
						matched++
						matchedResp = append(matchedResp, apiMember)
						break
					}
				}
			}
		}

		if matched == len(expected) {
			return matchedResp, true
		}

		if attemptCount >= maxAttempts {
			resp.AddError(
				"Polling Timeout",
				fmt.Sprintf(
					"Not all target group members reached desired status within timeout (matched %d / %d)",
					matched,
					len(expected),
				),
			)
			return nil, false
		}

		time.Sleep(2 * time.Second)
	}
}

func (r *loadBalancerTargetGroupMembersResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config loadBalancerTargetGroupMemberListResourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if len(config.Members) == 0 {
		return
	}

	seen := make(map[string]struct{})

	for _, member := range config.Members {
		address := member.Address.ValueString()
		if _, exists := seen[address]; exists {
			common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
				fmt.Sprintf("address '%s' is duplicated.", address),
			)
			return
		}

		seen[address] = struct{}{}
	}
}

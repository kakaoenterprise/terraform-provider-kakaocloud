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
	_ resource.ResourceWithConfigure   = &loadBalancerTargetGroupResource{}
	_ resource.ResourceWithImportState = &loadBalancerTargetGroupResource{}
)

func NewLoadBalancerTargetGroupResource() resource.Resource {
	return &loadBalancerTargetGroupResource{}
}

type loadBalancerTargetGroupResource struct {
	kc *common.KakaoCloudClient
}

func (r *loadBalancerTargetGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_load_balancer_target_group"
}

func (r *loadBalancerTargetGroupResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: docs.GetResourceDescription("LoadBalancerTargetGroup"),
		Attributes: utils.MergeResourceSchemaAttributes(
			loadBalancerTargetGroupResourceSchema,
			map[string]schema.Attribute{
				"timeouts": timeouts.AttributesAll(ctx),
			},
		),
	}
}

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

	createReq := mapLoadBalancerTargetGroupToCreateRequest(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	body := loadbalancer.BodyCreateTargetGroup{TargetGroup: *createReq}

	respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*loadbalancer.BnsLoadBalancerV1ApiCreateTargetGroupModelResponseTargetGroupModel, *http.Response, error) {
			return r.kc.ApiClient.LoadBalancerTargetGroupAPI.
				CreateTargetGroup(ctx).
				XAuthToken(r.kc.XAuthToken).
				BodyCreateTargetGroup(body).
				Execute()
		},
	)

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

	plan.Id = types.StringValue(respModel.TargetGroup.Id)

	ok := mapLoadBalancerTargetGroupFromCreateResponse(ctx, &plan, &respModel.TargetGroup, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	result, ok := r.pollTargetGroupUntilStatus(
		ctx,
		plan.Id.ValueString(),
		[]string{ProvisioningStatusActive, ProvisioningStatusError},
		&resp.Diagnostics,
	)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	if result.ProvisioningStatus.IsSet() && result.ProvisioningStatus.Get() != nil && string(*result.ProvisioningStatus.Get()) == "ERROR" {
		resp.Diagnostics.AddError(
			"Target Group Creation Failed",
			fmt.Sprintf("Target group %s failed to create and is in ERROR state", plan.Id.ValueString()),
		)
		return
	}

	common.CheckResourceAvailableStatus(ctx, r, (*string)(result.ProvisioningStatus.Get()), []string{ProvisioningStatusActive}, &resp.Diagnostics)

	originalHealthMonitor := plan.HealthMonitor
	originalSessionPersistence := plan.SessionPersistence

	ok = mapLoadBalancerTargetGroupFromGetResponse(ctx, &plan, result, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	if plan.HealthMonitor.IsNull() && !originalHealthMonitor.IsNull() && !originalHealthMonitor.IsUnknown() {
		plan.HealthMonitor = preserveHealthMonitorUserConfig(ctx, originalHealthMonitor)
	}

	if plan.SessionPersistence.IsNull() && !originalSessionPersistence.IsNull() && !originalSessionPersistence.IsUnknown() {
		plan.SessionPersistence = originalSessionPersistence
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
}

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

	ok := mapLoadBalancerTargetGroupFromGetResponse(ctx, &state, &respModel.TargetGroup, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

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

	updateReq := mapLoadBalancerTargetGroupToUpdateRequest(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	body := loadbalancer.NewBodyUpdateTargetGroup(*updateReq)

	respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*loadbalancer.BnsLoadBalancerV1ApiUpdateTargetGroupModelResponseTargetGroupModel, *http.Response, error) {
			return r.kc.ApiClient.LoadBalancerTargetGroupAPI.
				UpdateTargetGroup(ctx, state.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				BodyUpdateTargetGroup(*body).
				Execute()
		},
	)

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

	ok := mapLoadBalancerTargetGroupFromUpdateResponse(ctx, &plan, &respModel.TargetGroup, &state, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

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

	if result.ProvisioningStatus.IsSet() && result.ProvisioningStatus.Get() != nil && string(*result.ProvisioningStatus.Get()) == "ERROR" {
		resp.Diagnostics.AddError(
			"Target Group Update Failed",
			fmt.Sprintf("Target group %s failed to update and is in ERROR state", plan.Id.ValueString()),
		)
		return
	}

	common.CheckResourceAvailableStatus(ctx, r, (*string)(result.ProvisioningStatus.Get()), []string{ProvisioningStatusActive}, &resp.Diagnostics)

	originalHealthMonitor := plan.HealthMonitor
	originalSessionPersistence := plan.SessionPersistence

	ok = mapLoadBalancerTargetGroupFromGetResponse(ctx, &plan, result, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	if plan.HealthMonitor.IsNull() && !originalHealthMonitor.IsNull() && !originalHealthMonitor.IsUnknown() {
		plan.HealthMonitor = preserveHealthMonitorUserConfig(ctx, originalHealthMonitor)
	}

	if !originalSessionPersistence.IsNull() && !originalSessionPersistence.IsUnknown() {

		plan.SessionPersistence = originalSessionPersistence
	} else {

		plan.SessionPersistence = types.ObjectNull(loadBalancerTargetGroupSessionPersistenceAttrType)
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

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

	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (interface{}, *http.Response, error) {
			httpResp, err := r.kc.ApiClient.LoadBalancerTargetGroupAPI.
				DeleteTargetGroup(ctx, state.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				Execute()
			return nil, httpResp, err
		},
	)

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

		return
	}

	if err != nil {
		common.AddApiActionError(ctx, nil, httpResp, "delete", err, &resp.Diagnostics)
		return
	}

	common.PollUntilDeletion(ctx, r, 2*time.Second, &resp.Diagnostics,
		func(ctx context.Context) (bool, *http.Response, error) {
			_, httpResp, err := r.kc.ApiClient.LoadBalancerTargetGroupAPI.
				GetTargetGroup(ctx, state.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				Execute()
			if httpResp != nil && httpResp.StatusCode == 404 {
				return true, httpResp, nil
			}
			return false, httpResp, err
		},
	)
}

func (r *loadBalancerTargetGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {

	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

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

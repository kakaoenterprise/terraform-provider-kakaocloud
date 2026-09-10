// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package loadbalancer

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-kakaocloud/internal/common"
	"terraform-provider-kakaocloud/internal/utils"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/kakaoenterprise/kc-sdk-go/v2/services/loadbalancer"
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

type targetGroupExpectedState struct {
	Name                  *types.String
	Description           *types.String
	LoadBalancerAlgorithm *types.String
	SessionPersistence    *types.Object
}

const targetGroupSynchronizedStatus = "SYNCHRONIZED"

func (r *loadBalancerTargetGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_load_balancer_target_group"
}

func (r *loadBalancerTargetGroupResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
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

	unlock, ok := common.AcquireLockForID(ctx, plan.LoadBalancerId.ValueString(), &resp.Diagnostics)
	if !ok {
		return
	}
	defer unlock()

	_, ok = pollLoadBalancerStatus(ctx, plan.LoadBalancerId.ValueString(), false, r, r.kc, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	createReq := mapLoadBalancerTargetGroupToCreateRequest(ctx, &plan, &resp.Diagnostics)
	if createReq == nil || resp.Diagnostics.HasError() {
		return
	}

	body := loadbalancer.CreateTargetGroupRequest{TargetGroup: *createReq}

	respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*loadbalancer.CreateTargetGroupResponse, *http.Response, error) {
			return r.kc.ApiClient.LoadBalancerTargetGroupAPI.
				CreateTargetGroup(ctx).
				XAuthToken(r.kc.XAuthToken).
				CreateTargetGroupRequest(body).
				Execute()
		},
	)

	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "CreateTargetGroup", err, &resp.Diagnostics)
		return
	}

	plan.Id = types.StringValue(respModel.TargetGroup.Id)

	result, ok := r.pollTargetGroupUntilStatus(
		ctx,
		plan.Id.ValueString(),
		loadBalancerAvailableAndErrorStatuses,
		nil,
		&resp.Diagnostics,
	)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	common.CheckResourceAvailableStatus(ctx, r, (*string)(result.ProvisioningStatus.Get()), []string{common.LoadBalancerProvisioningStatusActive}, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	ok = mapLoadBalancerTargetGroupFromGetResponse(ctx, &plan.loadBalancerTargetGroupBaseModel, result, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
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
		func() (*loadbalancer.GetTargetGroupResponse, *http.Response, error) {
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
		common.AddApiActionError(ctx, r, httpResp, "GetTargetGroup", err, &resp.Diagnostics)
		return
	}

	ok := mapLoadBalancerTargetGroupFromGetResponse(ctx, &state.loadBalancerTargetGroupBaseModel, &respModel.TargetGroup, &resp.Diagnostics)
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

	unlock, ok := common.AcquireLockForID(ctx, plan.LoadBalancerId.ValueString(), &resp.Diagnostics)
	if !ok {
		return
	}
	defer unlock()

	_, ok = pollLoadBalancerStatus(ctx, plan.LoadBalancerId.ValueString(), true, r, r.kc, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	expected := buildTargetGroupExpectedState(&plan, &state)

	updateReq := loadbalancer.NewUpdateTargetGroup()

	if expected.Name != nil {
		updateReq.SetName(expected.Name.ValueString())
	}

	if expected.Description != nil {
		if expected.Description.IsNull() {
			updateReq.SetDescriptionNil()
		} else if !expected.Description.IsUnknown() {
			updateReq.SetDescription(expected.Description.ValueString())
		}
	}

	if expected.LoadBalancerAlgorithm != nil {
		algorithm := loadbalancer.LoadBalancerPoolAlgorithm(expected.LoadBalancerAlgorithm.ValueString())
		updateReq.SetLoadBalancerAlgorithm(algorithm)
	}

	if expected.SessionPersistence != nil {
		if expected.SessionPersistence.IsNull() {
			updateReq.SetSessionPersistenceNil()
		} else if !expected.SessionPersistence.IsUnknown() {
			var sessionPersistence loadBalancerTargetGroupSessionPersistence
			convertDiags := expected.SessionPersistence.As(ctx, &sessionPersistence, basetypes.ObjectAsOptions{})
			resp.Diagnostics.Append(convertDiags...)
			if resp.Diagnostics.HasError() {
				return
			}

			sessionPersistenceReq := loadbalancer.NewSessionPersistenceRequest(
				loadbalancer.LoadBalancerSessionPersistenceType(sessionPersistence.Type.ValueString()),
				int32(sessionPersistence.PersistenceTimeout.ValueInt64()),
			)

			if !sessionPersistence.CookieName.IsNull() && !sessionPersistence.CookieName.IsUnknown() {
				sessionPersistenceReq.SetCookieName(sessionPersistence.CookieName.ValueString())
			}
			if !sessionPersistence.PersistenceGranularity.IsNull() && !sessionPersistence.PersistenceGranularity.IsUnknown() {
				sessionPersistenceReq.SetPersistenceGranularity(sessionPersistence.PersistenceGranularity.ValueString())
			}

			updateReq.SetSessionPersistence(*sessionPersistenceReq)
		}
	}

	body := loadbalancer.NewUpdateTargetGroupRequest(*updateReq)

	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*loadbalancer.UpdateTargetGroupResponse, *http.Response, error) {
			return r.kc.ApiClient.LoadBalancerTargetGroupAPI.
				UpdateTargetGroup(ctx, state.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				UpdateTargetGroupRequest(*body).
				Execute()
		},
	)

	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "UpdateTargetGroup", err, &resp.Diagnostics)
		return
	}

	result, ok := r.pollTargetGroupUntilStatus(
		ctx,
		plan.Id.ValueString(),
		loadBalancerAvailableAndErrorStatuses,
		&expected,
		&resp.Diagnostics,
	)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	common.CheckResourceAvailableStatus(ctx, r, (*string)(result.ProvisioningStatus.Get()), []string{common.LoadBalancerProvisioningStatusActive}, &resp.Diagnostics)

	ok = mapLoadBalancerTargetGroupFromGetResponse(ctx, &plan.loadBalancerTargetGroupBaseModel, result, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
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

	unlock, ok := common.AcquireLockForID(ctx, state.LoadBalancerId.ValueString(), &resp.Diagnostics)
	if !ok {
		return
	}
	defer unlock()

	_, ok = pollLoadBalancerStatus(ctx, state.LoadBalancerId.ValueString(), true, r, r.kc, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (interface{}, *http.Response, error) {
			httpResp, err := r.kc.ApiClient.LoadBalancerTargetGroupAPI.
				DeleteTargetGroup(ctx, state.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				Execute()
			return nil, httpResp, err
		},
	)

	if httpResp != nil && httpResp.StatusCode == 404 {

		return
	}

	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "DeleteTargetGroup", err, &resp.Diagnostics)
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
	targetGroupID string,
	targetStatuses []string,
	expected *targetGroupExpectedState,
	diags *diag.Diagnostics,
) (*loadbalancer.LoadBalancerPool, bool) {
	pollingTargetStatuses := targetStatuses
	if expected != nil {
		pollingTargetStatuses = []string{
			targetGroupSynchronizedStatus,
			common.LoadBalancerProvisioningStatusError,
		}
	}

	return common.PollUntilResult(
		ctx,
		r,
		2*time.Second,
		"target group",
		targetGroupID,
		pollingTargetStatuses,
		diags,
		func(ctx context.Context) (*loadbalancer.LoadBalancerPool, *http.Response, error) {
			respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, diags,
				func() (*loadbalancer.GetTargetGroupResponse, *http.Response, error) {
					return r.kc.ApiClient.LoadBalancerTargetGroupAPI.
						GetTargetGroup(ctx, targetGroupID).
						XAuthToken(r.kc.XAuthToken).
						Execute()
				},
			)
			if err != nil {
				return nil, httpResp, err
			}
			return &respModel.TargetGroup, httpResp, nil
		},
		func(model *loadbalancer.LoadBalancerPool) string {
			if !model.ProvisioningStatus.IsSet() || model.ProvisioningStatus.Get() == nil {
				return ""
			}

			status := string(*model.ProvisioningStatus.Get())
			if status != common.LoadBalancerProvisioningStatusActive || expected == nil {
				return status
			}

			if targetGroupMatchesExpectedState(ctx, model, expected, diags) {
				return targetGroupSynchronizedStatus
			}

			return status
		},
	)
}

func buildTargetGroupExpectedState(
	plan *loadBalancerTargetGroupResourceModel,
	state *loadBalancerTargetGroupResourceModel,
) targetGroupExpectedState {
	expected := targetGroupExpectedState{}

	if !plan.Name.Equal(state.Name) {
		expected.Name = new(plan.Name)
	}
	if !plan.Description.IsUnknown() && !plan.Description.Equal(state.Description) {
		expected.Description = new(plan.Description)
	}
	if !plan.LoadBalancerAlgorithm.Equal(state.LoadBalancerAlgorithm) {
		expected.LoadBalancerAlgorithm = new(plan.LoadBalancerAlgorithm)
	}
	if !plan.SessionPersistence.Equal(state.SessionPersistence) {
		expected.SessionPersistence = new(plan.SessionPersistence)
	}

	return expected
}

func targetGroupMatchesExpectedState(
	ctx context.Context,
	result *loadbalancer.LoadBalancerPool,
	expected *targetGroupExpectedState,
	diags *diag.Diagnostics,
) bool {
	var actual loadBalancerTargetGroupBaseModel
	if !mapLoadBalancerTargetGroupFromGetResponse(ctx, &actual, result, diags) {
		return false
	}

	if expected.Name != nil && !expected.Name.Equal(actual.Name) {
		return false
	}
	if expected.Description != nil && !expected.Description.Equal(actual.Description) {
		return false
	}
	if expected.LoadBalancerAlgorithm != nil &&
		!expected.LoadBalancerAlgorithm.Equal(actual.LoadBalancerAlgorithm) {
		return false
	}
	if expected.SessionPersistence != nil &&
		!expected.SessionPersistence.Equal(actual.SessionPersistence) {
		return false
	}

	return true
}

func (r *loadBalancerTargetGroupResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config loadBalancerTargetGroupResourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.SessionPersistence.IsNull() || config.SessionPersistence.IsUnknown() {
		return
	}

	var sessionPersistence loadBalancerTargetGroupSessionPersistence
	diags.Append(config.SessionPersistence.As(ctx, &sessionPersistence, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return
	}

	if config.Protocol.ValueString() != string(loadbalancer.LOADBALANCERPOOLPROTOCOL_HTTP) &&
		config.Protocol.ValueString() != string(loadbalancer.LOADBALANCERPOOLPROTOCOL_TCP) &&
		config.Protocol.ValueString() != string(loadbalancer.LOADBALANCERPOOLPROTOCOL_UDP) {
		common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
			fmt.Sprintf("Session persistence is not supported for protocol '%s'", config.Protocol.ValueString()),
		)
	}

	if config.Protocol.ValueString() == string(loadbalancer.LOADBALANCERPOOLPROTOCOL_HTTP) {
		if sessionPersistence.Type.ValueString() != "HTTP_COOKIE" && sessionPersistence.Type.ValueString() != "APP_COOKIE" {
			common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
				fmt.Sprintf("Session persistence type '%s' is not valid for protocol '%s'.", sessionPersistence.Type.ValueString(), config.Protocol.ValueString()),
			)
		}
	} else {
		if sessionPersistence.Type.ValueString() != "SOURCE_IP" {
			common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
				fmt.Sprintf("Session persistence type '%s' is not valid for protocol '%s'.", sessionPersistence.Type.ValueString(), config.Protocol.ValueString()),
			)
		}
	}

	if sessionPersistence.Type.ValueString() != "APP_COOKIE" && !sessionPersistence.CookieName.IsNull() {
		common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
			fmt.Sprintf("'cookie_name' must not be set when session persistence type is '%s'.", sessionPersistence.Type.ValueString()),
		)
	}
	if sessionPersistence.Type.ValueString() == "APP_COOKIE" && sessionPersistence.CookieName.IsNull() {
		common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
			fmt.Sprintf("'cookie_name' must be set when session persistence type is '%s'.", sessionPersistence.Type.ValueString()),
		)
	}
}

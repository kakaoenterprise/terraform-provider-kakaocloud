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

	mutex := common.LockForID(plan.LoadBalancerId.ValueString())
	mutex.Lock()
	defer mutex.Unlock()

	timeout, diags := plan.Timeouts.Create(ctx, common.DefaultCreateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ok := CheckLoadBalancerStatus(ctx, plan.LoadBalancerId.ValueString(), false, r, r.kc, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

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

	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "CreateTargetGroup", err, &resp.Diagnostics)
		return
	}

	plan.Id = types.StringValue(respModel.TargetGroup.Id)

	result, ok := r.pollTargetGroupUntilStatus(
		ctx,
		plan.Id.ValueString(),
		[]string{common.LoadBalancerProvisioningStatusActive, common.LoadBalancerProvisioningStatusError},
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

	mutex := common.LockForID(plan.LoadBalancerId.ValueString())
	mutex.Lock()
	defer mutex.Unlock()

	timeout, diags := plan.Timeouts.Update(ctx, common.DefaultUpdateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ok := CheckLoadBalancerStatus(ctx, plan.LoadBalancerId.ValueString(), true, r, r.kc, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	updateReq := loadbalancer.NewEditTargetGroup()

	if !plan.Name.Equal(state.Name) {
		updateReq.SetName(plan.Name.ValueString())
	}

	if !plan.Description.IsNull() && !plan.Description.Equal(state.Description) {
		updateReq.SetDescription(plan.Description.ValueString())
	}

	if !plan.LoadBalancerAlgorithm.Equal(state.LoadBalancerAlgorithm) {
		algorithm := loadbalancer.TargetGroupAlgorithm(plan.LoadBalancerAlgorithm.ValueString())
		updateReq.SetLoadBalancerAlgorithm(algorithm)
	}

	if !plan.SessionPersistence.Equal(state.SessionPersistence) {
		if plan.SessionPersistence.IsNull() {

			updateReq.SetSessionPersistenceNil()
		} else {

			var sessionPersistence loadBalancerTargetGroupSessionPersistenceModel
			diags.Append(plan.SessionPersistence.As(ctx, &sessionPersistence, basetypes.ObjectAsOptions{})...)
			if !diags.HasError() {
				sessionPersistenceReq := loadbalancer.SessionPersistenceModel{
					Type:               sessionPersistence.Type.ValueString(),
					PersistenceTimeout: int32(sessionPersistence.PersistenceTimeout.ValueInt64()),
				}

				if !sessionPersistence.CookieName.IsNull() {
					sessionPersistenceReq.CookieName.Set(loadbalancer.PtrString(sessionPersistence.CookieName.ValueString()))
				}
				if !sessionPersistence.PersistenceGranularity.IsNull() {
					sessionPersistenceReq.PersistenceGranularity.Set(loadbalancer.PtrString(sessionPersistence.PersistenceGranularity.ValueString()))
				}

				updateReq.SetSessionPersistence(sessionPersistenceReq)
			}
		}
	}

	body := loadbalancer.NewBodyUpdateTargetGroup(*updateReq)

	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*loadbalancer.BnsLoadBalancerV1ApiUpdateTargetGroupModelResponseTargetGroupModel, *http.Response, error) {
			return r.kc.ApiClient.LoadBalancerTargetGroupAPI.
				UpdateTargetGroup(ctx, state.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				BodyUpdateTargetGroup(*body).
				Execute()
		},
	)

	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "UpdateTargetGroup", err, &resp.Diagnostics)
		return
	}

	time.Sleep(5 * time.Second)

	result, ok := r.pollTargetGroupUntilStatus(
		ctx,
		plan.Id.ValueString(),
		[]string{common.LoadBalancerProvisioningStatusActive, common.LoadBalancerProvisioningStatusError},
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

	mutex := common.LockForID(state.LoadBalancerId.ValueString())
	mutex.Lock()
	defer mutex.Unlock()

	timeout, diags := state.Timeouts.Delete(ctx, common.DefaultDeleteTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ok := CheckLoadBalancerStatus(ctx, state.LoadBalancerId.ValueString(), true, r, r.kc, &resp.Diagnostics)
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
	targetGroupId string,
	targetStatuses []string,
	diags *diag.Diagnostics,
) (*loadbalancer.BnsLoadBalancerV1ApiGetTargetGroupModelTargetGroupModel, bool) {
	return common.PollUntilResult(
		ctx,
		r,
		2*time.Second,
		"target group",
		targetGroupId,
		targetStatuses,
		diags,
		func(ctx context.Context) (*loadbalancer.BnsLoadBalancerV1ApiGetTargetGroupModelTargetGroupModel, *http.Response, error) {
			respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, diags,
				func() (*loadbalancer.TargetGroupResponseModel, *http.Response, error) {
					return r.kc.ApiClient.LoadBalancerTargetGroupAPI.
						GetTargetGroup(ctx, targetGroupId).
						XAuthToken(r.kc.XAuthToken).
						Execute()
				},
			)
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

	var sessionPersistence loadBalancerTargetGroupSessionPersistenceModel
	diags.Append(config.SessionPersistence.As(ctx, &sessionPersistence, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return
	}

	if config.Protocol.ValueString() != string(loadbalancer.TARGETGROUPPROTOCOL_HTTP) &&
		config.Protocol.ValueString() != string(loadbalancer.TARGETGROUPPROTOCOL_TCP) &&
		config.Protocol.ValueString() != string(loadbalancer.TARGETGROUPPROTOCOL_UDP) {
		common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
			fmt.Sprintf("Session persistence is not supported for protocol '%s'", config.Protocol.ValueString()),
		)
	}

	if config.Protocol.ValueString() == string(loadbalancer.TARGETGROUPPROTOCOL_HTTP) {
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

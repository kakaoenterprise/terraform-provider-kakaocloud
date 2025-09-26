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
	_ resource.Resource                = &loadBalancerListenerResource{}
	_ resource.ResourceWithConfigure   = &loadBalancerListenerResource{}
	_ resource.ResourceWithImportState = &loadBalancerListenerResource{}
)

func NewLoadBalancerListenerResource() resource.Resource {
	return &loadBalancerListenerResource{}
}

type loadBalancerListenerResource struct {
	kc *common.KakaoCloudClient
}

func (r *loadBalancerListenerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *loadBalancerListenerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_load_balancer_listener"
}

func (r *loadBalancerListenerResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: docs.GetResourceDescription("LoadBalancerListener"),
		Attributes: utils.MergeResourceSchemaAttributes(
			listenerResourceSchemaAttributes,
			map[string]schema.Attribute{
				"timeouts": timeouts.AttributesAll(ctx),
			},
		),
	}
}

func (r *loadBalancerListenerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan loadBalancerListenerResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.SniContainerRefs.IsNull() && !plan.SniContainerRefs.IsUnknown() {
		resp.Diagnostics.AddError(
			"Invalid Configuration",
			"sni_container_refs cannot be set during listener creation. Please create the listener first, then update it to add SNI container references.",
		)
		return
	}

	timeout, diags := plan.Timeouts.Create(ctx, common.DefaultCreateTimeout)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	createReq := loadbalancer.CreateListener{
		LoadBalancerId: plan.LoadBalancerId.ValueString(),
		Protocol:       loadbalancer.Protocol(plan.Protocol.ValueString()),
		ProtocolPort:   int32(plan.ProtocolPort.ValueInt64()),
	}

	if !plan.TargetGroupId.IsNull() && !plan.TargetGroupId.IsUnknown() {
		createReq.SetTargetGroupId(plan.TargetGroupId.ValueString())
	}

	if !plan.DefaultTlsContainerRef.IsNull() && !plan.DefaultTlsContainerRef.IsUnknown() {
		createReq.SetDefaultTlsContainerRef(plan.DefaultTlsContainerRef.ValueString())
	}

	if !plan.TlsMinVersion.IsNull() && !plan.TlsMinVersion.IsUnknown() {
		createReq.SetTlsMinVersion(loadbalancer.TLSVersion(plan.TlsMinVersion.ValueString()))
	}

	body := loadbalancer.BodyCreateListener{
		Listener: createReq,
	}

	lbl, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*loadbalancer.BnsLoadBalancerV1ApiCreateListenerModelResponseListenerModel, *http.Response, error) {
			return r.kc.ApiClient.LoadBalancerListenerAPI.CreateListener(ctx).XAuthToken(r.kc.XAuthToken).BodyCreateListener(body).Execute()
		},
	)

	if httpResp != nil && httpResp.StatusCode == http.StatusConflict {
		lbl, httpResp, err = ExecuteWithLoadBalancerConflictRetry(ctx, r.kc, &resp.Diagnostics,
			func() (*loadbalancer.BnsLoadBalancerV1ApiCreateListenerModelResponseListenerModel, *http.Response, error) {
				return r.kc.ApiClient.LoadBalancerListenerAPI.CreateListener(ctx).XAuthToken(r.kc.XAuthToken).BodyCreateListener(body).Execute()
			},
		)
	}

	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "CreateListener", err, &resp.Diagnostics)
		return
	}

	plan.Id = types.StringValue(lbl.Listener.Id)

	result, ok := r.pollListenerUntilStatus(
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

	finalResult, ok := r.updateListenerFields(ctx, plan.Id.ValueString(), &plan, resp)
	if !ok || resp.Diagnostics.HasError() {

		return
	}

	ok = mapLoadBalancerListenerBaseModel(ctx, &plan.loadBalancerListenerBaseModel, finalResult, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	plan.TargetGroupId = utils.ConvertNullableString(finalResult.DefaultTargetGroupId)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *loadBalancerListenerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state loadBalancerListenerResourceModel
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
		func() (*loadbalancer.BnsLoadBalancerV1ApiGetListenerModelResponseListenerModel, *http.Response, error) {
			return r.kc.ApiClient.LoadBalancerListenerAPI.GetListener(ctx, state.Id.ValueString()).XAuthToken(r.kc.XAuthToken).Execute()
		},
	)

	if httpResp != nil && httpResp.StatusCode == 404 {
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "GetListener", err, &resp.Diagnostics)
		return
	}
	loadBalancerListenerResult := respModel.Listener
	ok := mapLoadBalancerListenerBaseModel(ctx, &state.loadBalancerListenerBaseModel, &loadBalancerListenerResult, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	state.TargetGroupId = utils.ConvertNullableString(loadBalancerListenerResult.DefaultTargetGroupId)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *loadBalancerListenerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state loadBalancerListenerResourceModel
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

	editReq := loadbalancer.EditListener{}

	if !plan.DefaultTlsContainerRef.Equal(state.DefaultTlsContainerRef) {

		if !plan.DefaultTlsContainerRef.IsNull() && !plan.DefaultTlsContainerRef.IsUnknown() && plan.DefaultTlsContainerRef.ValueString() != "" {
			editReq.SetDefaultTlsContainerRef(plan.DefaultTlsContainerRef.ValueString())
		}
	}

	if !plan.SniContainerRefs.Equal(state.SniContainerRefs) {
		var sniRefs []string
		diags := plan.SniContainerRefs.ElementsAs(ctx, &sniRefs, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		editReq.SetSniContainerRefs(sniRefs)
	}
	if !plan.ConnectionLimit.Equal(state.ConnectionLimit) {
		editReq.SetConnectionLimit(int32(plan.ConnectionLimit.ValueInt64()))
	}
	if !plan.TargetGroupId.Equal(state.TargetGroupId) {

		if !plan.TargetGroupId.IsNull() && !plan.TargetGroupId.IsUnknown() && plan.TargetGroupId.ValueString() != "" {
			editReq.SetTargetGroupId(plan.TargetGroupId.ValueString())
		}
	}
	if !plan.TlsMinVersion.Equal(state.TlsMinVersion) {
		tlsVersionEnum := loadbalancer.TLSVersion(plan.TlsMinVersion.ValueString())
		editReq.SetTlsMinVersion(tlsVersionEnum)
	}

	if !plan.TimeoutClientData.Equal(state.TimeoutClientData) {

		if !plan.TimeoutClientData.IsNull() && !plan.TimeoutClientData.IsUnknown() {
			editReq.SetTimeoutClientData(int32(plan.TimeoutClientData.ValueInt64()))
		}
	} else {

		currentValue := state.TimeoutClientData.ValueInt64()
		if currentValue < 1000 {
			editReq.SetTimeoutClientData(1000)
		} else {
			editReq.SetTimeoutClientData(int32(currentValue))
		}
	}

	if !plan.InsertHeaders.Equal(state.InsertHeaders) {

		sdkHeaders := &loadbalancer.InsertHeaderModel{}

		if !plan.InsertHeaders.IsNull() && !plan.InsertHeaders.IsUnknown() {

			attrs := plan.InsertHeaders.Attributes()
			xForwardedFor := attrs["x_forwarded_for"].(types.String)
			xForwardedProto := attrs["x_forwarded_proto"].(types.String)
			xForwardedPort := attrs["x_forwarded_port"].(types.String)

			if !xForwardedFor.IsNull() {
				enumVal := loadbalancer.XForwardedFor(xForwardedFor.ValueString())
				sdkHeaders.XForwardedFor = *loadbalancer.NewNullableXForwardedFor(&enumVal)
			}
			if !xForwardedProto.IsNull() {
				enumVal := loadbalancer.XForwardedProto(xForwardedProto.ValueString())
				sdkHeaders.XForwardedProto = *loadbalancer.NewNullableXForwardedProto(&enumVal)
			}
			if !xForwardedPort.IsNull() {
				enumVal := loadbalancer.XForwardedPort(xForwardedPort.ValueString())
				sdkHeaders.XForwardedPort = *loadbalancer.NewNullableXForwardedPort(&enumVal)
			}
		}

		editReq.InsertHeaders = *loadbalancer.NewNullableInsertHeaderModel(sdkHeaders)
	}

	body := *loadbalancer.NewBodyUpdateListener(editReq)

	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (interface{}, *http.Response, error) {
			return r.kc.ApiClient.LoadBalancerListenerAPI.UpdateListener(ctx, state.Id.ValueString()).XAuthToken(r.kc.XAuthToken).BodyUpdateListener(body).Execute()
		},
	)

	if httpResp != nil && httpResp.StatusCode == http.StatusConflict {
		_, httpResp, err = ExecuteWithLoadBalancerConflictRetry(ctx, r.kc, &resp.Diagnostics,
			func() (interface{}, *http.Response, error) {
				return r.kc.ApiClient.LoadBalancerListenerAPI.UpdateListener(ctx, state.Id.ValueString()).XAuthToken(r.kc.XAuthToken).BodyUpdateListener(body).Execute()
			},
		)
	}

	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "UpdateListener", err, &resp.Diagnostics)
		return
	}

	result, ok := r.pollListenerUntilStatus(
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

	state.DefaultTlsContainerRef = plan.DefaultTlsContainerRef
	state.SniContainerRefs = plan.SniContainerRefs
	state.TlsMinVersion = plan.TlsMinVersion

	ok = mapLoadBalancerListenerBaseModel(ctx, &state.loadBalancerListenerBaseModel, result, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	if !plan.TargetGroupId.IsNull() && !plan.TargetGroupId.IsUnknown() {
		state.TargetGroupId = plan.TargetGroupId
	} else {

		state.TargetGroupId = utils.ConvertNullableString(result.DefaultTargetGroupId)
	}

	state.Timeouts = plan.Timeouts
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *loadBalancerListenerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state loadBalancerListenerResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (interface{}, *http.Response, error) {
			httpResp, err := r.kc.ApiClient.LoadBalancerListenerAPI.DeleteListener(ctx, state.Id.ValueString()).XAuthToken(r.kc.XAuthToken).Execute()
			return nil, httpResp, err
		},
	)

	if httpResp != nil && httpResp.StatusCode == http.StatusConflict {
		_, httpResp, err = ExecuteWithLoadBalancerConflictRetry(ctx, r.kc, &resp.Diagnostics,
			func() (interface{}, *http.Response, error) {
				httpResp, err := r.kc.ApiClient.LoadBalancerListenerAPI.DeleteListener(ctx, state.Id.ValueString()).XAuthToken(r.kc.XAuthToken).Execute()
				return nil, httpResp, err
			},
		)
	}

	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "DeleteListener", err, &resp.Diagnostics)
		return
	}

	common.PollUntilDeletion(ctx, r, 5*time.Second, &resp.Diagnostics, func(ctx context.Context) (bool, *http.Response, error) {
		_, httpResp, err := r.kc.ApiClient.LoadBalancerListenerAPI.
			GetListener(ctx, state.Id.ValueString()).
			XAuthToken(r.kc.XAuthToken).
			Execute()
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			return true, httpResp, nil
		}
		return false, httpResp, err
	})
}

func (r *loadBalancerListenerResource) pollListenerUntilStatus(
	ctx context.Context,
	listenerId string,
	targetStatuses []string,
	resp *diag.Diagnostics,
) (*loadbalancer.BnsLoadBalancerV1ApiGetListenerModelListenerModel, bool) {
	return common.PollUntilResult(
		ctx,
		r,
		5*time.Second,
		targetStatuses,
		resp,
		func(ctx context.Context) (*loadbalancer.BnsLoadBalancerV1ApiGetListenerModelListenerModel, *http.Response, error) {
			respModel, httpResp, err := r.kc.ApiClient.LoadBalancerListenerAPI.
				GetListener(ctx, listenerId).
				XAuthToken(r.kc.XAuthToken).
				Execute()
			if err != nil {
				return nil, httpResp, err
			}
			return &respModel.Listener, httpResp, nil
		},
		func(lb *loadbalancer.BnsLoadBalancerV1ApiGetListenerModelListenerModel) string {
			return string(*lb.ProvisioningStatus.Get())
		},
	)
}

func (r *loadBalancerListenerResource) updateListenerFields(
	ctx context.Context,
	listenerId string,
	plan *loadBalancerListenerResourceModel,
	resp *resource.CreateResponse,
) (*loadbalancer.BnsLoadBalancerV1ApiGetListenerModelListenerModel, bool) {

	needsUpdate := false
	editReq := loadbalancer.EditListener{}

	if !plan.TimeoutClientData.IsNull() && !plan.TimeoutClientData.IsUnknown() {
		editReq.SetTimeoutClientData(int32(plan.TimeoutClientData.ValueInt64()))
		needsUpdate = true
	}

	if !plan.ConnectionLimit.IsNull() && !plan.ConnectionLimit.IsUnknown() {
		editReq.SetConnectionLimit(int32(plan.ConnectionLimit.ValueInt64()))
		needsUpdate = true
	}

	if !plan.InsertHeaders.IsNull() && !plan.InsertHeaders.IsUnknown() {

		sdkHeaders := &loadbalancer.InsertHeaderModel{}
		attrs := plan.InsertHeaders.Attributes()
		xForwardedFor := attrs["x_forwarded_for"].(types.String)
		xForwardedProto := attrs["x_forwarded_proto"].(types.String)
		xForwardedPort := attrs["x_forwarded_port"].(types.String)

		if !xForwardedFor.IsNull() {
			enumVal := loadbalancer.XForwardedFor(xForwardedFor.ValueString())
			sdkHeaders.XForwardedFor = *loadbalancer.NewNullableXForwardedFor(&enumVal)
		}
		if !xForwardedProto.IsNull() {
			enumVal := loadbalancer.XForwardedProto(xForwardedProto.ValueString())
			sdkHeaders.XForwardedProto = *loadbalancer.NewNullableXForwardedProto(&enumVal)
		}
		if !xForwardedPort.IsNull() {
			enumVal := loadbalancer.XForwardedPort(xForwardedPort.ValueString())
			sdkHeaders.XForwardedPort = *loadbalancer.NewNullableXForwardedPort(&enumVal)
		}

		editReq.InsertHeaders = *loadbalancer.NewNullableInsertHeaderModel(sdkHeaders)
		needsUpdate = true
	}

	if !plan.SniContainerRefs.IsNull() && !plan.SniContainerRefs.IsUnknown() {
		var sniRefs []string
		diags := plan.SniContainerRefs.ElementsAs(ctx, &sniRefs, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return nil, false
		}
		editReq.SetSniContainerRefs(sniRefs)
		needsUpdate = true
	}

	if !needsUpdate {

		respModel, _, err := r.kc.ApiClient.LoadBalancerListenerAPI.GetListener(ctx, listenerId).
			XAuthToken(r.kc.XAuthToken).Execute()
		if err != nil {
			resp.Diagnostics.AddError("GetListener", fmt.Sprintf("Failed to get listener: %v", err))
			return nil, false
		}
		return &respModel.Listener, true
	}

	body := *loadbalancer.NewBodyUpdateListener(editReq)

	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (interface{}, *http.Response, error) {
			return r.kc.ApiClient.LoadBalancerListenerAPI.UpdateListener(ctx, listenerId).XAuthToken(r.kc.XAuthToken).BodyUpdateListener(body).Execute()
		},
	)

	if httpResp != nil && httpResp.StatusCode == http.StatusConflict {
		_, httpResp, err = ExecuteWithLoadBalancerConflictRetry(ctx, r.kc, &resp.Diagnostics,
			func() (interface{}, *http.Response, error) {
				return r.kc.ApiClient.LoadBalancerListenerAPI.UpdateListener(ctx, listenerId).XAuthToken(r.kc.XAuthToken).BodyUpdateListener(body).Execute()
			},
		)
	}

	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "UpdateListener", err, &resp.Diagnostics)
		return nil, false
	}

	result, ok := r.pollListenerUntilStatus(ctx, listenerId,
		[]string{ProvisioningStatusActive, ProvisioningStatusError}, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return nil, false
	}

	return result, true
}

func (r *loadBalancerListenerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

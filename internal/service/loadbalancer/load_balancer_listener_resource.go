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
		Description: "Manages a KakaoCloud Load Balancer.",
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

	// Validate that sni_container_refs is not set during creation
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

	// First try with normal auth retry, then with conflict retry if needed
	lbl, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*loadbalancer.BnsLoadBalancerV1ApiCreateListenerModelResponseListenerModel, *http.Response, error) {
			return r.kc.ApiClient.LoadBalancerListenerAPI.CreateListener(ctx).XAuthToken(r.kc.XAuthToken).BodyCreateListener(body).Execute()
		},
	)

	// If we get a 409 conflict, retry with loadbalancer-specific conflict logic
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

	// Phase 2: Update with fields that require update operations
	finalResult, ok := r.updateListenerFields(ctx, plan.Id.ValueString(), &plan, resp)
	if !ok || resp.Diagnostics.HasError() {
		// If update fails, the entire create operation should fail
		// Don't save the state with invalid data
		return
	}

	// Map the final result to state
	ok = mapLoadBalancerListenerBaseModel(ctx, &plan.loadBalancerListenerBaseModel, finalResult, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	// Map resource-specific fields
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

	// Map resource-specific fields
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

	// Only proceed if one of the updatable attributes has changed
	timeout, diags := plan.Timeouts.Update(ctx, common.DefaultUpdateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Create an empty request model
	editReq := loadbalancer.EditListener{}

	if !plan.DefaultTlsContainerRef.Equal(state.DefaultTlsContainerRef) {
		// Only set if the value is not null/unknown and not empty
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
		// Only set if the value is not null/unknown and not empty
		if !plan.TargetGroupId.IsNull() && !plan.TargetGroupId.IsUnknown() && plan.TargetGroupId.ValueString() != "" {
			editReq.SetTargetGroupId(plan.TargetGroupId.ValueString())
		}
	}
	if !plan.TlsMinVersion.Equal(state.TlsMinVersion) {
		tlsVersionEnum := loadbalancer.TLSVersion(plan.TlsMinVersion.ValueString())
		editReq.SetTlsMinVersion(tlsVersionEnum)
	}

	if !plan.TimeoutClientData.Equal(state.TimeoutClientData) {
		// If value changed, set the new value
		if !plan.TimeoutClientData.IsNull() && !plan.TimeoutClientData.IsUnknown() {
			editReq.SetTimeoutClientData(int32(plan.TimeoutClientData.ValueInt64()))
		}
	} else {
		// If value didn't change, use the existing value (but fix for validation)
		currentValue := state.TimeoutClientData.ValueInt64()
		if currentValue < 1000 {
			editReq.SetTimeoutClientData(1000) // Fix for minimum value
		} else {
			editReq.SetTimeoutClientData(int32(currentValue))
		}
	}

	// Check if the insert_headers block has changed.
	if !plan.InsertHeaders.Equal(state.InsertHeaders) {
		// This is the model we will build and send to the API.
		sdkHeaders := &loadbalancer.InsertHeaderModel{}

		// Check if the user has the block configured in their plan.
		if !plan.InsertHeaders.IsNull() && !plan.InsertHeaders.IsUnknown() {
			// Use direct attribute access instead of As() for more reliability.
			attrs := plan.InsertHeaders.Attributes()
			xForwardedFor := attrs["x_forwarded_for"].(types.String)
			xForwardedProto := attrs["x_forwarded_proto"].(types.String)
			xForwardedPort := attrs["x_forwarded_port"].(types.String)

			// Populate the SDK model, checking each value.
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

	// First try with normal auth retry, then with conflict retry if needed
	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (interface{}, *http.Response, error) {
			return r.kc.ApiClient.LoadBalancerListenerAPI.UpdateListener(ctx, state.Id.ValueString()).XAuthToken(r.kc.XAuthToken).BodyUpdateListener(body).Execute()
		},
	)

	// If we get a 409 conflict, retry with loadbalancer-specific conflict logic
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

	// Wait for the load balancer to become active again
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

	// TLS certificate fields are write-only (not returned in GET response)
	// Preserve them from the plan BEFORE base model mapping to avoid state drift
	state.DefaultTlsContainerRef = plan.DefaultTlsContainerRef
	state.SniContainerRefs = plan.SniContainerRefs
	state.TlsMinVersion = plan.TlsMinVersion

	ok = mapLoadBalancerListenerBaseModel(ctx, &state.loadBalancerListenerBaseModel, result, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	// Map resource-specific fields
	// Only update target_group_id if it's explicitly set in the plan
	if !plan.TargetGroupId.IsNull() && !plan.TargetGroupId.IsUnknown() {
		state.TargetGroupId = plan.TargetGroupId
	} else {
		// If not set in plan, get the current value from API response
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

	// First try with normal auth retry, then with conflict retry if needed
	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (interface{}, *http.Response, error) {
			httpResp, err := r.kc.ApiClient.LoadBalancerListenerAPI.DeleteListener(ctx, state.Id.ValueString()).XAuthToken(r.kc.XAuthToken).Execute()
			return nil, httpResp, err
		},
	)

	// If we get a 409 conflict, retry with loadbalancer-specific conflict logic
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

// Helper function to handle listener updates for fields that require update operations
func (r *loadBalancerListenerResource) updateListenerFields(
	ctx context.Context,
	listenerId string,
	plan *loadBalancerListenerResourceModel,
	resp *resource.CreateResponse,
) (*loadbalancer.BnsLoadBalancerV1ApiGetListenerModelListenerModel, bool) {

	needsUpdate := false
	editReq := loadbalancer.EditListener{}

	// Check which fields need updating (only fields that require update operations)
	if !plan.TimeoutClientData.IsNull() && !plan.TimeoutClientData.IsUnknown() {
		editReq.SetTimeoutClientData(int32(plan.TimeoutClientData.ValueInt64()))
		needsUpdate = true
	}

	if !plan.ConnectionLimit.IsNull() && !plan.ConnectionLimit.IsUnknown() {
		editReq.SetConnectionLimit(int32(plan.ConnectionLimit.ValueInt64()))
		needsUpdate = true
	}

	if !plan.InsertHeaders.IsNull() && !plan.InsertHeaders.IsUnknown() {
		// Handle insert_headers
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

	// If no updates needed, return current state
	if !needsUpdate {
		// Get current state
		respModel, _, err := r.kc.ApiClient.LoadBalancerListenerAPI.GetListener(ctx, listenerId).
			XAuthToken(r.kc.XAuthToken).Execute()
		if err != nil {
			resp.Diagnostics.AddError("GetListener", fmt.Sprintf("Failed to get listener: %v", err))
			return nil, false
		}
		return &respModel.Listener, true
	}

	// Perform the update
	body := *loadbalancer.NewBodyUpdateListener(editReq)

	// First try with normal auth retry, then with conflict retry if needed
	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (interface{}, *http.Response, error) {
			return r.kc.ApiClient.LoadBalancerListenerAPI.UpdateListener(ctx, listenerId).XAuthToken(r.kc.XAuthToken).BodyUpdateListener(body).Execute()
		},
	)

	// If we get a 409 conflict, retry with loadbalancer-specific conflict logic
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

	// Wait for update to complete
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

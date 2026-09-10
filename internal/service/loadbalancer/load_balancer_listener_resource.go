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
	"github.com/kakaoenterprise/kc-sdk-go/v2/services/loadbalancer"
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

	createReq := loadbalancer.CreateListener{
		LoadBalancerId: plan.LoadBalancerId.ValueString(),
		Protocol:       loadbalancer.Protocol(plan.Protocol.ValueString()),
		ProtocolPort:   plan.ProtocolPort.ValueInt32(),
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

	body := loadbalancer.CreateListenerRequest{
		Listener: createReq,
	}

	lbl, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*loadbalancer.CreateListenerResponse, *http.Response, error) {
			return r.kc.ApiClient.LoadBalancerListenerAPI.CreateListener(ctx).XAuthToken(r.kc.XAuthToken).CreateListenerRequest(body).Execute()
		},
	)

	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "CreateListener", err, &resp.Diagnostics)
		return
	}

	plan.Id = types.StringValue(lbl.Listener.Id)

	result, ok := r.pollListenerUntilStatus(
		ctx,
		plan.Id.ValueString(),
		loadBalancerAvailableAndErrorStatuses,
		&resp.Diagnostics,
	)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	common.CheckResourceAvailableStatus(ctx, r, (*string)(result.ProvisioningStatus.Get()), []string{common.LoadBalancerProvisioningStatusActive}, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	finalResult, ok := r.updateListenerAttributes(ctx, &plan, nil, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {

		return
	}

	if finalResult != nil {
		result = finalResult
	}

	ok = mapLoadBalancerListenerModel(ctx, &plan.loadBalancerListenerBaseModel, result, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

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
		func() (*loadbalancer.GetListenerResponse, *http.Response, error) {
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
	ok := mapLoadBalancerListenerModel(ctx, &state.loadBalancerListenerBaseModel, new(respModel.Listener), &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

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

	unlock, ok := common.AcquireLockForID(ctx, plan.LoadBalancerId.ValueString(), &resp.Diagnostics)
	if !ok {
		return
	}
	defer unlock()

	_, ok = pollLoadBalancerStatus(ctx, plan.LoadBalancerId.ValueString(), true, r, r.kc, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	result, ok := r.updateListenerAttributes(ctx, &plan, &state, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {

		return
	}

	ok = mapLoadBalancerListenerModel(ctx, &plan.loadBalancerListenerBaseModel, result, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *loadBalancerListenerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state loadBalancerListenerResourceModel
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
			httpResp, err := r.kc.ApiClient.LoadBalancerListenerAPI.DeleteListener(ctx, state.Id.ValueString()).XAuthToken(r.kc.XAuthToken).Execute()
			return nil, httpResp, err
		},
	)

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
) (*loadbalancer.LoadBalancerListener, bool) {
	return common.PollUntilResult(
		ctx,
		r,
		5*time.Second,
		"listener",
		listenerId,
		targetStatuses,
		resp,
		func(ctx context.Context) (*loadbalancer.LoadBalancerListener, *http.Response, error) {
			respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, resp,
				func() (*loadbalancer.GetListenerResponse, *http.Response, error) {
					return r.kc.ApiClient.LoadBalancerListenerAPI.GetListener(ctx, listenerId).XAuthToken(r.kc.XAuthToken).Execute()
				},
			)
			if err != nil {
				return nil, httpResp, err
			}
			return &respModel.Listener, httpResp, nil
		},
		func(lb *loadbalancer.LoadBalancerListener) string {
			return string(*lb.ProvisioningStatus.Get())
		},
	)
}

func (r *loadBalancerListenerResource) updateListenerAttributes(
	ctx context.Context,
	plan, state *loadBalancerListenerResourceModel,
	diag *diag.Diagnostics,
) (*loadbalancer.LoadBalancerListener, bool) {
	needsUpdate := false
	editReq := loadbalancer.NewUpdateListener()

	if state != nil && !plan.DefaultTlsContainerRef.Equal(state.DefaultTlsContainerRef) {
		if !plan.DefaultTlsContainerRef.IsNull() && !plan.DefaultTlsContainerRef.IsUnknown() {
			editReq.SetDefaultTlsContainerRef(plan.DefaultTlsContainerRef.ValueString())
		} else {
			editReq.SetDefaultTlsContainerRefNil()
		}
		needsUpdate = true
	}

	if state == nil || !plan.SniContainerRefs.Equal(state.SniContainerRefs) {
		var sniRefs []string

		if !plan.SniContainerRefs.IsNull() && !plan.SniContainerRefs.IsUnknown() {
			diags := plan.SniContainerRefs.ElementsAs(ctx, &sniRefs, false)
			diag.Append(diags...)
			if diag.HasError() {
				return nil, false
			}
			editReq.SetSniContainerRefs(sniRefs)
			needsUpdate = true
		} else {
			if state != nil {
				editReq.SetSniContainerRefs(sniRefs)
				needsUpdate = true
			}
		}
	}

	if state != nil && !plan.TargetGroupId.Equal(state.DefaultTargetGroupId) {
		if !plan.TargetGroupId.IsNull() && !plan.TargetGroupId.IsUnknown() {
			editReq.SetTargetGroupId(plan.TargetGroupId.ValueString())
		} else {
			editReq.SetTargetGroupIdNil()
		}
		needsUpdate = true
	}

	if state != nil && !plan.TlsMinVersion.IsNull() && !plan.TlsMinVersion.IsUnknown() && !plan.TlsMinVersion.Equal(state.TlsMinVersion) {
		tlsVersionEnum := loadbalancer.TLSVersion(plan.TlsMinVersion.ValueString())
		editReq.SetTlsMinVersion(tlsVersionEnum)
		needsUpdate = true
	}

	if state == nil || !plan.TimeoutClientData.Equal(state.TimeoutClientData) {
		if !plan.TimeoutClientData.IsNull() && !plan.TimeoutClientData.IsUnknown() {
			editReq.SetTimeoutClientData(plan.TimeoutClientData.ValueInt32())
			needsUpdate = true
		}
	}

	if state == nil || !plan.InsertHeaders.Equal(state.InsertHeaders) {
		sdkHeaders := loadbalancer.NewInsertHeaderRequest()

		if !plan.InsertHeaders.IsNull() && !plan.InsertHeaders.IsUnknown() {
			attrs := plan.InsertHeaders.Attributes()
			xForwardedFor := attrs["x_forwarded_for"].(types.String)
			xForwardedProto := attrs["x_forwarded_proto"].(types.String)
			xForwardedPort := attrs["x_forwarded_port"].(types.String)

			if !xForwardedFor.IsNull() && !xForwardedFor.IsUnknown() {
				enumVal := loadbalancer.XForwardedFor(xForwardedFor.ValueString())
				sdkHeaders.SetXForwardedFor(enumVal)
			}
			if !xForwardedProto.IsNull() && !xForwardedProto.IsUnknown() {
				enumVal := loadbalancer.XForwardedProto(xForwardedProto.ValueString())
				sdkHeaders.SetXForwardedProto(enumVal)
			}
			if !xForwardedPort.IsNull() && !xForwardedPort.IsUnknown() {
				enumVal := loadbalancer.XForwardedProto(xForwardedPort.ValueString())
				sdkHeaders.SetXForwardedPort(enumVal)
			}
			editReq.SetInsertHeaders(*sdkHeaders)
			needsUpdate = true
		}
	}

	if needsUpdate {
		body := *loadbalancer.NewUpdateListenerRequest(*editReq)

		_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, diag,
			func() (interface{}, *http.Response, error) {
				return r.kc.ApiClient.LoadBalancerListenerAPI.UpdateListener(ctx, plan.Id.ValueString()).
					XAuthToken(r.kc.XAuthToken).UpdateListenerRequest(body).Execute()
			},
		)
		if err != nil {
			common.AddApiActionError(ctx, r, httpResp, "UpdateListener", err, diag)
			return nil, false
		}

		time.Sleep(5 * time.Second)
	}

	result, ok := r.pollListenerUntilStatus(
		ctx,
		plan.Id.ValueString(),
		loadBalancerAvailableAndErrorStatuses,
		diag,
	)
	if !ok || diag.HasError() {
		return nil, false
	}

	common.CheckResourceAvailableStatus(ctx, r, (*string)(result.ProvisioningStatus.Get()), []string{common.LoadBalancerProvisioningStatusActive}, diag)
	if diag.HasError() {
		return nil, false
	}

	return result, true
}

func (r *loadBalancerListenerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *loadBalancerListenerResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config loadBalancerListenerResourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.Protocol.IsUnknown() {
		return
	}

	if config.Protocol.ValueString() == string(loadbalancer.PROTOCOL_TERMINATED_HTTPS) {
		if config.DefaultTlsContainerRef.IsNull() || config.TlsMinVersion.IsNull() {
			common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
				"tls_min_version and default_tls_container_ref must be provided for TERMINATED_HTTPS protocol.")
		}
	} else {
		if !config.DefaultTlsContainerRef.IsNull() ||
			!config.TlsMinVersion.IsNull() ||
			!config.SniContainerRefs.IsNull() {
			common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
				"tls_min_version and default_tls_container_ref and sni_container_refs must only be specified when using the TERMINATED_HTTPS protocol.")
		}
	}

	if config.Protocol.ValueString() != string(loadbalancer.PROTOCOL_HTTP) &&
		config.Protocol.ValueString() != string(loadbalancer.PROTOCOL_TERMINATED_HTTPS) {
		if !config.InsertHeaders.IsNull() && !config.InsertHeaders.IsUnknown() {
			common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
				"insert_headers can only be specified when using the HTTP or TERMINATED_HTTPS protocol.")
		}
	}
}

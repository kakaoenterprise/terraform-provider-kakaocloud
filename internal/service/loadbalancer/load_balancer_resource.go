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
	_ resource.Resource                = &loadBalancerResource{}
	_ resource.ResourceWithConfigure   = &loadBalancerResource{}
	_ resource.ResourceWithImportState = &loadBalancerResource{}
)

func NewLoadBalancerResource() resource.Resource {
	return &loadBalancerResource{}
}

type loadBalancerResource struct {
	kc *common.KakaoCloudClient
}

func (r *loadBalancerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *loadBalancerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_load_balancer"
}

func (r *loadBalancerResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: utils.MergeResourceSchemaAttributes(
			loadBalancerResourceSchemaAttributes,
			map[string]schema.Attribute{
				"timeouts": timeouts.AttributesAll(ctx),
			},
		),
	}
}

func (r *loadBalancerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan loadBalancerResourceModel
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

	var result *loadbalancer.BnsLoadBalancerV1ApiGetLoadBalancerModelLoadBalancerModel
	var ok bool

	createReq := loadbalancer.CreateLoadBalancerModel{
		Name:             plan.Name.ValueString(),
		SubnetId:         plan.SubnetId.ValueString(),
		AvailabilityZone: loadbalancer.AvailabilityZone(plan.AvailabilityZone.ValueString()),
		FlavorId:         plan.FlavorId.ValueString(),
	}

	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		createReq.SetDescription(plan.Description.ValueString())
	}

	body := loadbalancer.BodyCreateLoadBalancer{LoadBalancer: createReq}

	lb, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*loadbalancer.BnsLoadBalancerV1ApiCreateLoadBalancerModelResponseLoadBalancerModel, *http.Response, error) {
			return r.kc.ApiClient.LoadBalancerAPI.CreateLoadBalancer(ctx).XAuthToken(r.kc.XAuthToken).BodyCreateLoadBalancer(body).Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "CreateLoadBalancer", err, &resp.Diagnostics)
		return
	}

	plan.Id = types.StringValue(lb.LoadBalancer.Id)

	result, ok = r.pollLoadBalancerUntilStatus(
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

	if !plan.AccessLogs.IsNull() && !plan.AccessLogs.IsUnknown() {
		result, ok = r.updateLoadBalancerAccessLogs(ctx, &plan, &resp.Diagnostics)
		if !ok {
			return
		}
	}

	ok = mapLoadBalancer(ctx, &plan.loadBalancerBaseModel, result, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *loadBalancerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state loadBalancerResourceModel
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
		func() (*loadbalancer.BnsLoadBalancerV1ApiGetLoadBalancerModelResponseLoadBalancerModel, *http.Response, error) {
			return r.kc.ApiClient.LoadBalancerAPI.
				GetLoadBalancer(ctx, state.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				Execute()
		},
	)

	if httpResp != nil && httpResp.StatusCode == 404 {
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "GetLoadBalancer", err, &resp.Diagnostics)
		return
	}

	loadBalancerResult := respModel.LoadBalancer
	ok := mapLoadBalancer(ctx, &state.loadBalancerBaseModel, &loadBalancerResult, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	if state.FlavorId.IsNull() {
		lbfs, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
			func() (*loadbalancer.FlavorListModel, *http.Response, error) {
				resp, httpResp, err := r.kc.ApiClient.LoadBalancerEtcAPI.ListLoadBalancerTypes(ctx).XAuthToken(r.kc.XAuthToken).Execute()
				return resp, httpResp, err
			},
		)
		if err != nil {
			common.AddApiActionError(ctx, r, httpResp, "ListLoadBalancerTypes", err, &resp.Diagnostics)
			return
		}

		for _, lbf := range lbfs.Flavors {
			if lbf.Name.Get() != nil && *lbf.Name.Get() == state.Type.ValueString() {
				state.FlavorId = types.StringValue(lbf.Id)
				break
			}
		}
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *loadBalancerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state loadBalancerResourceModel
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

	var result *loadbalancer.BnsLoadBalancerV1ApiGetLoadBalancerModelLoadBalancerModel
	var ok bool

	mutex := common.LockForID(plan.Id.ValueString())
	mutex.Lock()
	defer mutex.Unlock()

	ok = CheckLoadBalancerStatus(ctx, plan.Id.ValueString(), true, r, r.kc, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	if !plan.Name.Equal(state.Name) || !plan.Description.Equal(state.Description) {

		editReq := loadbalancer.EditLoadBalancerModel{}

		if !plan.Name.Equal(state.Name) {
			editReq.SetName(plan.Name.ValueString())
		}

		if !plan.Description.Equal(state.Description) {
			if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
				editReq.SetDescription(plan.Description.ValueString())
			} else {

				editReq.SetDescription("")
			}
		}

		body := *loadbalancer.NewBodyUpdateLoadBalancer(editReq)

		_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
			func() (*loadbalancer.BnsLoadBalancerV1ApiUpdateLoadBalancerModelResponseLoadBalancerModel, *http.Response, error) {
				return r.kc.ApiClient.LoadBalancerAPI.
					UpdateLoadBalancer(ctx, state.Id.ValueString()).
					XAuthToken(r.kc.XAuthToken).
					BodyUpdateLoadBalancer(body).
					Execute()
			},
		)

		if err != nil {
			common.AddApiActionError(ctx, r, httpResp, "UpdateLoadBalancer", err, &resp.Diagnostics)
			return
		}

		result, ok = r.pollLoadBalancerUntilStatus(
			ctx,
			state.Id.ValueString(),
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
	}

	if !plan.AccessLogs.Equal(state.AccessLogs) {
		result, ok = r.updateLoadBalancerAccessLogs(ctx, &plan, &resp.Diagnostics)
		if !ok {
			return
		}
	}

	if result != nil {
		ok = mapLoadBalancer(ctx, &plan.loadBalancerBaseModel, result, &resp.Diagnostics)
		if !ok || resp.Diagnostics.HasError() {
			return
		}
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *loadBalancerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state loadBalancerResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	mutex := common.LockForID(state.Id.ValueString())
	mutex.Lock()
	defer mutex.Unlock()

	ok := CheckLoadBalancerStatus(ctx, state.Id.ValueString(), true, r, r.kc, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (struct{}, *http.Response, error) {
			httpResp, err := r.kc.ApiClient.LoadBalancerAPI.
				DeleteLoadBalancer(ctx, state.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				Execute()
			return struct{}{}, httpResp, err
		},
	)

	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "DeleteLoadBalancer", err, &resp.Diagnostics)
		return
	}

	common.PollUntilDeletion(ctx, r, 5*time.Second, &resp.Diagnostics, func(ctx context.Context) (bool, *http.Response, error) {
		_, httpResp, err := r.kc.ApiClient.LoadBalancerAPI.
			GetLoadBalancer(ctx, state.Id.ValueString()).
			XAuthToken(r.kc.XAuthToken).
			Execute()
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			return true, httpResp, nil
		}
		return false, httpResp, err
	})
}

func (r *loadBalancerResource) pollLoadBalancerUntilStatus(
	ctx context.Context,
	loadBalancerId string,
	targetStatuses []string,
	resp *diag.Diagnostics,
) (*loadbalancer.BnsLoadBalancerV1ApiGetLoadBalancerModelLoadBalancerModel, bool) {
	return common.PollUntilResult(
		ctx,
		r,
		5*time.Second,
		"load balancer",
		loadBalancerId,
		targetStatuses,
		resp,
		func(ctx context.Context) (*loadbalancer.BnsLoadBalancerV1ApiGetLoadBalancerModelLoadBalancerModel, *http.Response, error) {
			respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, resp,
				func() (*loadbalancer.BnsLoadBalancerV1ApiGetLoadBalancerModelResponseLoadBalancerModel, *http.Response, error) {
					return r.kc.ApiClient.LoadBalancerAPI.
						GetLoadBalancer(ctx, loadBalancerId).
						XAuthToken(r.kc.XAuthToken).
						Execute()
				},
			)
			if err != nil {
				return nil, httpResp, err
			}
			return &respModel.LoadBalancer, httpResp, nil
		},
		func(lb *loadbalancer.BnsLoadBalancerV1ApiGetLoadBalancerModelLoadBalancerModel) string {
			return string(*lb.ProvisioningStatus.Get())
		},
	)
}

func (r *loadBalancerResource) updateLoadBalancerAccessLogs(
	ctx context.Context,
	plan *loadBalancerResourceModel,
	diag *diag.Diagnostics,
) (*loadbalancer.BnsLoadBalancerV1ApiGetLoadBalancerModelLoadBalancerModel, bool) {
	var accessLog accessLogModel
	body := loadbalancer.NewBodyUpdateAccessLog()

	if plan.AccessLogs.IsNull() {
		body.SetAccessLogsNil()
	} else {
		diags := plan.AccessLogs.As(ctx, &accessLog, basetypes.ObjectAsOptions{})
		diag.Append(diags...)
		if diag.HasError() {
			return nil, false
		}

		accessLogReq := loadbalancer.EditLoadBalancerAccessLogModel{
			Bucket:    accessLog.Bucket.ValueString(),
			AccessKey: accessLog.AccessKey.ValueString(),
			SecretKey: accessLog.SecretKey.ValueString(),
		}

		body.SetAccessLogs(accessLogReq)
	}

	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, diag,
		func() (*loadbalancer.BnsLoadBalancerV1ApiUpdateAccessLogModelResponseLoadBalancerModel, *http.Response, error) {
			return r.kc.ApiClient.LoadBalancerAPI.UpdateAccessLog(ctx, plan.Id.ValueString()).XAuthToken(r.kc.XAuthToken).BodyUpdateAccessLog(*body).Execute()
		},
	)

	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "UpdateLoadBalancerAccessLog", err, diag)
		return nil, false
	}

	result, ok := r.pollLoadBalancerUntilStatus(
		ctx,
		plan.Id.ValueString(),
		[]string{common.LoadBalancerProvisioningStatusActive, common.LoadBalancerProvisioningStatusError},
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

func (r *loadBalancerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *loadBalancerResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config loadBalancerResourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.validateAvailabilityZoneConfig(config, resp)
}

func (r *loadBalancerResource) validateAvailabilityZoneConfig(config loadBalancerResourceModel, resp *resource.ValidateConfigResponse) {
	common.ValidateAvailabilityZone(
		path.Root("availability_zone"),
		config.AvailabilityZone,
		r.kc,
		&resp.Diagnostics,
	)
}

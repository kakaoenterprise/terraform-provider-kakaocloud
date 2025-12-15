// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package network

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
	"github.com/kakaoenterprise/kc-sdk-go/services/bcs"
	"github.com/kakaoenterprise/kc-sdk-go/services/loadbalancer"
	"github.com/kakaoenterprise/kc-sdk-go/services/network"
)

var (
	_ resource.ResourceWithConfigure      = &publicIpResource{}
	_ resource.ResourceWithImportState    = &publicIpResource{}
	_ resource.ResourceWithValidateConfig = &publicIpResource{}
)

func NewPublicIpResource() resource.Resource {
	return &publicIpResource{}
}

type publicIpResource struct {
	kc *common.KakaoCloudClient
}

func (r *publicIpResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config publicIpResourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.validateRelatedResourceConfig(ctx, config, resp)
}

func (r *publicIpResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *publicIpResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_public_ip"
}

func (r *publicIpResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: utils.MergeAttributes[schema.Attribute](
			publicIpResourceSchema,
			map[string]schema.Attribute{
				"timeouts": timeouts.AttributesAll(ctx),
			},
		),
	}
}
func (r *publicIpResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan publicIpResourceModel
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

	createReq := network.CreatePublicIpModel{}

	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		createReq.SetDescription(plan.Description.ValueString())
	}
	body := network.BodyCreatePublicIp{PublicIp: createReq}

	respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*network.BnsNetworkV1ApiCreatePublicIpModelResponsePublicIpModel, *http.Response, error) {
			return r.kc.ApiClient.PublicIPAPI.CreatePublicIp(ctx).XAuthToken(r.kc.XAuthToken).BodyCreatePublicIp(body).Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "CreatePublicIp", err, &resp.Diagnostics)
		return
	}

	plan.Id = types.StringValue(respModel.PublicIp.Id)

	_, ok := r.pollPublicIpUtilsStatus(
		ctx,
		plan.Id.ValueString(),
		[]string{common.PublicIpAvailable},
		&resp.Diagnostics,
	)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	if !plan.RelatedResource.IsNull() {
		var relatedResource resourceModel
		if d := plan.RelatedResource.As(ctx, &relatedResource, basetypes.ObjectAsOptions{}); d.HasError() {
			resp.Diagnostics.Append(d...)
			return
		}

		publicIpId := plan.Id.ValueString()

		deviceId := relatedResource.DeviceId.ValueString()
		deviceType := relatedResource.DeviceType.ValueString()
		networkInterfaceId := relatedResource.Id.ValueString()

		if !r.attachPublicIPByType(ctx, &resp.Diagnostics, deviceType, networkInterfaceId, deviceId, publicIpId) {
			_ = r.cleanupCreateFailurePublicIP(ctx, plan.Id.ValueString(), &resp.Diagnostics)
			return
		}

		_, ok = r.pollPublicIpUtilsStatus(
			ctx,
			plan.Id.ValueString(),
			[]string{common.PublicIpInUse},
			&resp.Diagnostics,
		)

		if !ok || resp.Diagnostics.HasError() {
			return
		}
	}

	result, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*network.BnsNetworkV1ApiGetPublicIpModelResponsePublicIpModel, *http.Response, error) {
			return r.kc.ApiClient.PublicIPAPI.
				GetPublicIp(ctx, plan.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "GetPublicIp", err, &resp.Diagnostics)
		return
	}

	ok = mapPublicIpBaseModel(ctx, &plan.publicIpBaseModel, &result.PublicIp, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *publicIpResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state publicIpResourceModel
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
		func() (*network.BnsNetworkV1ApiGetPublicIpModelResponsePublicIpModel, *http.Response, error) {
			return r.kc.ApiClient.PublicIPAPI.
				GetPublicIp(ctx, state.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				Execute()
		},
	)

	if httpResp != nil && httpResp.StatusCode == 404 {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "GetPublicIp", err, &resp.Diagnostics)
		return
	}
	result := respModel.PublicIp

	ok := mapPublicIpBaseModel(ctx, &state.publicIpBaseModel, &result, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *publicIpResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state publicIpResourceModel

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

	editReq := network.EditPublicIpModel{}

	var newDescription types.String

	if !(plan.Description.IsUnknown() || plan.Description.IsNull()) && !plan.Description.Equal(state.Description) {
		newDescription = plan.Description
		editReq.SetDescription(plan.Description.ValueString())

		body := *network.NewBodyUpdatePublicIp(editReq)
		_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
			func() (*network.BnsNetworkV1ApiUpdatePublicIpModelResponsePublicIpModel, *http.Response, error) {
				return r.kc.ApiClient.PublicIPAPI.
					UpdatePublicIp(ctx, plan.Id.ValueString()).
					XAuthToken(r.kc.XAuthToken).
					BodyUpdatePublicIp(body).
					Execute()
			},
		)
		if err != nil {
			common.AddApiActionError(ctx, r, httpResp, "UpdatePublicIp", err, &resp.Diagnostics)
			return
		}
	}

	needReattach, d := shouldReattachRelatedResource(ctx, plan.RelatedResource, state.RelatedResource)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	if needReattach {
		if !state.RelatedResource.IsNull() {
			var stateRelatedResource resourceModel
			diags := state.RelatedResource.As(ctx, &stateRelatedResource, basetypes.ObjectAsOptions{})
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}

			publicIpId := state.Id.ValueString()
			networkInterfaceId := stateRelatedResource.Id.ValueString()
			deviceId := stateRelatedResource.DeviceId.ValueString()
			deviceType := stateRelatedResource.DeviceType.ValueString()

			if !r.detachPublicIPByType(ctx, &resp.Diagnostics, deviceType, networkInterfaceId, deviceId) {
				return
			}

			_, ok := r.pollPublicIpUtilsStatus(
				ctx,
				publicIpId,
				[]string{common.PublicIpAvailable},
				&resp.Diagnostics,
			)
			if !ok || resp.Diagnostics.HasError() {
				return
			}
		}

		if !plan.RelatedResource.IsNull() {
			var planRelatedResource resourceModel
			if d := plan.RelatedResource.As(ctx, &planRelatedResource, basetypes.ObjectAsOptions{}); d.HasError() {
				resp.Diagnostics.Append(d...)
				return
			}

			publicIpId := plan.Id.ValueString()

			deviceId := planRelatedResource.DeviceId.ValueString()
			deviceType := planRelatedResource.DeviceType.ValueString()
			networkInterfaceId := planRelatedResource.Id.ValueString()

			if !r.attachPublicIPByType(ctx, &resp.Diagnostics, deviceType, networkInterfaceId, deviceId, publicIpId) {
				return
			}

			_, ok := r.pollPublicIpUtilsStatus(
				ctx,
				plan.Id.ValueString(),
				[]string{common.PublicIpInUse},
				&resp.Diagnostics,
			)
			if !ok || resp.Diagnostics.HasError() {
				return
			}
		}
	}

	respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*network.BnsNetworkV1ApiGetPublicIpModelResponsePublicIpModel, *http.Response, error) {
			return r.kc.ApiClient.PublicIPAPI.
				GetPublicIp(ctx, state.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "GetPublicIp", err, &resp.Diagnostics)
		return
	}

	ok := mapPublicIpBaseModel(ctx, &plan.publicIpBaseModel, &respModel.PublicIp, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	if !newDescription.IsNull() && !newDescription.IsUnknown() {
		plan.Description = newDescription
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *publicIpResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state publicIpResourceModel
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

	if !state.RelatedResource.IsNull() {
		var relatedResource resourceModel
		diags := state.RelatedResource.As(ctx, &relatedResource, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		publicIpId := state.Id.ValueString()
		networkInterfaceId := relatedResource.Id.ValueString()
		deviceId := relatedResource.DeviceId.ValueString()
		deviceType := relatedResource.DeviceType.ValueString()

		if !r.detachPublicIPByType(ctx, &resp.Diagnostics, deviceType, networkInterfaceId, deviceId) {
			return
		}

		_, ok := r.pollPublicIpUtilsStatus(
			ctx,
			publicIpId,
			[]string{common.PublicIpAvailable},
			&resp.Diagnostics,
		)

		if !ok || resp.Diagnostics.HasError() {
			return
		}
	}

	_, httpResp, err := common.ExecuteWithRetryAndAuth[struct{}](
		ctx,
		r.kc,
		&resp.Diagnostics,
		func() (struct{}, *http.Response, error) {
			resp, err := r.kc.ApiClient.PublicIPAPI.
				DeletePublicIp(ctx, state.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				Execute()

			return struct{}{}, resp, err
		},
	)

	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			return
		}
		common.AddApiActionError(ctx, r, httpResp, "DeletePublicIp", err, &resp.Diagnostics)
		return
	}

	common.PollUntilDeletion(ctx, r, 2*time.Second, &resp.Diagnostics, func(ctx context.Context) (bool, *http.Response, error) {
		_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
			func() (*network.BnsNetworkV1ApiGetPublicIpModelResponsePublicIpModel, *http.Response, error) {
				_, httpResp, err := r.kc.ApiClient.PublicIPAPI.
					GetPublicIp(ctx, state.Id.ValueString()).
					XAuthToken(r.kc.XAuthToken).
					Execute()
				return nil, httpResp, err
			},
		)

		return false, httpResp, err
	})
}

func (r *publicIpResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*common.KakaoCloudClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *kakaocloud.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.kc = client
}

func (r *publicIpResource) attachPublicIPByType(
	ctx context.Context,
	respDiags *diag.Diagnostics,
	deviceType, networkInterfaceId, deviceId, publicIpId string,
) bool {
	var (
		httpResp *http.Response
		err      error
	)

	switch deviceType {
	case "instance":
		_, httpResp, err = common.ExecuteWithRetryAndAuth(ctx, r.kc, respDiags,
			func() (*bcs.BcsInstanceV1ApiAssociatePublicIpModelResponsePublicIpModel, *http.Response, error) {
				return r.kc.ApiClient.InstancePublicIPAPI.
					AssociatePublicIp(ctx, deviceId, networkInterfaceId, publicIpId).
					XAuthToken(r.kc.XAuthToken).
					Execute()
			},
		)

	case "load-balancer":
		_, httpResp, err = common.ExecuteWithRetryAndAuth(ctx, r.kc, respDiags,
			func() (*loadbalancer.BnsLoadBalancerV1ApiAssociatePublicIpModelResponsePublicIpModel, *http.Response, error) {
				return r.kc.ApiClient.LoadBalancerAPI.
					AssociatePublicIp(ctx, deviceId, publicIpId).
					XAuthToken(r.kc.XAuthToken).
					Execute()
			},
		)

	default:
		common.AddGeneralError(ctx, r, respDiags,
			fmt.Sprintf("Unsupported device_type: %s", deviceType),
		)
		return false
	}

	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "AssociatePublicIp", err, respDiags)
		return false
	}

	return true
}

func (r *publicIpResource) detachPublicIPByType(
	ctx context.Context,
	respDiags *diag.Diagnostics,
	deviceType, networkInterfaceId, deviceId string,
) bool {
	var (
		httpResp *http.Response
		err      error
	)

	switch deviceType {
	case "instance":
		_, httpResp, err = common.ExecuteWithRetryAndAuth(ctx, r.kc, respDiags,
			func() (*bcs.BcsInstanceV1ApiRemovePublicIpModelResponsePublicIpModel, *http.Response, error) {
				return r.kc.ApiClient.InstancePublicIPAPI.
					RemovePublicIp(ctx, deviceId, networkInterfaceId).
					XAuthToken(r.kc.XAuthToken).
					Execute()
			},
		)

	case "load-balancer":
		_, httpResp, err = common.ExecuteWithRetryAndAuth(ctx, r.kc, respDiags,
			func() (*loadbalancer.BnsLoadBalancerV1ApiRemovePublicIpModelResponsePublicIpModel, *http.Response, error) {
				return r.kc.ApiClient.LoadBalancerAPI.
					RemovePublicIp(ctx, deviceId).
					XAuthToken(r.kc.XAuthToken).
					Execute()
			},
		)

	default:
		common.AddGeneralError(ctx, r, respDiags,
			fmt.Sprintf("Unsupported device_type: %s", deviceType),
		)
		return false
	}

	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "AssociatePublicIp", err, respDiags)
		return false
	}

	return true
}

func (r *publicIpResource) validateRelatedResourceConfig(
	ctx context.Context,
	config publicIpResourceModel,
	resp *resource.ValidateConfigResponse,
) {
	if config.RelatedResource.IsNull() || config.RelatedResource.IsUnknown() {
		return
	}

	var rr resourceModel
	diags := config.RelatedResource.As(ctx, &rr, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	deviceType := rr.DeviceType.ValueString()

	switch deviceType {
	case "instance":
		if rr.Id.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("related_resource").AtName("id"),
				"Missing required field(s)",
				"`related_resource.id` is required when `related_resource.device_type = instance`.",
			)
		}

	case "load-balancer":
		if !(rr.Id.IsNull() || rr.Id.IsUnknown()) {
			resp.Diagnostics.AddAttributeError(
				path.Root("related_resource").AtName("id"),
				"Invalid field",
				"`related_resource.id` must not be set when `related_resource.device_type = load-balancer`.",
			)
		}
	default:
		return
	}
}

func shouldReattachRelatedResource(ctx context.Context, planObj, stateObj types.Object) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	extract := func(obj types.Object) (resourceModel, bool) {
		var rr resourceModel
		if obj.IsNull() || obj.IsUnknown() {
			return rr, false
		}
		if d := obj.As(ctx, &rr, basetypes.ObjectAsOptions{}); d.HasError() {
			diags.Append(d...)
			return rr, false
		}
		return rr, true
	}

	planRelatedResource, hasPlan := extract(planObj)
	stateRelatedResource, hasState := extract(stateObj)

	if hasPlan != hasState {
		return true, diags
	}
	if !hasPlan && !hasState {
		return false, diags
	}

	if !planRelatedResource.DeviceType.Equal(stateRelatedResource.DeviceType) {
		return true, diags
	}
	if !planRelatedResource.DeviceId.Equal(stateRelatedResource.DeviceId) {
		return true, diags
	}

	if planRelatedResource.DeviceType.ValueString() == "instance" &&
		!planRelatedResource.Id.Equal(stateRelatedResource.Id) {
		return true, diags
	}
	return false, diags
}

func (r *publicIpResource) cleanupCreateFailurePublicIP(
	ctx context.Context,
	publicIpId string,
	diags *diag.Diagnostics,
) error {
	_, httpResp, err := common.ExecuteWithRetryAndAuth[struct{}](ctx, r.kc, diags,
		func() (struct{}, *http.Response, error) {
			resp, err := r.kc.ApiClient.PublicIPAPI.
				DeletePublicIp(ctx, publicIpId).
				XAuthToken(r.kc.XAuthToken).
				Execute()
			return struct{}{}, resp, err
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "DeletePublicIp(rollback)", err, diags)
		return err
	}

	common.PollUntilDeletion(ctx, r, 2*time.Second, diags, func(ctx context.Context) (bool, *http.Response, error) {
		_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, diags,
			func() (*network.BnsNetworkV1ApiGetPublicIpModelResponsePublicIpModel, *http.Response, error) {
				_, hr, err := r.kc.ApiClient.PublicIPAPI.
					GetPublicIp(ctx, publicIpId).
					XAuthToken(r.kc.XAuthToken).
					Execute()
				return nil, hr, err
			},
		)
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			return true, httpResp, nil
		}
		return false, httpResp, err
	})

	return nil
}

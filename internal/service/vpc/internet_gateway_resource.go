// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package vpc

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"terraform-provider-kakaocloud/internal/common"
	"terraform-provider-kakaocloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kakaoenterprise/kc-sdk-go/v2/services/vpc"
)

var (
	_ resource.ResourceWithConfigure   = &internetGatewayResource{}
	_ resource.ResourceWithImportState = &internetGatewayResource{}
)

func NewInternetGatewayResource() resource.Resource {
	return &internetGatewayResource{}
}

type internetGatewayResource struct {
	kc *common.KakaoCloudClient
}

func (r *internetGatewayResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_internet_gateway"
}

func (r *internetGatewayResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: utils.MergeResourceSchemaAttributes(
			internetGatewayResourceSchemaAttributes,
			map[string]schema.Attribute{
				"timeouts": timeouts.AttributesAll(ctx),
			},
		),
	}
}

func (r *internetGatewayResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan internetGatewayResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout, diags := plan.Timeouts.Create(ctx, common.DefaultCreateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	createIGW := vpc.NewCreateInternetGateway()
	if !plan.Name.IsNull() && !plan.Name.IsUnknown() {
		createIGW.SetName(plan.Name.ValueString())
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		createIGW.SetDescription(plan.Description.ValueString())
	}
	body := vpc.NewCreateInternetGatewayRequest(*createIGW)

	createResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*vpc.CreateInternetGatewayResponse, *http.Response, error) {
			return r.kc.ApiClient.VPCInternetGatewayAPI.
				CreateInternetGateway(ctx).
				XAuthToken(r.kc.XAuthToken).
				CreateInternetGatewayRequest(*body).
				Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "CreateInternetGateway", err, &resp.Diagnostics)
		return
	}

	plan.Id = types.StringValue(createResp.Igw.Id)

	result, ok := r.pollUntilAvailable(ctx, plan.Id.ValueString(), nil, nil, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	if !mapInternetGatewayModel(ctx, &plan.internetGatewayBaseModel, result, &resp.Diagnostics) || resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *internetGatewayResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state internetGatewayResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	readTimeout, diags := state.Timeouts.Read(ctx, common.DefaultReadTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, readTimeout)
	defer cancel()

	igwResp, httpResp, err := r.get(ctx, state.Id.ValueString(), &resp.Diagnostics)
	if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "GetInternetGateway", err, &resp.Diagnostics)
		return
	}

	if !mapInternetGatewayModel(ctx, &state.internetGatewayBaseModel, &igwResp.Igw, &resp.Diagnostics) || resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *internetGatewayResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state internetGatewayResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateTimeout, diags := plan.Timeouts.Update(ctx, common.DefaultUpdateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	if !plan.Name.Equal(state.Name) || !plan.Description.Equal(state.Description) {
		updateIGW := vpc.NewUpdateInternetGateway()
		if !plan.Name.Equal(state.Name) {
			updateIGW.SetName(plan.Name.ValueString())
		}
		if !plan.Description.Equal(state.Description) {
			updateIGW.SetDescription(plan.Description.ValueString())
		}
		body := vpc.NewUpdateInternetGatewayRequest(*updateIGW)

		_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
			func() (*vpc.UpdateInternetGatewayResponse, *http.Response, error) {
				return r.kc.ApiClient.VPCInternetGatewayAPI.
					UpdateInternetGateway(ctx, plan.Id.ValueString()).
					XAuthToken(r.kc.XAuthToken).
					UpdateInternetGatewayRequest(*body).
					Execute()
			},
		)
		if err != nil {
			common.AddApiActionError(ctx, r, httpResp, "UpdateInternetGateway", err, &resp.Diagnostics)
			return
		}
	}

	result, ok := r.pollUntilAvailable(
		ctx,
		plan.Id.ValueString(),
		new(plan.Name.ValueString()),
		new(plan.Description.ValueString()),
		&resp.Diagnostics,
	)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	if !mapInternetGatewayModel(ctx, &plan.internetGatewayBaseModel, result, &resp.Diagnostics) || resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *internetGatewayResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state internetGatewayResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteTimeout, diags := state.Timeouts.Delete(ctx, common.DefaultDeleteTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (struct{}, *http.Response, error) {
			httpResp, err := r.kc.ApiClient.VPCInternetGatewayAPI.
				DeleteInternetGateway(ctx, state.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				Execute()
			return struct{}{}, httpResp, err
		},
	)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			return
		}
		common.AddApiActionError(ctx, r, httpResp, "DeleteInternetGateway", err, &resp.Diagnostics)
		return
	}

	common.PollUntilDeletion(ctx, r, 5*time.Second, &resp.Diagnostics,
		func(ctx context.Context) (bool, *http.Response, error) {
			_, httpResp, err := r.get(ctx, state.Id.ValueString(), &resp.Diagnostics)
			return false, httpResp, err
		},
	)
}

func (r *internetGatewayResource) get(ctx context.Context, id string, diags *diag.Diagnostics) (*vpc.GetInternetGatewayResponse, *http.Response, error) {
	return common.ExecuteWithRetryAndAuth(ctx, r.kc, diags,
		func() (*vpc.GetInternetGatewayResponse, *http.Response, error) {
			return r.kc.ApiClient.VPCInternetGatewayAPI.
				GetInternetGateway(ctx, id).
				XAuthToken(r.kc.XAuthToken).
				Execute()
		},
	)
}

func (r *internetGatewayResource) pollUntilAvailable(
	ctx context.Context,
	id string,
	expectedName *string,
	expectedDescription *string,
	diags *diag.Diagnostics,
) (*vpc.Igw, bool) {
	return common.PollUntilResult(
		ctx,
		r,
		5*time.Second,
		"igw",
		id,
		[]string{common.VpcProvisioningStatusActive, common.VpcProvisioningStatusError},
		diags,
		func(ctx context.Context) (*vpc.Igw, *http.Response, error) {
			result, httpResp, err := r.get(ctx, id, diags)
			if err != nil {
				return nil, httpResp, err
			}
			return &result.Igw, httpResp, nil
		},
		func(result *vpc.Igw) string {
			if result == nil || result.ProvisioningStatus.Get() == nil {
				return ""
			}

			status := *result.ProvisioningStatus.Get()
			if status != common.VpcProvisioningStatusActive || expectedName == nil || expectedDescription == nil {
				return status
			}

			actualName := result.Name

			actualDescription := ""
			if result.Description.Get() != nil {
				actualDescription = *result.Description.Get()
			}

			if actualName != *expectedName || actualDescription != *expectedDescription {
				return "UPDATING"
			}

			return status
		},
	)
}

func (r *internetGatewayResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *internetGatewayResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package vpc

import (
	"context"
	"fmt"
	"net/http"
	"strings"
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
	_ resource.ResourceWithConfigure   = &internetGatewayAttachmentResource{}
	_ resource.ResourceWithImportState = &internetGatewayAttachmentResource{}
)

func NewInternetGatewayAttachmentResource() resource.Resource {
	return &internetGatewayAttachmentResource{}
}

type internetGatewayAttachmentResource struct {
	kc *common.KakaoCloudClient
}

func (r *internetGatewayAttachmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_internet_gateway_attachment"
}

func (r *internetGatewayAttachmentResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: utils.MergeResourceSchemaAttributes(
			internetGatewayAttachmentResourceSchemaAttributes,
			map[string]schema.Attribute{
				"timeouts": timeouts.AttributesAll(ctx),
			},
		),
	}
}

func (r *internetGatewayAttachmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan internetGatewayAttachmentResourceModel
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

	unlock, ok := common.AcquireLockForID(ctx, plan.InternetGatewayId.ValueString(), &resp.Diagnostics)
	if !ok {
		return
	}
	defer unlock()

	unlock2, ok := common.AcquireLockForID(ctx, plan.VpcId.ValueString(), &resp.Diagnostics)
	if !ok {
		return
	}
	defer unlock2()

	ok = pollVpcStatus(ctx, r, r.kc, plan.VpcId.ValueString(), &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	attachment := vpc.NewCreateIgwAttachment(plan.VpcId.ValueString())
	body := vpc.NewCreateIgwAttachmentRequest(*attachment)

	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*vpc.CreateIgwAttachmentResponse, *http.Response, error) {
			return r.kc.ApiClient.VPCInternetGatewayAPI.
				CreateIgwAttachment(ctx, plan.InternetGatewayId.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				CreateIgwAttachmentRequest(*body).
				Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "CreateIgwAttachment", err, &resp.Diagnostics)
		return
	}

	result, ok := r.pollUntilAttached(ctx, plan.InternetGatewayId.ValueString(), plan.VpcId.ValueString(), &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	ok = pollVpcStatus(ctx, r, r.kc, plan.VpcId.ValueString(), &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	if !mapInternetGatewayAttachmentModel(&plan.internetGatewayAttachmentModel, result, &resp.Diagnostics) {
		return
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *internetGatewayAttachmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state internetGatewayAttachmentResourceModel
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

	const maxAttachmentRetries = 2

	var igwResp *vpc.GetInternetGatewayResponse
	for attempt := 0; attempt <= maxAttachmentRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-time.After(3 * time.Second):
			case <-ctx.Done():
				resp.Diagnostics.AddError(
					"Error Retrying Internet Gateway Attachment Read",
					fmt.Sprintf("Context cancelled before retrying GetInternetGateway: %s", ctx.Err()),
				)
				return
			}
		}

		var httpResp *http.Response
		var err error
		igwResp, httpResp, err = r.get(ctx, state.InternetGatewayId.ValueString(), &resp.Diagnostics)
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		if err != nil {
			common.AddApiActionError(ctx, r, httpResp, "GetInternetGateway", err, &resp.Diagnostics)
			return
		}

		if igwResp.Igw.Attachment.Get() != nil {
			break
		}
	}

	attachment := igwResp.Igw.Attachment.Get()
	if attachment == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	if !state.Id.IsNull() && !state.Id.IsUnknown() && state.Id.ValueString() != attachment.Id {
		resp.State.RemoveResource(ctx)
		return
	}

	vpcInfo := igwResp.Igw.Vpc.Get()
	if vpcInfo == nil || vpcInfo.Id != state.VpcId.ValueString() {
		resp.State.RemoveResource(ctx)
		return
	}

	if !mapInternetGatewayAttachmentModel(&state.internetGatewayAttachmentModel, &igwResp.Igw, &resp.Diagnostics) {
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *internetGatewayAttachmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan internetGatewayAttachmentResourceModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *internetGatewayAttachmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state internetGatewayAttachmentResourceModel
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

	unlock, ok := common.AcquireLockForID(ctx, state.InternetGatewayId.ValueString(), &resp.Diagnostics)
	if !ok {
		return
	}
	defer unlock()

	unlock2, ok := common.AcquireLockForID(ctx, state.VpcId.ValueString(), &resp.Diagnostics)
	if !ok {
		return
	}
	defer unlock2()

	ok = pollVpcStatus(ctx, r, r.kc, state.VpcId.ValueString(), &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (struct{}, *http.Response, error) {
			httpResp, err := r.kc.ApiClient.VPCInternetGatewayAPI.
				DeleteIgwAttachment(ctx, state.InternetGatewayId.ValueString(), state.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				Execute()
			return struct{}{}, httpResp, err
		},
	)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			return
		}
		common.AddApiActionError(ctx, r, httpResp, "DeleteIgwAttachment", err, &resp.Diagnostics)
		return
	}

	common.PollUntilDeletion(ctx, r, 5*time.Second, &resp.Diagnostics,
		func(ctx context.Context) (bool, *http.Response, error) {
			result, httpResp, err := r.get(ctx, state.InternetGatewayId.ValueString(), &resp.Diagnostics)
			if err != nil {
				return false, httpResp, err
			}
			return result.Igw.Attachment.Get() == nil, httpResp, nil
		},
	)
}

func (r *internetGatewayAttachmentResource) get(ctx context.Context, id string, diags *diag.Diagnostics) (*vpc.GetInternetGatewayResponse, *http.Response, error) {
	return common.ExecuteWithRetryAndAuth(ctx, r.kc, diags,
		func() (*vpc.GetInternetGatewayResponse, *http.Response, error) {
			return r.kc.ApiClient.VPCInternetGatewayAPI.
				GetInternetGateway(ctx, id).
				XAuthToken(r.kc.XAuthToken).
				Execute()
		},
	)
}

func (r *internetGatewayAttachmentResource) pollUntilAttached(
	ctx context.Context,
	id string,
	vpcID string,
	diags *diag.Diagnostics,
) (*vpc.Igw, bool) {
	result, ok := common.PollUntilResult(
		ctx,
		r,
		5*time.Second,
		"internet gateway attachment",
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
			if result == nil || result.Attachment.Get() == nil {
				return "ATTACHING"
			}
			if result.Vpc.Get() == nil || result.Vpc.Get().Id != vpcID {
				return "ATTACHING"
			}
			status := result.Attachment.Get().ProvisioningStatus.Get()
			if status == nil {
				return "ATTACHING"
			}
			return *status
		},
	)
	if !ok || diags.HasError() {
		return nil, false
	}

	attachment := result.Attachment.Get()
	common.CheckResourceAvailableStatus(
		ctx,
		r,
		attachment.ProvisioningStatus.Get(),
		[]string{common.VpcProvisioningStatusActive},
		diags,
	)
	return result, !diags.HasError()
}

func mapInternetGatewayAttachmentModel(model *internetGatewayAttachmentModel, result *vpc.Igw, diags *diag.Diagnostics) bool {
	if result == nil || result.Attachment.Get() == nil {
		diags.AddError(
			"Internet Gateway Attachment Not Found",
			"GetInternetGateway returned no attachment.",
		)
		return false
	}

	attachment := result.Attachment.Get()
	model.Id = types.StringValue(attachment.Id)
	model.ProvisioningStatus = utils.ConvertNullableString(attachment.ProvisioningStatus)
	return true
}

func (r *internetGatewayAttachmentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *internetGatewayAttachmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Expected import ID in the format <internet_gateway_id>/<vpc_id>.",
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("internet_gateway_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("vpc_id"), parts[1])...)
}

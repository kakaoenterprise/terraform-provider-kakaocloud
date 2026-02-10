// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package tgw

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-kakaocloud/internal/common"
	"terraform-provider-kakaocloud/internal/utils"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kakaoenterprise/kc-sdk-go/services/tgw"
)

var (
	_ resource.ResourceWithConfigure   = &transitGatewayAttachmentApprovalResource{}
	_ resource.ResourceWithImportState = &transitGatewayAttachmentApprovalResource{}
)

func NewTransitGatewayAttachmentApprovalResource() resource.Resource {
	return &transitGatewayAttachmentApprovalResource{}
}

type transitGatewayAttachmentApprovalResource struct {
	kc *common.KakaoCloudClient
}

func (r *transitGatewayAttachmentApprovalResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_transit_gateway_attachment_approval"
}

func (r *transitGatewayAttachmentApprovalResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: utils.MergeResourceSchemaAttributes(
			transitGatewayAttachmentApprovalResourceSchemaAttributes,
			map[string]schema.Attribute{
				"timeouts": timeouts.AttributesAll(ctx),
			},
		),
	}
}

func (r *transitGatewayAttachmentApprovalResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan transitGatewayAttachmentApprovalResourceModel
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

	attachmentId := plan.AttachmentId.ValueString()

	approveResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*tgw.BnsTgwV1ApiApproveTgwAttachmentModelCreateTgwAttachmentResponseModel, *http.Response, error) {
			return r.kc.ApiClient.AttachmentsAPI.ApproveTgwAttachment(ctx, attachmentId).
				XAuthToken(r.kc.XAuthToken).
				Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "ApproveTgwAttachment", err, &resp.Diagnostics)
		return
	}

	plan.Id = types.StringValue(approveResp.Attachment.Id)

	result, ok := common.PollUntilResult(
		ctx, r, 10*time.Second, "transit gateway route", attachmentId, []string{common.TgwStatusActive, common.TgwStatusError, common.TgwStatusInUse, common.TgwStatusInactive, common.TgwStatusAvaliable}, &resp.Diagnostics,
		func(ctx context.Context) (*tgw.BnsTgwV1ApiGetTgwAttachmentModelTgwAttachmentResponseModel, *http.Response, error) {
			attachResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
				func() (*tgw.GetTgwAttachmentResponseModel, *http.Response, error) {
					return r.kc.ApiClient.AttachmentsAPI.GetTgwAttachment(ctx, attachmentId).
						XAuthToken(r.kc.XAuthToken).
						Execute()
				},
			)
			if err != nil {
				return nil, httpResp, err
			}
			return &attachResp.Attachment, httpResp, nil
		},
		func(v *tgw.BnsTgwV1ApiGetTgwAttachmentModelTgwAttachmentResponseModel) string {
			if v.ProvisioningStatus.IsSet() {
				return string(*v.ProvisioningStatus.Get())
			}
			return ""
		},
	)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	if result.ProvisioningStatus.IsSet() {
		status := string(*result.ProvisioningStatus.Get())
		common.CheckResourceAvailableStatus(ctx, r, &status, []string{common.TgwStatusActive}, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	ok = mapAttachmentApprovalToModel(ctx, &plan, result, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *transitGatewayAttachmentApprovalResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state transitGatewayAttachmentApprovalResourceModel
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
		func() (*tgw.GetTgwAttachmentResponseModel, *http.Response, error) {
			return r.kc.ApiClient.AttachmentsAPI.GetTgwAttachment(ctx, state.AttachmentId.ValueString()).
				XAuthToken(r.kc.XAuthToken).Execute()
		},
	)

	if httpResp != nil && httpResp.StatusCode == 404 {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "GetTgwAttachment", err, &resp.Diagnostics)
		return
	}

	status := string(*respModel.Attachment.ProvisioningStatus.Get())
	if status == common.TgwStatusPendingApprove {
		common.AddGeneralError(ctx, r, &resp.Diagnostics, "The attachment is still pending approval.")
		if resp.Diagnostics.HasError() {
			return
		}
	}

	ok := mapAttachmentApprovalToModel(ctx, &state, &respModel.Attachment, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *transitGatewayAttachmentApprovalResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	common.AddGeneralError(
		ctx, r, &resp.Diagnostics,
		"Updates are not supported for transit_gateway_attachment_approval. The attachment_id requires replacement.",
	)
}

func (r *transitGatewayAttachmentApprovalResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

func (r *transitGatewayAttachmentApprovalResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *transitGatewayAttachmentApprovalResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("attachment_id"), req.ID)...)
}

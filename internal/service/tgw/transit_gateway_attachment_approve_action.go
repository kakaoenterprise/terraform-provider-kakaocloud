// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package tgw

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"terraform-provider-kakaocloud/internal/common"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/kakaoenterprise/kc-sdk-go/v2/services/tgw"
)

var _ action.ActionWithConfigure = &transitGatewayAttachmentApproveAction{}

var transitGatewayAttachmentApproveActionPollInterval = 10 * time.Second

func NewTransitGatewayAttachmentApproveAction() action.Action {
	return &transitGatewayAttachmentApproveAction{}
}

type transitGatewayAttachmentApproveAction struct {
	transitGatewayAttachmentActionBase
}

func (a *transitGatewayAttachmentApproveAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_transit_gateway_attachment_approve"
}

func (a *transitGatewayAttachmentApproveAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = transitGatewayAttachmentActionSchema()
}

func (a *transitGatewayAttachmentApproveAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	a.configure(req, resp)
}

func (a *transitGatewayAttachmentApproveAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var config transitGatewayAttachmentActionModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, common.DefaultCreateTimeout)
	defer cancel()

	attachmentId := config.AttachmentId.ValueString()
	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, a.kc, &resp.Diagnostics,
		func() (*tgw.ApproveTgwAttachmentResponse, *http.Response, error) {
			return a.kc.ApiClient.TransitGatewayAttachmentAPI.
				ApproveTgwAttachment(ctx, attachmentId).
				XAuthToken(a.kc.XAuthToken).
				Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, a, httpResp, "ApproveTgwAttachment", err, &resp.Diagnostics)
		return
	}

	stopProgress := common.StartActionProgress(ctx, resp.SendProgress, fmt.Sprintf("Waiting for transit gateway attachment %s approval to complete", attachmentId))
	defer stopProgress()

	result, ok := common.PollUntilResult(
		ctx,
		a,
		transitGatewayAttachmentApproveActionPollInterval,
		"transit gateway attachment",
		attachmentId,
		[]string{
			string(tgw.TGWATTACHMENTPROVISIONINGSTATUS_ACTIVE),
			string(tgw.TGWATTACHMENTPROVISIONINGSTATUS_IN_USE),
			string(tgw.TGWATTACHMENTPROVISIONINGSTATUS_INACTIVE),
			string(tgw.TGWATTACHMENTPROVISIONINGSTATUS_AVAILABLE),
			string(tgw.TGWATTACHMENTPROVISIONINGSTATUS_ERROR),
			string(tgw.TGWATTACHMENTPROVISIONINGSTATUS_REJECTED),
			string(tgw.TGWATTACHMENTPROVISIONINGSTATUS_DELETED),
		},
		&resp.Diagnostics,
		func(ctx context.Context) (*tgw.TgwAttachment, *http.Response, error) {
			attachmentResponse, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, a.kc, &resp.Diagnostics,
				func() (*tgw.GetTgwAttachmentResponse, *http.Response, error) {
					return a.kc.ApiClient.TransitGatewayAttachmentAPI.
						GetTgwAttachment(ctx, attachmentId).
						XAuthToken(a.kc.XAuthToken).
						Execute()
				},
			)
			if err != nil {
				return nil, httpResp, err
			}
			return &attachmentResponse.Attachment, httpResp, nil
		},
		func(attachment *tgw.TgwAttachment) string {
			if attachment.ProvisioningStatus.IsSet() && attachment.ProvisioningStatus.Get() != nil {
				return string(*attachment.ProvisioningStatus.Get())
			}
			return ""
		},
	)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	if !result.ProvisioningStatus.IsSet() || result.ProvisioningStatus.Get() == nil {
		common.AddGeneralError(ctx, a, &resp.Diagnostics, "attachment provisioning status is missing after approval")
		return
	}

	status := *result.ProvisioningStatus.Get()
	if status != tgw.TGWATTACHMENTPROVISIONINGSTATUS_ACTIVE {
		common.AddGeneralError(ctx, a, &resp.Diagnostics, fmt.Sprintf("attachment approval finished with unexpected status %q", status))
		return
	}
}

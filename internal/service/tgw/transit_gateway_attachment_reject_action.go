// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package tgw

import (
	"context"
	"fmt"
	"net/http"

	"terraform-provider-kakaocloud/internal/common"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/kakaoenterprise/kc-sdk-go/v2/services/tgw"
)

var _ action.ActionWithConfigure = &transitGatewayAttachmentRejectAction{}

func NewTransitGatewayAttachmentRejectAction() action.Action {
	return &transitGatewayAttachmentRejectAction{}
}

type transitGatewayAttachmentRejectAction struct {
	transitGatewayAttachmentActionBase
}

func (a *transitGatewayAttachmentRejectAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_transit_gateway_attachment_reject"
}

func (a *transitGatewayAttachmentRejectAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = transitGatewayAttachmentActionSchema()
}

func (a *transitGatewayAttachmentRejectAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	a.configure(req, resp)
}

func (a *transitGatewayAttachmentRejectAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var config transitGatewayAttachmentActionModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, common.DefaultDeleteTimeout)
	defer cancel()

	attachmentId := config.AttachmentId.ValueString()
	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, a.kc, &resp.Diagnostics,
		func() (*tgw.RejectTgwAttachmentResponse, *http.Response, error) {
			return a.kc.ApiClient.TransitGatewayAttachmentAPI.
				RejectTgwAttachment(ctx, attachmentId).
				XAuthToken(a.kc.XAuthToken).
				Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, a, httpResp, "RejectTgwAttachment", err, &resp.Diagnostics)
		return
	}

	stopProgress := common.StartActionProgress(ctx, resp.SendProgress, fmt.Sprintf("Waiting for transit gateway attachment %s rejection to delete the attachment", attachmentId))
	defer stopProgress()

	common.PollUntilDeletion(ctx, a, transitGatewayAttachmentActionPollInterval, &resp.Diagnostics, func(ctx context.Context) (bool, *http.Response, error) {
		attachmentResponse, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, a.kc, &resp.Diagnostics,
			func() (*tgw.GetTgwAttachmentResponse, *http.Response, error) {
				return a.kc.ApiClient.TransitGatewayAttachmentAPI.
					GetTgwAttachment(ctx, attachmentId).
					XAuthToken(a.kc.XAuthToken).
					Execute()
			},
		)
		if err != nil {
			return false, httpResp, err
		}

		if attachmentResponse.Attachment.ProvisioningStatus.IsSet() && attachmentResponse.Attachment.ProvisioningStatus.Get() != nil &&
			*attachmentResponse.Attachment.ProvisioningStatus.Get() == tgw.TGWATTACHMENTPROVISIONINGSTATUS_ERROR {
			common.AddGeneralError(ctx, a, &resp.Diagnostics, "attachment rejection finished with status \"ERROR\"")
			return true, httpResp, nil
		}

		return false, httpResp, nil
	})
}

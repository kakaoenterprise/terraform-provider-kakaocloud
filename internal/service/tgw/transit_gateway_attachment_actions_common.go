// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package tgw

import (
	"fmt"
	"time"

	"terraform-provider-kakaocloud/internal/common"

	"github.com/hashicorp/terraform-plugin-framework/action"
	actionschema "github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var transitGatewayAttachmentActionPollInterval = 3 * time.Second

type transitGatewayAttachmentActionModel struct {
	AttachmentId types.String `tfsdk:"attachment_id"`
}

type transitGatewayAttachmentActionBase struct {
	kc *common.KakaoCloudClient
}

func (a *transitGatewayAttachmentActionBase) configure(req action.ConfigureRequest, resp *action.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*common.KakaoCloudClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Action Configure Type",
			fmt.Sprintf("Expected *common.KakaoCloudClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	a.kc = client
}

func transitGatewayAttachmentActionSchema() actionschema.Schema {
	return actionschema.Schema{
		Attributes: map[string]actionschema.Attribute{
			"attachment_id": actionschema.StringAttribute{
				Required:   true,
				Validators: common.UuidValidator(),
			},
		},
	}
}

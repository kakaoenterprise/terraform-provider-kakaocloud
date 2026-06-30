// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package mysql

import (
	"context"
	"fmt"
	"slices"
	"terraform-provider-kakaocloud/internal/common"

	"github.com/hashicorp/terraform-plugin-framework/action"
	actionschema "github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type mysqlActionBase struct {
	kc *common.KakaoCloudClient
}

func (a *mysqlActionBase) configure(req action.ConfigureRequest, resp *action.ConfigureResponse) {
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

func mysqlActionInstanceGroupIDAttribute() actionschema.StringAttribute {
	return actionschema.StringAttribute{
		Required:   true,
		Validators: common.UuidValidator(),
	}
}

func mysqlActionInstanceIDsFromSet(ctx context.Context, set types.Set, respDiags *diag.Diagnostics) ([]string, bool) {
	var instanceIDs []string
	respDiags.Append(set.ElementsAs(ctx, &instanceIDs, false)...)
	if respDiags.HasError() {
		return nil, false
	}

	slices.Sort(instanceIDs)
	return instanceIDs, true
}

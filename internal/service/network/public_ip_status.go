// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package network

import (
	"context"
	"net/http"
	"terraform-provider-kakaocloud/internal/common"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/kakaoenterprise/kc-sdk-go/v2/services/network"
)

func (r *publicIpResource) pollPublicIpUtilsStatus(
	ctx context.Context,
	publicIpId string,
	targetStatuses []string,
	resp *diag.Diagnostics,
) (*network.PublicIp, bool) {
	return common.PollUntilResult(
		ctx,
		r,
		10*time.Second,
		"public ip",
		publicIpId,
		targetStatuses,
		resp,
		func(ctx context.Context) (*network.PublicIp, *http.Response, error) {
			respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, resp,
				func() (*network.GetPublicIpResponse, *http.Response, error) {
					return r.kc.ApiClient.NetworkPublicIPAPI.
						GetPublicIp(ctx, publicIpId).
						XAuthToken(r.kc.XAuthToken).
						Execute()
				},
			)
			if err != nil {
				return nil, httpResp, err
			}
			return &respModel.PublicIp, httpResp, nil
		},
		func(v *network.PublicIp) string {
			if v.Status.Get() == nil {
				return ""
			}
			return string(*v.Status.Get())
		},
	)
}

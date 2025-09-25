// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package network

import (
	"context"
	"net/http"
	"terraform-provider-kakaocloud/internal/common"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/kakaoenterprise/kc-sdk-go/services/network"
)

func (r *publicIpResource) pollPublicIpUtilsStatus(
	ctx context.Context,
	publicIpId string,
	targetStatuses []string,
	resp *diag.Diagnostics,
) (*network.BnsNetworkV1ApiGetPublicIpModelFloatingIpModel, bool) {
	return common.PollUntilResult(
		ctx,
		r,
		10*time.Second,
		targetStatuses,
		resp,
		func(ctx context.Context) (*network.BnsNetworkV1ApiGetPublicIpModelFloatingIpModel, *http.Response, error) {
			respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, resp,
				func() (*network.BnsNetworkV1ApiGetPublicIpModelResponsePublicIpModel, *http.Response, error) {
					return r.kc.ApiClient.PublicIPAPI.
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
		func(v *network.BnsNetworkV1ApiGetPublicIpModelFloatingIpModel) string { return string(*v.Status.Get()) },
	)
}

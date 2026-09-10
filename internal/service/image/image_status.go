// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package image

import (
	"context"
	"net/http"
	"terraform-provider-kakaocloud/internal/common"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/kakaoenterprise/kc-sdk-go/v2/services/image"
)

func (r *imageResource) pollImageUtilsStatus(
	ctx context.Context,
	imageId string,
	targetStatuses []string,
	resp *diag.Diagnostics,
) (*image.Image, bool) {
	return common.PollUntilResult(
		ctx,
		r,
		2*time.Second,
		"image",
		imageId,
		targetStatuses,
		resp,
		func(ctx context.Context) (*image.Image, *http.Response, error) {
			respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, resp,
				func() (*image.GetImageResponse, *http.Response, error) {
					return r.kc.ApiClient.ImageAPI.
						GetImage(ctx, imageId).
						XAuthToken(r.kc.XAuthToken).
						Execute()
				},
			)
			if err != nil {
				return nil, httpResp, err
			}
			return &respModel.Image, httpResp, nil
		},
		func(v *image.Image) string { return *v.Status.Get() },
	)
}

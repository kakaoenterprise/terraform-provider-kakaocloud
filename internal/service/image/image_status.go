// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package image

import (
	"context"
	"net/http"
	"terraform-provider-kakaocloud/internal/common"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/kakaoenterprise/kc-sdk-go/services/image"
)

// BcsImageV1ApiGetImageModelResponseImageModel
func (r *imageResource) pollImageUtilsStatus(
	ctx context.Context,
	imageId string,
	targetStatuses []string,
	resp *diag.Diagnostics,
) (*image.BcsImageV1ApiGetImageModelImageModel, bool) {
	return common.PollUntilResult(
		ctx,
		r,
		2*time.Second,
		targetStatuses,
		resp,
		func(ctx context.Context) (*image.BcsImageV1ApiGetImageModelImageModel, *http.Response, error) {
			respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, resp,
				func() (*image.BcsImageV1ApiGetImageModelResponseImageModel, *http.Response, error) {
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
		func(v *image.BcsImageV1ApiGetImageModelImageModel) string { return *v.Status.Get() },
	)
}

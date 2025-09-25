// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kubernetesengine

import (
	"context"
	"net/http"
	"terraform-provider-kakaocloud/internal/common"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/kakaoenterprise/kc-sdk-go/services/kubernetesengine"
)

func (r *clusterResource) pollClusterUtilStatus(
	ctx context.Context,
	clusterName string,
	targetStatuses []string,
	diag *diag.Diagnostics,
) (*kubernetesengine.KubernetesEngineV1ApiGetClusterModelClusterResponseModel, bool) {
	return common.PollUntilResult(
		ctx,
		r,
		10*time.Second,
		targetStatuses,
		diag,
		func(ctx context.Context) (*kubernetesengine.KubernetesEngineV1ApiGetClusterModelClusterResponseModel, *http.Response, error) {
			respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, diag,
				func() (*kubernetesengine.GetK8sClusterResponseModel, *http.Response, error) {
					return r.kc.ApiClient.ClustersAPI.
						GetCluster(ctx, clusterName).
						XAuthToken(r.kc.XAuthToken).
						Execute()
				},
			)
			if err != nil {
				return nil, httpResp, err
			}
			return &respModel.Cluster, httpResp, nil
		},
		func(v *kubernetesengine.KubernetesEngineV1ApiGetClusterModelClusterResponseModel) string {
			return string(v.Status.Phase)
		},
	)
}

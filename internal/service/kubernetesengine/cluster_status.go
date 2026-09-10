// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package kubernetesengine

import (
	"context"
	"net/http"
	"terraform-provider-kakaocloud/internal/common"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/kakaoenterprise/kc-sdk-go/v2/services/kubernetesengine"
)

func (r *clusterResource) pollClusterUtilStatus(
	ctx context.Context,
	clusterName string,
	targetStatuses []string,
	diag *diag.Diagnostics,
) (*kubernetesengine.Cluster, bool) {
	return common.PollUntilResult(
		ctx,
		r,
		10*time.Second,
		"cluster",
		clusterName,
		targetStatuses,
		diag,
		func(ctx context.Context) (*kubernetesengine.Cluster, *http.Response, error) {
			respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, diag,
				func() (*kubernetesengine.GetClusterResponse, *http.Response, error) {
					return r.kc.ApiClient.KubernetesEngineClusterAPI.
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
		func(v *kubernetesengine.Cluster) string {
			return string(v.Status.Phase)
		},
	)
}

func (r *nodePoolResource) pollClusterUtilAvailableStatus(
	ctx context.Context,
	clusterName string,
	diag *diag.Diagnostics,
) (*kubernetesengine.Cluster, bool) {
	return common.PollUntilResult(
		ctx,
		r,
		10*time.Second,
		"cluster",
		clusterName,
		ClusterAvailableStatuses,
		diag,
		func(ctx context.Context) (*kubernetesengine.Cluster, *http.Response, error) {
			respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, diag,
				func() (*kubernetesengine.GetClusterResponse, *http.Response, error) {
					return r.kc.ApiClient.KubernetesEngineClusterAPI.
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
		func(v *kubernetesengine.Cluster) string {
			return string(v.Status.Phase)
		},
	)
}

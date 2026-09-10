package // Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
kubernetesengine

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"terraform-provider-kakaocloud/internal/common"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/kakaoenterprise/kc-sdk-go/v2/services/kubernetesengine"
)

var (
	NodePoolStatuesReady = []string{
		string(kubernetesengine.NODEPOOLSTATUS_RUNNING),
		string(kubernetesengine.NODEPOOLSTATUS_RUNNING__SCHEDULING_DISABLE),
	}

	NodePoolStatusesReadyOrFailed = []string{
		string(kubernetesengine.NODEPOOLSTATUS_RUNNING),
		string(kubernetesengine.NODEPOOLSTATUS_RUNNING__SCHEDULING_DISABLE),
		string(kubernetesengine.NODEPOOLSTATUS_FAILED),
		string(kubernetesengine.NODEPOOLSTATUS_DELETING),
	}

	NodePoolStatusesReadyToDelete = []string{
		string(kubernetesengine.NODEPOOLSTATUS_RUNNING),
		string(kubernetesengine.NODEPOOLSTATUS_RUNNING__SCHEDULING_DISABLE),
		string(kubernetesengine.NODEPOOLSTATUS_FAILED),
		string(kubernetesengine.NODEPOOLSTATUS_PENDING),
	}

	ClusterAvailableStatuses = []string{
		common.ClusterStatusProvisioned,
		common.ClusterStatusFailed,
		common.ClusterStatusDeleting,
	}

	NodePoolClusterAvailableStatuses = []string{
		common.ClusterStatusProvisioned,
		common.ClusterStatusFailed,
	}
)

func waitNodePool(
	ctx context.Context,
	client *common.KakaoCloudClient,
	resource interface{},
	clusterName string,
	nodePoolName string,
	targetStatuses []string,
	diags *diag.Diagnostics,
) (*kubernetesengine.NodePool, bool) {
	return pollNodePool(
		ctx,
		client,
		resource,
		clusterName,
		nodePoolName,
		targetStatuses,
		diags,
	)
}

func pollNodePool(
	ctx context.Context,
	client *common.KakaoCloudClient,
	resource interface{},
	clusterName string,
	nodePoolName string,
	targetStatuses []string,
	diags *diag.Diagnostics,
) (*kubernetesengine.NodePool, bool) {
	result, ok := common.PollUntilResult(
		ctx,
		resource,
		5*time.Second,
		"node pool",
		nodePoolName,
		targetStatuses,
		diags,
		func(ctx context.Context) (*kubernetesengine.NodePool, *http.Response, error) {
			model, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, client, diags,
				func() (*kubernetesengine.GetNodePoolResponse, *http.Response, error) {
					return client.ApiClient.KubernetesEngineNodePoolAPI.
						GetNodePool(ctx, clusterName, nodePoolName).
						XAuthToken(client.XAuthToken).
						Execute()
				})
			if err != nil {
				return nil, httpResp, err
			}
			return &model.NodePool, httpResp, nil
		},
		func(v *kubernetesengine.NodePool) string {
			return string(v.Status.Phase)
		},
	)
	if !ok {
		for _, d := range diags.Errors() {
			if strings.Contains(d.Detail(), "context deadline exceeded") {
				common.AddGeneralError(ctx, resource, diags,
					fmt.Sprintf("Node Pool %s did not reach one of the following states: '%v'.", nodePoolName, NodePoolStatuesReady))
				return result, false
			}
		}
	}
	return result, ok
}

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kubernetesengine

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-kakaocloud/internal/common"
	. "terraform-provider-kakaocloud/internal/utils"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kakaoenterprise/kc-sdk-go/services/kubernetesengine"
)

var (
	_ resource.ResourceWithConfigure      = &clusterNodeResource{}
	_ resource.ResourceWithImportState    = &clusterNodeResource{}
	_ resource.ResourceWithValidateConfig = &clusterNodeResource{}
)

func NewClusterNodeResource() resource.Resource { return &clusterNodeResource{} }

type clusterNodeResource struct {
	kc *common.KakaoCloudClient
}

func (r *clusterNodeResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config clusterNodeResourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	isRemoveKnown := !config.IsRemove.IsNull() && !config.IsRemove.IsUnknown()
	isCordonKnown := !config.IsCordon.IsNull() && !config.IsCordon.IsUnknown()

	if !isRemoveKnown && !isCordonKnown {
		resp.Diagnostics.AddError(
			"Missing operation selector",
			"Either 'is_remove' or 'is_cordon' must be set. These two are mutually exclusive.",
		)
		return
	}

	if isRemoveKnown && isCordonKnown {
		resp.Diagnostics.AddError(
			"Conflicting arguments",
			"'is_remove' and 'is_cordon' cannot be set together. Set exactly one.",
		)
		return
	}
}

func (r *clusterNodeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.AddError("Import Not Supported", "This resource cannot be imported.")
}

func (r *clusterNodeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kubernetes_engine_cluster_node"
}

func (r *clusterNodeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: MergeResourceSchemaAttributes(
			nodeResourceSchemaAttributes,
			map[string]schema.Attribute{
				"timeouts": timeouts.AttributesAll(ctx),
			},
		),
	}
}

func (r *clusterNodeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan clusterNodeResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	timeout, diags := plan.Timeouts.Create(ctx, common.DefaultCreateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cluster := plan.ClusterName.ValueString()

	nodeNames := StringsFromSet(ctx, plan.NodeNames, &resp.Diagnostics)

	isRemoveKnown := !plan.IsRemove.IsNull() && !plan.IsRemove.IsUnknown()
	isCordonKnown := !plan.IsCordon.IsNull() && !plan.IsCordon.IsUnknown()

	switch {
	case isRemoveKnown:
		isRemove := plan.IsRemove.ValueBool()

		body := kubernetesengine.DeleteK8sClusterNodesRequestModel{
			Cluster: kubernetesengine.KubernetesEngineV1ApiDeleteClusterNodesModelClusterRequestModel{
				IsRemove:  isRemove,
				NodeNames: nodeNames,
			},
		}

		_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
			func() (interface{}, *http.Response, error) {
				req := r.kc.ApiClient.ClustersAPI.
					DeleteClusterNodes(ctx, cluster).
					XAuthToken(r.kc.XAuthToken).
					DeleteK8sClusterNodesRequestModel(body)

				httpResp, err := req.Execute()
				return nil, httpResp, err
			},
		)
		if err != nil {
			common.AddApiActionError(ctx, r, httpResp, "DeleteClusterNodes", err, &resp.Diagnostics)
			return
		}

		type pollResult struct{ AllGone bool }

		_, ok := common.PollUntilResult(
			ctx,
			r,
			2*time.Second,
			[]string{"done"},
			&resp.Diagnostics,
			func(ctx context.Context) (pollResult, *http.Response, error) {
				modelResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
					func() (*kubernetesengine.GetK8sClusterNodesResponseModel, *http.Response, error) {
						return r.kc.ApiClient.ClustersAPI.
							ListClusterNodes(ctx, cluster).
							XAuthToken(r.kc.XAuthToken).
							Execute()
					},
				)
				if err != nil {
					return pollResult{}, httpResp, err
				}
				present := map[string]bool{}
				for _, n := range nodeNames {
					present[n] = false
				}
				for _, it := range modelResp.Nodes {
					if _, hit := present[it.Name]; hit {
						present[it.Name] = true
					}
				}
				for _, p := range present {
					if p {
						return pollResult{AllGone: false}, httpResp, nil
					}
				}
				return pollResult{AllGone: true}, httpResp, nil
			},
			func(pr pollResult) string {
				if pr.AllGone {
					return "done"
				}
				return "pending"
			},
		)
		if !ok {
			return
		}

	case isCordonKnown:
		desired := plan.IsCordon.ValueBool()

		body := kubernetesengine.UpdateK8sClusterNodesCordonRequestModel{
			Cluster: kubernetesengine.KubernetesEngineV1ApiSetClusterNodesCordonModelClusterRequestModel{
				IsCordon:  desired,
				NodeNames: nodeNames,
			},
		}

		_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
			func() (interface{}, *http.Response, error) {
				req := r.kc.ApiClient.ClustersAPI.
					SetClusterNodesCordon(ctx, cluster).
					XAuthToken(r.kc.XAuthToken).
					UpdateK8sClusterNodesCordonRequestModel(body)

				httpResp, err := req.Execute()
				return nil, httpResp, err
			},
		)
		if err != nil {
			common.AddApiActionError(ctx, r, httpResp, "CordonClusterNodes", err, &resp.Diagnostics)
			return
		}

		_, ok := common.PollUntilResult(
			ctx, r, 2*time.Second, []string{"done"}, &resp.Diagnostics,
			func(ctx context.Context) (bool, *http.Response, error) {
				modelResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
					func() (*kubernetesengine.GetK8sClusterNodesResponseModel, *http.Response, error) {
						return r.kc.ApiClient.ClustersAPI.
							ListClusterNodes(ctx, cluster).
							XAuthToken(r.kc.XAuthToken).
							Execute()
					},
				)
				if err != nil {
					return false, httpResp, err
				}

				hit := 0
				want := len(nodeNames)
				set := map[string]struct{}{}
				for _, n := range nodeNames {
					set[n] = struct{}{}
				}
				for _, it := range modelResp.Nodes {
					if _, ok := set[it.Name]; ok && it.IsCordon == desired {
						hit++
					}
				}
				return hit == want, httpResp, nil
			},
			func(done bool) string {
				if done {
					return "done"
				}
				return "pending"
			},
		)
		if !ok {
			return
		}

	default:
		resp.Diagnostics.AddError(
			"Missing operation selector",
			"Either 'is_remove' or 'is_cordon' must be set. These two are mutually exclusive.",
		)
		return
	}

	plan.Id = types.StringValue(
		fmt.Sprintf("%s/%d", cluster, time.Now().UnixNano()),
	)

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *clusterNodeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state clusterNodeResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *clusterNodeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"This resource does not support update. Please recreate the resource if needed.",
	)
}

func (r *clusterNodeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.State.RemoveResource(ctx)
}

func (r *clusterNodeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*common.KakaoCloudClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *kakaocloud.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.kc = client
}

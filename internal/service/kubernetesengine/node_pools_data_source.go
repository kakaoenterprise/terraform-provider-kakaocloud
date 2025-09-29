// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package kubernetesengine

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-kakaocloud/internal/common"
	. "terraform-provider-kakaocloud/internal/utils"

	datasourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/kakaoenterprise/kc-sdk-go/services/kubernetesengine"
)

var (
	_ datasource.DataSource              = &nodePoolsDataSource{}
	_ datasource.DataSourceWithConfigure = &nodePoolsDataSource{}
)

func NewNodePoolsDataSource() datasource.DataSource { return &nodePoolsDataSource{} }

type nodePoolsDataSource struct{ kc *common.KakaoCloudClient }

func (d *nodePoolsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
	d.kc = client
}

func (d *nodePoolsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kubernetes_engine_node_pools"
}

func (d *nodePoolsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Data source to retrieve the list of node pools for a specific cluster",
		Attributes: map[string]schema.Attribute{
			"cluster_name": schema.StringAttribute{Required: true, Description: "클러스터 이름"},
			"node_pools": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: MergeAttributes[schema.Attribute](
						map[string]schema.Attribute{
							"name": schema.StringAttribute{Computed: true},
						},
						nodePoolDataSourceAttributes,
					),
				},
			},
			"timeouts": datasourceTimeouts.Attributes(ctx),
		},
	}
}

func (d *nodePoolsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config nodePoolsDataSourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	readTimeout, diags := config.Timeouts.Read(ctx, common.DefaultReadTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, readTimeout)
	defer cancel()

	listResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, d.kc, &resp.Diagnostics,
		func() (struct{ NodePools []struct{ Name string } }, *http.Response, error) {
			api := d.kc.ApiClient.NodePoolsAPI.ListNodePools(ctx, config.ClusterName.ValueString()).XAuthToken(d.kc.XAuthToken)
			resp, httpResp, err := api.Execute()
			if err != nil {
				return struct{ NodePools []struct{ Name string } }{}, httpResp, err
			}

			tmp := struct{ NodePools []struct{ Name string } }{NodePools: make([]struct{ Name string }, len(resp.NodePools))}
			for i, np := range resp.NodePools {
				tmp.NodePools[i] = struct{ Name string }{Name: np.Name}
			}
			return tmp, httpResp, nil
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, d, httpResp, "ListNodePools", err, &resp.Diagnostics)
		return
	}

	for _, np := range listResp.NodePools {
		var model NodePoolBaseModelDS

		detail, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, d.kc, &resp.Diagnostics,
			func() (struct {
				NodePool kubernetesengine.KubernetesEngineV1ApiGetNodePoolModelNodePoolResponseModel
			}, *http.Response, error) {
				apiResp, httpResp, err := d.kc.ApiClient.NodePoolsAPI.
					GetNodePool(ctx, config.ClusterName.ValueString(), np.Name).
					XAuthToken(d.kc.XAuthToken).
					Execute()
				if err != nil {
					return struct {
						NodePool kubernetesengine.KubernetesEngineV1ApiGetNodePoolModelNodePoolResponseModel
					}{}, httpResp, err
				}
				return struct {
					NodePool kubernetesengine.KubernetesEngineV1ApiGetNodePoolModelNodePoolResponseModel
				}{NodePool: apiResp.NodePool}, httpResp, nil
			},
		)
		if err != nil {
			common.AddApiActionError(ctx, d, httpResp, "GetNodePool", err, &resp.Diagnostics)
			return
		}
		res := detail.NodePool
		ok := mapNodePoolFromResponseDS(ctx, &model, &res, &resp.Diagnostics)
		if !ok || resp.Diagnostics.HasError() {
			return
		}
		config.NodePools = append(config.NodePools, model)
	}

	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
}

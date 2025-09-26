// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package kubernetesengine

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-kakaocloud/internal/common"
	. "terraform-provider-kakaocloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/kakaoenterprise/kc-sdk-go/services/kubernetesengine"
)

var (
	_ datasource.DataSource              = &nodePoolDataSource{}
	_ datasource.DataSourceWithConfigure = &nodePoolDataSource{}
)

func NewNodePoolDataSource() datasource.DataSource { return &nodePoolDataSource{} }

type nodePoolDataSource struct {
	kc *common.KakaoCloudClient
}

func (d *nodePoolDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *nodePoolDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kubernetes_engine_node_pool"
}

func (d *nodePoolDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "kakaocloud 특정 클러스터의 노드 풀을 조회하는 데이터 소스",
		Attributes: MergeAttributes[schema.Attribute](
			nodePoolDataSourceAttributes,
			map[string]schema.Attribute{
				"cluster_name": schema.StringAttribute{Required: true, Description: "클러스터 이름"},
				"name":         schema.StringAttribute{Required: true, Description: "노드 풀 이름"},
				"timeouts":     timeouts.Attributes(ctx),
			},
		),
	}
}

func (d *nodePoolDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config nodePoolDataSourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	timeout, diags := config.Timeouts.Read(ctx, common.DefaultReadTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, d.kc, &resp.Diagnostics,
		func() (struct {
			NodePool kubernetesengine.KubernetesEngineV1ApiGetNodePoolModelNodePoolResponseModel
		}, *http.Response, error) {
			apiResp, httpResp, err := d.kc.ApiClient.NodePoolsAPI.
				GetNodePool(ctx, config.ClusterName.ValueString(), config.Name.ValueString()).
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

	result := respModel.NodePool
	ok := mapNodePoolFromResponse(ctx, &config.NodePoolBaseModel, &result, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
}

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package kubernetesengine

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-kakaocloud/internal/common"
	"terraform-provider-kakaocloud/internal/docs"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kakaoenterprise/kc-sdk-go/services/kubernetesengine"
)

var (
	_ datasource.DataSource              = &clusterNodeDataSource{}
	_ datasource.DataSourceWithConfigure = &clusterNodeDataSource{}
)

func NewClusterNodeDataSource() datasource.DataSource { return &clusterNodeDataSource{} }

type clusterNodeDataSource struct {
	kc *common.KakaoCloudClient
}

func (d *clusterNodeDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *clusterNodeDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kubernetes_engine_cluster_nodes"
}

func (d *clusterNodeDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: docs.GetDataSourceDescription("KubernetesEngineClusterNodes"),
		Attributes: map[string]schema.Attribute{
			"cluster_name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(4, 20),
				},
			},
			"node_pool_name": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(4, 20),
				},
			},
			"nodes": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: nodeDataSourceSchemaAttributes,
				},
			},
			"timeouts": timeouts.Attributes(ctx),
		},
	}
}

func (d *clusterNodeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config clusterNodeDataSourceModel

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

	clusterName := config.ClusterName.ValueString()

	var mapped []nodeBaseModel
	if !config.NodePoolName.IsNull() && !config.NodePoolName.IsUnknown() {
		nodePoolName := config.NodePoolName.ValueString()

		modelResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, d.kc, &resp.Diagnostics,
			func() (*kubernetesengine.GetK8sClusterNodePoolNodesResponseModel, *http.Response, error) {
				return d.kc.ApiClient.NodePoolsAPI.
					ListNodePoolNodes(ctx, clusterName, nodePoolName).
					XAuthToken(d.kc.XAuthToken).
					Execute()
			},
		)
		if err != nil {
			common.AddApiActionError(ctx, d, httpResp, "ListNodePoolNodes", err, &resp.Diagnostics)
			return
		}

		mapped = make([]nodeBaseModel, 0, len(modelResp.Nodes))
		for _, v := range modelResp.Nodes {
			var n nodeBaseModel
			if ok := mapNodePoolNodeBaseModel(&n, &v, &resp.Diagnostics); !ok || resp.Diagnostics.HasError() {
				return
			}
			mapped = append(mapped, n)
		}

		config.NodePoolName = types.StringValue(nodePoolName)
	} else {
		modelResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, d.kc, &resp.Diagnostics,
			func() (*kubernetesengine.GetK8sClusterNodesResponseModel, *http.Response, error) {
				return d.kc.ApiClient.ClustersAPI.
					ListClusterNodes(ctx, clusterName).
					XAuthToken(d.kc.XAuthToken).
					Execute()
			},
		)
		if err != nil {
			common.AddApiActionError(ctx, d, httpResp, "ListClusterNodes", err, &resp.Diagnostics)
			return
		}

		mapped = make([]nodeBaseModel, 0, len(modelResp.Nodes))
		for _, v := range modelResp.Nodes {
			var n nodeBaseModel
			if ok := mapClusterNodeBaseModel(&n, &v, &resp.Diagnostics); !ok || resp.Diagnostics.HasError() {
				return
			}
			mapped = append(mapped, n)
		}
	}

	config.Nodes = mapped

	respDiags := resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(respDiags...)
}

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
	"github.com/jinzhu/copier"
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
		Attributes: map[string]schema.Attribute{
			"cluster_name": schema.StringAttribute{
				Required:   true,
				Validators: common.NameValidator(20),
			},
			"node_pools": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: MergeAttributes[schema.Attribute](
						map[string]schema.Attribute{
							"cluster_name": schema.StringAttribute{
								Computed: true,
							},
							"name": schema.StringAttribute{
								Computed: true,
							},
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
		func() (*kubernetesengine.GetK8sClusterNodePoolsResponseModel, *http.Response, error) {
			return d.kc.ApiClient.NodePoolsAPI.ListNodePools(ctx, config.ClusterName.ValueString()).XAuthToken(d.kc.XAuthToken).Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, d, httpResp, "ListNodePools", err, &resp.Diagnostics)
		return
	}

	var nodePools []kubernetesengine.KubernetesEngineV1ApiGetNodePoolModelNodePoolResponseModel
	err = copier.Copy(&nodePools, &listResp.NodePools)
	if err != nil {
		common.AddGeneralError(ctx, d, &resp.Diagnostics,
			fmt.Sprintf("Failed to convert nodePools: %v", err))
		return
	}

	for _, nodePool := range nodePools {
		var model NodePoolBaseModel
		ok := mapNodePoolFromResponse(ctx, &model, &nodePool, &resp.Diagnostics)
		if !ok || resp.Diagnostics.HasError() {
			return
		}
		config.NodePools = append(config.NodePools, model)
	}

	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
}

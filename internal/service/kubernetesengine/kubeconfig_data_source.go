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
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &kubernetesKubeconfigDataSource{}
	_ datasource.DataSourceWithConfigure = &kubernetesKubeconfigDataSource{}
)

func NewKubernetesKubeconfigDataSource() datasource.DataSource {
	return &kubernetesKubeconfigDataSource{}
}

type kubernetesKubeconfigDataSource struct {
	kc *common.KakaoCloudClient
}

func (d *kubernetesKubeconfigDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kubernetes_engine_kubeconfig"
}

func (d *kubernetesKubeconfigDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: MergeDataSourceSchemaAttributes(
			map[string]schema.Attribute{
				"timeouts": timeouts.Attributes(ctx),
			},
			kubernetesKubeconfigDataSourceSchemaAttributes,
		),
	}
}

func (d *kubernetesKubeconfigDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config kubernetesKubeconfigDataSourceModel
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

	kubeYAML, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, d.kc, &resp.Diagnostics,
		func() (string, *http.Response, error) {
			return d.kc.ApiClient.
				ClustersAPI.
				GetClusterKubeconfig(ctx, clusterName).
				XAuthToken(d.kc.XAuthToken).
				Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, d, httpResp, "GetClusterKubeconfig", err, &resp.Diagnostics)
		return
	}

	var state kubernetesKubeconfigDataSourceModel
	state.ClusterName = types.StringValue(clusterName)
	state.KubeconfigYAML = types.StringValue(kubeYAML)

	mapKubeconfigYAMLToModel(ctx, kubeYAML, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	state.Timeouts = config.Timeouts

	if diags := resp.State.Set(ctx, &state); diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
}

func (d *kubernetesKubeconfigDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

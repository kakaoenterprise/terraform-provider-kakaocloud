// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package kubernetesengine

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-kakaocloud/internal/common"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/kakaoenterprise/kc-sdk-go/services/kubernetesengine"
)

var (
	_ datasource.DataSource              = &upgradableVersionDataSource{}
	_ datasource.DataSourceWithConfigure = &upgradableVersionDataSource{}
)

func NewClusterUpgradableVersionDataSource() datasource.DataSource {
	return &upgradableVersionDataSource{}
}

type upgradableVersionDataSource struct {
	kc *common.KakaoCloudClient
}

func (d *upgradableVersionDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {

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

func (d *upgradableVersionDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kubernetes_engine_upgradable_versions"

}

func (d *upgradableVersionDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	versionAttrs := upgradableVersionsDataSourceSchemaAttributes

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"cluster_name": schema.StringAttribute{
				Required:   true,
				Validators: common.NameValidator(20),
			},
			"current": schema.SingleNestedAttribute{
				Computed:   true,
				Attributes: versionAttrs,
			},
			"upgradable": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: versionAttrs,
				},
			},
			"timeouts": timeouts.Attributes(ctx),
		},
	}
}

func (d *upgradableVersionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config upgradableVersionsDataSourceModel

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
		func() (*kubernetesengine.GetK8sClusterUpgradableVersionsResponseModel, *http.Response, error) {
			return d.kc.ApiClient.ClustersAPI.
				ListClusterUpgradableVersions(ctx, config.ClusterName.ValueString()).
				XAuthToken(d.kc.XAuthToken).
				Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, d, httpResp, "ListClusterUpgradableVersions", err, &resp.Diagnostics)
		return
	}

	result := respModel.Upgrade
	ok := d.mapUpgradableVersions(ctx, &config, &result, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
}

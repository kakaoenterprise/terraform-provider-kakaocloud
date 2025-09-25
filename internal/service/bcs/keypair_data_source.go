// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bcs

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-kakaocloud/internal/common"
	. "terraform-provider-kakaocloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/kakaoenterprise/kc-sdk-go/services/bcs"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &keypairDataSource{}
	_ datasource.DataSourceWithConfigure = &keypairDataSource{}
)

func NewKeypairDataSource() datasource.DataSource {
	return &keypairDataSource{}
}

// keypairDataSource is the data source implementation.
type keypairDataSource struct {
	kc *common.KakaoCloudClient
}

func (d *keypairDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_keypair"
}

func (d *keypairDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "kakaocloud의 특정 키페어를 이름으로 조회하는 데이터 소스",
		Attributes: MergeDataSourceSchemaAttributes(
			map[string]schema.Attribute{
				"name": schema.StringAttribute{
					Required:    true,
					Description: "키페어 이름",
				},
				"timeouts": timeouts.Attributes(ctx),
			},
			keypairDataSourceSchemaAttributes,
		),
	}
}

func (d *keypairDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config keypairDataSourceModel

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

	keypairResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, d.kc, &resp.Diagnostics,
		func() (*bcs.BcsInstanceV1ApiGetKeypairModelResponseKeypairModel, *http.Response, error) {
			return d.kc.ApiClient.KeypairAPI.GetKeypair(ctx, config.Name.ValueString()).
				XAuthToken(d.kc.XAuthToken).
				Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, d, httpResp, "GetKeypair", err, &resp.Diagnostics)
		return
	}

	result := keypairResp.Keypair

	ok := mapKeypairBaseModel(ctx, &config.keypairBaseModel, &result, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (d *keypairDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*common.KakaoCloudClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *common.KakaoCloudClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	d.kc = client
}

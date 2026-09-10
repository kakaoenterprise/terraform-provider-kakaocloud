// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package vpc

import (
	"context"
	"fmt"
	"net/http"

	"terraform-provider-kakaocloud/internal/common"
	. "terraform-provider-kakaocloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/kakaoenterprise/kc-sdk-go/v2/services/vpc"
)

var (
	_ datasource.DataSource              = &internetGatewayDataSource{}
	_ datasource.DataSourceWithConfigure = &internetGatewayDataSource{}
)

func NewInternetGatewayDataSource() datasource.DataSource {
	return &internetGatewayDataSource{}
}

type internetGatewayDataSource struct {
	kc *common.KakaoCloudClient
}

func (d *internetGatewayDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_internet_gateway"
}

func (d *internetGatewayDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: MergeDataSourceSchemaAttributes(
			map[string]schema.Attribute{
				"id": schema.StringAttribute{
					Required:   true,
					Validators: common.UuidValidator(),
				},
				"timeouts": timeouts.Attributes(ctx),
			},
			internetGatewayDataSourceSchemaAttributes,
		),
	}
}

func (d *internetGatewayDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config internetGatewayDataSourceModel

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

	igwResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, d.kc, &resp.Diagnostics,
		func() (*vpc.GetInternetGatewayResponse, *http.Response, error) {
			return d.kc.ApiClient.VPCInternetGatewayAPI.
				GetInternetGateway(ctx, config.Id.ValueString()).
				XAuthToken(d.kc.XAuthToken).
				Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, d, httpResp, "GetInternetGateway", err, &resp.Diagnostics)
		return
	}

	if !mapInternetGatewayModel(ctx, &config.internetGatewayBaseModel, &igwResp.Igw, &resp.Diagnostics) || resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
}

func (d *internetGatewayDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

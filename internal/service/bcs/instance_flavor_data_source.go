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

var (
	_ datasource.DataSource              = &instanceFlavorDataSource{}
	_ datasource.DataSourceWithConfigure = &instanceFlavorDataSource{}
)

func NewInstanceFlavorDataSource() datasource.DataSource {
	return &instanceFlavorDataSource{}
}

type instanceFlavorDataSource struct {
	kc *common.KakaoCloudClient
}

func (d *instanceFlavorDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_instance_flavor"
}

func (d *instanceFlavorDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: MergeDataSourceSchemaAttributes(
			map[string]schema.Attribute{
				"id": schema.StringAttribute{
					Required:   true,
					Validators: common.UuidValidator(),
				},
				"timeouts": timeouts.Attributes(ctx),
			},
			instanceFlavorDataSourceSchemaAttributes,
		),
	}
}

func (d *instanceFlavorDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config instanceFlavorDataSourceModel

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

	flavorResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, d.kc, &resp.Diagnostics,
		func() (*bcs.ResponseFlavorModel, *http.Response, error) {
			return d.kc.ApiClient.FlavorAPI.GetInstanceType(ctx, config.Id.ValueString()).XAuthToken(d.kc.XAuthToken).Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, d, httpResp, "GetInstanceType", err, &resp.Diagnostics)
		return
	}

	result := flavorResp.Flavor

	ok := mapInstanceFlavorBaseModel(ctx, &config.instanceFlavorBaseModel, &result, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (d *instanceFlavorDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {

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

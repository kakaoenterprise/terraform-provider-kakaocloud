// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package loadbalancer

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-kakaocloud/internal/common"
	"terraform-provider-kakaocloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/kakaoenterprise/kc-sdk-go/services/loadbalancer"
)

var (
	_ datasource.DataSource              = &beyondLoadBalancerDataSource{}
	_ datasource.DataSourceWithConfigure = &beyondLoadBalancerDataSource{}
)

func NewBeyondLoadBalancerDataSource() datasource.DataSource {
	return &beyondLoadBalancerDataSource{}
}

type beyondLoadBalancerDataSource struct {
	kc *common.KakaoCloudClient
}

func (d *beyondLoadBalancerDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_beyond_load_balancer"
}

func (d *beyondLoadBalancerDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: utils.MergeDataSourceSchemaAttributes(
			map[string]schema.Attribute{
				"id": schema.StringAttribute{
					Required:   true,
					Validators: common.UuidValidator(),
				},
				"timeouts": timeouts.Attributes(ctx),
			},
			beyondLoadBalancerDatasourceSchema,
		),
	}
}

func (d *beyondLoadBalancerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data beyondLoadBalancerDataSourceModel

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	timeout, diags := data.Timeouts.Read(ctx, common.DefaultReadTimeout)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	blbResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, d.kc, &resp.Diagnostics,
		func() (*loadbalancer.BnsLoadBalancerV1ApiGetHaGroupModelResponseBeyondLoadBalancerModel, *http.Response, error) {
			return d.kc.ApiClient.BeyondLoadBalancerAPI.GetHaGroup(ctx, data.Id.ValueString()).XAuthToken(d.kc.XAuthToken).Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, d, httpResp, "GetHaGroup", err, &resp.Diagnostics)
		return
	}

	ok := mapBeyondLoadBalancerBaseModel(ctx, &data.beyondLoadBalancerBaseModel, &blbResp.BeyondLoadBalancer, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (d *beyondLoadBalancerDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

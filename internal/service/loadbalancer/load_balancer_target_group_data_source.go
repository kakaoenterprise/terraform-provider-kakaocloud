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
	_ datasource.DataSource              = &loadBalancerTargetGroupDataSource{}
	_ datasource.DataSourceWithConfigure = &loadBalancerTargetGroupDataSource{}
)

func NewLoadBalancerTargetGroupDataSource() datasource.DataSource {
	return &loadBalancerTargetGroupDataSource{}
}

type loadBalancerTargetGroupDataSource struct {
	kc *common.KakaoCloudClient
}

func (d *loadBalancerTargetGroupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*common.KakaoCloudClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected DataSource Configure Type",
			fmt.Sprintf("Expected *common.KakaoCloudClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	d.kc = client
}

func (d *loadBalancerTargetGroupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_load_balancer_target_group"
}

func (d *loadBalancerTargetGroupDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: utils.MergeDataSourceSchemaAttributes(
			map[string]schema.Attribute{
				"id": schema.StringAttribute{
					Required:   true,
					Validators: common.UuidValidator(),
				},
				"timeouts": timeouts.Attributes(ctx),
			},
			loadBalancerTargetGroupDataSourceSchemaAttributes,
		),
	}
}

func (d *loadBalancerTargetGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data loadBalancerTargetGroupDataSourceModel
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

	respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, d.kc, &resp.Diagnostics,
		func() (*loadbalancer.TargetGroupResponseModel, *http.Response, error) {
			return d.kc.ApiClient.LoadBalancerTargetGroupAPI.GetTargetGroup(ctx, data.Id.ValueString()).XAuthToken(d.kc.XAuthToken).Execute()
		},
	)

	if err != nil {
		common.AddApiActionError(ctx, d, httpResp, "GetTargetGroup", err, &resp.Diagnostics)
		return
	}

	ok := mapLoadBalancerTargetGroupFromGetResponse(ctx, &data.loadBalancerTargetGroupBaseModel, &respModel.TargetGroup, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

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
	_ datasource.DataSource              = &loadBalancerL7PolicyRuleDataSource{}
	_ datasource.DataSourceWithConfigure = &loadBalancerL7PolicyRuleDataSource{}
)

func NewLoadBalancerL7PolicyRuleDataSource() datasource.DataSource {
	return &loadBalancerL7PolicyRuleDataSource{}
}

type loadBalancerL7PolicyRuleDataSource struct {
	kc *common.KakaoCloudClient
}

func (d *loadBalancerL7PolicyRuleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_load_balancer_l7_policy_rule"
}

func (d *loadBalancerL7PolicyRuleDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: utils.MergeDataSourceSchemaAttributes(
			loadBalancerL7PolicyRuleDataSourceSchema,
			map[string]schema.Attribute{
				"timeouts": timeouts.Attributes(ctx),
			},
		),
	}
}

func (d *loadBalancerL7PolicyRuleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *loadBalancerL7PolicyRuleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config loadBalancerL7PolicyRuleDataSourceModel

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

	ruleResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, d.kc, &resp.Diagnostics,
		func() (*loadbalancer.Responsel7PolicyRuleModel, *http.Response, error) {
			return d.kc.ApiClient.LoadBalancerL7PoliciesAPI.GetL7PolicyRule(ctx, config.L7PolicyId.ValueString(), config.Id.ValueString()).XAuthToken(d.kc.XAuthToken).Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, d, httpResp, "GetL7PolicyRule", err, &resp.Diagnostics)
		return
	}

	ok := mapLoadBalancerL7PolicyRuleBaseModel(&config.loadBalancerL7PolicyRuleBaseModel, &ruleResp.L7Rule, config.L7PolicyId.ValueString(), &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
}

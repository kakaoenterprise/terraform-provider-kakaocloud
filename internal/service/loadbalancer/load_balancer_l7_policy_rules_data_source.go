// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package loadbalancer

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-kakaocloud/internal/docs"

	"terraform-provider-kakaocloud/internal/common"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/kakaoenterprise/kc-sdk-go/services/loadbalancer"
)

var (
	_ datasource.DataSource              = &loadBalancerL7PolicyRulesDataSource{}
	_ datasource.DataSourceWithConfigure = &loadBalancerL7PolicyRulesDataSource{}
)

func NewLoadBalancerL7PolicyRulesDataSource() datasource.DataSource {
	return &loadBalancerL7PolicyRulesDataSource{}
}

type loadBalancerL7PolicyRulesDataSource struct {
	kc *common.KakaoCloudClient
}

func (d *loadBalancerL7PolicyRulesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_load_balancer_l7_policy_rules"
}

func (d *loadBalancerL7PolicyRulesDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: docs.GetDataSourceDescription("LoadBalancerL7PolicyRules"),
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the L7 policy to list rules for",
			},
			"l7_rules": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of L7 policy rules",
				NestedObject: schema.NestedAttributeObject{
					Attributes: loadBalancerL7PolicyRuleListDataSourceSchema,
				},
			},
			"timeouts": timeouts.Attributes(ctx),
		},
	}
}

func (d *loadBalancerL7PolicyRulesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *loadBalancerL7PolicyRulesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config loadBalancerL7PolicyRuleListDataSourceModel

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

	l7policyApi := d.kc.ApiClient.LoadBalancerL7PoliciesAPI.GetL7Policy(ctx, config.Id.ValueString()).XAuthToken(d.kc.XAuthToken)

	l7policyResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, d.kc, &resp.Diagnostics,
		func() (*loadbalancer.BnsLoadBalancerV1ApiGetL7PolicyModelResponseL7PolicyModel, *http.Response, error) {
			return l7policyApi.Execute()
		},
	)

	if err != nil {
		common.AddApiActionError(ctx, d, httpResp, "GetL7Policy", err, &resp.Diagnostics)
		return
	}

	config = mapLoadBalancerL7PolicyRuleListFromGetPolicyResponse(*l7policyResp, config.Id.ValueString(), config.Timeouts)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

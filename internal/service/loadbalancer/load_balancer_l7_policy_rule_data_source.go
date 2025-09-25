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

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &loadBalancerL7PolicyRuleDataSource{}
	_ datasource.DataSourceWithConfigure = &loadBalancerL7PolicyRuleDataSource{}
)

// NewLoadBalancerL7PolicyRuleDataSource is a helper function to simplify the provider implementation.
func NewLoadBalancerL7PolicyRuleDataSource() datasource.DataSource {
	return &loadBalancerL7PolicyRuleDataSource{}
}

type loadBalancerL7PolicyRuleDataSource struct {
	kc *common.KakaoCloudClient
}

// Metadata returns the data source type name.
func (d *loadBalancerL7PolicyRuleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_load_balancer_l7_policy_rule"
}

// Schema defines the schema for the data source.
func (d *loadBalancerL7PolicyRuleDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to get information about a specific KakaoCloud Load Balancer L7 Policy Rule.",
		Attributes: utils.MergeDataSourceSchemaAttributes(
			loadBalancerL7PolicyRuleDataSourceSchema,
			map[string]schema.Attribute{
				"timeouts": timeouts.Attributes(ctx),
			},
		),
	}
}

// Configure adds the provider configured client to the data source.
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

// Read refreshes the Terraform state with the latest data.
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

	// Get the specific L7 policy rule
	ruleResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, d.kc, &resp.Diagnostics,
		func() (*loadbalancer.Responsel7PolicyRuleModel, *http.Response, error) {
			return d.kc.ApiClient.LoadBalancerL7PoliciesAPI.GetL7PolicyRule(ctx, config.L7PolicyId.ValueString(), config.Id.ValueString()).XAuthToken(d.kc.XAuthToken).Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, d, httpResp, "GetL7PolicyRule", err, &resp.Diagnostics)
		return
	}

	// Map the API response to the Terraform model
	ok := mapLoadBalancerL7PolicyRuleDataSourceFromGetRuleResponse(ctx, &config, &ruleResp.L7Rule, config.L7PolicyId.ValueString(), &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
}

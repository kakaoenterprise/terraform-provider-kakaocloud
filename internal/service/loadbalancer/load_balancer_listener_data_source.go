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
	_ datasource.DataSource              = &loadBalancerListenerDataSource{}
	_ datasource.DataSourceWithConfigure = &loadBalancerListenerDataSource{}
)

// NewLoadBalancerListenersDataSource is a helper function to simplify the provider implementation.
func NewLoadBalancerListenerDataSource() datasource.DataSource {
	return &loadBalancerListenerDataSource{}
}

type loadBalancerListenerDataSource struct {
	kc *common.KakaoCloudClient
}

// Metadata returns the data source type name.
func (d *loadBalancerListenerDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_load_balancer_listener"
}

// Configure adds the provider configured client to the data source.
func (d *loadBalancerListenerDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Schema defines the schema for the data source.
func (d *loadBalancerListenerDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to get information about KakaoCloud Load Balancer Listener lists.",
		Attributes: utils.MergeDataSourceSchemaAttributes(
			map[string]schema.Attribute{
				"id": schema.StringAttribute{
					Required:   true,
					Validators: common.UuidValidator(),
				},
				"timeouts": timeouts.Attributes(ctx),
			},
			listenerDataSourceSchemaAttributes,
		),
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *loadBalancerListenerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config loadBalancerListenerDataSourceModel

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

	lbl, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, d.kc, &resp.Diagnostics,
		func() (*loadbalancer.BnsLoadBalancerV1ApiGetListenerModelResponseListenerModel, *http.Response, error) {
			return d.kc.ApiClient.LoadBalancerListenerAPI.GetListener(ctx, config.Id.ValueString()).XAuthToken(d.kc.XAuthToken).Execute()
		},
	)

	if err != nil {
		common.AddApiActionError(ctx, d, httpResp, "GetListener", err, &resp.Diagnostics)
		return
	}

	ok := mapLoadBalancerListenerBaseModel(ctx, &config.loadBalancerListenerBaseModel, &lbl.Listener, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
}

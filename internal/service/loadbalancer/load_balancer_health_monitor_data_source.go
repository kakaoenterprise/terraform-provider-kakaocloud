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

// Ensure the implementation satisfies the expected interfaces
var (
	_ datasource.DataSource              = &loadBalancerHealthMonitorDataSource{}
	_ datasource.DataSourceWithConfigure = &loadBalancerHealthMonitorDataSource{}
)

func NewLoadBalancerHealthMonitorDataSource() datasource.DataSource {
	return &loadBalancerHealthMonitorDataSource{}
}

type loadBalancerHealthMonitorDataSource struct {
	kc *common.KakaoCloudClient
}

func (d *loadBalancerHealthMonitorDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *loadBalancerHealthMonitorDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_load_balancer_health_monitor"
}

func (d *loadBalancerHealthMonitorDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Get information about a KakaoCloud Load Balancer Health Monitor.",
		Attributes: utils.MergeDataSourceSchemaAttributes(
			map[string]schema.Attribute{
				"id": schema.StringAttribute{
					Required:   true,
					Validators: common.UuidValidator(),
				},
				"timeouts": timeouts.Attributes(ctx),
			},
			loadBalancerHealthMonitorDataSourceSchema,
		),
	}
}

func (d *loadBalancerHealthMonitorDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data loadBalancerHealthMonitorDataSourceModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the health monitor
	healthMonitor, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, d.kc, &resp.Diagnostics,
		func() (*loadbalancer.BnsLoadBalancerV1ApiGetTargetGroupHealthMonitorModelResponseHealthMonitorModel, *http.Response, error) {
			return d.kc.ApiClient.LoadBalancerTargetGroupAPI.
				GetTargetGroupHealthMonitor(ctx, data.Id.ValueString()).
				XAuthToken(d.kc.XAuthToken).
				Execute()
		},
	)

	// 404 → Not found
	if httpResp != nil && httpResp.StatusCode == 404 {
		resp.Diagnostics.AddError(
			"Health Monitor Not Found",
			fmt.Sprintf("Health monitor with ID %s was not found.", data.Id.ValueString()),
		)
		return
	}

	if err != nil {
		common.AddApiActionError(ctx, d, httpResp, "GetHealthMonitor", err, &resp.Diagnostics)
		return
	}

	// Map response back to Terraform model
	ok := mapHealthMonitorFromGetResponse(ctx, &data.loadBalancerHealthMonitorBaseModel, &healthMonitor.HealthMonitor, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	// Set the state
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

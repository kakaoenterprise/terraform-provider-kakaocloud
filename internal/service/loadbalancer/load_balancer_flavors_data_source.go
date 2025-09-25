// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package loadbalancer

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-kakaocloud/internal/common"

	"github.com/jinzhu/copier"
	"github.com/kakaoenterprise/kc-sdk-go/services/loadbalancer"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

// Ensure the implementation satisfies the expected interfaces
var (
	_ datasource.DataSource              = &loadBalancerFlavorsDataSource{}
	_ datasource.DataSourceWithConfigure = &loadBalancerFlavorsDataSource{}
)

// NewLoadBalancerFlavorsDataSource returns a new load balancer flavors data source
func NewLoadBalancerFlavorsDataSource() datasource.DataSource {
	return &loadBalancerFlavorsDataSource{}
}

// loadBalancerFlavorsDataSource implements the data source for load balancer flavors
type loadBalancerFlavorsDataSource struct {
	kc *common.KakaoCloudClient
}

func (d *loadBalancerFlavorsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Metadata returns the data source type name
func (d *loadBalancerFlavorsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_load_balancer_flavors"
}

// Schema defines the schema for the data source
func (d *loadBalancerFlavorsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a list of available load balancer flavors.",
		Attributes: map[string]schema.Attribute{
			"flavors": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: loadBalancerFlavorBaseSchemaAttributes,
				},
			},
			"timeouts": timeouts.Attributes(ctx),
		},
	}
}

// Read gets all load balancer flavors
func (d *loadBalancerFlavorsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data loadBalancerFlavorsDataSourceModel

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

	// The result, response, and error must be captured in variables.
	lbfs, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, d.kc, &resp.Diagnostics,
		func() (interface{}, *http.Response, error) {
			resp, httpResp, err := d.kc.ApiClient.LoadBalancerEtcAPI.ListLoadBalancerTypes(ctx).XAuthToken(d.kc.XAuthToken).Execute()
			return resp, httpResp, err
		},
	)

	if err != nil {
		common.AddApiActionError(ctx, d, httpResp, "ListLoadBalancerTypes", err, &resp.Diagnostics)
		return
	}

	// Use type assertion to get the Types field
	// The API response should have a Types field containing the list
	lbfsTyped, ok := lbfs.(*loadbalancer.FlavorListModel)
	if !ok {
		resp.Diagnostics.AddError("Type assertion failed", "Failed to cast API response to expected type")
		return
	}

	var lbfsResult []loadbalancer.FlavorModel
	err = copier.Copy(&lbfsResult, &lbfsTyped.Flavors)
	if err != nil {
		resp.Diagnostics.AddError("List 변환 실패", fmt.Sprintf("lbfsResult 변환 실패: %v", err))
		return
	}
	// Convert API response directly to our model
	for _, flavor := range lbfsResult {
		var lbfModel loadBalancerFlavorBaseModel
		ok := mapLoadBalancerFlavor(ctx, &lbfModel, &flavor, &resp.Diagnostics)
		if !ok {
			return
		}
		data.LoadBalancerFlavors = append(data.LoadBalancerFlavors, lbfModel)
	}

	// Set state
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

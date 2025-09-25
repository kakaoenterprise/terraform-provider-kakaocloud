// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package loadbalancer

import (
	"context"
	"fmt"
	"net/http"

	"github.com/jinzhu/copier"

	"terraform-provider-kakaocloud/internal/common"
	"terraform-provider-kakaocloud/internal/utils"

	"github.com/kakaoenterprise/kc-sdk-go/services/loadbalancer"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &loadBalancerL7PoliciesDataSource{}
	_ datasource.DataSourceWithConfigure = &loadBalancerL7PoliciesDataSource{}
)

// NewLoadBalancerL7PoliciesDataSource is a helper function to simplify the provider implementation.
func NewLoadBalancerL7PoliciesDataSource() datasource.DataSource {
	return &loadBalancerL7PoliciesDataSource{}
}

type loadBalancerL7PoliciesDataSource struct {
	kc *common.KakaoCloudClient
}

// Metadata returns the data source type name.
func (d *loadBalancerL7PoliciesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_load_balancer_l7_policies"
}

// Schema defines the schema for the data source.
func (d *loadBalancerL7PoliciesDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to get information about KakaoCloud Load Balancer L7 Policy lists.",
		Attributes: map[string]schema.Attribute{
			"load_balancer_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the load balancer",
			},
			"listener_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the listener",
			},
			"filter": schema.ListNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required: true,
						},
						"value": schema.StringAttribute{
							Optional: true,
						},
					},
				},
			},
			"l7_policies": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: utils.MergeDataSourceSchemaAttributes(
						map[string]schema.Attribute{
							"id": schema.StringAttribute{
								Computed: true,
							},
						},
						loadBalancerL7PolicyDataSourceSchemaAttributes,
					),
				},
			},
			"timeouts": timeouts.Attributes(ctx),
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *loadBalancerL7PoliciesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config loadBalancerL7PoliciesDataSourceModel

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

	// Initialize the API call with required path parameters
	l7PolicyApi := d.kc.ApiClient.LoadBalancerL7PoliciesAPI.ListL7Policies(
		ctx,
		config.LoadBalancerId.ValueString(),
		config.ListenerId.ValueString(),
	)

	// Apply filters
	for _, f := range config.Filter {
		if f.Name.IsNull() || f.Name.IsUnknown() {
			continue
		}

		filterName := f.Name.ValueString()

		if !f.Value.IsNull() && !f.Value.IsUnknown() {
			v := f.Value.ValueString()
			switch filterName {
			case "position":
				if pos, err := ParseInt32(v); err == nil {
					l7PolicyApi = l7PolicyApi.Position(*pos)
				} else {
					resp.Diagnostics.AddError(
						"Invalid position value",
						err.Error(),
					)
				}
			case "action":
				if action, err := ToL7PolicyAction(v); err == nil {
					l7PolicyApi = l7PolicyApi.Action(*action)
				} else {
					resp.Diagnostics.AddError(
						"Invalid action value",
						err.Error(),
					)
				}
			case "provisioning_status":
				if ps, err := ToProvisioningStatus(v); err == nil {
					l7PolicyApi = l7PolicyApi.ProvisioningStatus(*ps)
				} else {
					resp.Diagnostics.AddError(
						"Invalid provisioning_status value",
						err.Error(),
					)
				}
			case "operating_status":
				if os, err := ToLoadBalancerOperatingStatus(v); err == nil {
					l7PolicyApi = l7PolicyApi.OperatingStatus(*os)
				} else {
					resp.Diagnostics.AddError(
						"Invalid operating_status value",
						err.Error(),
					)
				}
			case "name":
				l7PolicyApi = l7PolicyApi.Name(v)
			default:
				resp.Diagnostics.AddError(
					"Invalid filter name",
					fmt.Sprintf("filter %q is not supported", filterName),
				)
			}
		}
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Execute the API call
	lbL7PoliciesResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, d.kc, &resp.Diagnostics,
		func() (*loadbalancer.L7PolicyListModel, *http.Response, error) {
			return l7PolicyApi.Limit(1000).XAuthToken(d.kc.XAuthToken).Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, d, httpResp, "ListL7Policies", err, &resp.Diagnostics)
		return
	}

	var lbL7PoliciesResult []loadbalancer.BnsLoadBalancerV1ApiGetL7PolicyModelL7PolicyModel
	err = copier.Copy(&lbL7PoliciesResult, &lbL7PoliciesResp.L7Policies)
	if err != nil {
		common.AddGeneralError(ctx, d, &resp.Diagnostics,
			fmt.Sprintf("lblsResult transform fail: %v", err))
		return
	}

	for _, v := range lbL7PoliciesResult {
		var tmplb7Policy loadBalancerL7PolicyBaseModel
		ok := mapLoadBalancerL7PolicyFromGetResponse(ctx, &tmplb7Policy, &v, &resp.Diagnostics)
		if !ok || resp.Diagnostics.HasError() {
			return
		}

		config.L7Policies = append(config.L7Policies, tmplb7Policy)
	}

	// Save data into Terraform state
	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
}

// Configure adds the provider configured client to the data source.
func (d *loadBalancerL7PoliciesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

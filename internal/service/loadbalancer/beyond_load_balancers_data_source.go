// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package loadbalancer

import (
	"context"
	"fmt"
	"net/http"

	"github.com/jinzhu/copier"
	"github.com/kakaoenterprise/kc-sdk-go/services/loadbalancer"

	"terraform-provider-kakaocloud/internal/common"

	"terraform-provider-kakaocloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &beyondLoadBalancersDataSource{}
	_ datasource.DataSourceWithConfigure = &beyondLoadBalancersDataSource{}
)

// NewBeyondLoadBalancersDataSource is a helper function to simplify the provider implementation.
func NewBeyondLoadBalancersDataSource() datasource.DataSource {
	return &beyondLoadBalancersDataSource{}
}

type beyondLoadBalancersDataSource struct {
	kc *common.KakaoCloudClient
}

// Metadata returns the data source type name.
func (d *beyondLoadBalancersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_beyond_load_balancers"
}

// Schema defines the schema for the data source.
func (d *beyondLoadBalancersDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to get information about KakaoCloud Beyond Load Balancer lists.",
		Attributes: map[string]schema.Attribute{
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
			"beyond_load_balancers": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: utils.MergeDataSourceSchemaAttributes(
						map[string]schema.Attribute{
							"id": schema.StringAttribute{
								Computed: true,
							},
						},
						beyondLoadBalancerDatasourceSchema,
					),
				},
			},
			"timeouts": timeouts.Attributes(ctx),
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *beyondLoadBalancersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config beyondLoadBalancersDataSourceModel

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

	// The SDK function and response object will need to be verified against the actual Go SDK.
	blbApi := d.kc.ApiClient.BeyondLoadBalancerAPI.ListHaGroups(ctx)

	for _, f := range config.Filter {
		if f.Name.IsNull() || f.Name.IsUnknown() {
			continue
		}

		filterName := f.Name.ValueString()

		if !f.Value.IsNull() && !f.Value.IsUnknown() {
			v := f.Value.ValueString()

			switch filterName {
			case "id":
				blbApi = blbApi.Id(v)
			case "name":
				blbApi = blbApi.Name(v)
			case "dns_name":
				blbApi = blbApi.DnsName(v)
			case "scheme":
				if s, err := ToScheme(v); err == nil {
					blbApi = blbApi.Scheme(*s)
				} else {
					resp.Diagnostics.AddError(
						"Invalid scheme value",
						err.Error(),
					)
				}
			case "provisioning_status":
				if ps, err := ToProvisioningStatus(v); err == nil {
					blbApi = blbApi.ProvisioningStatus(*ps)
				} else {
					resp.Diagnostics.AddError(
						"Invalid provisioning_status value",
						err.Error(),
					)
				}
			case "operating_status":
				if os, err := ToLoadBalancerOperatingStatus(v); err == nil {
					blbApi = blbApi.OperatingStatus(*os)
				} else {
					resp.Diagnostics.AddError(
						"Invalid operating_status value",
						err.Error(),
					)
				}
			case "type":
				if t, err := ToLoadBalancerType(v); err == nil {
					blbApi = blbApi.Type_(*t)
				} else {
					resp.Diagnostics.AddError(
						"Invalid type value",
						err.Error(),
					)
				}
			case "vpc_name":
				blbApi = blbApi.VpcName(v)
			case "vpc_id":
				blbApi = blbApi.VpcId(v)
			case "created_at":
				if err := common.ValidateRFC3339(v); err == nil {
					blbApi = blbApi.CreatedAt(v)
				} else {
					resp.Diagnostics.AddError(
						"Invalid created_at value",
						err.Error(),
					)
				}
			case "updated_at":
				if err := common.ValidateRFC3339(v); err == nil {
					blbApi = blbApi.UpdatedAt(v)
				} else {
					resp.Diagnostics.AddError(
						"Invalid updated_at value",
						err.Error(),
					)
				}
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
	blbResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, d.kc, &resp.Diagnostics,
		func() (*loadbalancer.BeyondLoadBalancerListModel, *http.Response, error) {
			return blbApi.Limit(1000).XAuthToken(d.kc.XAuthToken).Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, d, httpResp, "ListHaGroups", err, &resp.Diagnostics)
		return
	}

	var blbsResult []loadbalancer.BnsLoadBalancerV1ApiGetHaGroupModelBeyondLoadBalancerModel
	err = copier.Copy(&blbsResult, &blbResp.BeyondLoadBalancers)
	if err != nil {
		common.AddGeneralError(ctx, d, &resp.Diagnostics,
			fmt.Sprintf("lblsResult transform fail: %v", err))
		return
	}

	for _, v := range blbsResult {
		var tmpBlb beyondLoadBalancerBaseModel
		ok := mapBeyondLoadBalancerBaseModel(ctx, &tmpBlb, &v, &resp.Diagnostics)
		if !ok || resp.Diagnostics.HasError() {
			return
		}

		config.BeyondLoadBalancers = append(config.BeyondLoadBalancers, tmpBlb)
	}

	// Save data into Terraform state
	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
}

// Configure adds the provider configured client to the data source.
func (d *beyondLoadBalancersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

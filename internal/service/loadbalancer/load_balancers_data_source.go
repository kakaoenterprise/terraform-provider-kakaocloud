// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package loadbalancer

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-kakaocloud/internal/common"
	"terraform-provider-kakaocloud/internal/docs"

	"terraform-provider-kakaocloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/jinzhu/copier"
	"github.com/kakaoenterprise/kc-sdk-go/services/loadbalancer"
)

var (
	_ datasource.DataSource              = &loadBalancersDataSource{}
	_ datasource.DataSourceWithConfigure = &loadBalancersDataSource{}
)

func NewLoadBalancersDataSource() datasource.DataSource {
	return &loadBalancersDataSource{}
}

type loadBalancersDataSource struct {
	kc *common.KakaoCloudClient
}

func (d *loadBalancersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *loadBalancersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_load_balancers"
}

func (d *loadBalancersDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: docs.GetDataSourceDescription("LoadBalancers"),
		Attributes: map[string]schema.Attribute{
			"filter": schema.ListNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required: true,
						},
						"value": schema.StringAttribute{
							Required: true,
						},
					},
				},
			},
			"load_balancers": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: utils.MergeDataSourceSchemaAttributes(
						map[string]schema.Attribute{
							"id": schema.StringAttribute{
								Computed: true,
							},
						},
						loadBalancerDataSourceSchemaAttributes,
					),
				},
			},
			"timeouts": timeouts.Attributes(ctx),
		},
	}
}

func (d *loadBalancersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data loadBalancersDataSourceModel

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

	lbApi := d.kc.ApiClient.LoadBalancerAPI.ListLoadBalancers(ctx)

	for _, f := range data.Filter {
		if f.Name.IsNull() || f.Name.IsUnknown() || f.Value.IsNull() || f.Value.IsUnknown() {
			continue
		}
		filterName := f.Name.ValueString()
		filterValue := f.Value.ValueString()

		if !f.Value.IsNull() && !f.Value.IsUnknown() {
			switch filterName {
			case "id":
				lbApi = lbApi.Id(filterValue)
			case "name":
				lbApi = lbApi.Name(filterValue)
			case "type":
				lbApi = lbApi.Type_(filterValue)
			case "private_vip":
				lbApi = lbApi.PrivateVip(filterValue)
			case "public_vip":
				lbApi = lbApi.PublicVip(filterValue)
			case "provisioning_status":
				if ps, err := ToProvisioningStatus(filterValue); err == nil {
					lbApi = lbApi.ProvisioningStatus(*ps)
				} else {
					resp.Diagnostics.AddError(
						"Invalid provisioning_status value",
						err.Error(),
					)
				}
			case "operating_status":
				if os, err := ToLoadBalancerOperatingStatus(filterValue); err == nil {
					lbApi = lbApi.OperatingStatus(*os)
				} else {
					resp.Diagnostics.AddError(
						"Invalid operating_status value",
						err.Error(),
					)
				}
			case "subnet_id":
				lbApi = lbApi.SubnetId(filterValue)
			case "subnet_cidr_block":
				lbApi = lbApi.SubnetCidrBlock(filterValue)
			case "vpc_id":
				lbApi = lbApi.VpcId(filterValue)
			case "vpc_name":
				lbApi = lbApi.VpcName(filterValue)
			case "availability_zone":
				if az, err := ToAvailabilityZone(filterValue); err == nil {
					lbApi = lbApi.AvailabilityZone(*az)
				} else {
					resp.Diagnostics.AddError(
						"Invalid availability_zone value",
						err.Error(),
					)
				}
			case "beyond_load_balancer_name":
				lbApi = lbApi.BeyondLoadBalancerName(filterValue)
			case "created_at":
				if err := common.ValidateRFC3339(filterValue); err == nil {
					lbApi = lbApi.CreatedAt(filterValue)
				} else {
					resp.Diagnostics.AddError(
						"Invalid created_at value",
						err.Error(),
					)
				}
			case "updated_at":
				if err := common.ValidateRFC3339(filterValue); err == nil {
					lbApi = lbApi.UpdatedAt(filterValue)
				} else {
					resp.Diagnostics.AddError(
						"Invalid updated_at value",
						err.Error(),
					)
				}
			default:
				resp.Diagnostics.AddWarning("Unsupported Filter", fmt.Sprintf("The filter '%s' is not supported for this data source.", filterName))
			}
		}
	}

	if resp.Diagnostics.HasError() {
		return
	}
	lbs, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, d.kc, &resp.Diagnostics,
		func() (interface{}, *http.Response, error) {
			apiResp, httpResp, err := lbApi.Limit(1000).XAuthToken(d.kc.XAuthToken).Execute()
			return apiResp, httpResp, err
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, d, httpResp, "ListLoadBalancers", err, &resp.Diagnostics)
		return
	}

	lbsTyped, ok := lbs.(*loadbalancer.LoadBalancerListModel)
	if !ok {
		resp.Diagnostics.AddError("Type assertion failed", "Failed to cast API response to expected type")
		return
	}

	var loadBalancersResult []loadbalancer.BnsLoadBalancerV1ApiGetLoadBalancerModelLoadBalancerModel
	err = copier.Copy(&loadBalancersResult, &lbsTyped.LoadBalancers)
	if err != nil {
		resp.Diagnostics.AddError("List conversion failed", fmt.Sprintf("loadBalancersResult conversion failed: %v", err))
		return
	}

	for _, lb := range loadBalancersResult {
		var lbModel loadBalancerDataSourceBaseModel
		ok := mapLoadBalancerBaseForDataSource(ctx, &lbModel, &lb, &resp.Diagnostics)
		if !ok {
			return
		}
		data.LoadBalancers = append(data.LoadBalancers, lbModel)
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package vpc

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-kakaocloud/internal/common"
	. "terraform-provider-kakaocloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/jinzhu/copier"
	"github.com/kakaoenterprise/kc-sdk-go/services/vpc"
)

var (
	_ datasource.DataSource              = &routeTablesDataSource{}
	_ datasource.DataSourceWithConfigure = &routeTablesDataSource{}
)

func NewRouteTablesDataSource() datasource.DataSource {
	return &routeTablesDataSource{}
}

type routeTablesDataSource struct {
	kc *common.KakaoCloudClient
}

func (d *routeTablesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_route_tables"
}

func (d *routeTablesDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
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
			"route_tables": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: MergeDataSourceSchemaAttributes(
						map[string]schema.Attribute{
							"id": schema.StringAttribute{
								Computed: true,
							},
						},
						routeTableDataSourceSchemaAttributes,
					),
				},
			},
			"timeouts": timeouts.Attributes(ctx),
		},
	}
}

func (d *routeTablesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config routeTablesDataSourceModel

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

	routeTableApi := d.kc.ApiClient.VPCRouteTableAPI.ListRouteTables(ctx)

	for _, f := range config.Filter {
		if f.Name.IsNull() || f.Name.IsUnknown() {
			continue
		}

		filterName := f.Name.ValueString()

		if !f.Value.IsNull() && !f.Value.IsUnknown() {
			v := f.Value.ValueString()
			switch filterName {
			case "id":
				routeTableApi = routeTableApi.Id(v)
			case "name":
				routeTableApi = routeTableApi.Name(v)
			case "vpc_id":
				routeTableApi = routeTableApi.VpcId(v)
			case "vpc_name":
				routeTableApi = routeTableApi.VpcName(v)
			case "provisioning_status":
				if ps, err := ToProvisioningStatus(v); err == nil {
					routeTableApi = routeTableApi.ProvisioningStatus(*ps)
				} else {
					resp.Diagnostics.AddError(
						"Invalid provisioning_status value",
						err.Error(),
					)
				}
			case "vpc_provisioning_status":
				if ps, err := ToProvisioningStatus(v); err == nil {
					routeTableApi = routeTableApi.VpcProvisioningStatus(*ps)
				} else {
					resp.Diagnostics.AddError(
						"Invalid vpc_provisioning_status value",
						err.Error(),
					)
				}
			case "subnet_id":
				routeTableApi = routeTableApi.SubnetId(v)
			case "subnet_name":
				routeTableApi = routeTableApi.SubnetName(v)
			case "association_count":
				routeTableApi = routeTableApi.AssociationCount(v)
			case "destination":
				routeTableApi = routeTableApi.Destination(v)
			case "created_at":
				if err := common.ValidateRFC3339(v); err == nil {
					routeTableApi = routeTableApi.CreatedAt(v)
				} else {
					resp.Diagnostics.AddError(
						"Invalid created_at value",
						err.Error(),
					)
				}
			case "updated_at":
				if err := common.ValidateRFC3339(v); err == nil {
					routeTableApi = routeTableApi.UpdatedAt(v)
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
	routeTableResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, d.kc, &resp.Diagnostics,
		func() (*vpc.RouteTableListModel, *http.Response, error) {
			return routeTableApi.Limit(1000).XAuthToken(d.kc.XAuthToken).Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, d, httpResp, "ListRouteTables", err, &resp.Diagnostics)
		return
	}

	var routeTableResult []vpc.BnsVpcV1ApiGetRouteTableModelRouteTableModel
	err = copier.Copy(&routeTableResult, &routeTableResp.VpcRouteTables)
	if err != nil {
		common.AddGeneralError(ctx, d, &resp.Diagnostics,
			fmt.Sprintf("Failed to convert routeTableResult: %v", err))
		return
	}

	for _, v := range routeTableResult {
		var tmpRouteTable routeTableBaseModel
		ok := mapRouteTableBaseModel(ctx, &tmpRouteTable, &v, &resp.Diagnostics)
		if !ok || resp.Diagnostics.HasError() {
			return
		}

		config.RouteTables = append(config.RouteTables, tmpRouteTable)
	}

	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (d *routeTablesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {

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

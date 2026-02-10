// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package tgw

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-kakaocloud/internal/common"
	"terraform-provider-kakaocloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/kakaoenterprise/kc-sdk-go/services/tgw"
)

var (
	_ datasource.DataSourceWithConfigure = &transitGatewayRoutesDataSource{}
)

func NewTransitGatewayRoutesDataSource() datasource.DataSource {
	return &transitGatewayRoutesDataSource{}
}

type transitGatewayRoutesDataSource struct {
	kc *common.KakaoCloudClient
}

func (d *transitGatewayRoutesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_transit_gateway_routes"
}

func (d *transitGatewayRoutesDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: utils.MergeDataSourceSchemaAttributes(
			getTransitGatewayRoutesDataSourceSchema(),
			map[string]schema.Attribute{
				"timeouts": timeouts.Attributes(ctx),
			},
		),
	}
}

func (d *transitGatewayRoutesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config transitGatewayRoutesDataSourceModel

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

	routeApi := d.kc.ApiClient.RouteTablesAPI.ListTgwRoutes(ctx, config.RouteTableId.ValueString())

	for _, f := range config.Filter {
		if f.Name.IsNull() || f.Name.IsUnknown() {
			continue
		}

		filterName := f.Name.ValueString()

		if !f.Value.IsNull() && !f.Value.IsUnknown() {
			v := f.Value.ValueString()

			switch filterName {
			case "destination_cidr_block":
				routeApi = routeApi.DestinationCidrBlock(v)
			case "route_type":
				routeApi = routeApi.RouteType(v)
			case "provisioning_status":
				if ps, err := ToTgwProvisioningStatus(v); err == nil {
					routeApi = routeApi.ProvisioningStatus(*ps)
				} else {
					resp.Diagnostics.AddError(
						"Invalid provisioning_status value",
						err.Error(),
					)
				}
			case "resource_type":
				routeApi = routeApi.ResourceType(v)
			case "resource_id":
				routeApi = routeApi.ResourceId(v)
			case "resource_name":
				routeApi = routeApi.ResourceName(v)
			case "resource_provisioning_status":
				routeApi = routeApi.ResourceProvisioningStatus(v)
			case "resource_attachment_id":
				routeApi = routeApi.ResourceAttachmentId(v)
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

	routesResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, d.kc, &resp.Diagnostics,
		func() (*tgw.GetTgwRouteTableRoutesResponseModel, *http.Response, error) {
			return routeApi.Limit(1000).XAuthToken(d.kc.XAuthToken).Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, d, httpResp, "ListTgwRoutes", err, &resp.Diagnostics)
		return
	}

	config.TransitGatewayRoutes = make([]transitGatewayRouteBaseModel, 0)
	for _, item := range routesResp.Routes {
		var routeModel transitGatewayRouteBaseModel
		ok := mapTransitGatewayRouteListModel(ctx, &routeModel, &item, config.RouteTableId.ValueString(), &resp.Diagnostics)
		if !ok || resp.Diagnostics.HasError() {
			return
		}
		config.TransitGatewayRoutes = append(config.TransitGatewayRoutes, routeModel)
	}

	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (d *transitGatewayRoutesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

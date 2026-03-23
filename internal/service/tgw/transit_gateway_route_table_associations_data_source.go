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
	_ datasource.DataSource              = &transitGatewayRouteTableAssociationsDataSource{}
	_ datasource.DataSourceWithConfigure = &transitGatewayRouteTableAssociationsDataSource{}
)

func NewTransitGatewayRouteTableAssociationsDataSource() datasource.DataSource {
	return &transitGatewayRouteTableAssociationsDataSource{}
}

type transitGatewayRouteTableAssociationsDataSource struct {
	kc *common.KakaoCloudClient
}

func (d *transitGatewayRouteTableAssociationsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_transit_gateway_route_table_associations"
}

func (d *transitGatewayRouteTableAssociationsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: utils.MergeDataSourceSchemaAttributes(
			map[string]schema.Attribute{
				"timeouts": timeouts.Attributes(ctx),
			},
			transitGatewayRouteTableAssociationsDataSourceSchemaAttributes,
		),
	}
}

func (d *transitGatewayRouteTableAssociationsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config transitGatewayRouteTableAssociationsDataSourceModel

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

	associationApi := d.kc.ApiClient.RouteTablesAPI.ListTgwRouteTableAssociations(ctx, config.RouteTableId.ValueString())

	for _, f := range config.Filter {
		if f.Name.IsNull() || f.Name.IsUnknown() {
			continue
		}

		filterName := f.Name.ValueString()

		if !f.Value.IsNull() && !f.Value.IsUnknown() {
			v := f.Value.ValueString()

			switch filterName {
			case "resource_id":
				associationApi = associationApi.ResourceId(v)
			case "resource_name":
				associationApi = associationApi.ResourceName(v)
			case "resource_provisioning_status":
				associationApi = associationApi.ResourceProvisioningStatus(v)
			case "resource_attachment_id":
				associationApi = associationApi.ResourceAttachmentId(v)
			case "resource_type":
				if rt, err := ToResourceType(v); err == nil {
					associationApi = associationApi.ResourceType(*rt)
				} else {
					resp.Diagnostics.AddError(
						"Invalid resource_type value",
						err.Error(),
					)
				}
			case "provisioning_status":
				if ps, err := ToTgwProvisioningStatus(v); err == nil {
					associationApi = associationApi.ProvisioningStatus(*ps)
				} else {
					resp.Diagnostics.AddError(
						"Invalid provisioning_status value",
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

	listResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, d.kc, &resp.Diagnostics,
		func() (*tgw.GetTgwRouteTableAssociationsResponseModel, *http.Response, error) {
			return associationApi.Limit(1000).XAuthToken(d.kc.XAuthToken).Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, d, httpResp, "ListTgwRouteTableAssociations", err, &resp.Diagnostics)
		return
	}

	config.Associations = make([]transitGatewayRouteTableAssociationBaseModel, 0)
	for _, assoc := range listResp.Associations {
		var assocModel transitGatewayRouteTableAssociationBaseModel
		ok := mapTransitGatewayRouteTableAssociationBaseModel(ctx, &assocModel, &assoc, config.RouteTableId.ValueString(), &resp.Diagnostics)
		if !ok || resp.Diagnostics.HasError() {
			return
		}
		config.Associations = append(config.Associations, assocModel)
	}

	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (d *transitGatewayRouteTableAssociationsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

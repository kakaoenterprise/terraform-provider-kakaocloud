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
	"github.com/kakaoenterprise/kc-sdk-go/v2/services/vpc"
)

var (
	_ datasource.DataSource              = &internetGatewaysDataSource{}
	_ datasource.DataSourceWithConfigure = &internetGatewaysDataSource{}
)

func NewInternetGatewaysDataSource() datasource.DataSource {
	return &internetGatewaysDataSource{}
}

type internetGatewaysDataSource struct {
	kc *common.KakaoCloudClient
}

func (d *internetGatewaysDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_internet_gateways"
}

func (d *internetGatewaysDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
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
			"igws": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: MergeDataSourceSchemaAttributes(
						map[string]schema.Attribute{
							"id": schema.StringAttribute{Computed: true},
						},
						internetGatewayDataSourceSchemaAttributes,
					),
				},
			},
			"timeouts": timeouts.Attributes(ctx),
		},
	}
}

func (d *internetGatewaysDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config internetGatewaysDataSourceModel

	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	readTimeout, diags := config.Timeouts.Read(ctx, common.DefaultReadTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, readTimeout)
	defer cancel()

	igwAPI := d.kc.ApiClient.VPCInternetGatewayAPI.ListInternetGateways(ctx)

	for _, filter := range config.Filter {
		if filter.Name.IsNull() || filter.Name.IsUnknown() || filter.Value.IsNull() || filter.Value.IsUnknown() {
			continue
		}

		name := filter.Name.ValueString()
		value := filter.Value.ValueString()

		switch name {
		case "id":
			igwAPI = igwAPI.Id(value)
		case "name":
			igwAPI = igwAPI.Name(value)
		case "vpc_id":
			igwAPI = igwAPI.VpcId(value)
		case "vpc_name":
			igwAPI = igwAPI.VpcName(value)
		case "vpc_cidr_block":
			igwAPI = igwAPI.VpcCidrBlock(value)
		case "attachment_id":
			igwAPI = igwAPI.AttachmentId(value)
		case "attachment_provisioning_status":
			igwAPI = igwAPI.AttachmentProvisioningStatus(value)
		case "nat_ip":
			igwAPI = igwAPI.NatIp(value)
		case "provisioning_status":
			igwAPI = igwAPI.ProvisioningStatus(value)
		case "created_at":
			if err := common.ValidateRFC3339(value); err != nil {
				resp.Diagnostics.AddError("Invalid created_at value", err.Error())
				continue
			}
			igwAPI = igwAPI.CreatedAt(value)
		case "updated_at":
			if err := common.ValidateRFC3339(value); err != nil {
				resp.Diagnostics.AddError("Invalid updated_at value", err.Error())
				continue
			}
			igwAPI = igwAPI.UpdatedAt(value)
		default:
			resp.Diagnostics.AddError(
				"Invalid filter name",
				fmt.Sprintf("filter %q is not supported", name),
			)
		}
	}

	if resp.Diagnostics.HasError() {
		return
	}

	igwResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, d.kc, &resp.Diagnostics,
		func() (*vpc.ListInternetGatewaysResponse, *http.Response, error) {
			return igwAPI.Limit(1000).XAuthToken(d.kc.XAuthToken).Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, d, httpResp, "ListInternetGateways", err, &resp.Diagnostics)
		return
	}

	config.Igws = make([]internetGatewayBaseModel, 0, len(igwResp.Igws))
	for i := range igwResp.Igws {
		var model internetGatewayBaseModel
		if !mapInternetGatewayModel(ctx, &model, &igwResp.Igws[i], &resp.Diagnostics) || resp.Diagnostics.HasError() {
			return
		}
		config.Igws = append(config.Igws, model)
	}

	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
}

func (d *internetGatewaysDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

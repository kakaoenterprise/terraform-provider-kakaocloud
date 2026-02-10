// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package tgw

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-kakaocloud/internal/common"
	. "terraform-provider-kakaocloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/kakaoenterprise/kc-sdk-go/services/tgw"
)

var (
	_ datasource.DataSource              = &transitGatewayAttachmentsDataSource{}
	_ datasource.DataSourceWithConfigure = &transitGatewayAttachmentsDataSource{}
)

func NewTransitGatewayAttachmentsDataSource() datasource.DataSource {
	return &transitGatewayAttachmentsDataSource{}
}

type transitGatewayAttachmentsDataSource struct {
	kc *common.KakaoCloudClient
}

func (d *transitGatewayAttachmentsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_transit_gateway_attachments"
}

func (d *transitGatewayAttachmentsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
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
			"transit_gateway_attachments": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: MergeDataSourceSchemaAttributes(
						map[string]schema.Attribute{
							"id": schema.StringAttribute{
								Computed: true,
							},
						},
						transitGatewayAttachmentDataSourceSchemaAttributes,
					),
				},
			},
			"timeouts": timeouts.Attributes(ctx),
		},
	}
}

func (d *transitGatewayAttachmentsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config transitGatewayAttachmentsDataSourceModel

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

	attachmentApi := d.kc.ApiClient.AttachmentsAPI.ListTgwAttachments(ctx)

	for _, f := range config.Filter {
		if f.Name.IsNull() || f.Name.IsUnknown() {
			continue
		}

		filterName := f.Name.ValueString()

		if !f.Value.IsNull() && !f.Value.IsUnknown() {
			v := f.Value.ValueString()

			switch filterName {
			case "id":
				attachmentApi = attachmentApi.Id(v)
			case "tgw_id":
				attachmentApi = attachmentApi.TgwId(v)
			case "tgw_name":
				attachmentApi = attachmentApi.TgwName(v)
			case "provisioning_status":
				if ps, err := ToTgwProvisioningStatus(v); err == nil {
					attachmentApi = attachmentApi.ProvisioningStatus(*ps)
				} else {
					resp.Diagnostics.AddError(
						"Invalid provisioning_status value",
						err.Error(),
					)
				}
			case "resource_id":
				attachmentApi = attachmentApi.ResourceId(v)
			case "resource_name":
				attachmentApi = attachmentApi.ResourceName(v)
			case "route_table_id":
				attachmentApi = attachmentApi.RouteTableId(v)
			case "route_table_name":
				attachmentApi = attachmentApi.RouteTableName(v)
			case "created_at":
				if err := common.ValidateRFC3339(v); err == nil {
					attachmentApi = attachmentApi.CreatedAt(v)
				} else {
					resp.Diagnostics.AddError(
						"Invalid created_at value",
						err.Error(),
					)
				}
			case "updated_at":
				if err := common.ValidateRFC3339(v); err == nil {
					attachmentApi = attachmentApi.UpdatedAt(v)
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

	attachResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, d.kc, &resp.Diagnostics,
		func() (*tgw.GetTgwAttachmentsResponseModel, *http.Response, error) {
			return attachmentApi.Limit(1000).XAuthToken(d.kc.XAuthToken).Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, d, httpResp, "ListTgwAttachments", err, &resp.Diagnostics)
		return
	}

	config.TransitGatewayAttachments = make([]transitGatewayAttachmentBaseModel, 0)
	for _, item := range attachResp.Attachments {
		var attachModel transitGatewayAttachmentBaseModel
		ok := mapTransitGatewayAttachmentModelFromList(ctx, &attachModel, &item, &resp.Diagnostics)
		if !ok || resp.Diagnostics.HasError() {
			return
		}
		config.TransitGatewayAttachments = append(config.TransitGatewayAttachments, attachModel)
	}

	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
}

func (d *transitGatewayAttachmentsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

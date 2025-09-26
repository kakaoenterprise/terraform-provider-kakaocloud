// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package network

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"terraform-provider-kakaocloud/internal/common"
	"terraform-provider-kakaocloud/internal/docs"
	. "terraform-provider-kakaocloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/kakaoenterprise/kc-sdk-go/services/network"
)

var (
	_ datasource.DataSource              = &publicIpsDataSource{}
	_ datasource.DataSourceWithConfigure = &publicIpsDataSource{}
)

func NewPublicIpsDataSource() datasource.DataSource {
	return &publicIpsDataSource{}
}

type publicIpsDataSource struct {
	kc *common.KakaoCloudClient
}

func (d *publicIpsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *publicIpsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_public_ips"
}

func (d *publicIpsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: docs.GetDataSourceDescription("PublicIps"),
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
			"public_ips": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: MergeAttributes[schema.Attribute](
						map[string]schema.Attribute{
							"id": schema.StringAttribute{
								Required:    true,
								Description: "Image ID",
							},
						},
						publicIpDataSourceSchemaAttributes,
					),
				},
			},
			"timeouts": timeouts.Attributes(ctx),
		},
	}
}

func (d *publicIpsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config publicIpsDataSourceModel

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

	publicIpApi := d.kc.ApiClient.PublicIPAPI.ListPublicIps(ctx).XAuthToken(d.kc.XAuthToken)

	for _, f := range config.Filter {
		if f.Name.IsNull() || f.Name.IsUnknown() {
			continue
		}

		filterName := f.Name.ValueString()

		if !f.Value.IsNull() && !f.Value.IsUnknown() {
			v := f.Value.ValueString()
			switch filterName {
			case "id":
				publicIpApi = publicIpApi.Id(v)
			case "status":
				if status, err := ToStatusType(v); err == nil {
					publicIpApi = publicIpApi.Status(*status)
				} else {
					resp.Diagnostics.AddError(
						"Invalid status",
						err.Error(),
					)
				}
			case "public_ip":
				publicIpApi = publicIpApi.PublicIp(v)
			case "related_resource_id":
				publicIpApi = publicIpApi.RelatedResourceId(v)
			case "related_resource_name":
				publicIpApi = publicIpApi.RelatedResourceName(v)
			case "created_at":
				if err := common.ValidateRFC3339(v); err == nil {
					publicIpApi = publicIpApi.CreatedAt(v)
				} else {
					resp.Diagnostics.AddError(
						"Invalid created_at value",
						err.Error(),
					)
				}
			case "updated_at":
				if err := common.ValidateRFC3339(v); err == nil {
					publicIpApi = publicIpApi.UpdatedAt(v)
				} else {
					resp.Diagnostics.AddError(
						"Invalid created_at value",
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

	publicIpResp, httpResp, err := common.
		ExecuteWithRetryAndAuth(ctx, d.kc, &resp.Diagnostics,
			func() (*network.PublicIpListModel, *http.Response, error) {
				return publicIpApi.Limit(1000).XAuthToken(d.kc.XAuthToken).Execute()
			},
		)
	if err != nil {
		common.AddApiActionError(ctx, d, httpResp, "ListPublicIps", err, &resp.Diagnostics)
		return
	}

	for _, v := range publicIpResp.PublicIps {
		var tmpPublicIp publicIpBaseModel
		ok := d.mapPublicIps(ctx, &tmpPublicIp, &v, &resp.Diagnostics)
		if !ok || resp.Diagnostics.HasError() {
			return
		}

		config.PublicIps = append(config.PublicIps, tmpPublicIp)
	}

	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
}

func ToStatusType(v string) (*network.PublicIpStatus, error) {
	status := network.PublicIpStatus(strings.ToLower(v))

	for _, allowed := range network.AllowedPublicIpStatusEnumValues {
		if status == allowed {
			return &status, nil
		}
	}

	allowedStrings := make([]string, len(network.AllowedPublicIpStatusEnumValues))
	for i, s := range network.AllowedPublicIpStatusEnumValues {
		allowedStrings[i] = string(s)
	}

	return nil, fmt.Errorf(
		"status '%s' is not allowed (allowed: %s)",
		v,
		strings.Join(allowedStrings, ", "),
	)
}

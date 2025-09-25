// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpc

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"terraform-provider-kakaocloud/internal/common"
	. "terraform-provider-kakaocloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/kakaoenterprise/kc-sdk-go/services/vpc"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &vpcsDataSource{}
	_ datasource.DataSourceWithConfigure = &vpcsDataSource{}
)

func NewVpcsDataSource() datasource.DataSource {
	return &vpcsDataSource{}
}

type vpcsDataSource struct {
	kc *common.KakaoCloudClient
}

func (d *vpcsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpcs"
}

func (d *vpcsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "kakaocloud VPC 목록을 조회하는 데이터 소스",
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
			"vpcs": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: MergeDataSourceSchemaAttributes(
						map[string]schema.Attribute{
							"id": schema.StringAttribute{
								Computed:    true,
								Description: "VPC ID",
							},
						},
						vpcDataSourceSchemaAttributes,
					),
				},
			},
			"timeouts": timeouts.Attributes(ctx),
		},
	}
}

func (d *vpcsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config vpcsDataSourceModel

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

	vpcApi := d.kc.ApiClient.VPCAPI.ListVpcs(ctx)

	for _, f := range config.Filter {
		if f.Name.IsNull() || f.Name.IsUnknown() {
			continue
		}

		filterName := f.Name.ValueString()

		if !f.Value.IsNull() && !f.Value.IsUnknown() {
			v := f.Value.ValueString()

			switch filterName {
			case "id":
				vpcApi = vpcApi.Id(v)
			case "name":
				vpcApi = vpcApi.Name(v)
			case "cidr_block":
				vpcApi = vpcApi.CidrBlock(v)
			case "provisioning_status":
				if ps, err := ToProvisioningStatus(v); err == nil {
					vpcApi = vpcApi.ProvisioningStatus(*ps)
				} else {
					resp.Diagnostics.AddError(
						"Invalid provisioning_status value",
						err.Error(),
					)
				}
			case "is_default":
				if b, err := strconv.ParseBool(v); err == nil {
					vpcApi = vpcApi.IsDefault(b)
				} else {
					resp.Diagnostics.AddError(
						"Invalid is_default value",
						fmt.Sprintf("expected true/false but got %q (error: %s)", v, err),
					)
				}
			case "created_at":
				if err := common.ValidateRFC3339(v); err == nil {
					vpcApi = vpcApi.CreatedAt(v)
				} else {
					resp.Diagnostics.AddError(
						"Invalid created_at value",
						err.Error(),
					)
				}
			case "updated_at":
				if err := common.ValidateRFC3339(v); err == nil {
					vpcApi = vpcApi.UpdatedAt(v)
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
	vpcResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, d.kc, &resp.Diagnostics,
		func() (*vpc.VPCListModel, *http.Response, error) {
			return vpcApi.Limit(1000).XAuthToken(d.kc.XAuthToken).Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, d, httpResp, "ListVpcs", err, &resp.Diagnostics)
		return
	}

	for _, v := range vpcResp.Vpcs {
		var tmpVpc vpcBaseModel
		ok := mapVpcListModel(ctx, &tmpVpc, &v, &resp.Diagnostics)
		if !ok || resp.Diagnostics.HasError() {
			return
		}

		config.Vpcs = append(config.Vpcs, tmpVpc)
	}

	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (d *vpcsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Add a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
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

func ToProvisioningStatus(v string) (*vpc.ProvisioningStatus, error) {
	ps := vpc.ProvisioningStatus(strings.ToUpper(v))

	for _, allowed := range vpc.AllowedProvisioningStatusEnumValues {
		if ps == allowed {
			return &ps, nil
		}
	}
	return nil, fmt.Errorf("invalid provisioning status: %s (allowed: %v)", v, vpc.AllowedProvisioningStatusEnumValues)
}

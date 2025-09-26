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
	"terraform-provider-kakaocloud/internal/docs"
	. "terraform-provider-kakaocloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/jinzhu/copier"
	"github.com/kakaoenterprise/kc-sdk-go/services/vpc"
)

var (
	_ datasource.DataSource              = &subnetsDataSource{}
	_ datasource.DataSourceWithConfigure = &subnetsDataSource{}
)

func NewSubnetsDataSource() datasource.DataSource {
	return &subnetsDataSource{}
}

type subnetsDataSource struct {
	kc *common.KakaoCloudClient
}

func (d *subnetsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subnets"
}

func (d *subnetsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: docs.GetDataSourceDescription("Subnets"),
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
			"subnets": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: MergeDataSourceSchemaAttributes(
						map[string]schema.Attribute{
							"id": schema.StringAttribute{
								Computed:    true,
								Description: "Subnet ID",
							},
						},
						subnetDataSourceSchemaAttributes,
					),
				},
			},
			"timeouts": timeouts.Attributes(ctx),
		},
	}
}

func (d *subnetsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config subnetsDataSourceModel

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

	subnetApi := d.kc.ApiClient.VPCSubnetAPI.ListSubnets(ctx)

	for _, f := range config.Filter {
		if f.Name.IsNull() || f.Name.IsUnknown() {
			continue
		}

		filterName := f.Name.ValueString()

		if !f.Value.IsNull() && !f.Value.IsUnknown() {
			v := f.Value.ValueString()

			switch filterName {
			case "id":
				subnetApi = subnetApi.Id(v)

			case "name":
				subnetApi = subnetApi.Name(v)

			case "availability_zone":
				if az, err := ToAvailabilityZone(v); err == nil {
					subnetApi = subnetApi.AvailabilityZone(*az)
				} else {
					resp.Diagnostics.AddError("Invalid availability_zone", err.Error())
				}

			case "provisioning_status":
				if ps, err := ToProvisioningStatus(v); err == nil {
					subnetApi = subnetApi.ProvisioningStatus(*ps)
				} else {
					resp.Diagnostics.AddError("Invalid provisioning_status", err.Error())
				}

			case "operating_status":
				if os, err := ToSubnetOperatingStatus(v); err == nil {
					subnetApi = subnetApi.OperatingStatus(*os)
				} else {
					resp.Diagnostics.AddError("Invalid operating_status", err.Error())
				}

			case "cidr_block":
				subnetApi = subnetApi.CidrBlock(v)

			case "vpc_id":
				subnetApi = subnetApi.VpcId(v)

			case "vpc_name":
				subnetApi = subnetApi.VpcName(v)

			case "route_table_id":
				subnetApi = subnetApi.RouteTableId(v)

			case "route_table_name":
				subnetApi = subnetApi.RouteTableName(v)

			case "is_shared":
				if b, err := strconv.ParseBool(v); err == nil {
					subnetApi = subnetApi.IsShared(b)
				} else {
					resp.Diagnostics.AddError("Invalid is_shared value", fmt.Sprintf("expected true/false but got %q (error: %s)", v, err))
				}

			case "created_at":
				if err := common.ValidateRFC3339(v); err == nil {
					subnetApi = subnetApi.CreatedAt(v)
				} else {
					resp.Diagnostics.AddError("Invalid created_at value", err.Error())
				}

			case "updated_at":
				if err := common.ValidateRFC3339(v); err == nil {
					subnetApi = subnetApi.UpdatedAt(v)
				} else {
					resp.Diagnostics.AddError("Invalid updated_at value", err.Error())
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
	subnetResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, d.kc, &resp.Diagnostics,
		func() (*vpc.SubnetListModel, *http.Response, error) {
			return subnetApi.Limit(1000).XAuthToken(d.kc.XAuthToken).Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, d, httpResp, "ListSubnets", err, &resp.Diagnostics)
		return
	}

	var subnetsResult []vpc.BnsVpcV1ApiGetSubnetModelSubnetModel
	err = copier.Copy(&subnetsResult, &subnetResp.Subnets)
	if err != nil {
		common.AddGeneralError(ctx, d, &resp.Diagnostics,
			fmt.Sprintf("subnetsResult transform fail: %v", err))
		return
	}

	for _, v := range subnetsResult {
		var tmpSubnet subnetBaseModel
		ok := mapSubnetBaseModel(&tmpSubnet, &v, &resp.Diagnostics)
		if !ok || resp.Diagnostics.HasError() {
			return
		}

		config.Subnets = append(config.Subnets, tmpSubnet)
	}

	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (d *subnetsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {

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

func ToAvailabilityZone(v string) (*vpc.AvailabilityZone, error) {
	az := vpc.AvailabilityZone(v)

	for _, allowed := range vpc.AllowedAvailabilityZoneEnumValues {
		if az == allowed {
			return &az, nil
		}
	}
	return nil, fmt.Errorf("invalid availability_zone: %s (allowed: %v)", v, vpc.AllowedAvailabilityZoneEnumValues)
}

func ToSubnetOperatingStatus(v string) (*vpc.SubnetOperatingStatus, error) {
	os := vpc.SubnetOperatingStatus(strings.ToUpper(v))
	for _, allowed := range vpc.AllowedSubnetOperatingStatusEnumValues {
		if os == allowed {
			return &os, nil
		}
	}
	return nil, fmt.Errorf("invalid subnet operating status: %s (allowed: %v)", v, vpc.AllowedSubnetOperatingStatusEnumValues)
}

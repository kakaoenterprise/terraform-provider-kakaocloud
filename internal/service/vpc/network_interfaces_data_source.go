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

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &networkInterfacesDataSource{}
	_ datasource.DataSourceWithConfigure = &networkInterfacesDataSource{}
)

func NewNetworkInterfacesDataSource() datasource.DataSource {
	return &networkInterfacesDataSource{}
}

type networkInterfacesDataSource struct {
	kc *common.KakaoCloudClient
}

func (d *networkInterfacesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network_interfaces"
}

func (d *networkInterfacesDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "kakaocloud Network Interface 목록 조회하는 데이터 소스",
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
			"network_interfaces": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: MergeDataSourceSchemaAttributes(
						map[string]schema.Attribute{
							"id": schema.StringAttribute{
								Computed:    true,
								Description: "Network Interface ID",
							},
						},
						networkInterfaceDataSourceBaseSchema,
					),
				},
			},
			"timeouts": timeouts.Attributes(ctx),
		},
	}
}

func (d *networkInterfacesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config networkInterfacesDataSourceModel

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

	networkInterfaceApi := d.kc.ApiClient.NetworkInterfaceAPI.ListNetworkInterfaces(ctx)

	for _, f := range config.Filter {
		if f.Name.IsNull() || f.Name.IsUnknown() {
			continue
		}

		filterName := f.Name.ValueString()

		if !f.Value.IsNull() && !f.Value.IsUnknown() {
			v := f.Value.ValueString()

			switch filterName {
			case "id":
				networkInterfaceApi = networkInterfaceApi.Id(v)

			case "name":
				networkInterfaceApi = networkInterfaceApi.Name(v)

			case "status":
				if status, err := ToNetworkInterfaceStatus(v); err == nil {
					networkInterfaceApi = networkInterfaceApi.Status(*status)
				} else {
					resp.Diagnostics.AddError("Invalid status", err.Error())
				}

			case "private_ip":
				networkInterfaceApi = networkInterfaceApi.PrivateIp(v)

			case "public_ip":
				networkInterfaceApi = networkInterfaceApi.PublicIp(v)

			case "device_id":
				networkInterfaceApi = networkInterfaceApi.DeviceId(v)

			case "device_owner":
				networkInterfaceApi = networkInterfaceApi.DeviceOwner(v)

			case "subnet_id":
				networkInterfaceApi = networkInterfaceApi.SubnetId(v)

			case "mac_address":
				networkInterfaceApi = networkInterfaceApi.MacAddress(v)

			case "security_group_id":
				networkInterfaceApi = networkInterfaceApi.SecurityGroupId(v)

			case "security_group_name":
				networkInterfaceApi = networkInterfaceApi.SecurityGroupName(v)

			case "created_at":
				if err := common.ValidateRFC3339(v); err == nil {
					networkInterfaceApi = networkInterfaceApi.CreatedAt(v)
				} else {
					resp.Diagnostics.AddError("Invalid created_at value", err.Error())
				}

			case "updated_at":
				if err := common.ValidateRFC3339(v); err == nil {
					networkInterfaceApi = networkInterfaceApi.UpdatedAt(v)
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
	networkInterfaceResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, d.kc, &resp.Diagnostics,
		func() (*vpc.NetworkInterfaceListModel, *http.Response, error) {
			return networkInterfaceApi.Limit(1000).XAuthToken(d.kc.XAuthToken).Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, d, httpResp, "ListNetworkInterfaces", err, &resp.Diagnostics)
		return
	}

	var networkInterfaceResult []vpc.BnsVpcV1ApiGetNetworkInterfaceModelNetworkInterfaceModel
	err = copier.Copy(&networkInterfaceResult, &networkInterfaceResp.NetworkInterfaces)
	if err != nil {
		common.AddGeneralError(ctx, d, &resp.Diagnostics,
			fmt.Sprintf("networkInterfaceResult transform fail: %v", err))
		return
	}

	for _, v := range networkInterfaceResult {
		var tmpNi networkInterfaceBaseModel
		ok := mapNetworkInterfaceBaseModel(ctx, &tmpNi, &v, &resp.Diagnostics)
		if !ok || resp.Diagnostics.HasError() {
			return
		}

		config.NetworkInterfaces = append(config.NetworkInterfaces, tmpNi)
	}

	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (d *networkInterfacesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func ToNetworkInterfaceStatus(s string) (*vpc.NetworkInterfaceStatus, error) {
	status := vpc.NetworkInterfaceStatus(s)
	switch status {
	case vpc.NETWORKINTERFACESTATUS_AVAILABLE, vpc.NETWORKINTERFACESTATUS_IN_USE:
		return &status, nil
	default:
		return nil, fmt.Errorf("invalid network interface status: %q", s)
	}
}

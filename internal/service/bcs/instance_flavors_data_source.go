// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bcs

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"terraform-provider-kakaocloud/internal/common"
	. "terraform-provider-kakaocloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/jinzhu/copier"
	"github.com/kakaoenterprise/kc-sdk-go/services/bcs"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &instanceFlavorsDataSource{}
	_ datasource.DataSourceWithConfigure = &instanceFlavorsDataSource{}
)

func NewInstanceFlavorsDataSource() datasource.DataSource {
	return &instanceFlavorsDataSource{}
}

type instanceFlavorsDataSource struct {
	kc *common.KakaoCloudClient
}

func (d *instanceFlavorsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_instance_flavors"
}

func (d *instanceFlavorsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "kakaocloud Flavor 목록을 조회하는 데이터 소스",
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
			"instance_flavors": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: MergeDataSourceSchemaAttributes(
						map[string]schema.Attribute{
							"id": schema.StringAttribute{
								Computed:    true,
								Description: "Flavor ID",
							},
						},
						instanceFlavorDataSourceSchemaAttributes,
					),
				},
			},
			"timeouts": timeouts.Attributes(ctx),
		},
	}
}

func (d *instanceFlavorsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config instanceFlavorsDataSourceModel

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

	flavorApi := d.kc.ApiClient.FlavorAPI.ListInstanceTypes(ctx)

	for _, f := range config.Filter {
		if f.Name.IsNull() || f.Name.IsUnknown() {
			continue
		}

		filterName := f.Name.ValueString()

		if !f.Value.IsNull() && !f.Value.IsUnknown() {
			v := f.Value.ValueString()

			switch filterName {
			case "id":
				flavorApi = flavorApi.Id(v)
			case "name":
				flavorApi = flavorApi.Name(v)
			case "is_burstable":
				if b, err := strconv.ParseBool(v); err == nil {
					flavorApi = flavorApi.IsBurstable(b)
				} else {
					resp.Diagnostics.AddError("Invalid is_burstable value", fmt.Sprintf("expected true/false but got %q (error: %s)", v, err))
				}
			case "vcpus":
				if i, err := strconv.Atoi(v); err == nil {
					flavorApi = flavorApi.Vcpus(int32(i))
				} else {
					resp.Diagnostics.AddError("Invalid vcpus value", fmt.Sprintf("expected integer but got %q (error: %s)", v, err))
				}
			case "architecture":
				flavorApi = flavorApi.Architecture(v)
			case "memory_mb":
				if i, err := strconv.Atoi(v); err == nil {
					flavorApi = flavorApi.MemoryMb(int32(i))
				} else {
					resp.Diagnostics.AddError("Invalid memory_mb value", fmt.Sprintf("expected integer but got %q (error: %s)", v, err))
				}
			case "instance_type":
				if instanceType, err := ToInstanceType(v); err == nil {
					flavorApi = flavorApi.InstanceType(*instanceType)
				} else {
					resp.Diagnostics.AddError("Invalid instance_type", err.Error())
				}
			case "instance_family":
				flavorApi = flavorApi.InstanceFamily(v)
			case "instance_size":
				flavorApi = flavorApi.InstanceSize(v)
			case "manufacturer":
				flavorApi = flavorApi.Manufacturer(v)
			case "maximum_network_interfaces":
				if i, err := strconv.Atoi(v); err == nil {
					flavorApi = flavorApi.MaximumNetworkInterfaces(int32(i))
				} else {
					resp.Diagnostics.AddError("Invalid maximum_network_interfaces value", fmt.Sprintf("expected integer but got %q (error: %s)", v, err))
				}
			case "processor":
				flavorApi = flavorApi.Processor(v)
			case "created_at":
				if err := common.ValidateRFC3339(v); err == nil {
					flavorApi = flavorApi.CreatedAt(v)
				} else {
					resp.Diagnostics.AddError("Invalid created_at value", err.Error())
				}
			case "updated_at":
				if err := common.ValidateRFC3339(v); err == nil {
					flavorApi = flavorApi.UpdatedAt(v)
				} else {
					resp.Diagnostics.AddError("Invalid updated_at value", err.Error())
				}
			}
		}
	}

	if resp.Diagnostics.HasError() {
		return
	}

	flavorResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, d.kc, &resp.Diagnostics,
		func() (*bcs.FlavorListModel, *http.Response, error) {
			return flavorApi.Limit(1000).XAuthToken(d.kc.XAuthToken).Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, d, httpResp, "ListInstanceType", err, &resp.Diagnostics)
		return
	}

	var instanceFlavorResult []bcs.BcsInstanceV1ApiGetInstanceTypeModelFlavorModel
	err = copier.Copy(&instanceFlavorResult, &flavorResp.Flavors)
	if err != nil {
		resp.Diagnostics.AddError("List 변환 실패", fmt.Sprintf("instanceFlavorResult 변환 실패: %v", err))
		return
	}

	for _, v := range instanceFlavorResult {
		var tmpFlavor instanceFlavorBaseModel
		ok := mapInstanceFlavorBaseModel(ctx, &tmpFlavor, &v, &resp.Diagnostics)
		if !ok || resp.Diagnostics.HasError() {
			return
		}

		config.InstanceFlavors = append(config.InstanceFlavors, tmpFlavor)
	}

	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (d *instanceFlavorsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

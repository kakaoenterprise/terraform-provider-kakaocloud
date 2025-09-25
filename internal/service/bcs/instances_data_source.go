// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bcs

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
	"github.com/kakaoenterprise/kc-sdk-go/services/bcs"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &instancesDataSource{}
	_ datasource.DataSourceWithConfigure = &instancesDataSource{}
)

func NewInstancesDataSource() datasource.DataSource {
	return &instancesDataSource{}
}

type instancesDataSource struct {
	kc *common.KakaoCloudClient
}

func (d *instancesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_instances"
}

func (d *instancesDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "kakaocloud 인스턴스 목록을 조회하는 데이터 소스",
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
			"instances": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: MergeDataSourceSchemaAttributes(
						map[string]schema.Attribute{
							"id": schema.StringAttribute{
								Computed:    true,
								Description: "Instance ID",
							},
						},
						instanceDataSourceSchemaAttributes,
					),
				},
			},
			"timeouts": timeouts.Attributes(ctx),
		},
	}
}

func (d *instancesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config instancesDataSourceModel

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

	instanceApi := d.kc.ApiClient.InstanceAPI.ListInstances(ctx)

	for _, f := range config.Filter {
		if f.Name.IsNull() || f.Name.IsUnknown() {
			continue
		}

		filterName := f.Name.ValueString()

		if !f.Value.IsNull() && !f.Value.IsUnknown() {
			v := f.Value.ValueString()
			switch filterName {
			case "name":
				instanceApi = instanceApi.Name(v)
			case "id":
				instanceApi = instanceApi.Id(v)
			case "status":
				instanceApi = instanceApi.Status(v)
			case "vm_state":
				instanceApi = instanceApi.VmState(v)
			case "flavor_name":
				instanceApi = instanceApi.FlavorName(v)
			case "image_name":
				instanceApi = instanceApi.ImageName(v)
			case "private_ip":
				instanceApi = instanceApi.PrivateIp(v)
			case "public_ip":
				instanceApi = instanceApi.PublicIp(v)
			case "availability_zone":
				if az, err := ToAvailabilityZone(v); err == nil {
					instanceApi = instanceApi.AvailabilityZone(*az)
				} else {
					resp.Diagnostics.AddError(
						"Invalid availability_zone",
						err.Error(),
					)
				}
			case "instance_type":
				if instanceType, err := ToInstanceType(v); err == nil {
					instanceApi = instanceApi.InstanceType(*instanceType)
				} else {
					resp.Diagnostics.AddError(
						"Invalid instance_type",
						err.Error(),
					)
				}
			case "user_id":
				instanceApi = instanceApi.UserId(v)
			case "hostname":
				instanceApi = instanceApi.Hostname(v)
			case "os_type":
				instanceApi = instanceApi.OsType(v)
			case "is_hadoop":
				if b, err := strconv.ParseBool(v); err == nil {
					instanceApi = instanceApi.IsHadoop(b)
				} else {
					resp.Diagnostics.AddError(
						"Invalid is_hadoop value",
						fmt.Sprintf("expected true/false but got %q (error: %s)", v, err),
					)
				}
			case "is_k8se":
				if b, err := strconv.ParseBool(v); err == nil {
					instanceApi = instanceApi.IsK8se(b)
				} else {
					resp.Diagnostics.AddError(
						"Invalid is_k8se value",
						fmt.Sprintf("expected true/false but got %q (error: %s)", v, err),
					)
				}
			case "created_at":
				if err := common.ValidateRFC3339(v); err == nil {
					instanceApi = instanceApi.CreatedAt(v)
				} else {
					resp.Diagnostics.AddError(
						"Invalid created_at value",
						err.Error(),
					)
				}
			case "updated_at":
				if err := common.ValidateRFC3339(v); err == nil {
					instanceApi = instanceApi.UpdatedAt(v)
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

	instanceResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, d.kc, &resp.Diagnostics,
		func() (*bcs.InstanceListModel, *http.Response, error) {
			return instanceApi.Limit(1000).XAuthToken(d.kc.XAuthToken).Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, d, httpResp, "ListInstances", err, &resp.Diagnostics)
		return
	}

	for _, v := range instanceResp.Instances {
		var tmpInstance instanceBaseModel
		ok := mapInstanceListModel(ctx, &tmpInstance, &v, &resp.Diagnostics)
		if !ok || resp.Diagnostics.HasError() {
			return
		}

		config.Instances = append(config.Instances, tmpInstance)
	}

	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (d *instancesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func ToAvailabilityZone(v string) (*bcs.AvailabilityZone, error) {
	az := bcs.AvailabilityZone(v)

	for _, allowed := range bcs.AllowedAvailabilityZoneEnumValues {
		if az == allowed {
			return &az, nil
		}
	}
	return nil, fmt.Errorf("invalid availability_zone: %s (allowed: %v)", v, bcs.AllowedAvailabilityZoneEnumValues)
}

func ToInstanceType(v string) (*bcs.InstanceType, error) {
	instanceType := bcs.InstanceType(strings.ToLower(v))

	for _, allowed := range bcs.AllowedInstanceTypeEnumValues {
		if instanceType == allowed {
			return &instanceType, nil
		}
	}
	return nil, fmt.Errorf("invalid instance_type: %s (allowed: %v)", v, bcs.AllowedInstanceTypeEnumValues)
}

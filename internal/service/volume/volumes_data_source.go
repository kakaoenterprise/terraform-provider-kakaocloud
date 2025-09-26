// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package volume

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"terraform-provider-kakaocloud/internal/common"
	"terraform-provider-kakaocloud/internal/docs"
	. "terraform-provider-kakaocloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/kakaoenterprise/kc-sdk-go/services/volume"
)

var (
	_ datasource.DataSource              = &volumesDataSource{}
	_ datasource.DataSourceWithConfigure = &volumesDataSource{}
)

func NewVolumesDataSource() datasource.DataSource {
	return &volumesDataSource{}
}

type volumesDataSource struct {
	kc *common.KakaoCloudClient
}

func (d *volumesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_volumes"
}

func (d *volumesDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: docs.GetDataSourceDescription("Volumes"),
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
			"volumes": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: MergeDataSourceSchemaAttributes(
						map[string]schema.Attribute{
							"id": schema.StringAttribute{
								Computed:    true,
								Description: "Volume ID",
							},
						},
						volumeDataSourceSchemaAttributes,
					),
				},
			},
			"timeouts": timeouts.Attributes(ctx),
		},
	}
}

func (d *volumesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config volumesDataSourceModel

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

	volumeApi := d.kc.ApiClient.VolumeAPI.ListVolumes(ctx)

	for _, f := range config.Filter {
		if f.Name.IsNull() || f.Name.IsUnknown() {
			continue
		}

		filterName := f.Name.ValueString()

		if !f.Value.IsNull() && !f.Value.IsUnknown() {
			v := f.Value.ValueString()

			switch filterName {
			case "name":
				volumeApi = volumeApi.Name(v)
			case "id":
				volumeApi = volumeApi.Id(v)
			case "status":
				volumeApi = volumeApi.Status(v)
			case "instance_id":
				volumeApi = volumeApi.InstanceId(v)
			case "mount_point":
				volumeApi = volumeApi.MountPoint(v)
			case "type":
				volumeApi = volumeApi.Type_(v)
			case "size":
				if i, err := strconv.Atoi(v); err == nil {
					volumeApi = volumeApi.Size(int32(i))
				} else {
					resp.Diagnostics.AddError(
						"Invalid size filter",
						fmt.Sprintf("Expected integer for 'size', got: %q (error: %s)", v, err),
					)
				}
			case "availability_zone":
				if az, err := ToAvailabilityZone(v); err == nil {
					volumeApi = volumeApi.AvailabilityZone(*az)
				} else {
					resp.Diagnostics.AddError(
						"Invalid availability_zone",
						err.Error(),
					)
				}
			case "instance_name":
				volumeApi = volumeApi.InstanceName(v)
			case "volume_type":
				volumeApi = volumeApi.VolumeType(v)
			case "attach_status":
				volumeApi = volumeApi.AttachStatus(v)
			case "is_bootable":
				if b, err := strconv.ParseBool(v); err == nil {
					volumeApi = volumeApi.IsBootable(b)
				} else {
					resp.Diagnostics.AddError(
						"Invalid is_bootable value",
						fmt.Sprintf("expected true/false but got %q (error: %s)", v, err),
					)
				}
			case "is_encrypted":
				if b, err := strconv.ParseBool(v); err == nil {
					volumeApi = volumeApi.IsEncrypted(b)
				} else {
					resp.Diagnostics.AddError(
						"Invalid is_encrypted value",
						fmt.Sprintf("expected true/false but got %q (error: %s)", v, err),
					)
				}
			case "is_root":
				if b, err := strconv.ParseBool(v); err == nil {
					volumeApi = volumeApi.IsRoot(b)
				} else {
					resp.Diagnostics.AddError(
						"Invalid is_root value",
						fmt.Sprintf("expected true/false but got %q (error: %s)", v, err),
					)
				}
			case "created_at":
				if err := common.ValidateRFC3339(v); err == nil {
					volumeApi = volumeApi.CreatedAt(v)
				} else {
					resp.Diagnostics.AddError(
						"Invalid created_at value",
						err.Error(),
					)
				}
			case "updated_at":
				if err := common.ValidateRFC3339(v); err == nil {
					volumeApi = volumeApi.UpdatedAt(v)
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

	volumesResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, d.kc, &resp.Diagnostics,
		func() (*volume.VolumeListModel, *http.Response, error) {
			return volumeApi.Limit(1000).XAuthToken(d.kc.XAuthToken).Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, d, httpResp, "ListVolumes", err, &resp.Diagnostics)
		return
	}

	for _, v := range volumesResp.Volumes {
		var tmpVolume volumeBaseModel
		ok := mapVolumeListModel(ctx, &tmpVolume, &v, &resp.Diagnostics)
		if !ok || resp.Diagnostics.HasError() {
			return
		}

		config.Volumes = append(config.Volumes, tmpVolume)
	}

	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (d *volumesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {

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

func ToAvailabilityZone(v string) (*volume.AvailabilityZone, error) {
	az := volume.AvailabilityZone(v)

	for _, allowed := range volume.AllowedAvailabilityZoneEnumValues {
		if az == allowed {
			return &az, nil
		}
	}
	return nil, fmt.Errorf("invalid availability_zone: %s (allowed: %v)", v, volume.AllowedAvailabilityZoneEnumValues)
}

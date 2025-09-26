// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package image

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
	"github.com/kakaoenterprise/kc-sdk-go/services/image"
)

var (
	_ datasource.DataSource              = &imagesDataSource{}
	_ datasource.DataSourceWithConfigure = &imagesDataSource{}
)

func NewImagesDataSource() datasource.DataSource {
	return &imagesDataSource{}
}

type imagesDataSource struct {
	kc *common.KakaoCloudClient
}

func (d *imagesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *imagesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_images"
}

func (d *imagesDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: docs.GetDataSourceDescription("Images"),
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
			"images": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: MergeDataSourceSchemaAttributes(
						map[string]schema.Attribute{
							"id": schema.StringAttribute{
								Computed:    true,
								Description: "Image ID",
							},
						},
						imageDataSourceSchemaAttributes,
					),
				},
			},
			"timeouts": timeouts.Attributes(ctx),
		},
	}
}

func (d *imagesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config imagesDataSourceModel

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

	imageApi := d.kc.ApiClient.ImageAPI.ListImages(ctx)

	for _, f := range config.Filter {
		if f.Name.IsNull() || f.Name.IsUnknown() {
			continue
		}

		filterName := f.Name.ValueString()

		if !f.Value.IsNull() && !f.Value.IsUnknown() {
			v := f.Value.ValueString()

			switch filterName {
			case "id":
				imageApi = imageApi.Id(v)
			case "name":
				imageApi = imageApi.Name(v)
			case "image_type":
				if imageType, err := ToImageType(v); err == nil {
					imageApi = imageApi.ImageType(*imageType)
				} else {
					resp.Diagnostics.AddError(
						"Invalid image_type",
						err.Error(),
					)
				}
			case "instance_type":
				if instanceType, err := ToInstanceType(v); err == nil {
					imageApi = imageApi.InstanceType(*instanceType)
				} else {
					resp.Diagnostics.AddError(
						"Invalid instance_type",
						err.Error(),
					)
				}
			case "size":
				i, err := strconv.ParseInt(v, 10, 64)
				if err != nil {
					resp.Diagnostics.AddError("Invalid size value",
						fmt.Sprintf("failed to parse size %q (error: %v)", v, err))
					return
				}
				imageApi = imageApi.Size(i)
			case "min_disk":
				if i, err := strconv.ParseInt(v, 10, 64); err == nil {
					imageApi = imageApi.MinDisk(int32(i))
				} else {
					resp.Diagnostics.AddError(
						"Invalid min_disk value",
						fmt.Sprintf("expected int32 but got %q (error: %s)", v, err),
					)
				}
			case "disk_format":
				imageApi = imageApi.DiskFormat(v)
			case "status":
				imageApi = imageApi.Status(v)
			case "os_type":
				imageApi = imageApi.OsType(v)
			case "visibility":
				imageApi = imageApi.Visibility(v)
			case "image_member_status":
				imageApi = imageApi.ImageMemberStatus(v)
			case "created_at":
				if err := common.ValidateRFC3339(v); err == nil {
					imageApi = imageApi.CreatedAt(v)
				} else {
					resp.Diagnostics.AddError(
						"Invalid created_at value",
						err.Error(),
					)
				}
			case "updated_at":
				if err := common.ValidateRFC3339(v); err == nil {
					imageApi = imageApi.UpdatedAt(v)
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

	imageResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, d.kc, &resp.Diagnostics,
		func() (*image.ImageListModel, *http.Response, error) {
			return imageApi.Limit(1000).XAuthToken(d.kc.XAuthToken).Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, d, httpResp, "ListImages", err, &resp.Diagnostics)
		return
	}

	for _, v := range imageResp.Images {
		var tmpImage imageBaseModel
		ok := d.mapImages(ctx, &tmpImage, &v, &resp.Diagnostics)
		if !ok || resp.Diagnostics.HasError() {
			return
		}

		config.Images = append(config.Images, tmpImage)
	}

	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func ToInstanceType(v string) (*image.ImageInstanceType, error) {
	instanceType := image.ImageInstanceType(v)

	for _, allowed := range image.AllowedImageInstanceTypeEnumValues {
		if instanceType == allowed {
			return &instanceType, nil
		}
	}

	return nil, fmt.Errorf("instance type '%s' is not allowed (allowed: %v)", v, image.AllowedImageInstanceTypeEnumValues)
}

func ToImageType(v string) (*image.ImageVisibilityType, error) {
	imageType := image.ImageVisibilityType(strings.ToLower(v))

	for _, allowed := range image.AllowedImageVisibilityTypeEnumValues {
		if imageType == allowed {
			return &imageType, nil
		}
	}

	return nil, fmt.Errorf("image type '%s' is not allowed (allowed: %v)", v, image.AllowedImageVisibilityTypeEnumValues)
}

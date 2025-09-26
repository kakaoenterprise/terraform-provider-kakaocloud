// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package kubernetesengine

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"terraform-provider-kakaocloud/internal/common"
	. "terraform-provider-kakaocloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/kakaoenterprise/kc-sdk-go/services/image"
	"github.com/kakaoenterprise/kc-sdk-go/services/kubernetesengine"
)

var (
	_ datasource.DataSource              = &kubernetesImagesDataSource{}
	_ datasource.DataSourceWithConfigure = &kubernetesImagesDataSource{}
)

func NewKubernetesImagesDataSource() datasource.DataSource { return &kubernetesImagesDataSource{} }

type kubernetesImagesDataSource struct {
	kc *common.KakaoCloudClient
}

func (d *kubernetesImagesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *kubernetesImagesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kubernetes_engine_images"
}

func (d *kubernetesImagesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
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
						kubernetesImageDataSourceSchemaAttributes,
					),
				},
			},
			"timeouts": timeouts.Attributes(ctx),
		},
	}
}

func (d *kubernetesImagesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
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

	imageApi := d.kc.ApiClient.ImagesAPI.ListNodePoolImages(ctx)

	for _, f := range config.Filter {
		if f.Name.IsNull() || f.Name.IsUnknown() {
			continue
		}

		filterName := f.Name.ValueString()

		if !f.Value.IsNull() && !f.Value.IsUnknown() {
			v := f.Value.ValueString()
			v = url.QueryEscape(v)
			switch filterName {
			case "os_distro":
				imageApi = imageApi.OsDistro(v)
			case "instance_type":
				if instanceType, err := ToInstanceType(v); err == nil {
					imageApi = imageApi.InstanceType(*instanceType)
				} else {
					common.AddInvalidParamEnum(&resp.Diagnostics, "instance_type", image.AllowedImageInstanceTypeEnumValues)
				}
			case "is_gpu_type":
				if b, err := strconv.ParseBool(v); err == nil {
					imageApi = imageApi.IsGpuType(b)
				} else {
					common.AddInvalidParamType(&resp.Diagnostics, "is_gpu_type", "Boolean", v)
				}
			case "k8s_version":
				imageApi = imageApi.K8sVersion(v)
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
		func() (*kubernetesengine.GetK8sImagesResponseModel, *http.Response, error) {
			return imageApi.XAuthToken(d.kc.XAuthToken).Execute()
		},
	)

	if err != nil {
		common.AddApiActionError(ctx, d, httpResp, "ListNodePoolImages", err, &resp.Diagnostics)
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

func ToInstanceType(v string) (*kubernetesengine.ImageInstanceType, error) {
	instanceType := kubernetesengine.ImageInstanceType(strings.ToLower(v))

	for _, allowed := range kubernetesengine.AllowedImageInstanceTypeEnumValues {
		if instanceType == allowed {
			return &instanceType, nil
		}
	}

	return nil, fmt.Errorf("instance type '%s' is not allowed (allowed: %v)", v, image.AllowedImageInstanceTypeEnumValues)
}

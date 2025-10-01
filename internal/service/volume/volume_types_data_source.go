// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package volume

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-kakaocloud/internal/common"
	"terraform-provider-kakaocloud/internal/docs"
	. "terraform-provider-kakaocloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kakaoenterprise/kc-sdk-go/services/volume"
)

var (
	_ datasource.DataSource              = &volumeTypesDataSource{}
	_ datasource.DataSourceWithConfigure = &volumeTypesDataSource{}
)

func NewVolumeTypesDataSource() datasource.DataSource {
	return &volumeTypesDataSource{}
}

type volumeTypesDataSource struct {
	kc *common.KakaoCloudClient
}

func (d *volumeTypesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_volume_type"
}

func (d *volumeTypesDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	desc := docs.Volume("VolumeTypeModel")
	resp.Schema = schema.Schema{
		Description: docs.GetDataSourceDescription("VolumeType"),
		Attributes: map[string]schema.Attribute{
			"volume_types": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: desc.String("id"),
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: desc.String("name"),
						},
						"description": schema.StringAttribute{
							Computed:    true,
							Description: desc.String("description"),
						},
					},
				},
			},
			"timeouts": timeouts.Attributes(ctx),
		},
	}
}

func (d *volumeTypesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config volumeTypesDataSourceModel

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

	respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, d.kc, &resp.Diagnostics,
		func() (*volume.VolumeTypeListModel, *http.Response, error) {
			return d.kc.ApiClient.VolumeAPI.ListVolumeTypes(ctx).XAuthToken(d.kc.XAuthToken).Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, d, httpResp, "ListVolumeTypes", err, &resp.Diagnostics)
		return
	}

	for _, v := range respModel.VolumeTypes {
		config.VolumeTypes = append(config.VolumeTypes, volumeTypeModel{
			Id:          types.StringValue(v.Id),
			Name:        types.StringValue(v.Name),
			Description: ConvertNullableString(v.Description),
		})
	}

	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (d *volumeTypesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {

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

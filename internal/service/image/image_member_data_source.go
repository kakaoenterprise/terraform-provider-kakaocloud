// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package image

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-kakaocloud/internal/common"
	"terraform-provider-kakaocloud/internal/docs"
	"terraform-provider-kakaocloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/kakaoenterprise/kc-sdk-go/services/image"
)

var (
	_ datasource.DataSource              = &imageMemberDataSource{}
	_ datasource.DataSourceWithConfigure = &imageMemberDataSource{}
)

func NewImageMemberDataSource() datasource.DataSource {
	return &imageMemberDataSource{}
}

type imageMemberDataSource struct {
	kc *common.KakaoCloudClient
}

func (d *imageMemberDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *imageMemberDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_image_member"
}

func (d *imageMemberDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: docs.GetDataSourceDescription("ImageMember"),
		Attributes: utils.MergeAttributes[schema.Attribute](
			imageMemberDataSourceSchema,
			map[string]schema.Attribute{
				"timeouts": timeouts.Attributes(ctx),
			},
		),
	}
}

func (d *imageMemberDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config imageMemberDataSourceModel

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

	imageMemberResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, d.kc, &resp.Diagnostics,
		func() (*image.ImageMemberListModel, *http.Response, error) {
			return d.kc.ApiClient.ImageAPI.
				ListImageSharedProjects(ctx, config.Id.ValueString()).
				XAuthToken(d.kc.XAuthToken).
				Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, d, httpResp, "ListImageSharedProjects", err, &resp.Diagnostics)
		return
	}

	ok := mapImageMemberModel(ctx, &config.imageMemberBaseModel, imageMemberResp.Members, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)

}

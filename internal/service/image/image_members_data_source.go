// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package image

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-kakaocloud/internal/common"
	"terraform-provider-kakaocloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/kakaoenterprise/kc-sdk-go/services/image"
)

var (
	_ datasource.DataSource              = &imageMembersDataSource{}
	_ datasource.DataSourceWithConfigure = &imageMembersDataSource{}
)

func NewImageMembersDataSource() datasource.DataSource {
	return &imageMembersDataSource{}
}

type imageMembersDataSource struct {
	kc *common.KakaoCloudClient
}

func (d *imageMembersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *imageMembersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_image_members"
}

func (d *imageMembersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "KakaoCloud Image Member Schema",
		Attributes: map[string]schema.Attribute{
			"image_id": schema.StringAttribute{
				Required:   true,
				Validators: common.UuidValidator(),
			},
			"members": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: utils.MergeAttributes[schema.Attribute](
						map[string]schema.Attribute{
							"id": schema.StringAttribute{
								Computed:    true,
								Description: "Image Member ID",
							},
						},
						imageMemberDataSourceSchemaAttributes,
					),
				},
			},
			"timeouts": timeouts.Attributes(ctx),
		},
	}
}

func (d *imageMembersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
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
				ListImageSharedProjects(ctx, config.ImageId.ValueString()).
				XAuthToken(d.kc.XAuthToken).
				Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, d, httpResp, "ListImageSharedProjects", err, &resp.Diagnostics)
		return
	}

	for _, v := range imageMemberResp.Members {
		var tmpImageMember imageMemberBaseModel
		ok := mapImageMemberModel(ctx, &tmpImageMember, &v, &resp.Diagnostics)
		if !ok || resp.Diagnostics.HasError() {
			return
		}

		config.Members = append(config.Members, tmpImageMember)
	}

	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)

}

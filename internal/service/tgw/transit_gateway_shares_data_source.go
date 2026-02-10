// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package tgw

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-kakaocloud/internal/common"
	"terraform-provider-kakaocloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kakaoenterprise/kc-sdk-go/services/tgw"
)

var (
	_ datasource.DataSource              = &transitGatewaySharedProjectsDataSource{}
	_ datasource.DataSourceWithConfigure = &transitGatewaySharedProjectsDataSource{}
)

func NewTransitGatewaySharedProjectsDataSource() datasource.DataSource {
	return &transitGatewaySharedProjectsDataSource{}
}

type transitGatewaySharedProjectsDataSource struct {
	kc *common.KakaoCloudClient
}

func (d *transitGatewaySharedProjectsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_transit_gateway_shares"
}

func (d *transitGatewaySharedProjectsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: utils.MergeDataSourceSchemaAttributes(
			map[string]schema.Attribute{
				"timeouts": timeouts.Attributes(ctx),
			},
			transitGatewaySharedProjectsDataSourceSchemaAttributes,
		),
	}
}

func (d *transitGatewaySharedProjectsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config transitGatewaySharedProjectsDataSourceModel

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
		func() (*tgw.GetTgwProjectsResponseModel, *http.Response, error) {
			return d.kc.ApiClient.TgwsAPI.ListTgwSharedProjects(ctx, config.TgwId.ValueString()).
				XAuthToken(d.kc.XAuthToken).
				Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, d, httpResp, "ListTgwSharedProjects", err, &resp.Diagnostics)
		return
	}

	config.SharedProjects = make([]sharedProjectModel, 0)
	for _, project := range respModel.Projects {
		projectModel := sharedProjectModel{
			Id:          types.StringValue(project.Id),
			Name:        types.StringValue(project.Name),
			Nickname:    types.StringValue(project.Nickname),
			Description: types.StringValue(project.Description),
			DomainId:    types.StringValue(project.DomainId),
			IsEnabled:   utils.ConvertNullableBool(project.IsEnabled),
			CreatedAt:   utils.ConvertNullableTime(project.CreatedAt),
			DisabledAt:  utils.ConvertNullableTime(project.DisabledAt),
		}

		config.SharedProjects = append(config.SharedProjects, projectModel)
	}

	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (d *transitGatewaySharedProjectsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*common.KakaoCloudClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *common.KakaoCloudClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.kc = client
}

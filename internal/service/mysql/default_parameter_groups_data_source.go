// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package mysql

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-kakaocloud/internal/common"
	"terraform-provider-kakaocloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	mysqlsdk "github.com/kakaoenterprise/kc-sdk-go/services/mysql"
)

var (
	_ datasource.DataSource              = &defaultParameterGroupsDataSource{}
	_ datasource.DataSourceWithConfigure = &defaultParameterGroupsDataSource{}
)

func NewDefaultParameterGroupsDataSource() datasource.DataSource {
	return &defaultParameterGroupsDataSource{}
}

type defaultParameterGroupsDataSource struct{ kc *common.KakaoCloudClient }

func (d *defaultParameterGroupsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mysql_default_parameter_groups"
}

func (d *defaultParameterGroupsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"default_parameter_groups": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: defaultParameterGroupListSchemaAttributes,
				},
			},
			"timeouts": timeouts.Attributes(ctx),
		},
	}
}

func (d *defaultParameterGroupsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data defaultParameterGroupsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	timeout, diags := data.Timeouts.Read(ctx, common.DefaultReadTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	modelResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, d.kc, &resp.Diagnostics,
		func() (*mysqlsdk.GetMySQLDefaultParameterGroupsResponseModel, *http.Response, error) {
			request := d.kc.ApiClient.MySQLDefaultParameterGroupsAPI.
				ListMysqlDefaultParameterGroups(ctx).
				XAuthToken(d.kc.XAuthToken).
				ShowInstanceGroupsInfo(true)
			return request.Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, d, httpResp, "ListMysqlDefaultParameterGroups", err, &resp.Diagnostics)
		return
	}

	items := make([]defaultParameterGroupListModel, 0, len(modelResp.DefaultParameterGroups))
	for _, item := range modelResp.DefaultParameterGroups {
		mapped, ok := toDefaultParameterGroupListModel(item)
		if !ok {
			return
		}
		items = append(items, mapped)
	}
	value, diags := utils.ConvertListFromModel(ctx, items, defaultParameterGroupListAttrTypes, func(item defaultParameterGroupListModel) any {
		return item
	})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.DefaultParameterGroups = value
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *defaultParameterGroupsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*common.KakaoCloudClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", fmt.Sprintf("Expected *common.KakaoCloudClient, got: %T. Please report this issue to the provider developers.", req.ProviderData))
		return
	}
	d.kc = client
}

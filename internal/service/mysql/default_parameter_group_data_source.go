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
	_ datasource.DataSource              = &defaultParameterGroupDataSource{}
	_ datasource.DataSourceWithConfigure = &defaultParameterGroupDataSource{}
)

func NewDefaultParameterGroupDataSource() datasource.DataSource {
	return &defaultParameterGroupDataSource{}
}

type defaultParameterGroupDataSource struct{ kc *common.KakaoCloudClient }

func (d *defaultParameterGroupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mysql_default_parameter_group"
}

func (d *defaultParameterGroupDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: utils.MergeDataSourceSchemaAttributes(defaultParameterGroupSingleSchemaAttributes, map[string]schema.Attribute{
			"id":       schema.StringAttribute{Required: true, Validators: common.UuidValidator()},
			"timeouts": timeouts.Attributes(ctx),
		}),
	}
}

func (d *defaultParameterGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data defaultParameterGroupDataSourceModel
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
		func() (*mysqlsdk.GetMySQLDefaultParameterGroupResponseModel, *http.Response, error) {
			return d.kc.ApiClient.MySQLDefaultParameterGroupsAPI.
				GetMysqlDefaultParameterGroup(ctx, data.Id.ValueString()).
				XAuthToken(d.kc.XAuthToken).
				Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, d, httpResp, "GetMysqlDefaultParameterGroup", err, &resp.Diagnostics)
		return
	}

	item, ok := toDefaultParameterGroupSingleModel(ctx, modelResp.DefaultParameterGroup, &resp.Diagnostics)
	if !ok {
		return
	}
	data.defaultParameterGroupSingleModel = item
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *defaultParameterGroupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

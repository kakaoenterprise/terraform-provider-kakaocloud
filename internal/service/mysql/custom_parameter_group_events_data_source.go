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
	_ datasource.DataSource              = &customParameterGroupEventsDataSource{}
	_ datasource.DataSourceWithConfigure = &customParameterGroupEventsDataSource{}
)

func NewCustomParameterGroupEventsDataSource() datasource.DataSource {
	return &customParameterGroupEventsDataSource{}
}

type customParameterGroupEventsDataSource struct{ kc *common.KakaoCloudClient }

func (d *customParameterGroupEventsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mysql_custom_parameter_group_events"
}

func (d *customParameterGroupEventsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"custom_parameter_group_id": schema.StringAttribute{Required: true, Validators: common.UuidValidator()},
			"events": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: mysqlParameterGroupEventSchemaAttributes,
				},
			},
			"timeouts": timeouts.Attributes(ctx),
		},
	}
}

func (d *customParameterGroupEventsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data customParameterGroupEventsDataSourceModel
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
		func() (*mysqlsdk.GetMySQLCustomParameterGroupEventsResponseModel, *http.Response, error) {
			return d.kc.ApiClient.MySQLCustomParameterGroupsAPI.
				ListMysqlCustomParameterGroupEvents(ctx, data.CustomParameterGroupId.ValueString()).
				XAuthToken(d.kc.XAuthToken).
				Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, d, httpResp, "ListMysqlCustomParameterGroupEvents", err, &resp.Diagnostics)
		return
	}

	items := make([]mysqlParameterGroupEventModel, 0, len(modelResp.Events))
	for _, item := range modelResp.Events {
		items = append(items, toCustomParameterGroupEventModel(item))
	}

	value, diags := utils.ConvertListFromModel(ctx, items, mysqlParameterGroupEventAttrTypes, func(item mysqlParameterGroupEventModel) any {
		return item
	})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Events = value
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *customParameterGroupEventsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

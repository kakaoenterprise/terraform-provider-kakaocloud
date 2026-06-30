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
	_ datasource.DataSource              = &defaultParameterGroupEventsDataSource{}
	_ datasource.DataSourceWithConfigure = &defaultParameterGroupEventsDataSource{}
)

func NewDefaultParameterGroupEventsDataSource() datasource.DataSource {
	return &defaultParameterGroupEventsDataSource{}
}

type defaultParameterGroupEventsDataSource struct{ kc *common.KakaoCloudClient }

func (d *defaultParameterGroupEventsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mysql_default_parameter_group_events"
}

func (d *defaultParameterGroupEventsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"default_parameter_group_id": schema.StringAttribute{Required: true, Validators: common.UuidValidator()},
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

func (d *defaultParameterGroupEventsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data defaultParameterGroupEventsDataSourceModel
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
		func() (*mysqlsdk.GetMySQLDefaultParameterGroupEventsResponseModel, *http.Response, error) {
			return d.kc.ApiClient.MySQLDefaultParameterGroupsAPI.
				ListMysqlDefaultParameterGroupEvents(ctx, data.DefaultParameterGroupId.ValueString()).
				XAuthToken(d.kc.XAuthToken).
				Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, d, httpResp, "ListMysqlDefaultParameterGroupEvents", err, &resp.Diagnostics)
		return
	}

	items := make([]mysqlParameterGroupEventModel, 0, len(modelResp.Events))
	for _, item := range modelResp.Events {
		items = append(items, toDefaultParameterGroupEventModel(item))
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

func (d *defaultParameterGroupEventsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

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
	_ datasource.DataSource              = &instanceGroupsUsingDefaultParameterGroupDataSource{}
	_ datasource.DataSourceWithConfigure = &instanceGroupsUsingDefaultParameterGroupDataSource{}
)

func NewInstanceGroupsUsingDefaultParameterGroupDataSource() datasource.DataSource {
	return &instanceGroupsUsingDefaultParameterGroupDataSource{}
}

type instanceGroupsUsingDefaultParameterGroupDataSource struct{ kc *common.KakaoCloudClient }

func (d *instanceGroupsUsingDefaultParameterGroupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mysql_instance_groups_using_default_parameter_group"
}

func (d *instanceGroupsUsingDefaultParameterGroupDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"default_parameter_group_id": schema.StringAttribute{Required: true, Validators: common.UuidValidator()},
			"instance_groups": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: mysqlParameterGroupInstanceGroupSchemaAttributes,
				},
			},
			"timeouts": timeouts.Attributes(ctx),
		},
	}
}

func (d *instanceGroupsUsingDefaultParameterGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data instanceGroupsUsingDefaultParameterGroupDataSourceModel
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
		func() (*mysqlsdk.MysqlV1ApiListMysqlInstanceGroupsUsingDefaultParameterGroupModelGetMySQLInstanceGroupsUsingDefaultParameterGroupResponseModel, *http.Response, error) {
			return d.kc.ApiClient.MySQLDefaultParameterGroupsAPI.
				ListMysqlInstanceGroupsUsingDefaultParameterGroup(ctx, data.DefaultParameterGroupId.ValueString()).
				XAuthToken(d.kc.XAuthToken).
				Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, d, httpResp, "ListMysqlInstanceGroupsUsingDefaultParameterGroup", err, &resp.Diagnostics)
		return
	}

	items := make([]mysqlParameterGroupInstanceGroupModel, 0, len(modelResp.InstanceGroups))
	for _, item := range modelResp.InstanceGroups {
		items = append(items, toParameterGroupInstanceGroupModel(item))
	}

	value, diags := utils.ConvertListFromModel(ctx, items, mysqlParameterGroupInstanceGroupAttrTypes, func(item mysqlParameterGroupInstanceGroupModel) any {
		return item
	})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.InstanceGroups = value
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *instanceGroupsUsingDefaultParameterGroupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

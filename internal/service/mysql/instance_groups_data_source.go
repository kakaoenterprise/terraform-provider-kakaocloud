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
	_ datasource.DataSource              = &instanceGroupsDataSource{}
	_ datasource.DataSourceWithConfigure = &instanceGroupsDataSource{}
)

func NewInstanceGroupsDataSource() datasource.DataSource { return &instanceGroupsDataSource{} }

type instanceGroupsDataSource struct {
	kc *common.KakaoCloudClient
}

func (d *instanceGroupsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mysql_instance_groups"
}

func (d *instanceGroupsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"instance_groups": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: instanceGroupListSchemaAttributes,
				},
			},
			"timeouts": timeouts.Attributes(ctx),
		},
	}
}

func (d *instanceGroupsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data instanceGroupsDataSourceModel

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
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
		func() (*mysqlsdk.GetMySQLInstanceGroupsResponseModel, *http.Response, error) {
			return d.kc.ApiClient.MySQLInstanceGroupsAPI.
				ListMysqlInstanceGroups(ctx).
				XAuthToken(d.kc.XAuthToken).
				Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, d, httpResp, "ListMysqlInstanceGroups", err, &resp.Diagnostics)
		return
	}

	items := make([]instanceGroupListModel, 0, len(modelResp.InstanceGroups))
	for _, item := range modelResp.InstanceGroups {
		instanceGroupModel, ok := toInstanceGroupModelFromList(ctx, item, &resp.Diagnostics)
		if !ok {
			return
		}
		items = append(items, instanceGroupModel)
	}

	instanceGroups, diags := utils.ConvertListFromModel(ctx, items, instanceGroupListAttrTypes, func(item instanceGroupListModel) any {
		return item
	})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.InstanceGroups = instanceGroups

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (d *instanceGroupsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

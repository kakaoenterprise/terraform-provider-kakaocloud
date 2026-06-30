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
	_ datasource.DataSource              = &backupsDataSource{}
	_ datasource.DataSourceWithConfigure = &backupsDataSource{}
)

func NewBackupsDataSource() datasource.DataSource { return &backupsDataSource{} }

type backupsDataSource struct {
	kc *common.KakaoCloudClient
}

func (d *backupsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mysql_backups"
}

func (d *backupsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"instance_group_id": schema.StringAttribute{
				Optional:   true,
				Validators: common.UuidValidator(),
			},
			"backups": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: backupSchemaAttributes,
				},
			},
			"timeouts": timeouts.Attributes(ctx),
		},
	}
}

func (d *backupsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data backupsDataSourceModel

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
		func() (*mysqlsdk.GetMySQLBackupsResponseModel, *http.Response, error) {
			request := d.kc.ApiClient.MySQLBackupsAPI.
				ListMysqlBackups(ctx).
				XAuthToken(d.kc.XAuthToken)
			if !data.InstanceGroupId.IsNull() && !data.InstanceGroupId.IsUnknown() {
				request = request.InstanceGroupId(data.InstanceGroupId.ValueString())
			}
			return request.Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, d, httpResp, "ListMysqlBackups", err, &resp.Diagnostics)
		return
	}

	items := make([]backupModel, 0, len(modelResp.Backups))
	for _, item := range modelResp.Backups {
		backupModel, ok := toBackupModelFromList(ctx, item, &resp.Diagnostics)
		if !ok {
			return
		}
		items = append(items, backupModel)
	}

	backups, diags := utils.ConvertListFromModel(ctx, items, backupAttrTypes, func(item backupModel) any {
		return item
	})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Backups = backups

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *backupsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

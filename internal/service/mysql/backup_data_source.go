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
	_ datasource.DataSource              = &backupDataSource{}
	_ datasource.DataSourceWithConfigure = &backupDataSource{}
)

func NewBackupDataSource() datasource.DataSource { return &backupDataSource{} }

type backupDataSource struct {
	kc *common.KakaoCloudClient
}

func (d *backupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mysql_backup"
}

func (d *backupDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: utils.MergeDataSourceSchemaAttributes(backupSchemaAttributes, map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:   true,
				Validators: common.UuidValidator(),
			},
			"timeouts": timeouts.Attributes(ctx),
		}),
	}
}

func (d *backupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data backupDataSourceModel

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
		func() (*mysqlsdk.GetMySQLBackupResponseModel, *http.Response, error) {
			return d.kc.ApiClient.MySQLBackupsAPI.
				GetMysqlBackup(ctx, data.Id.ValueString()).
				XAuthToken(d.kc.XAuthToken).
				Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, d, httpResp, "GetMysqlBackup", err, &resp.Diagnostics)
		return
	}

	backupModel, ok := toBackupModelFromGet(ctx, modelResp.Backup, &resp.Diagnostics)
	if !ok {
		return
	}

	data.backupModel = backupModel

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *backupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

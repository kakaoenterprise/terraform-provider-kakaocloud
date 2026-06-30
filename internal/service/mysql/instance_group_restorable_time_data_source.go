// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package mysql

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-kakaocloud/internal/common"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	mysqlsdk "github.com/kakaoenterprise/kc-sdk-go/services/mysql"
)

var (
	_ datasource.DataSource              = &instanceGroupRestorableTimeDataSource{}
	_ datasource.DataSourceWithConfigure = &instanceGroupRestorableTimeDataSource{}
)

func NewInstanceGroupRestorableTimeDataSource() datasource.DataSource {
	return &instanceGroupRestorableTimeDataSource{}
}

type instanceGroupRestorableTimeDataSource struct {
	kc *common.KakaoCloudClient
}

func (d *instanceGroupRestorableTimeDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mysql_instance_group_restorable_time"
}

func (d *instanceGroupRestorableTimeDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"instance_group_id": schema.StringAttribute{
				Required:   true,
				Validators: common.UuidValidator(),
			},
			"restorable_time": schema.SingleNestedAttribute{
				Computed:   true,
				Attributes: restorableTimeSchemaAttributes,
			},
			"timeouts": timeouts.Attributes(ctx),
		},
	}
}

func (d *instanceGroupRestorableTimeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data instanceGroupRestorableTimeDataSourceModel

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
		func() (*mysqlsdk.GetMySQLInstanceGroupRestorableTimeResponseModel, *http.Response, error) {
			return d.kc.ApiClient.MySQLInstanceGroupsAPI.
				GetMysqlRestorableTime(ctx, data.InstanceGroupId.ValueString()).
				XAuthToken(d.kc.XAuthToken).
				Execute()
		},
	)
	if err != nil {

		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound && err.Error() == "not found restorable time" {
			data.RestorableTime = types.ObjectNull(restorableTimeAttrTypes)
			resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
			return
		}
		common.AddApiActionError(ctx, d, httpResp, "GetMysqlRestorableTime", err, &resp.Diagnostics)
		return
	}

	restorableTime, diags := types.ObjectValueFrom(ctx, restorableTimeAttrTypes, restorableTimeModel{
		FromTime: types.StringValue(modelResp.RestorableTime.FromTime),
		ToTime:   types.StringValue(modelResp.RestorableTime.ToTime),
	})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.RestorableTime = restorableTime

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *instanceGroupRestorableTimeDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

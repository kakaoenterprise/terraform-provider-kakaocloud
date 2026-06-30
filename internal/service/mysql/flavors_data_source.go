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
	"github.com/hashicorp/terraform-plugin-framework/types"
	mysqlsdk "github.com/kakaoenterprise/kc-sdk-go/services/mysql"
)

var (
	_ datasource.DataSource              = &flavorsDataSource{}
	_ datasource.DataSourceWithConfigure = &flavorsDataSource{}
)

func NewFlavorsDataSource() datasource.DataSource { return &flavorsDataSource{} }

type flavorsDataSource struct {
	kc *common.KakaoCloudClient
}

func (d *flavorsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mysql_flavors"
}

func (d *flavorsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"show_all": schema.BoolAttribute{
				Optional:    true,
				Description: "Whether to include all flavors, including deprecated flavors.",
			},
			"flavors": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: flavorSchemaAttributes,
				},
			},
			"timeouts": timeouts.Attributes(ctx),
		},
	}
}

func (d *flavorsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data flavorsDataSourceModel

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
		func() (*mysqlsdk.GetMySQLFlavorsResponseModel, *http.Response, error) {
			req := d.kc.ApiClient.MySQLFlavorsAPI.
				ListMysqlInstanceTypesFlavors(ctx).
				XAuthToken(d.kc.XAuthToken)
			if !data.ShowAll.IsNull() && !data.ShowAll.IsUnknown() {
				req = req.ShowAll(data.ShowAll.ValueBool())
			}
			return req.Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, d, httpResp, "ListMysqlInstanceTypesFlavors", err, &resp.Diagnostics)
		return
	}

	for _, flavor := range modelResp.Flavors {
		data.Flavors = append(data.Flavors, toFlavorModel(flavor))
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (d *flavorsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func toFlavorModel(flavor mysqlsdk.FlavorResponseModel) flavorModel {
	availabilityZones, _ := utils.ListFromStrings(context.Background(), flavor.AvailabilityZones)

	return flavorModel{
		Id:                types.StringValue(flavor.Id),
		Name:              types.StringValue(flavor.Name),
		Type:              types.StringValue(flavor.Type),
		Vcpus:             types.Int32Value(flavor.Vcpus),
		Memory:            types.Int32Value(flavor.Memory),
		MemoryMb:          types.Int32Value(flavor.MemoryMb),
		Group:             types.StringValue(flavor.Group),
		Family:            types.StringValue(flavor.Family),
		AvailabilityZones: availabilityZones,
		Deprecated:        types.BoolValue(flavor.Deprecated),
	}
}

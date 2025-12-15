// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package vpc

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-kakaocloud/internal/common"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kakaoenterprise/kc-sdk-go/services/vpc"
)

var (
	_ datasource.DataSource              = &subnetShareDataSource{}
	_ datasource.DataSourceWithConfigure = &subnetShareDataSource{}
)

func NewSubnetShareDataSource() datasource.DataSource {
	return &subnetShareDataSource{}
}

type subnetShareDataSource struct {
	kc *common.KakaoCloudClient
}

func (d *subnetShareDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subnet_share"
}

func (d *subnetShareDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:   true,
				Validators: common.UuidValidator(),
			},
			"projects": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
			},
			"timeouts": timeouts.Attributes(ctx),
		},
	}
}

func (d *subnetShareDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config subnetShareDataSourceModel

	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	timeout, diags := config.Timeouts.Read(ctx, common.DefaultReadTimeout)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	subnetShareResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, d.kc, &resp.Diagnostics,
		func() (*vpc.ResponseSubnetSharedProjectListModel, *http.Response, error) {
			return d.kc.ApiClient.VPCSubnetAPI.ListSubnetSharedProjects(ctx, config.Id.ValueString()).XAuthToken(d.kc.XAuthToken).Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, d, httpResp, "ListSubnetSharedProjects", err, &resp.Diagnostics)
		return
	}

	projectIDs := make([]attr.Value, 0, len(subnetShareResp.Projects))
	for _, p := range subnetShareResp.Projects {
		projectIDs = append(projectIDs, types.StringValue(p.Id))
	}

	config.Projects, diags = types.ListValue(types.StringType, projectIDs)
	resp.Diagnostics.Append(diags...)

	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (d *subnetShareDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {

	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*common.KakaoCloudClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *kakaocloud.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.kc = client
}

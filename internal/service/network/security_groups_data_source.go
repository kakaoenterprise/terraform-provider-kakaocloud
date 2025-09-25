// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package network

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-kakaocloud/internal/common"
	. "terraform-provider-kakaocloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/kakaoenterprise/kc-sdk-go/services/network"
)

var (
	_ datasource.DataSource              = &securityGroupsDataSource{}
	_ datasource.DataSourceWithConfigure = &securityGroupsDataSource{}
)

func NewSecurityGroupsDataSource() datasource.DataSource {
	return &securityGroupsDataSource{}
}

type securityGroupsDataSource struct {
	kc *common.KakaoCloudClient
}

func (d *securityGroupsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *securityGroupsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_security_groups"
}

func (d *securityGroupsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Data source to list KakaoCloud Security Groups with optional filters.",
		Attributes: map[string]schema.Attribute{
			"filter": schema.ListNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required: true,
						},
						"value": schema.StringAttribute{
							Optional: true,
						},
					},
				},
			},
			"security_groups": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: MergeAttributes[schema.Attribute](
						map[string]schema.Attribute{
							"id": schema.StringAttribute{
								Required:    true,
								Description: "Security Group ID",
							},
						},
						securityGroupDataSourceSchemaAttributes,
					),
				},
			},
			"timeouts": timeouts.Attributes(ctx),
		},
	}
}

func (d *securityGroupsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config securityGroupsDataSourceModel

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

	sgApi := d.kc.ApiClient.SecurityGroupAPI.ListSecurityGroups(ctx)

	for _, f := range config.Filter {
		if f.Name.IsNull() || f.Name.IsUnknown() || f.Value.IsNull() || f.Value.IsUnknown() {
			continue
		}

		filterName := f.Name.ValueString()
		v := f.Value.ValueString()

		switch filterName {
		case "id":
			sgApi = sgApi.Id(v)
		case "name":
			sgApi = sgApi.Name(v)
		case "created_at":
			if err := common.ValidateRFC3339(v); err == nil {
				sgApi = sgApi.CreatedAt(v)
			} else {
				resp.Diagnostics.AddError("Invalid created_at value", err.Error())
			}
		case "updated_at":
			if err := common.ValidateRFC3339(v); err == nil {
				sgApi = sgApi.UpdatedAt(v)
			} else {
				resp.Diagnostics.AddError("Invalid updated_at value", err.Error())
			}
		default:
			resp.Diagnostics.AddWarning(
				"Invalid filter name",
				fmt.Sprintf("filter %q is not supported for security groups", filterName),
			)
		}
	}

	if resp.Diagnostics.HasError() {
		return
	}

	sgResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, d.kc, &resp.Diagnostics,
		func() (*network.SecurityGroupListModel, *http.Response, error) {
			return sgApi.Limit(1000).XAuthToken(d.kc.XAuthToken).Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, d, httpResp, "ListSecurityGroups", err, &resp.Diagnostics)
		return
	}

	var resultSgs []securityGroupBaseModel
	for i := range sgResp.SecurityGroups {
		var base securityGroupBaseModel
		mapSecurityGroupBaseModelFromList(ctx, &base, &sgResp.SecurityGroups[i], &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
		resultSgs = append(resultSgs, base)
	}

	config.SecurityGroups = resultSgs

	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
}

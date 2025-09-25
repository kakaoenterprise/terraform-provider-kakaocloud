// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bcs

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-kakaocloud/internal/common"
	. "terraform-provider-kakaocloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jinzhu/copier"
	"github.com/kakaoenterprise/kc-sdk-go/services/bcs"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &keypairsDataSource{}
	_ datasource.DataSourceWithConfigure = &keypairsDataSource{}
)

func NewKeypairsDataSource() datasource.DataSource {
	return &keypairsDataSource{}
}

// keypairsDataSource is the data source implementation.
type keypairsDataSource struct {
	kc *common.KakaoCloudClient
}

func (d *keypairsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_keypairs"
}

func (d *keypairsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "kakaocloud의 모든 키페어 정보를 필터링하여 조회하는 데이터 소스",
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
			"keypairs": schema.ListNestedAttribute{
				Computed:    true,
				Description: "조회된 키페어 목록",
				NestedObject: schema.NestedAttributeObject{
					Attributes: MergeDataSourceSchemaAttributes(
						map[string]schema.Attribute{
							"name": schema.StringAttribute{
								Computed:    true,
								Description: "키페어 이름",
							},
						},
						keypairDataSourceSchemaAttributes,
					),
				},
			},
			"timeouts": timeouts.Attributes(ctx),
		},
	}
}

func (d *keypairsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config keypairsDataSourceModel

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

	keypairApi := d.kc.ApiClient.KeypairAPI.ListKeypairs(ctx)

	for _, f := range config.Filter {
		if f.Name.IsNull() || f.Name.IsUnknown() {
			continue
		}

		filterName := f.Name.ValueString()

		if !f.Value.IsNull() && !f.Value.IsUnknown() {
			v := f.Value.ValueString()

			switch filterName {
			case "id":
				keypairApi = keypairApi.Id(v)
			case "name":
				keypairApi = keypairApi.Name(v)
			case "type":
				keypairApi = keypairApi.Type_(v)
			case "fingerprint":
				keypairApi = keypairApi.Fingerprint(v)
			case "created_at":
				if err := common.ValidateRFC3339(v); err == nil {
					keypairApi = keypairApi.CreatedAt(v)
				} else {
					resp.Diagnostics.AddError(
						"Invalid created_at value",
						err.Error(),
					)
				}
			default:
				resp.Diagnostics.AddError(
					"Invalid filter name",
					fmt.Sprintf("filter %q is not supported", filterName),
				)
			}
		}
	}

	if resp.Diagnostics.HasError() {
		return
	}

	keypairResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, d.kc, &resp.Diagnostics,
		func() (*bcs.KeypairListModel, *http.Response, error) {
			return keypairApi.Limit(1000).XAuthToken(d.kc.XAuthToken).Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, d, httpResp, "ListKeypairs", err, &resp.Diagnostics)
		return
	}

	var keypairResult []bcs.BcsInstanceV1ApiGetKeypairModelKeypairModel
	err = copier.Copy(&keypairResult, &keypairResp.Keypairs)
	if err != nil {
		resp.Diagnostics.AddError("List 변환 실패", fmt.Sprintf("keypairResult 변환 실패: %v", err))
		return
	}

	for _, v := range keypairResult {
		var tmpKeypair keypairBaseModel
		ok := mapKeypairBaseModel(ctx, &tmpKeypair, &v, &resp.Diagnostics)
		if !ok || resp.Diagnostics.HasError() {
			return
		}

		config.Keypairs = append(config.Keypairs, tmpKeypair)

	}

	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (d *keypairsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*common.KakaoCloudClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *common.KakaoCloudClient, got: %T.", req.ProviderData),
		)
		return
	}
	d.kc = client
}

func mapKeypair(
	model *keypairBaseModel,
	keypairResult *bcs.BcsInstanceV1ApiListKeypairsModelKeypairModel,
) {
	model.Id = types.StringValue(keypairResult.GetId())
	model.Name = ConvertNullableString(keypairResult.Name)
	model.Fingerprint = ConvertNullableString(keypairResult.Fingerprint)
	model.PublicKey = ConvertNullableString(keypairResult.PublicKey)
	model.UserId = ConvertNullableString(keypairResult.UserId)
	model.Type = ConvertNullableString(keypairResult.Type)
	model.CreatedAt = ConvertNullableTime(keypairResult.CreatedAt)
}

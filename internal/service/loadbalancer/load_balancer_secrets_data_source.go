// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package loadbalancer

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-kakaocloud/internal/common"

	"github.com/jinzhu/copier"
	"github.com/kakaoenterprise/kc-sdk-go/v2/services/loadbalancer"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

var (
	_ datasource.DataSource              = &loadBalancerSecretsDataSource{}
	_ datasource.DataSourceWithConfigure = &loadBalancerSecretsDataSource{}
)

func NewLoadBalancerSecretsDataSource() datasource.DataSource {
	return &loadBalancerSecretsDataSource{}
}

type loadBalancerSecretsDataSource struct {
	kc *common.KakaoCloudClient
}

func (d *loadBalancerSecretsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *loadBalancerSecretsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_load_balancer_secrets"
}

func (d *loadBalancerSecretsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
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
			"secrets": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: loadBalancerSecretBaseSchemaAttributes,
				},
			},
			"timeouts": timeouts.Attributes(ctx),
		},
	}
}

func (d *loadBalancerSecretsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config loadBalancerSecretsDataSourceModel

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

	lbsApi := d.kc.ApiClient.LoadBalancerAPI.ListTlsCertificates(ctx)

	for _, f := range config.Filter {
		if f.Name.IsNull() || f.Name.IsUnknown() {
			continue
		}

		filterName := f.Name.ValueString()

		if !f.Value.IsNull() && !f.Value.IsUnknown() {
			v := f.Value.ValueString()

			switch filterName {
			case "name":
				lbsApi = lbsApi.Name(v)
			case "created_at":
				if err := common.ValidateRFC3339(v); err == nil {
					lbsApi = lbsApi.CreatedAt(v)
				} else {
					resp.Diagnostics.AddError(
						"Invalid created_at value",
						err.Error(),
					)
				}
			case "updated_at":
				if err := common.ValidateRFC3339(v); err == nil {
					lbsApi = lbsApi.UpdatedAt(v)
				} else {
					resp.Diagnostics.AddError(
						"Invalid updated_at value",
						err.Error(),
					)
				}
			case "expiration":
				lbsApi = lbsApi.Expiration(v)
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

	lbsResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, d.kc, &resp.Diagnostics,
		func() (*loadbalancer.ListTlsCertificatesResponse, *http.Response, error) {
			return lbsApi.Limit(1000).XAuthToken(d.kc.XAuthToken).Execute()
		},
	)

	if err != nil {
		common.AddApiActionError(ctx, d, httpResp, "ListLoadBalancerSecrets", err, &resp.Diagnostics)
		return
	}

	var lbssResult []loadbalancer.ListTlsCertificatesSecret
	err = copier.Copy(&lbssResult, &lbsResp.Secrets)
	if err != nil {
		common.AddGeneralError(ctx, d, &resp.Diagnostics,
			fmt.Sprintf("Failed to convert lbssResult: %v", err))
		return
	}

	config.LoadBalancerSecrets = make([]loadBalancerSecretBaseModel, 0, len(lbssResult))
	for _, secret := range lbssResult {
		var lbsModel loadBalancerSecretBaseModel
		ok := mapLoadBalancerSecretsModel(ctx, &lbsModel, &secret, &resp.Diagnostics)
		if !ok || resp.Diagnostics.HasError() {
			return
		}
		config.LoadBalancerSecrets = append(config.LoadBalancerSecrets, lbsModel)
	}

	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
}

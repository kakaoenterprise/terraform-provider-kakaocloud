// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package loadbalancer

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-kakaocloud/internal/common"

	"github.com/jinzhu/copier"
	"github.com/kakaoenterprise/kc-sdk-go/services/loadbalancer"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

// Ensure the implementation satisfies the expected interfaces
var (
	_ datasource.DataSource              = &loadBalancerSecretsDataSource{}
	_ datasource.DataSourceWithConfigure = &loadBalancerSecretsDataSource{}
)

// NewLoadBalancerSecretsDataSource returns a new load balancer secrets data source
func NewLoadBalancerSecretsDataSource() datasource.DataSource {
	return &loadBalancerSecretsDataSource{}
}

// loadBalancerSecretsDataSource implements the data source for load balancer secrets
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

// Metadata returns the data source type name
func (d *loadBalancerSecretsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_load_balancer_secrets"
}

// Schema defines the schema for the data source
func (d *loadBalancerSecretsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a list of available load balancer secrets.",
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

// Read gets all load balancer secrets
func (d *loadBalancerSecretsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data loadBalancerSecretsDataSourceModel

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

	// The SDK function and response object will need to be verified against the actual Go SDK.
	lbsApi := d.kc.ApiClient.LoadBalancerEtcAPI.ListTlsCertificates(ctx)

	for _, f := range data.Filter {
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
				lbsApi = lbsApi.CreatedAt(v)
			case "updated_at":
				lbsApi = lbsApi.UpdatedAt(v)
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
	// The result, response, and error must be captured in variables.
	lbsResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, d.kc, &resp.Diagnostics,
		func() (*loadbalancer.SecretListModel, *http.Response, error) {
			return lbsApi.Limit(1000).XAuthToken(d.kc.XAuthToken).Execute()
		},
	)

	if err != nil {
		common.AddApiActionError(ctx, d, httpResp, "ListLoadBalancerSecrets", err, &resp.Diagnostics)
		return
	}

	var lbssResult []loadbalancer.BnsLoadBalancerV1ApiListTlsCertificatesModelSecretModel
	err = copier.Copy(&lbssResult, &lbsResp.Secrets)
	if err != nil {
		resp.Diagnostics.AddError("List 변환 실패", fmt.Sprintf("lbssResult 변환 실패: %v", err))
		return
	}
	// Convert API response directly to our model
	for _, secret := range lbssResult {
		var lbsModel loadBalancerSecretBaseModel
		ok := mapLoadBalancerSecretsBaseModel(ctx, &lbsModel, &secret, &resp.Diagnostics)
		if !ok || resp.Diagnostics.HasError() {
			return
		}
		data.LoadBalancerSecrets = append(data.LoadBalancerSecrets, lbsModel)
	}

	// Set state
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

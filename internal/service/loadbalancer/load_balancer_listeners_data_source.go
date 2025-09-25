// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package loadbalancer

import (
	"context"
	"fmt"
	"net/http"

	"github.com/jinzhu/copier"
	"github.com/kakaoenterprise/kc-sdk-go/services/loadbalancer"

	"terraform-provider-kakaocloud/internal/common"

	//"terraform-provider-kakaocloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &loadBalancerListenersDataSource{}
	_ datasource.DataSourceWithConfigure = &loadBalancerListenersDataSource{}
)

// NewLoadBalancerListenersDataSource is a helper function to simplify the provider implementation.
func NewLoadBalancerListenersDataSource() datasource.DataSource {
	return &loadBalancerListenersDataSource{}
}

type loadBalancerListenersDataSource struct {
	kc *common.KakaoCloudClient
}

// Metadata returns the data source type name.
func (d *loadBalancerListenersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_load_balancer_listeners"
}

// Configure adds the provider configured client to the data source.
func (d *loadBalancerListenersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Schema defines the schema for the data source.
func (d *loadBalancerListenersDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to get information about KakaoCloud Load Balancer Listeners lists.",
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
			"listeners": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: listenerDataSourceSchemaAttributes,
				},
			},
			"timeouts": timeouts.Attributes(ctx),
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *loadBalancerListenersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config loadBalancerListenersDataSourceModel

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

	// The SDK function and response object will need to be verified against the actual Go SDK.
	lblApi := d.kc.ApiClient.LoadBalancerListenerAPI.ListListeners(ctx)
	for _, f := range config.Filter {
		if f.Name.IsNull() || f.Name.IsUnknown() {
			continue
		}

		filterName := f.Name.ValueString()

		if !f.Value.IsNull() && !f.Value.IsUnknown() {
			v := f.Value.ValueString()
			switch filterName {
			case "id":
				lblApi = lblApi.Id(v)
			case "load_balancer_id":
				lblApi = lblApi.LoadBalancerId(v)
			case "protocol":
				if s, err := ToLoadBalancerProtocol(v); err == nil {
					lblApi = lblApi.Protocol(*s)
				} else {
					resp.Diagnostics.AddError(
						"Invalid protocol value",
						err.Error(),
					)
				}
			case "protocol_port":
				lblApi = lblApi.ProtocolPort(v)
			case "provisioning_status":
				if ps, err := ToProvisioningStatus(v); err == nil {
					lblApi = lblApi.ProvisioningStatus(*ps)
				} else {
					resp.Diagnostics.AddError(
						"Invalid provisioning_status value",
						err.Error(),
					)
				}
			case "operating_status":
				if os, err := ToLoadBalancerOperatingStatus(v); err == nil {
					lblApi = lblApi.OperatingStatus(*os)
				} else {
					resp.Diagnostics.AddError(
						"Invalid operating_status value",
						err.Error(),
					)
				}
			case "secret_name":
				lblApi = lblApi.SecretName(v)
			case "secret_id":
				lblApi = lblApi.SecretId(v)
			case "tls_certificate_id":
				lblApi = lblApi.TlsCertificateId(v)
			case "created_at":
				if err := common.ValidateRFC3339(v); err == nil {
					lblApi = lblApi.CreatedAt(v)
				} else {
					resp.Diagnostics.AddError(
						"Invalid created_at value",
						err.Error(),
					)
				}
			case "updated_at":
				if err := common.ValidateRFC3339(v); err == nil {
					lblApi = lblApi.UpdatedAt(v)
				} else {
					resp.Diagnostics.AddError(
						"Invalid updated_at value",
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
	lblResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, d.kc, &resp.Diagnostics,
		func() (*loadbalancer.ListenerListModel, *http.Response, error) {
			return lblApi.Limit(1000).XAuthToken(d.kc.XAuthToken).Execute()
		},
	)

	if err != nil {
		common.AddApiActionError(ctx, d, httpResp, "ListLoadBalancerListeners", err, &resp.Diagnostics)
		return
	}

	var lblsResult []loadbalancer.BnsLoadBalancerV1ApiGetListenerModelListenerModel
	err = copier.Copy(&lblsResult, &lblResp.Listeners)
	if err != nil {
		resp.Diagnostics.AddError("List 변환 실패", fmt.Sprintf("lblsResult 변환 실패: %v", err))
		return
	}

	for _, v := range lblsResult {
		var tmpLbl loadBalancerListenerBaseModel
		ok := mapLoadBalancerListenerBaseModel(ctx, &tmpLbl, &v, &resp.Diagnostics)
		if !ok || resp.Diagnostics.HasError() {
			return
		}

		config.Listeners = append(config.Listeners, tmpLbl)
	}

	// Save data into Terraform state
	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
}

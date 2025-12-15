// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package loadbalancer

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-kakaocloud/internal/common"
	"terraform-provider-kakaocloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/jinzhu/copier"
	"github.com/kakaoenterprise/kc-sdk-go/services/loadbalancer"
)

var (
	_ datasource.DataSource              = &loadBalancerTargetGroupsDataSource{}
	_ datasource.DataSourceWithConfigure = &loadBalancerTargetGroupsDataSource{}
)

func NewLoadBalancerTargetGroupsDataSource() datasource.DataSource {
	return &loadBalancerTargetGroupsDataSource{}
}

type loadBalancerTargetGroupsDataSource struct {
	kc *common.KakaoCloudClient
}

func (d *loadBalancerTargetGroupsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*common.KakaoCloudClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected DataSource Configure Type",
			fmt.Sprintf("Expected *common.KakaoCloudClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	d.kc = client
}

func (d *loadBalancerTargetGroupsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_load_balancer_target_groups"
}

func (d *loadBalancerTargetGroupsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
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
							Required: true,
						},
					},
				},
			},
			"target_groups": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: utils.MergeDataSourceSchemaAttributes(
						map[string]schema.Attribute{
							"id": schema.StringAttribute{
								Computed: true,
							},
						},
						loadBalancerTargetGroupDataSourceSchemaAttributes,
					),
				},
			},
			"timeouts": timeouts.Attributes(ctx),
		},
	}
}

func (d *loadBalancerTargetGroupsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data loadBalancerTargetGroupListDataSourceModel
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

	targetGroupApi := d.kc.ApiClient.LoadBalancerTargetGroupAPI.ListTargetGroups(ctx)

	for _, f := range data.Filter {
		if f.Name.IsNull() || f.Name.IsUnknown() {
			continue
		}

		filterName := f.Name.ValueString()

		if !f.Value.IsNull() && !f.Value.IsUnknown() {
			v := f.Value.ValueString()
			switch filterName {
			case "id":
				targetGroupApi = targetGroupApi.Id(v)
			case "name":
				targetGroupApi = targetGroupApi.Name(v)
			case "protocol":
				if protocol, err := ToTargetGroupProtocol(v); err == nil {
					targetGroupApi = targetGroupApi.Protocol(*protocol)
				} else {
					resp.Diagnostics.AddError(
						"Invalid protocol value",
						err.Error(),
					)
				}
			case "availability_zone":
				if az, err := ToAvailabilityZone(v); err == nil {
					targetGroupApi = targetGroupApi.AvailabilityZone(*az)
				} else {
					resp.Diagnostics.AddError("Invalid availability_zone", err.Error())
				}
			case "load_balancer_algorithm":
				if algorithm, err := ToTargetGroupAlgorithm(v); err == nil {
					targetGroupApi = targetGroupApi.LoadBalancerAlgorithm(*algorithm)
				} else {
					resp.Diagnostics.AddError("Invalid load_balancer_algorithm", err.Error())
				}
			case "load_balancer_name":
				targetGroupApi = targetGroupApi.LoadBalancerName(v)
			case "load_balancer_id":
				targetGroupApi = targetGroupApi.LoadBalancerId(v)
			case "listener_protocol":
				if protocol, err := ToListenerProtocol(v); err == nil {
					targetGroupApi = targetGroupApi.ListenerProtocol(*protocol)
				} else {
					resp.Diagnostics.AddError("Invalid listener_protocol", err.Error())
				}
			case "vpc_name":
				targetGroupApi = targetGroupApi.VpcName(v)
			case "vpc_id":
				targetGroupApi = targetGroupApi.VpcId(v)
			case "subnet_name":
				targetGroupApi = targetGroupApi.SubnetName(v)
			case "subnet_id":
				targetGroupApi = targetGroupApi.SubnetId(v)
			case "health_monitor_id":
				targetGroupApi = targetGroupApi.HealthMonitorId(v)
			case "created_at":
				targetGroupApi = targetGroupApi.CreatedAt(v)
			case "updated_at":
				targetGroupApi = targetGroupApi.UpdatedAt(v)
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

	respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, d.kc, &resp.Diagnostics,
		func() (*loadbalancer.TargetGroupListModel, *http.Response, error) {
			return targetGroupApi.Limit(1000).XAuthToken(d.kc.XAuthToken).Execute()
		},
	)

	if err != nil {
		common.AddApiActionError(ctx, d, httpResp, "ListTargetGroups", err, &resp.Diagnostics)
		return
	}

	var targetGroupResult []loadbalancer.BnsLoadBalancerV1ApiGetTargetGroupModelTargetGroupModel
	err = copier.Copy(&targetGroupResult, &respModel.TargetGroups)
	if err != nil {
		common.AddGeneralError(ctx, d, &resp.Diagnostics,
			fmt.Sprintf("Failed to convert targetGroupResult: %v", err))
		return
	}

	for _, v := range targetGroupResult {
		var tmpTargetGroup loadBalancerTargetGroupBaseModel
		ok := mapLoadBalancerTargetGroupFromGetResponse(ctx, &tmpTargetGroup, &v, &resp.Diagnostics)
		if !ok || resp.Diagnostics.HasError() {
			return
		}

		data.TargetGroups = append(data.TargetGroups, tmpTargetGroup)
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

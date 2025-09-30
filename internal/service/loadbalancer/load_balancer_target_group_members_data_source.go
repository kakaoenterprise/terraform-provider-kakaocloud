// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package loadbalancer

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-kakaocloud/internal/common"
	"terraform-provider-kakaocloud/internal/docs"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/kakaoenterprise/kc-sdk-go/services/loadbalancer"
)

var _ datasource.DataSourceWithConfigure = &loadBalancerTargetGroupMembersDataSource{}

func NewLoadBalancerTargetGroupMembersDataSource() datasource.DataSource {
	return &loadBalancerTargetGroupMembersDataSource{}
}

type loadBalancerTargetGroupMembersDataSource struct {
	kc *common.KakaoCloudClient
}

func (d *loadBalancerTargetGroupMembersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_load_balancer_target_group_members"
}

func (d *loadBalancerTargetGroupMembersDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: docs.GetDataSourceDescription("LoadBalancerTargetGroupMembers"),
		Attributes: map[string]schema.Attribute{
			"target_group_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the target group to list members for",
				Validators:  common.UuidValidator(),
			},
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
			"members": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: loadBalancerTargetGroupMemberListDataSourceSchema["members"].(schema.ListNestedAttribute).NestedObject.Attributes,
				},
			},
			"timeouts": timeouts.Attributes(ctx),
		},
	}
}

func (d *loadBalancerTargetGroupMembersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	kc, ok := req.ProviderData.(*common.KakaoCloudClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected DataSource Configure Type",
			fmt.Sprintf("Expected *common.KakaoCloudClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.kc = kc
}

func (d *loadBalancerTargetGroupMembersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data loadBalancerTargetGroupMemberListDataSourceModel
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

	memberApi := d.kc.ApiClient.LoadBalancerTargetGroupAPI.ListTargetsInTargetGroup(ctx, data.TargetGroupId.ValueString())

	for _, f := range data.Filter {
		if f.Name.IsNull() || f.Name.IsUnknown() {
			continue
		}

		filterName := f.Name.ValueString()

		if !f.Value.IsNull() && !f.Value.IsUnknown() {
			v := f.Value.ValueString()

			switch filterName {
			case "ip":
				memberApi = memberApi.Ip(v)
			case "protocol_port":
				memberApi = memberApi.ProtocolPort(v)
			case "weight":
				memberApi = memberApi.Weight(v)
			case "provisioning_status":
				if ps, err := ToProvisioningStatus(v); err == nil {
					memberApi = memberApi.ProvisioningStatus(*ps)
				} else {
					resp.Diagnostics.AddError(
						"Invalid provisioning_status value",
						err.Error(),
					)
				}
			case "operating_status":
				if os, err := ToLoadBalancerOperatingStatus(v); err == nil {
					memberApi = memberApi.OperatingStatus(*os)
				} else {
					resp.Diagnostics.AddError(
						"Invalid operating_status value",
						err.Error(),
					)
				}
			case "instance_id":
				memberApi = memberApi.InstanceId(v)
			case "instance_name":
				memberApi = memberApi.InstanceName(v)
			case "vpc_id":
				memberApi = memberApi.VpcId(v)
			case "subnet_id":
				memberApi = memberApi.SubnetId(v)
			case "subnet_name":
				memberApi = memberApi.SubnetName(v)
			case "security_group_name":
				memberApi = memberApi.SecurityGroupName(v)
			case "created_at":
				if err := common.ValidateRFC3339(v); err == nil {
					memberApi = memberApi.CreatedAt(v)
				} else {
					resp.Diagnostics.AddError(
						"Invalid created_at value",
						err.Error(),
					)
				}
			case "updated_at":
				if err := common.ValidateRFC3339(v); err == nil {
					memberApi = memberApi.UpdatedAt(v)
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

	respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, d.kc, &resp.Diagnostics,
		func() (*loadbalancer.TargetGroupMemberListModel, *http.Response, error) {
			return memberApi.Limit(1000).XAuthToken(d.kc.XAuthToken).Execute()
		},
	)

	if httpResp != nil && httpResp.StatusCode == 404 {
		resp.Diagnostics.AddError("Target Group Not Found", "The specified target group was not found")
		return
	}

	if err != nil {
		common.AddApiActionError(ctx, d, httpResp, "ListTargetsInTargetGroup", err, &resp.Diagnostics)
		return
	}

	ok := mapLoadBalancerTargetGroupMemberListFromGetResponse(ctx, &data, respModel, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

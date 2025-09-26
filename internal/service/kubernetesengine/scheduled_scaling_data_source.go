// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package kubernetesengine

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-kakaocloud/internal/common"
	"terraform-provider-kakaocloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kakaoenterprise/kc-sdk-go/services/kubernetesengine"
)

var (
	_ datasource.DataSource              = &scheduledScalingDataSource{}
	_ datasource.DataSourceWithConfigure = &scheduledScalingDataSource{}
)

func NewScheduledScalingDataSource() datasource.DataSource { return &scheduledScalingDataSource{} }

type scheduledScalingDataSource struct {
	kc *common.KakaoCloudClient
}

func (d *scheduledScalingDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *scheduledScalingDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kubernetes_engine_scheduled_scalings"
}

func (d *scheduledScalingDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"cluster_name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Cluster name (path parameter).",
			},
			"node_pool_name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Node pool name (path parameter).",
			},
			"scheduled_scaling": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: utils.MergeDataSourceSchemaAttributes(
						map[string]schema.Attribute{
							"name": schema.StringAttribute{
								Computed: true,
							},
						},
						scheduledScalingDataSourceSchemaAttributes,
					),
				},
			},
			"timeouts": timeouts.Attributes(ctx),
		},
	}
}
func (d *scheduledScalingDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config scheduledScalingDataSourceModel

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

	if config.ClusterName.IsNull() || config.ClusterName.IsUnknown() || config.ClusterName.ValueString() == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("cluster_name"),
			"Missing required argument",
			"'cluster_name' must be provided.",
		)
	}
	if config.NodePoolName.IsNull() || config.NodePoolName.IsUnknown() || config.NodePoolName.ValueString() == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("node_pool_name"),
			"Missing required argument",
			"'node_pool_name' must be provided.",
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	clusterName := config.ClusterName.ValueString()
	nodePoolName := config.NodePoolName.ValueString()
	scheduledScalingApi := d.kc.ApiClient.ScalingAPI.ListNodePoolScheduledScalings(ctx, clusterName, nodePoolName)

	modelResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, d.kc, &resp.Diagnostics,
		func() (*kubernetesengine.GetK8sClusterNodePoolScalingScheduleResponseModel, *http.Response, error) {
			return scheduledScalingApi.XAuthToken(d.kc.XAuthToken).Execute()
		},
	)

	if err != nil {
		common.AddApiActionError(ctx, d, httpResp, "ListNodePoolScheduledScalings", err, &resp.Diagnostics)
		return
	}

	for _, v := range modelResp.ScheduledScaling {
		var tmpScheduledScaling scheduledScalingBaseModel
		ok := mapScheduledScalingBaseModel(ctx, &tmpScheduledScaling, &v, &resp.Diagnostics)
		if !ok || resp.Diagnostics.HasError() {
			return
		}

		config.ScheduledScaling = append(config.ScheduledScaling, tmpScheduledScaling)
	}

	config.ClusterName = types.StringValue(clusterName)
	config.NodePoolName = types.StringValue(nodePoolName)

	respDiags := resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(respDiags...)
}

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
	resp.TypeName = req.ProviderTypeName + "_kubernetes_engine_scheduled_scaling"
}

func (d *scheduledScalingDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: utils.MergeAttributes[schema.Attribute](
			map[string]schema.Attribute{
				"cluster_name": schema.StringAttribute{
					Required:   true,
					Validators: common.NameValidator(20),
				},
				"node_pool_name": schema.StringAttribute{
					Required:   true,
					Validators: common.NameValidator(20),
				},
				"name": schema.StringAttribute{
					Required:   true,
					Validators: common.NameValidator(20),
				},
				"timeouts": timeouts.Attributes(ctx),
			},
			scheduledScalingDataSourceSchemaAttributes,
		),
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

	clusterName := config.ClusterName.ValueString()
	nodePoolName := config.NodePoolName.ValueString()
	scheduleName := config.Name.ValueString()

	modelResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, d.kc, &resp.Diagnostics,
		func() (*kubernetesengine.GetK8sClusterNodePoolScalingScheduleResponseModel, *http.Response, error) {
			return d.kc.ApiClient.ScalingAPI.ListNodePoolScheduledScalings(ctx, clusterName, nodePoolName).
				XAuthToken(d.kc.XAuthToken).Execute()
		},
	)

	if err != nil {
		common.AddApiActionError(ctx, d, httpResp, "ListNodePoolScheduledScalings", err, &resp.Diagnostics)
		return
	}

	var found *kubernetesengine.ScheduledScaleResponseModel
	for i := range modelResp.ScheduledScaling {
		if modelResp.ScheduledScaling[i].Name == scheduleName {
			found = &modelResp.ScheduledScaling[i]
			break
		}
	}
	if found == nil {
		common.AddGeneralError(ctx, d, &resp.Diagnostics,
			fmt.Sprintf("The specified scheduled scaling does not exist: '%v'", scheduleName))
		return
	}

	ok := mapScheduledScalingBaseModel(ctx, clusterName, nodePoolName, &config.scheduledScalingBaseModel, found, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	respDiags := resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(respDiags...)
}

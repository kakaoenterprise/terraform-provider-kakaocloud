// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package kubernetesengine

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"terraform-provider-kakaocloud/internal/common"
	"terraform-provider-kakaocloud/internal/docs"
	"terraform-provider-kakaocloud/internal/utils"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/kakaoenterprise/kc-sdk-go/services/kubernetesengine"
)

var (
	_ resource.ResourceWithConfigure      = &scheduledScalingResource{}
	_ resource.ResourceWithImportState    = &scheduledScalingResource{}
	_ resource.ResourceWithValidateConfig = &scheduledScalingResource{}
)

func NewScheduledScalingResource() resource.Resource { return &scheduledScalingResource{} }

type scheduledScalingResource struct {
	kc *common.KakaoCloudClient
}

func (r *scheduledScalingResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config scheduledScalingResourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.validateScheduleRequirement(config, resp)
}

func (r *scheduledScalingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 3)
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			"Expected <cluster_name>/<node_pool_name>/<name>.",
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cluster_name"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("node_pool_name"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), parts[2])...)
}
func (r *scheduledScalingResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kubernetes_engine_scheduled_scaling"
}

func (r *scheduledScalingResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: docs.GetResourceDescription("KubernetesEngineScheduledScaling"),
		Attributes: utils.MergeResourceSchemaAttributes(
			scheduledScalingResourceSchema,
			map[string]schema.Attribute{
				"timeouts": timeouts.AttributesAll(ctx),
			},
		),
	}
}

type scheduleLookup struct {
	Found bool
	Item  *kubernetesengine.ScheduledScaleResponseModel
}

func (r *scheduledScalingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan, config scheduledScalingResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	timeout, diags := plan.Timeouts.Create(ctx, common.DefaultCreateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	createReq := kubernetesengine.ScheduleRequestModel{
		Name:         plan.Name.ValueString(),
		ScheduleType: kubernetesengine.SchedulingType(plan.ScheduleType.ValueString()),
		DesiredNodes: plan.DesiredNodes.ValueInt32(),
		StartTime:    stripSecondsForAPI(plan.StartTime.ValueString()),
	}
	if !plan.Schedule.IsNull() && !plan.Schedule.IsUnknown() && plan.Schedule.ValueString() != "" {
		createReq.SetSchedule(plan.Schedule.ValueString())
	}

	body := kubernetesengine.CreateK8sClusterNodePoolScalingScheduleRequestModel{
		ScheduledScaling: createReq,
	}

	clusterName := plan.ClusterName.ValueString()
	nodePoolName := plan.NodePoolName.ValueString()

	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (interface{}, *http.Response, error) {
			return r.kc.ApiClient.ScalingAPI.
				CreateNodePoolScheduledScaling(ctx, clusterName, nodePoolName).
				XAuthToken(r.kc.XAuthToken).
				CreateK8sClusterNodePoolScalingScheduleRequestModel(body).
				Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "CreateNodePoolScheduledScaling", err, &resp.Diagnostics)
		return
	}

	result, ok := common.PollUntilResult(
		ctx,
		r,
		2*time.Second,
		[]string{"found"},
		&resp.Diagnostics,
		func(ctx context.Context) (scheduleLookup, *http.Response, error) {
			modelResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
				func() (*kubernetesengine.GetK8sClusterNodePoolScalingScheduleResponseModel, *http.Response, error) {
					return r.kc.ApiClient.ScalingAPI.
						ListNodePoolScheduledScalings(ctx, clusterName, nodePoolName).
						XAuthToken(r.kc.XAuthToken).
						Execute()
				},
			)
			if err != nil {
				return scheduleLookup{}, httpResp, err
			}

			for i := range modelResp.ScheduledScaling {
				if modelResp.ScheduledScaling[i].Name == plan.Name.ValueString() {
					return scheduleLookup{Found: true, Item: &modelResp.ScheduledScaling[i]}, httpResp, nil
				}
			}

			return scheduleLookup{Found: false, Item: nil}, httpResp, nil
		},
		func(v scheduleLookup) string {
			if v.Found {
				return "found"
			}
			return "pending"
		},
	)

	if !ok {
		return
	}

	created := result.Item

	ok = mapScheduledScalingBaseModel(ctx, &plan.scheduledScalingBaseModel, created, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *scheduledScalingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state scheduledScalingResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	timeout, diags := state.Timeouts.Read(ctx, common.DefaultReadTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	clusterName := state.ClusterName.ValueString()
	nodePoolName := state.NodePoolName.ValueString()
	nodeName := state.Name.ValueString()

	respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*kubernetesengine.GetK8sClusterNodePoolScalingScheduleResponseModel, *http.Response, error) {
			return r.kc.ApiClient.ScalingAPI.
				ListNodePoolScheduledScalings(ctx, clusterName, nodePoolName).
				XAuthToken(r.kc.XAuthToken).
				Execute()
		},
	)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		common.AddApiActionError(ctx, r, httpResp, "ListNodePoolScheduledScalings", err, &resp.Diagnostics)
		return
	}

	var found *kubernetesengine.ScheduledScaleResponseModel
	for i := range respModel.ScheduledScaling {
		if respModel.ScheduledScaling[i].Name == nodeName {
			found = &respModel.ScheduledScaling[i]
			break
		}
	}
	if found == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	ok := mapScheduledScalingBaseModel(ctx, &state.scheduledScalingBaseModel, found, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *scheduledScalingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"This resource does not support update. Please recreate the resource if needed.",
	)
}

func (r *scheduledScalingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state scheduledScalingResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	timeout, diags := state.Timeouts.Delete(ctx, common.DefaultDeleteTimeout)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	clusterName := state.ClusterName.ValueString()
	nodePoolName := state.NodePoolName.ValueString()
	scheduleName := state.Name.ValueString()

	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (interface{}, *http.Response, error) {
			httpResp, err := r.kc.ApiClient.ScalingAPI.
				DeleteNodePoolScheduledScaling(ctx, clusterName, nodePoolName, scheduleName).
				XAuthToken(r.kc.XAuthToken).
				Execute()
			return nil, httpResp, err
		},
	)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			return
		}
		common.AddApiActionError(ctx, r, httpResp, "DeleteVolume", err, &resp.Diagnostics)
		return
	}

	common.PollUntilDeletion(ctx, r, 2*time.Second, &resp.Diagnostics, func(ctx context.Context) (bool, *http.Response, error) {
		modelResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
			func() (*kubernetesengine.GetK8sClusterNodePoolScalingScheduleResponseModel, *http.Response, error) {
				return r.kc.ApiClient.ScalingAPI.
					ListNodePoolScheduledScalings(ctx, clusterName, nodePoolName).
					XAuthToken(r.kc.XAuthToken).
					Execute()
			},
		)
		if err != nil {
			if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
				return true, httpResp, nil
			}
			return false, httpResp, err
		}

		for _, it := range modelResp.ScheduledScaling {
			if it.Name == scheduleName {
				return false, httpResp, nil
			}
		}
		return true, httpResp, nil
	})

	resp.State.RemoveResource(ctx)
}

func (r *scheduledScalingResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.kc = client
}

func stripSecondsForAPI(val string) string {
	if strings.HasSuffix(val, ":00Z") {
		return strings.Replace(val, ":00Z", "Z", 1)
	}
	return val
}

func (r *scheduledScalingResource) validateScheduleRequirement(
	config scheduledScalingResourceModel,
	resp *resource.ValidateConfigResponse,
) {
	if resp.Diagnostics.HasError() {
		return
	}

	if !config.ScheduleType.IsNull() && !config.ScheduleType.IsUnknown() &&
		config.ScheduleType.ValueString() == "cron" {

		if config.Schedule.IsNull() || config.Schedule.IsUnknown() || config.Schedule.ValueString() == "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("schedule"),
				"Missing schedule for cron",
				"When scheduling_type is 'cron', 'schedule' must be provided.",
			)
		}
		return
	}

	if !config.Schedule.IsNull() && !config.Schedule.IsUnknown() && config.Schedule.ValueString() != "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("schedule"),
			"Ignored schedule",
			"schedule is only used when scheduling_type is 'cron'. The value will be ignored.",
		)
	}
}

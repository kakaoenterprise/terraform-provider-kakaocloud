// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package kubernetesengine

import (
	"context"
	"fmt"
	"github.com/kakaoenterprise/kc-sdk-go/services/bcs"
	"github.com/kakaoenterprise/kc-sdk-go/services/network"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/kakaoenterprise/kc-sdk-go/services/kubernetesengine"
	"terraform-provider-kakaocloud/internal/common"
	"terraform-provider-kakaocloud/internal/docs"
	"terraform-provider-kakaocloud/internal/utils"
)

var (
	_ resource.ResourceWithConfigure      = &nodePoolResource{}
	_ resource.ResourceWithImportState    = &nodePoolResource{}
	_ resource.ResourceWithValidateConfig = &nodePoolResource{}
	_ resource.ResourceWithModifyPlan     = &nodePoolResource{}
)

func NewNodePoolResource() resource.Resource { return &nodePoolResource{} }

type nodePoolResource struct {
	kc *common.KakaoCloudClient
}

func (r *nodePoolResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kubernetes_engine_node_pool"
}

func (r *nodePoolResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: docs.GetResourceDescription("KubernetesEngineNodePool"),
		Attributes: utils.MergeResourceSchemaAttributes(
			nodePoolResourceAttributes,
			map[string]schema.Attribute{
				"timeouts": timeouts.AttributesAll(ctx),
			},
		),
	}
}

func (r *nodePoolResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *nodePoolResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {

	id := req.ID
	parts := strings.SplitN(id, "/", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Expected format 'cluster_name/node_pool_name', got: %q", id),
		)
		return
	}
	clusterName, nodePoolName := parts[0], parts[1]

	var diags diag.Diagnostics
	diags = append(diags, resp.State.SetAttribute(ctx, path.Root("cluster_name"), clusterName)...)
	diags = append(diags, resp.State.SetAttribute(ctx, path.Root("name"), nodePoolName)...)
	diags = append(diags, resp.State.SetAttribute(ctx, path.Root("id"), id)...)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	detail, _, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (struct {
			NodePool kubernetesengine.KubernetesEngineV1ApiGetNodePoolModelNodePoolResponseModel
		}, *http.Response, error) {
			apiResp, hr, err := r.kc.ApiClient.NodePoolsAPI.
				GetNodePool(ctx, clusterName, nodePoolName).
				XAuthToken(r.kc.XAuthToken).
				Execute()
			if err != nil {
				return struct {
					NodePool kubernetesengine.KubernetesEngineV1ApiGetNodePoolModelNodePoolResponseModel
				}{}, hr, err
			}
			return struct {
				NodePool kubernetesengine.KubernetesEngineV1ApiGetNodePoolModelNodePoolResponseModel
			}{NodePool: apiResp.NodePool}, hr, nil
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, nil, "GetNodePool", err, &resp.Diagnostics)
		return
	}

	sgIDs := detail.NodePool.SecurityGroups
	if len(sgIDs) == 1 {

		return
	}

	const defaultPrefix = "k8s-cluster-ke-cluster"
	userSGs := make([]string, 0, len(sgIDs))

	for _, sgID := range sgIDs {
		sgResp, _, e := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
			func() (*network.BnsNetworkV1ApiGetSecurityGroupModelResponseSecurityGroupModel, *http.Response, error) {
				return r.kc.ApiClient.SecurityGroupAPI.
					GetSecurityGroup(ctx, sgID).
					XAuthToken(r.kc.XAuthToken).
					Execute()
			},
		)
		if e != nil {
			common.AddApiActionError(ctx, r, nil, "GetSecurityGroup", e, &resp.Diagnostics)
			continue
		}

		var nameStr string
		if sgResp.SecurityGroup.Name.IsSet() && sgResp.SecurityGroup.Name.Get() != nil {
			nameStr = string(*sgResp.SecurityGroup.Name.Get())
		}

		if !strings.HasPrefix(nameStr, defaultPrefix) {

			userSGs = append(userSGs, sgID)
		}
	}

	diags = append(diags, resp.State.SetAttribute(ctx, path.Root("security_groups"), sgIDs)...)
	diags = append(diags, resp.State.SetAttribute(ctx, path.Root("request_security_groups"), userSGs)...)
	resp.Diagnostics.Append(diags...)
}

func (r *nodePoolResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan NodePoolResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var config NodePoolResourceModel
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

	flavorResp, httpRespFlavor, errFlavor := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*bcs.ResponseFlavorModel, *http.Response, error) {
			return r.kc.ApiClient.FlavorAPI.GetInstanceType(ctx, plan.FlavorId.ValueString()).
				XAuthToken(r.kc.XAuthToken).Execute()
		},
	)
	if errFlavor != nil {
		common.AddApiActionError(ctx, r, httpRespFlavor, "GetInstanceType", errFlavor, &resp.Diagnostics)
		return
	}
	if string(flavorResp.Flavor.GetInstanceType()) == common.InstanceTypeBM {

		hasErr := false
		if !config.VolumeSize.IsNull() && !config.VolumeSize.IsUnknown() {
			resp.Diagnostics.AddAttributeError(
				path.Root("volume_size"),
				"volume_size is not allowed for bare metal node pools",
				"Bare metal node pools do not support 'volume_size' configuration. Remove 'volume_size' from configuration.",
			)
			hasErr = true
		}
		if !config.Autoscaling.IsNull() && !config.Autoscaling.IsUnknown() {
			resp.Diagnostics.AddAttributeError(
				path.Root("autoscaling"),
				"Autoscaling is not supported for bare metal node pools",
				"This node pool flavor is bare metal. Remove the 'autoscaling' block from configuration.",
			)
			hasErr = true
		}
		if hasErr {
			return
		}
	}

	var initialNodeCount int32
	var autoscalingProvided bool
	var autoscalingEnabled bool
	var autoscalingModel NodePoolAutoscalingModel

	if !config.Autoscaling.IsNull() && !config.Autoscaling.IsUnknown() {
		autoscalingProvided = true
		_ = config.Autoscaling.As(ctx, &autoscalingModel, basetypes.ObjectAsOptions{})
		if !autoscalingModel.IsAutoscalerEnable.IsNull() && !autoscalingModel.IsAutoscalerEnable.IsUnknown() {
			autoscalingEnabled = autoscalingModel.IsAutoscalerEnable.ValueBool()
		}
	}

	if autoscalingProvided {
		if ok, msg := validateAutoscalingModel(autoscalingModel); !ok {
			resp.Diagnostics.AddError("Invalid autoscaling configuration", msg)
			return
		}
	}

	if autoscalingProvided && autoscalingEnabled {

		if !autoscalingModel.AutoscalerDesiredNodeCount.IsNull() && !autoscalingModel.AutoscalerDesiredNodeCount.IsUnknown() {
			initialNodeCount = autoscalingModel.AutoscalerDesiredNodeCount.ValueInt32()
		} else {
			resp.Diagnostics.AddAttributeError(
				path.Root("autoscaling").AtName("autoscaler_desired_node_count"),
				"Missing required autoscaling.desired value",
				"When enabling autoscaling during create, 'autoscaling.autoscaler_desired_node_count' is required and is used as the initial node count. Remove 'node_count' from configuration.",
			)
			return
		}
	} else {

		if plan.NodeCount.IsNull() || plan.NodeCount.IsUnknown() {
			resp.Diagnostics.AddAttributeError(
				path.Root("node_count"),
				"Missing required attribute on create",
				"'node_count' must be specified when creating a node pool (or enable autoscaling with a desired/min/max node count).",
			)
			return
		}
		initialNodeCount = plan.NodeCount.ValueInt32()
	}

	var vpcInfo NodePoolVpcInfoModelSet
	diags = plan.VpcInfo.As(ctx, &vpcInfo, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var subnetIds []string
	if !vpcInfo.Subnets.IsNull() && !vpcInfo.Subnets.IsUnknown() {
		var subnetObjs []SubnetModel
		_ = vpcInfo.Subnets.ElementsAs(ctx, &subnetObjs, false)
		for _, s := range subnetObjs {
			subnetIds = append(subnetIds, s.Id.ValueString())
		}
	}

	reqVpcInfo := kubernetesengine.VpcInfoRequestModel{
		Id:      vpcInfo.Id.ValueString(),
		Subnets: subnetIds,
	}

	var labels []kubernetesengine.LabelRequestModel
	if !plan.Labels.IsNull() && !plan.Labels.IsUnknown() {
		var tmp []NodePoolLabelModel
		_ = plan.Labels.ElementsAs(ctx, &tmp, false)
		for _, l := range tmp {
			labels = append(labels, kubernetesengine.LabelRequestModel{
				Key:   l.Key.ValueString(),
				Value: l.Value.ValueString(),
			})
		}
	}

	var taints []kubernetesengine.TaintRequestModel
	if !plan.Taints.IsNull() && !plan.Taints.IsUnknown() {
		var tmp []NodePoolTaintModel
		_ = plan.Taints.ElementsAs(ctx, &tmp, false)
		for _, t := range tmp {
			taints = append(taints, kubernetesengine.TaintRequestModel{
				Key:    t.Key.ValueString(),
				Value:  t.Value.ValueString(),
				Effect: kubernetesengine.NodePoolTaintEffect(t.Effect.ValueString()),
			})
		}
	}

	var userSGs []string
	hasUserSGs := false
	if !plan.RequestSecurityGroups.IsNull() && !plan.RequestSecurityGroups.IsUnknown() {
		_ = plan.RequestSecurityGroups.ElementsAs(ctx, &userSGs, false)
		if len(userSGs) > 0 {
			hasUserSGs = true
		}
	}

	createModel := kubernetesengine.NewKubernetesEngineV1ApiCreateNodePoolModelNodePoolRequestModel(
		plan.Name.ValueString(),
		plan.FlavorId.ValueString(),
		initialNodeCount,
		plan.SshKeyName.ValueString(),
		reqVpcInfo,
		plan.ImageId.ValueString(),
	)

	if !plan.VolumeSize.IsNull() && !plan.VolumeSize.IsUnknown() {
		createModel.SetVolumeSize(plan.VolumeSize.ValueInt32())
	}

	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		createModel.SetDescription(plan.Description.ValueString())
	}

	if !config.UserData.IsNull() && !config.UserData.IsUnknown() && config.UserData.ValueString() != "" {
		createModel.SetUserData(config.UserData.ValueString())
	}
	if !plan.IsHyperThreading.IsNull() && !plan.IsHyperThreading.IsUnknown() {
		createModel.SetIsHyperThreading(plan.IsHyperThreading.ValueBool())
	}
	if labels != nil {
		createModel.SetLabels(labels)
	}
	if taints != nil {
		createModel.SetTaints(taints)
	}

	if hasUserSGs {
		createModel.SetSecurityGroups(userSGs)
	}

	reqBody := kubernetesengine.CreateK8sClusterNodePoolRequestModel{NodePool: *createModel}

	_, httpRespCreate, errCreate := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics, func() (interface{}, *http.Response, error) {
		return r.kc.ApiClient.NodePoolsAPI.
			CreateNodePool(ctx, plan.ClusterName.ValueString()).
			XAuthToken(r.kc.XAuthToken).
			CreateK8sClusterNodePoolRequestModel(reqBody).
			Execute()
	})
	if errCreate != nil {
		common.AddApiActionError(ctx, r, httpRespCreate, "CreateNodePool", errCreate, &resp.Diagnostics)
		return
	}

	plan.Id = types.StringValue(plan.ClusterName.ValueString() + "/" + plan.Name.ValueString())
	result, ok := r.waitNodePoolReadyOrFailed(
		ctx,
		plan.ClusterName.ValueString(),
		plan.Name.ValueString(),
		&resp.Diagnostics,
	)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	if len(userSGs) > 0 {
		_, ok2 := r.waitNodePoolSecurityGroupsContains(ctx, plan.ClusterName.ValueString(), plan.Name.ValueString(), userSGs, &resp.Diagnostics)
		if !ok2 || resp.Diagnostics.HasError() {
			return
		}
	}

	if autoscalingProvided && result != nil && !result.IsBareMetal {

		if autoscalingModel.IsAutoscalerEnable.IsUnknown() || autoscalingModel.IsAutoscalerEnable.IsNull() {
			resp.Diagnostics.AddError("Invalid autoscaling configuration on create", "'autoscaling.is_autoscaler_enable' must be set when configuring autoscaling during create.")
			return
		}
		scaling := kubernetesengine.NewNodePoolScalingResourceRequestModel(autoscalingModel.IsAutoscalerEnable.ValueBool())
		if !autoscalingModel.AutoscalerDesiredNodeCount.IsUnknown() {
			if autoscalingModel.AutoscalerDesiredNodeCount.IsNull() {
				scaling.SetAutoscalerDesiredNodeCountNil()
			} else {
				scaling.SetAutoscalerDesiredNodeCount(autoscalingModel.AutoscalerDesiredNodeCount.ValueInt32())
			}
		}
		if !autoscalingModel.AutoscalerMaxNodeCount.IsUnknown() {
			if autoscalingModel.AutoscalerMaxNodeCount.IsNull() {
				scaling.SetAutoscalerMaxNodeCountNil()
			} else {
				scaling.SetAutoscalerMaxNodeCount(autoscalingModel.AutoscalerMaxNodeCount.ValueInt32())
			}
		}
		if !autoscalingModel.AutoscalerMinNodeCount.IsUnknown() {
			if autoscalingModel.AutoscalerMinNodeCount.IsNull() {
				scaling.SetAutoscalerMinNodeCountNil()
			} else {
				scaling.SetAutoscalerMinNodeCount(autoscalingModel.AutoscalerMinNodeCount.ValueInt32())
			}
		}
		if !autoscalingModel.AutoscalerScaleDownThreshold.IsUnknown() {
			if autoscalingModel.AutoscalerScaleDownThreshold.IsNull() {
				scaling.SetAutoscalerScaleDownThresholdNil()
			} else {
				scaling.SetAutoscalerScaleDownThreshold(autoscalingModel.AutoscalerScaleDownThreshold.ValueFloat32())
			}
		}
		if !autoscalingModel.AutoscalerScaleDownUnneededTime.IsUnknown() {
			if autoscalingModel.AutoscalerScaleDownUnneededTime.IsNull() {
				scaling.SetAutoscalerScaleDownUnneededTimeNil()
			} else {
				scaling.SetAutoscalerScaleDownUnneededTime(autoscalingModel.AutoscalerScaleDownUnneededTime.ValueInt32())
			}
		}
		if !autoscalingModel.AutoscalerScaleDownUnreadyTime.IsUnknown() {
			if autoscalingModel.AutoscalerScaleDownUnreadyTime.IsNull() {
				scaling.SetAutoscalerScaleDownUnreadyTimeNil()
			} else {
				scaling.SetAutoscalerScaleDownUnreadyTime(autoscalingModel.AutoscalerScaleDownUnreadyTime.ValueInt32())
			}
		}
		updBody := kubernetesengine.NewUpdateKubernetesEngineClusterNodePoolScalingResourceRequestModel(*scaling)
		_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics, func() (interface{}, *http.Response, error) {
			return r.kc.ApiClient.ScalingAPI.
				SetNodePoolResourceBasedAutoScaling(ctx, plan.ClusterName.ValueString(), plan.Name.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				UpdateKubernetesEngineClusterNodePoolScalingResourceRequestModel(*updBody).
				Execute()
		})
		if err != nil {
			common.AddApiActionError(ctx, r, httpResp, "SetNodePoolResourceBasedAutoScaling", err, &resp.Diagnostics)
			return
		}
		_, ok2 := r.waitNodePoolReadyOrFailed(ctx, plan.ClusterName.ValueString(), plan.Name.ValueString(), &resp.Diagnostics)
		if !ok2 || resp.Diagnostics.HasError() {
			return
		}
	}

	finalDetail, httpResp2, err2 := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (struct {
			NodePool kubernetesengine.KubernetesEngineV1ApiGetNodePoolModelNodePoolResponseModel
		}, *http.Response, error) {
			apiResp, hr, err := r.kc.ApiClient.NodePoolsAPI.
				GetNodePool(ctx, plan.ClusterName.ValueString(), plan.Name.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				Execute()
			if err != nil {
				return struct {
					NodePool kubernetesengine.KubernetesEngineV1ApiGetNodePoolModelNodePoolResponseModel
				}{}, hr, err
			}
			return struct {
				NodePool kubernetesengine.KubernetesEngineV1ApiGetNodePoolModelNodePoolResponseModel
			}{NodePool: apiResp.NodePool}, hr, nil
		},
	)
	if err2 != nil {
		common.AddApiActionError(ctx, r, httpResp2, "GetNodePool", err2, &resp.Diagnostics)
		return
	}
	res := finalDetail.NodePool
	result = &res

	var planUserSGs []string
	if !plan.RequestSecurityGroups.IsNull() && !plan.RequestSecurityGroups.IsUnknown() {
		_ = plan.RequestSecurityGroups.ElementsAs(ctx, &planUserSGs, false)
	}
	_ = mapNodePoolFromResponse(ctx, &plan.NodePoolBaseModel, result, &resp.Diagnostics, planUserSGs)

	if result.VpcInfo.Subnets != nil {
		subnetObjs := make([]attr.Value, 0, len(result.VpcInfo.Subnets))
		for _, s := range result.VpcInfo.Subnets {
			obj, _ := types.ObjectValue(
				subnetAttrTypes,
				map[string]attr.Value{
					"id":                types.StringValue(s.Id),
					"availability_zone": types.StringValue(string(s.AvailabilityZone)),
					"cidr_block":        types.StringValue(s.CidrBlock),
				},
			)
			subnetObjs = append(subnetObjs, obj)
		}

		subnetSet, _ := types.SetValue(types.ObjectType{AttrTypes: subnetAttrTypes}, subnetObjs)

		plan.VpcInfo, _ = types.ObjectValue(
			map[string]attr.Type{
				"id":      types.StringType,
				"subnets": types.SetType{ElemType: types.ObjectType{AttrTypes: subnetAttrTypes}},
			},
			map[string]attr.Value{
				"id":      types.StringValue(result.VpcInfo.Id),
				"subnets": subnetSet,
			},
		)
	}

	if !config.UserData.IsNull() && !config.UserData.IsUnknown() && config.UserData.ValueString() != "" {
		plan.UserData = config.UserData
	}

	plan.RequestSecurityGroups = config.RequestSecurityGroups

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *nodePoolResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state NodePoolResourceModel
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

	detail, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (struct {
			NodePool kubernetesengine.KubernetesEngineV1ApiGetNodePoolModelNodePoolResponseModel
		}, *http.Response, error) {
			apiResp, httpResp, err := r.kc.ApiClient.NodePoolsAPI.
				GetNodePool(ctx, state.ClusterName.ValueString(), state.Name.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				Execute()
			if err != nil {
				return struct {
					NodePool kubernetesengine.KubernetesEngineV1ApiGetNodePoolModelNodePoolResponseModel
				}{}, httpResp, err
			}
			return struct {
				NodePool kubernetesengine.KubernetesEngineV1ApiGetNodePoolModelNodePoolResponseModel
			}{NodePool: apiResp.NodePool}, httpResp, nil
		},
	)
	if httpResp != nil && httpResp.StatusCode == 404 {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "GetNodePool", err, &resp.Diagnostics)
		return
	}

	result := detail.NodePool

	var stateUserSGs []string
	if !state.RequestSecurityGroups.IsNull() && !state.RequestSecurityGroups.IsUnknown() {
		_ = state.RequestSecurityGroups.ElementsAs(ctx, &stateUserSGs, false)
	}
	_ = mapNodePoolFromResponse(ctx, &state.NodePoolBaseModel, &result, &resp.Diagnostics, stateUserSGs)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *nodePoolResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state NodePoolResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var config NodePoolResourceModel
	diags = req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	timeout, diags := plan.Timeouts.Update(ctx, common.DefaultUpdateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var immutableChanges []string
	if !plan.ClusterName.Equal(state.ClusterName) {
		immutableChanges = append(immutableChanges, "cluster_name")
	}
	if !plan.Name.Equal(state.Name) {
		immutableChanges = append(immutableChanges, "name")
	}
	if !plan.FlavorId.Equal(state.FlavorId) {
		immutableChanges = append(immutableChanges, "flavor_id")
	}
	if !plan.VolumeSize.Equal(state.VolumeSize) {
		immutableChanges = append(immutableChanges, "volume_size")
	}
	if !plan.SshKeyName.Equal(state.SshKeyName) {
		immutableChanges = append(immutableChanges, "ssh_key_name")
	}
	if !plan.ImageId.Equal(state.ImageId) {
		immutableChanges = append(immutableChanges, "image_id")
	}
	if !plan.IsHyperThreading.Equal(state.IsHyperThreading) {
		immutableChanges = append(immutableChanges, "is_hyper_threading")
	}
	if !plan.Taints.Equal(state.Taints) {
		immutableChanges = append(immutableChanges, "taints")
	}
	if !plan.VpcInfo.Equal(state.VpcInfo) {
		immutableChanges = append(immutableChanges, "vpc_info")
	}

	if len(immutableChanges) > 0 {
		resp.Diagnostics.AddError(
			"Immutable field update attempted",
			fmt.Sprintf(
				"The following fields cannot be updated in-place and require resource replacement: %v",
				immutableChanges,
			),
		)
		return
	}

	if !config.Autoscaling.IsNull() && !config.Autoscaling.IsUnknown() && (state.IsBareMetal.IsNull() || state.IsBareMetal.IsUnknown() || !state.IsBareMetal.ValueBool()) {
		var earlyAuto NodePoolAutoscalingModel
		_ = config.Autoscaling.As(ctx, &earlyAuto, basetypes.ObjectAsOptions{})
		if ok, msg := validateAutoscalingModel(earlyAuto); !ok {
			resp.Diagnostics.AddError("Invalid autoscaling configuration", msg)
			return
		}
	}

	upd := kubernetesengine.NewKubernetesEngineV1ApiUpdateNodePoolModelNodePoolRequestModel()

	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		upd.SetDescription(plan.Description.ValueString())
	}

	var userSGs []string
	shouldSendSGs := false
	if !plan.RequestSecurityGroups.IsNull() && !plan.RequestSecurityGroups.IsUnknown() {

		_ = plan.RequestSecurityGroups.ElementsAs(ctx, &userSGs, false)
		if len(userSGs) > 0 {
			shouldSendSGs = true
		} else {

			if !state.RequestSecurityGroups.IsNull() && !state.RequestSecurityGroups.IsUnknown() {
				var prevReq []string
				_ = state.RequestSecurityGroups.ElementsAs(ctx, &prevReq, false)
				if len(prevReq) > 0 {
					userSGs = []string{}
					shouldSendSGs = true
				}
			}
		}
	} else {

		if !state.RequestSecurityGroups.IsNull() && !state.RequestSecurityGroups.IsUnknown() {
			var prevReq []string
			_ = state.RequestSecurityGroups.ElementsAs(ctx, &prevReq, false)
			if len(prevReq) > 0 {
				userSGs = []string{}
				shouldSendSGs = true
			}
		}
	}
	if shouldSendSGs {
		upd.SetSecurityGroups(userSGs)
	}

	planLabels := map[string]string{}
	stateLabels := map[string]string{}

	if !plan.Labels.IsNull() && !plan.Labels.IsUnknown() {
		var pl []NodePoolLabelModel
		if err := plan.Labels.ElementsAs(ctx, &pl, false); err != nil {
			resp.Diagnostics.AddError(
				"Invalid Plan Labels",
				fmt.Sprintf("Failed to parse labels from plan: %s", err),
			)
			return
		}
		for _, l := range pl {
			planLabels[l.Key.ValueString()] = l.Value.ValueString()
		}
	}

	if !state.Labels.IsNull() && !state.Labels.IsUnknown() {
		var sl []NodePoolLabelModel
		if err := state.Labels.ElementsAs(ctx, &sl, false); err != nil {
			resp.Diagnostics.AddError(
				"Invalid State Labels",
				fmt.Sprintf("Failed to parse labels from state: %s", err),
			)
			return
		}
		for _, l := range sl {
			stateLabels[l.Key.ValueString()] = l.Value.ValueString()
		}
	}

	addOrUpdate := make([]kubernetesengine.LabelRequestModel, 0)
	removeKeys := make([]string, 0)

	for k, v := range planLabels {
		if sv, ok := stateLabels[k]; !ok || sv != v {
			addOrUpdate = append(addOrUpdate, *kubernetesengine.NewLabelRequestModel(k, v))
		}
	}

	for k := range stateLabels {
		if _, ok := planLabels[k]; !ok {
			removeKeys = append(removeKeys, k)
		}
	}
	if len(addOrUpdate) > 0 || len(removeKeys) > 0 {
		labelsReq := kubernetesengine.NewNodeLabelsRequestModel()
		if len(addOrUpdate) > 0 {
			labelsReq.SetAddOrUpdateLabels(addOrUpdate)
		}
		if len(removeKeys) > 0 {
			labelsReq.SetRemoveLabelKeys(removeKeys)
		}
		body := kubernetesengine.NewUpdateK8sClusterNodePoolNodeLabelsRequestModel(*labelsReq)
		_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics, func() (interface{}, *http.Response, error) {
			return r.kc.ApiClient.NodePoolsAPI.
				SetNodePoolNodeLabel(ctx, plan.ClusterName.ValueString(), state.Name.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				UpdateK8sClusterNodePoolNodeLabelsRequestModel(*body).
				Execute()
		})
		if err != nil {
			common.AddApiActionError(ctx, r, httpResp, "SetNodePoolNodeLabel", err, &resp.Diagnostics)
			return
		}

		_, _ = r.waitNodePoolReadyOrFailed(
			ctx,
			plan.ClusterName.ValueString(),
			state.Name.ValueString(),
			&resp.Diagnostics,
		)
	}

	effectiveAutoscalingEnabled := false

	if !state.Autoscaling.IsNull() && !state.Autoscaling.IsUnknown() {
		var stAuto NodePoolAutoscalingModel
		_ = state.Autoscaling.As(ctx, &stAuto, basetypes.ObjectAsOptions{})
		if !stAuto.IsAutoscalerEnable.IsNull() && !stAuto.IsAutoscalerEnable.IsUnknown() {
			effectiveAutoscalingEnabled = stAuto.IsAutoscalerEnable.ValueBool()
		}
	}

	if !plan.Autoscaling.IsNull() && !plan.Autoscaling.IsUnknown() {
		var plAuto NodePoolAutoscalingModel
		_ = plan.Autoscaling.As(ctx, &plAuto, basetypes.ObjectAsOptions{})
		if !plAuto.IsAutoscalerEnable.IsNull() && !plAuto.IsAutoscalerEnable.IsUnknown() {
			effectiveAutoscalingEnabled = plAuto.IsAutoscalerEnable.ValueBool()
		}
	}

	didSetCount := false
	if effectiveAutoscalingEnabled {

		if !req.State.Raw.IsNull() {
			if !plan.NodeCount.IsNull() && !plan.NodeCount.IsUnknown() {

				if state.NodeCount.IsNull() || state.NodeCount.IsUnknown() || plan.NodeCount.ValueInt32() != state.NodeCount.ValueInt32() {
					resp.Diagnostics.AddAttributeError(
						path.Root("node_count"),
						"node_count is read-only when autoscaling is enabled",
						"Disable autoscaling to modify node_count.",
					)
					return
				}
			}
		}
	} else {
		if !plan.NodeCount.IsNull() && !plan.NodeCount.IsUnknown() {
			upd.SetNodeCount(plan.NodeCount.ValueInt32())
			didSetCount = true
		}
	}

	didSetDesc := !plan.Description.IsNull() && !plan.Description.IsUnknown()
	if didSetDesc || didSetCount {
		reqBody := kubernetesengine.UpdateK8sClusterNodePoolRequestModel{NodePool: *upd}
		_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics, func() (interface{}, *http.Response, error) {
			return r.kc.ApiClient.NodePoolsAPI.
				UpdateNodePool(ctx, plan.ClusterName.ValueString(), state.Name.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				UpdateK8sClusterNodePoolRequestModel(reqBody).
				Execute()
		})
		if err != nil {
			common.AddApiActionError(ctx, r, httpResp, "UpdateNodePool", err, &resp.Diagnostics)
			return
		}

		_, _ = r.waitNodePoolReadyOrFailed(
			ctx,
			plan.ClusterName.ValueString(),
			state.Name.ValueString(),
			&resp.Diagnostics,
		)

		if len(userSGs) > 0 {
			_, ok := r.waitNodePoolSecurityGroupsContains(ctx, plan.ClusterName.ValueString(), state.Name.ValueString(), userSGs, &resp.Diagnostics)
			if !ok || resp.Diagnostics.HasError() {
				return
			}
		}
	}

	didUpdateUserData := false
	if !config.UserData.IsUnknown() && !config.UserData.IsNull() {
		if state.UserData.IsUnknown() || state.UserData.IsNull() || state.UserData.ValueString() != config.UserData.ValueString() {

			script := kubernetesengine.NewNodePoolScriptRequestModel(config.UserData.ValueString())
			usrBody := kubernetesengine.UpdateK8sClusterNodePoolUserScriptRequestModel{NodePool: *script}

			_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics, func() (interface{}, *http.Response, error) {
				return r.kc.ApiClient.NodePoolsAPI.
					SetNodePoolUserScript(ctx, plan.ClusterName.ValueString(), state.Name.ValueString()).
					XAuthToken(r.kc.XAuthToken).
					UpdateK8sClusterNodePoolUserScriptRequestModel(usrBody).
					Execute()
			})
			if err != nil {
				common.AddApiActionError(ctx, r, httpResp, "SetNodePoolUserScript", err, &resp.Diagnostics)
				return
			}

			_, _ = r.waitNodePoolReadyOrFailed(
				ctx,
				plan.ClusterName.ValueString(),
				state.Name.ValueString(),
				&resp.Diagnostics,
			)
			didUpdateUserData = true
		}
	}

	if !config.Autoscaling.IsNull() && !config.Autoscaling.IsUnknown() && (state.IsBareMetal.IsNull() || state.IsBareMetal.IsUnknown() || !state.IsBareMetal.ValueBool()) {
		var auto NodePoolAutoscalingModel
		_ = config.Autoscaling.As(ctx, &auto, basetypes.ObjectAsOptions{})

		if ok, msg := validateAutoscalingModel(auto); !ok {
			resp.Diagnostics.AddError("Invalid autoscaling configuration", msg)
			return
		}

		scaling := kubernetesengine.NewNodePoolScalingResourceRequestModel(auto.IsAutoscalerEnable.ValueBool())

		if !auto.AutoscalerDesiredNodeCount.IsUnknown() {
			if auto.AutoscalerDesiredNodeCount.IsNull() {
				scaling.SetAutoscalerDesiredNodeCountNil()
			} else {
				scaling.SetAutoscalerDesiredNodeCount(auto.AutoscalerDesiredNodeCount.ValueInt32())
			}
		}
		if !auto.AutoscalerMaxNodeCount.IsUnknown() {
			if auto.AutoscalerMaxNodeCount.IsNull() {
				scaling.SetAutoscalerMaxNodeCountNil()
			} else {
				scaling.SetAutoscalerMaxNodeCount(auto.AutoscalerMaxNodeCount.ValueInt32())
			}
		}
		if !auto.AutoscalerMinNodeCount.IsUnknown() {
			if auto.AutoscalerMinNodeCount.IsNull() {
				scaling.SetAutoscalerMinNodeCountNil()
			} else {
				scaling.SetAutoscalerMinNodeCount(auto.AutoscalerMinNodeCount.ValueInt32())
			}
		}
		if !auto.AutoscalerScaleDownThreshold.IsUnknown() {
			if auto.AutoscalerScaleDownThreshold.IsNull() {
				scaling.SetAutoscalerScaleDownThresholdNil()
			} else {
				scaling.SetAutoscalerScaleDownThreshold(auto.AutoscalerScaleDownThreshold.ValueFloat32())
			}
		}
		if !auto.AutoscalerScaleDownUnneededTime.IsUnknown() {
			if auto.AutoscalerScaleDownUnneededTime.IsNull() {
				scaling.SetAutoscalerScaleDownUnneededTimeNil()
			} else {
				scaling.SetAutoscalerScaleDownUnneededTime(auto.AutoscalerScaleDownUnneededTime.ValueInt32())
			}
		}
		if !auto.AutoscalerScaleDownUnreadyTime.IsUnknown() {
			if auto.AutoscalerScaleDownUnreadyTime.IsNull() {
				scaling.SetAutoscalerScaleDownUnreadyTimeNil()
			} else {
				scaling.SetAutoscalerScaleDownUnreadyTime(auto.AutoscalerScaleDownUnreadyTime.ValueInt32())
			}
		}

		updBody := kubernetesengine.NewUpdateKubernetesEngineClusterNodePoolScalingResourceRequestModel(*scaling)

		_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics, func() (interface{}, *http.Response, error) {
			return r.kc.ApiClient.ScalingAPI.
				SetNodePoolResourceBasedAutoScaling(ctx, plan.ClusterName.ValueString(), state.Name.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				UpdateKubernetesEngineClusterNodePoolScalingResourceRequestModel(*updBody).
				Execute()
		})
		if err != nil {
			common.AddApiActionError(ctx, r, httpResp, "SetNodePoolResourceBasedAutoScaling", err, &resp.Diagnostics)
			return
		}

		_, _ = r.waitNodePoolReadyOrFailed(
			ctx,
			plan.ClusterName.ValueString(),
			state.Name.ValueString(),
			&resp.Diagnostics,
		)
	}

	finalResp, httpResp2, err2 := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (struct {
			NodePool kubernetesengine.KubernetesEngineV1ApiGetNodePoolModelNodePoolResponseModel
		}, *http.Response, error) {
			apiResp, httpResp, err := r.kc.ApiClient.NodePoolsAPI.
				GetNodePool(ctx, plan.ClusterName.ValueString(), state.Name.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				Execute()
			if err != nil {
				return struct {
					NodePool kubernetesengine.KubernetesEngineV1ApiGetNodePoolModelNodePoolResponseModel
				}{}, httpResp, err
			}
			return struct {
				NodePool kubernetesengine.KubernetesEngineV1ApiGetNodePoolModelNodePoolResponseModel
			}{NodePool: apiResp.NodePool}, httpResp, nil
		},
	)
	if err2 != nil {
		common.AddApiActionError(ctx, r, httpResp2, "GetNodePool", err2, &resp.Diagnostics)
		return
	}

	latest := finalResp.NodePool

	newState := state
	_ = mapNodePoolFromResponse(ctx, &newState.NodePoolBaseModel, &latest, &resp.Diagnostics, userSGs)

	if newState.Id.IsNull() || newState.Id.ValueString() == "" {
		newState.Id = types.StringValue(plan.ClusterName.ValueString() + "/" + state.Name.ValueString())
	}

	if didUpdateUserData {
		newState.UserData = config.UserData
	}

	newState.RequestSecurityGroups = plan.RequestSecurityGroups

	if latest.SecurityGroups != nil {
		setVal, diags := types.SetValueFrom(ctx, types.StringType, latest.SecurityGroups)
		resp.Diagnostics.Append(diags...)
		newState.SecurityGroups = setVal
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *nodePoolResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state NodePoolResourceModel
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

	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (interface{}, *http.Response, error) {
			httpResp, err := r.kc.ApiClient.NodePoolsAPI.DeleteNodePool(ctx, state.ClusterName.ValueString(), state.Name.ValueString()).XAuthToken(r.kc.XAuthToken).Execute()
			return nil, httpResp, err
		},
	)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			return
		}
		common.AddApiActionError(ctx, r, httpResp, "DeleteNodePool", err, &resp.Diagnostics)
		return
	}

	common.PollUntilDeletion(ctx, r, 10*time.Second, &resp.Diagnostics, func(ctx context.Context) (bool, *http.Response, error) {
		_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
			func() (interface{}, *http.Response, error) {
				_, hr, err := r.kc.ApiClient.NodePoolsAPI.
					GetNodePool(ctx, state.ClusterName.ValueString(), state.Name.ValueString()).
					XAuthToken(r.kc.XAuthToken).
					Execute()
				return nil, hr, err
			},
		)
		if httpResp != nil && httpResp.StatusCode == 404 {
			return true, httpResp, nil
		}
		return false, httpResp, err
	})
}

func (r *nodePoolResource) ValidateConfig(
	ctx context.Context,
	req resource.ValidateConfigRequest,
	resp *resource.ValidateConfigResponse,
) {
	var config NodePoolResourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *nodePoolResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if resp.Diagnostics.HasError() {
		return
	}

	if req.Plan.Raw.IsNull() {
		return
	}

	isCreate := req.State.Raw.IsNull()

	var plan NodePoolResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var config NodePoolResourceModel
	d := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state NodePoolResourceModel
	if !isCreate {
		_ = req.State.Get(ctx, &state)
	}
	effectiveAutoscaling := false

	if !isCreate && !state.IsBareMetal.IsNull() && !state.IsBareMetal.IsUnknown() && state.IsBareMetal.ValueBool() {

		if !config.Autoscaling.IsNull() && !config.Autoscaling.IsUnknown() {

			if !state.Autoscaling.IsNull() && !state.Autoscaling.IsUnknown() {
				resp.Plan.SetAttribute(ctx, path.Root("autoscaling"), state.Autoscaling)
			} else {

				resp.Plan.SetAttribute(ctx, path.Root("autoscaling"), types.ObjectNull(nodePoolAutoscalingAttrTypes))
			}

			resp.Diagnostics.AddAttributeError(
				path.Root("autoscaling"),
				"Autoscaling is not supported for bare metal node pools",
				"This node pool is bare metal. Remove the 'autoscaling' block from configuration.",
			)
		}

		if !config.VolumeSize.IsNull() && !config.VolumeSize.IsUnknown() {

			resp.Plan.SetAttribute(ctx, path.Root("volume_size"), state.VolumeSize)

			resp.Diagnostics.AddAttributeError(
				path.Root("volume_size"),
				"volume_size is not allowed for bare metal node pools",
				"Bare metal node pools do not support 'volume_size' configuration.",
			)
		}
		return
	}

	if isCreate {
		resp.Plan.SetAttribute(ctx, path.Root("status").AtName("available_nodes"), types.Int32Unknown())
		resp.Plan.SetAttribute(ctx, path.Root("status").AtName("unavailable_nodes"), types.Int32Unknown())
	}

	if !config.Autoscaling.IsNull() && !config.Autoscaling.IsUnknown() {
		var cfAuto NodePoolAutoscalingModel
		_ = config.Autoscaling.As(ctx, &cfAuto, basetypes.ObjectAsOptions{})
		if !cfAuto.IsAutoscalerEnable.IsNull() && !cfAuto.IsAutoscalerEnable.IsUnknown() {
			effectiveAutoscaling = cfAuto.IsAutoscalerEnable.ValueBool()
		}
	} else if !isCreate && !state.Autoscaling.IsNull() && !state.Autoscaling.IsUnknown() {

		var stAuto NodePoolAutoscalingModel
		_ = state.Autoscaling.As(ctx, &stAuto, basetypes.ObjectAsOptions{})
		if !stAuto.IsAutoscalerEnable.IsNull() && !stAuto.IsAutoscalerEnable.IsUnknown() {
			effectiveAutoscaling = stAuto.IsAutoscalerEnable.ValueBool()
		}
	}

	if effectiveAutoscaling {

		if !config.NodeCount.IsNull() && !config.NodeCount.IsUnknown() {
			resp.Diagnostics.AddAttributeError(
				path.Root("node_count"),
				"node_count must be omitted when autoscaling is enabled",
				"Autoscaling is enabled (either in current state or this plan). The autoscaler manages the actual node count. Remove 'node_count' from configuration or disable autoscaling to manage a fixed node count.",
			)
			return
		}
		if isCreate {

			var cfAuto NodePoolAutoscalingModel
			if !config.Autoscaling.IsNull() && !config.Autoscaling.IsUnknown() {
				_ = config.Autoscaling.As(ctx, &cfAuto, basetypes.ObjectAsOptions{})
				if !cfAuto.AutoscalerDesiredNodeCount.IsNull() && !cfAuto.AutoscalerDesiredNodeCount.IsUnknown() {
					resp.Plan.SetAttribute(ctx, path.Root("node_count"), types.Int32Value(cfAuto.AutoscalerDesiredNodeCount.ValueInt32()))
				} else {
					resp.Plan.SetAttribute(ctx, path.Root("node_count"), types.Int32Unknown())
				}
			} else {
				resp.Plan.SetAttribute(ctx, path.Root("node_count"), types.Int32Unknown())
			}

			resp.Plan.SetAttribute(ctx, path.Root("status").AtName("available_nodes"), types.Int32Unknown())
			resp.Plan.SetAttribute(ctx, path.Root("status").AtName("unavailable_nodes"), types.Int32Unknown())
		} else {

			resp.Plan.SetAttribute(ctx, path.Root("node_count"), types.Int32Unknown())
		}
	} else {
		prevAutoEnabled := false
		if !isCreate && !state.Autoscaling.IsNull() && !state.Autoscaling.IsUnknown() {
			var stAuto NodePoolAutoscalingModel
			_ = state.Autoscaling.As(ctx, &stAuto, basetypes.ObjectAsOptions{})
			if !stAuto.IsAutoscalerEnable.IsNull() && !stAuto.IsAutoscalerEnable.IsUnknown() {
				prevAutoEnabled = stAuto.IsAutoscalerEnable.ValueBool()
			}
		}

		if prevAutoEnabled && (config.NodeCount.IsNull() || config.NodeCount.IsUnknown()) {
			resp.Plan.SetAttribute(ctx, path.Root("node_count"), types.Int32Unknown())
		}
	}

	if !isCreate {

		resp.Plan.SetAttribute(ctx, path.Root("security_groups"), types.SetUnknown(types.StringType))

		baseDefaults := make(map[string]struct{})
		if !state.SecurityGroups.IsNull() && !state.SecurityGroups.IsUnknown() {
			var stSGs []string
			_ = state.SecurityGroups.ElementsAs(ctx, &stSGs, false)
			for _, id := range stSGs {
				baseDefaults[id] = struct{}{}
			}
		}
		if !state.RequestSecurityGroups.IsNull() && !state.RequestSecurityGroups.IsUnknown() {
			var prevReq []string
			_ = state.RequestSecurityGroups.ElementsAs(ctx, &prevReq, false)
			for _, id := range prevReq {
				delete(baseDefaults, id)
			}
		}

		if !plan.RequestSecurityGroups.IsNull() && !plan.RequestSecurityGroups.IsUnknown() {
			var reqSGs []string
			_ = plan.RequestSecurityGroups.ElementsAs(ctx, &reqSGs, false)
			for _, id := range reqSGs {
				baseDefaults[id] = struct{}{}
			}
		}

		combined := make([]string, 0, len(baseDefaults))
		for id := range baseDefaults {
			combined = append(combined, id)
		}

		setVal, _ := types.SetValueFrom(ctx, types.StringType, combined)
		resp.Plan.SetAttribute(ctx, path.Root("security_groups"), setVal)
	}

}

func (r *nodePoolResource) waitNodePoolReadyOrFailed(
	ctx context.Context,
	clusterName string,
	nodePoolName string,
	diagnostics *diag.Diagnostics,
) (*kubernetesengine.KubernetesEngineV1ApiGetNodePoolModelNodePoolResponseModel, bool) {
	result, ok := common.PollUntilResult(
		ctx,
		r,
		10*time.Second,
		[]string{
			string(kubernetesengine.NODEPOOLSTATUS_RUNNING),
			string(kubernetesengine.NODEPOOLSTATUS_RUNNING__SCHEDULING_DISABLE),
			string(kubernetesengine.NODEPOOLSTATUS_FAILED),
		},
		diagnostics,
		func(ctx context.Context) (*kubernetesengine.KubernetesEngineV1ApiGetNodePoolModelNodePoolResponseModel, *http.Response, error) {
			model, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, diagnostics,
				func() (struct {
					NodePool kubernetesengine.KubernetesEngineV1ApiGetNodePoolModelNodePoolResponseModel
				}, *http.Response, error) {
					apiResp, hr, err := r.kc.ApiClient.NodePoolsAPI.
						GetNodePool(ctx, clusterName, nodePoolName).
						XAuthToken(r.kc.XAuthToken).
						Execute()
					if err != nil {
						return struct {
							NodePool kubernetesengine.KubernetesEngineV1ApiGetNodePoolModelNodePoolResponseModel
						}{}, hr, err
					}
					return struct {
						NodePool kubernetesengine.KubernetesEngineV1ApiGetNodePoolModelNodePoolResponseModel
					}{NodePool: apiResp.NodePool}, hr, nil
				},
			)
			if err != nil {
				return nil, httpResp, err
			}
			return &model.NodePool, httpResp, nil
		},
		func(v *kubernetesengine.KubernetesEngineV1ApiGetNodePoolModelNodePoolResponseModel) string {
			return string(v.Status.Phase)
		},
	)
	return result, ok
}

func (r *nodePoolResource) waitNodePoolSecurityGroupsContains(
	ctx context.Context,
	clusterName string,
	nodePoolName string,
	required []string,
	diagnostics *diag.Diagnostics,
) (*kubernetesengine.KubernetesEngineV1ApiGetNodePoolModelNodePoolResponseModel, bool) {
	if len(required) == 0 {
		return nil, true
	}

	result, ok := common.PollUntilResult(
		ctx,
		r,
		5*time.Second,
		[]string{"ok"},
		diagnostics,
		func(ctx context.Context) (*kubernetesengine.KubernetesEngineV1ApiGetNodePoolModelNodePoolResponseModel, *http.Response, error) {
			model, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, diagnostics,
				func() (struct {
					NodePool kubernetesengine.KubernetesEngineV1ApiGetNodePoolModelNodePoolResponseModel
				}, *http.Response, error) {
					apiResp, hr, err := r.kc.ApiClient.NodePoolsAPI.
						GetNodePool(ctx, clusterName, nodePoolName).
						XAuthToken(r.kc.XAuthToken).
						Execute()
					if err != nil {
						return struct {
							NodePool kubernetesengine.KubernetesEngineV1ApiGetNodePoolModelNodePoolResponseModel
						}{}, hr, err
					}
					return struct {
						NodePool kubernetesengine.KubernetesEngineV1ApiGetNodePoolModelNodePoolResponseModel
					}{NodePool: apiResp.NodePool}, hr, nil
				},
			)
			if err != nil {
				return nil, httpResp, err
			}
			return &model.NodePool, httpResp, nil
		},
		func(v *kubernetesengine.KubernetesEngineV1ApiGetNodePoolModelNodePoolResponseModel) string {
			present := map[string]struct{}{}
			for _, id := range v.SecurityGroups {
				present[id] = struct{}{}
			}
			for _, want := range required {
				if _, ok := present[want]; !ok {
					return "waiting"
				}
			}
			return "ok"
		},
	)
	return result, ok
}

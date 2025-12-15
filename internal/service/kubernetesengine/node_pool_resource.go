// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package kubernetesengine

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"terraform-provider-kakaocloud/internal/common"
	"terraform-provider-kakaocloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/kakaoenterprise/kc-sdk-go/services/bcs"
	"github.com/kakaoenterprise/kc-sdk-go/services/kubernetesengine"
	"github.com/kakaoenterprise/kc-sdk-go/services/network"
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

	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		common.AddImportFormatError(ctx, r, &resp.Diagnostics,
			"Expected import ID in the format: cluster_name/node_pool_name")
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cluster_name"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), parts[1])...)
}

func (r *nodePoolResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan, config NodePoolResourceModel
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

	instanceType, ok := r.getInstanceTypeFromFlavor(ctx, plan.FlavorId.ValueString(), &resp.Diagnostics)
	if !ok {
		return
	}
	if *instanceType == common.InstanceTypeBM {
		r.validateBM(ctx, config, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	var initialNodeCount int32
	autoscalingEnabled := false
	var autoscalingPlan NodePoolAutoscalingModel

	if !plan.Autoscaling.IsNull() && !plan.Autoscaling.IsUnknown() {
		_ = plan.Autoscaling.As(ctx, &autoscalingPlan, basetypes.ObjectAsOptions{})
		autoscalingEnabled = autoscalingPlan.IsAutoscalerEnable.ValueBool()
	}

	if !plan.RequestNodeCount.IsNull() {
		initialNodeCount = plan.RequestNodeCount.ValueInt32()
	} else {
		initialNodeCount = autoscalingPlan.AutoscalerDesiredNodeCount.ValueInt32()
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
	if !plan.RequestSecurityGroups.IsNull() {
		_ = plan.RequestSecurityGroups.ElementsAs(ctx, &userSGs, false)
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

	if !plan.UserData.IsNull() {
		createModel.SetUserData(plan.UserData.ValueString())
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

	if userSGs != nil {
		createModel.SetSecurityGroups(userSGs)
	}

	reqBody := kubernetesengine.CreateK8sClusterNodePoolRequestModel{NodePool: *createModel}

	clusterResp, ok := r.pollClusterUtilAvailableStatus(ctx, plan.ClusterName.ValueString(), &resp.Diagnostics)
	if !ok {
		return
	}
	status := string(clusterResp.Status.Phase)
	common.CheckResourceAvailableStatus(ctx, r, &status, []string{common.ClusterStatusProvisioned, common.ClusterStatusFailed}, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.MinorVersion.IsNull() && !plan.MinorVersion.IsUnknown() {
		if plan.MinorVersion.ValueString() != clusterResp.Version.MinorVersion {
			common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
				fmt.Sprintf("The version is inconsistent with the cluster’s current version: '%v'", clusterResp.Version.MinorVersion))
			return
		}
	}

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

	result, ok := r.checkNodePoolReadyAndGetResult(ctx, plan.ClusterName.ValueString(), plan.Name.ValueString(), &resp.Diagnostics)
	if !ok {
		return
	}

	if len(userSGs) > 0 {
		_, ok2 := r.waitNodePoolSecurityGroupsContains(ctx, plan.ClusterName.ValueString(), plan.Name.ValueString(), userSGs, &resp.Diagnostics)
		if !ok2 || resp.Diagnostics.HasError() {
			return
		}
	}

	if autoscalingEnabled {
		result = r.updateAutoScaling(ctx, &plan, autoscalingPlan, &resp.Diagnostics)
		if result == nil || resp.Diagnostics.HasError() {
			return
		}
	}

	ok = r.mapNodePoolResource(ctx, &plan, result, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

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
		func() (*kubernetesengine.GetK8sClusterNodePoolResponseModel, *http.Response, error) {
			return r.kc.ApiClient.NodePoolsAPI.
				GetNodePool(ctx, state.ClusterName.ValueString(), state.Name.ValueString()).
				XAuthToken(r.kc.XAuthToken).Execute()
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
	ok := r.mapNodePoolResource(ctx, &state, &result, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	if state.RequestNodeCount.IsNull() {
		if !result.Autoscaling.IsAutoscalerEnable {
			state.RequestNodeCount = types.Int32Value(result.NodeCount)
		}
	}

	if state.ImageId.IsNull() {
		state.ImageId = types.StringValue(result.Image.Id)
	}

	if state.RequestSecurityGroups.IsNull() {
		sgIDs := result.SecurityGroups
		if len(sgIDs) > 1 {

			const defaultPrefix = "k8s-cluster-ke-cluster"
			userSGs := make([]string, 0, len(sgIDs))

			for _, sgID := range sgIDs {
				sgResp, httpResp, e := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
					func() (*network.BnsNetworkV1ApiGetSecurityGroupModelResponseSecurityGroupModel, *http.Response, error) {
						return r.kc.ApiClient.SecurityGroupAPI.
							GetSecurityGroup(ctx, sgID).
							XAuthToken(r.kc.XAuthToken).
							Execute()
					},
				)
				if e != nil {
					common.AddApiActionError(ctx, r, httpResp, "GetSecurityGroup", e, &resp.Diagnostics)
					continue
				}

				var nameStr string
				if sgResp.SecurityGroup.Name.IsSet() && sgResp.SecurityGroup.Name.Get() != nil {
					nameStr = *sgResp.SecurityGroup.Name.Get()
				}

				if !strings.HasPrefix(nameStr, defaultPrefix) {

					userSGs = append(userSGs, sgID)
				}
			}

			setVal, setDiags := types.SetValueFrom(ctx, types.StringType, userSGs)
			diags.Append(setDiags...)
			state.RequestSecurityGroups = setVal
		}
	}

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

	timeout, diags := plan.Timeouts.Update(ctx, common.DefaultUpdateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var result *kubernetesengine.KubernetesEngineV1ApiGetNodePoolModelNodePoolResponseModel
	var ok bool

	_, ok = r.checkNodePoolReadyAndGetResult(ctx, plan.ClusterName.ValueString(), plan.Name.ValueString(), &resp.Diagnostics)
	if !ok {
		return
	}

	if !plan.MinorVersion.IsNull() && !plan.MinorVersion.IsUnknown() && !plan.MinorVersion.Equal(state.MinorVersion) {
		clusterResp, ok := r.pollClusterUtilAvailableStatus(ctx, plan.ClusterName.ValueString(), &resp.Diagnostics)
		if !ok {
			return
		}
		status := string(clusterResp.Status.Phase)
		common.CheckResourceAvailableStatus(ctx, r, &status, []string{common.ClusterStatusProvisioned, common.ClusterStatusFailed}, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}

		if plan.MinorVersion.ValueString() != clusterResp.Version.MinorVersion {
			common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
				fmt.Sprintf("The version is inconsistent with the cluster’s current version: '%v'", clusterResp.Version.MinorVersion))
			return
		}

		_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
			func() (interface{}, *http.Response, error) {
				return r.kc.ApiClient.NodePoolsAPI.UpgradeNodePool(ctx, plan.ClusterName.ValueString(), plan.Name.ValueString()).
					XAuthToken(r.kc.XAuthToken).
					Execute()
			},
		)
		if err != nil {
			common.AddApiActionError(ctx, r, httpResp, "UpgradeNodePool", err, &resp.Diagnostics)
			return
		}

		time.Sleep(5 * time.Second)
		_, ok = r.checkNodePoolReadyAndGetResult(ctx, plan.ClusterName.ValueString(), plan.Name.ValueString(), &resp.Diagnostics)
		if !ok {
			return
		}
	}

	if !plan.UserData.Equal(state.UserData) {
		var userData string
		if plan.UserData.IsNull() {
			userData = ""
		} else {
			userData = plan.UserData.ValueString()
		}

		script := kubernetesengine.NewNodePoolScriptRequestModel(userData)
		usrBody := kubernetesengine.UpdateK8sClusterNodePoolUserScriptRequestModel{NodePool: *script}

		_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics, func() (interface{}, *http.Response, error) {
			return r.kc.ApiClient.NodePoolsAPI.
				SetNodePoolUserScript(ctx, plan.ClusterName.ValueString(), plan.Name.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				UpdateK8sClusterNodePoolUserScriptRequestModel(usrBody).
				Execute()
		})
		if err != nil {
			common.AddApiActionError(ctx, r, httpResp, "SetNodePoolUserScript", err, &resp.Diagnostics)
			return
		}

		_, ok = r.checkNodePoolReadyAndGetResult(ctx, plan.ClusterName.ValueString(), plan.Name.ValueString(), &resp.Diagnostics)
		if !ok {
			return
		}
	}

	if !plan.Autoscaling.Equal(state.Autoscaling) {
		needUpdateAutoscaling := false

		var plAuto NodePoolAutoscalingModel
		if plan.Autoscaling.IsNull() || plan.Autoscaling.IsUnknown() {
			if !state.Autoscaling.IsNull() && !state.Autoscaling.IsUnknown() {
				var stAuto NodePoolAutoscalingModel
				_ = state.Autoscaling.As(ctx, &stAuto, basetypes.ObjectAsOptions{})
				if stAuto.IsAutoscalerEnable.ValueBool() {
					plAuto.IsAutoscalerEnable = types.BoolValue(false)
					needUpdateAutoscaling = true
				}
			}
		} else {
			_ = plan.Autoscaling.As(ctx, &plAuto, basetypes.ObjectAsOptions{})
			needUpdateAutoscaling = true
		}

		if needUpdateAutoscaling {
			result = r.updateAutoScaling(ctx, &plan, plAuto, &resp.Diagnostics)
			if result == nil || resp.Diagnostics.HasError() {
				return
			}
		}

		_, ok = r.checkNodePoolReadyAndGetResult(ctx, plan.ClusterName.ValueString(), plan.Name.ValueString(), &resp.Diagnostics)
		if !ok {
			return
		}
	}

	upd := kubernetesengine.NewKubernetesEngineV1ApiUpdateNodePoolModelNodePoolRequestModel()
	needUpdate := false

	if !plan.Description.IsNull() && !plan.Description.IsUnknown() && !plan.Description.Equal(state.Description) {
		upd.SetDescription(plan.Description.ValueString())
		needUpdate = true
	}

	if !plan.RequestNodeCount.IsNull() && !plan.RequestNodeCount.Equal(state.RequestNodeCount) &&
		!plan.RequestNodeCount.Equal(state.NodeCount) {
		upd.SetNodeCount(plan.RequestNodeCount.ValueInt32())
		needUpdate = true
	}

	var userSGs []string
	if !plan.RequestSecurityGroups.Equal(state.RequestSecurityGroups) {
		if !plan.RequestSecurityGroups.IsNull() {

			_ = plan.RequestSecurityGroups.ElementsAs(ctx, &userSGs, false)
		}
		if len(userSGs) > 0 {
			upd.SetSecurityGroups(userSGs)
		} else {
			upd.SetSecurityGroups([]string{})
		}
		needUpdate = true
	}

	if needUpdate {
		reqBody := kubernetesengine.UpdateK8sClusterNodePoolRequestModel{NodePool: *upd}
		_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics, func() (interface{}, *http.Response, error) {
			return r.kc.ApiClient.NodePoolsAPI.
				UpdateNodePool(ctx, plan.ClusterName.ValueString(), plan.Name.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				UpdateK8sClusterNodePoolRequestModel(reqBody).
				Execute()
		})
		if err != nil {
			common.AddApiActionError(ctx, r, httpResp, "UpdateNodePool", err, &resp.Diagnostics)
			return
		}

		_, ok = r.checkNodePoolReadyAndGetResult(ctx, plan.ClusterName.ValueString(), plan.Name.ValueString(), &resp.Diagnostics)
		if !ok {
			return
		}

		if len(userSGs) > 0 {
			_, ok := r.waitNodePoolSecurityGroupsContains(ctx, plan.ClusterName.ValueString(), plan.Name.ValueString(), userSGs, &resp.Diagnostics)
			if !ok || resp.Diagnostics.HasError() {
				return
			}
		}
	}

	if !plan.Labels.Equal(state.Labels) {

		planLabels := map[string]string{}
		stateLabels := map[string]string{}

		var labels []kubernetesengine.LabelRequestModel
		var pl []NodePoolLabelModel
		if !plan.Labels.IsNull() {
			if err := plan.Labels.ElementsAs(ctx, &pl, false); err != nil {
				resp.Diagnostics.AddError(
					"Invalid Plan Labels",
					fmt.Sprintf("Failed to parse labels from plan: %s", err),
				)
				return
			}
			for _, l := range pl {
				planLabels[l.Key.ValueString()] = l.Value.ValueString()
				labels = append(labels, kubernetesengine.LabelRequestModel{
					Key:   l.Key.ValueString(),
					Value: l.Value.ValueString(),
				})
			}
		}

		if !state.Labels.IsNull() {
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

		removeKeys := make([]string, 0)
		for k := range stateLabels {
			if _, ok := planLabels[k]; !ok {
				removeKeys = append(removeKeys, k)
			}
		}

		labelsReq := kubernetesengine.NewNodeLabelsRequestModel()
		labelsReq.SetAddOrUpdateLabels(labels)
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

		_, ok = r.checkNodePoolReadyAndGetResult(ctx, plan.ClusterName.ValueString(), plan.Name.ValueString(), &resp.Diagnostics)
		if !ok {
			return
		}
	}

	time.Sleep(5 * time.Second)
	result, ok = r.checkNodePoolReadyAndGetResult(ctx, plan.ClusterName.ValueString(), plan.Name.ValueString(), &resp.Diagnostics)
	if !ok {
		return
	}

	ok = r.mapNodePoolResource(ctx, &plan, result, &resp.Diagnostics)
	if !ok {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
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

	result, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*kubernetesengine.GetK8sClusterNodePoolResponseModel, *http.Response, error) {
			return r.kc.ApiClient.NodePoolsAPI.
				GetNodePool(ctx, state.ClusterName.ValueString(), state.Name.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "GetNodePool", err, &resp.Diagnostics)
		return
	}

	if result.NodePool.Status.Phase != kubernetesengine.NODEPOOLSTATUS_DELETING {
		if result.NodePool.Status.Phase != kubernetesengine.NODEPOOLSTATUS_RUNNING &&
			result.NodePool.Status.Phase != kubernetesengine.NODEPOOLSTATUS_RUNNING__SCHEDULING_DISABLE &&
			result.NodePool.Status.Phase != kubernetesengine.NODEPOOLSTATUS_FAILED &&
			result.NodePool.Status.Phase != kubernetesengine.NODEPOOLSTATUS_PENDING {
			_, ok := waitNodePool(
				ctx,
				r.kc,
				r,
				state.ClusterName.ValueString(),
				state.Name.ValueString(),
				NodePoolStatusesReadyToDelete,
				&resp.Diagnostics,
			)
			if !ok {
				return
			}
		}

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

	if !config.Autoscaling.IsNull() && !config.Autoscaling.IsUnknown() {
		var cfAuto NodePoolAutoscalingModel
		diags = config.Autoscaling.As(ctx, &cfAuto, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		if cfAuto.IsAutoscalerEnable.ValueBool() && !config.RequestNodeCount.IsNull() {
			common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
				"request_node_count must be omitted when autoscaling is enabled")
			return
		}
		if !cfAuto.IsAutoscalerEnable.ValueBool() && !cfAuto.IsAutoscalerEnable.IsUnknown() && config.RequestNodeCount.IsNull() {
			common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
				"request_node_count must be configured when autoscaling is disabled")
			return
		}

		r.validateAutoscalingModel(ctx, cfAuto, resp)
		if resp.Diagnostics.HasError() {
			return
		}
	} else {
		if config.RequestNodeCount.IsNull() {
			common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
				"request_node_count must be configured when autoscaling is disabled")
			return
		}
	}
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
		"node pool",
		nodePoolName,
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

func (r *nodePoolResource) checkNodePoolReadyAndGetResult(
	ctx context.Context,
	clusterName, nodePoolName string,
	diags *diag.Diagnostics,
) (*kubernetesengine.KubernetesEngineV1ApiGetNodePoolModelNodePoolResponseModel, bool) {
	result, ok := waitNodePool(
		ctx,
		r.kc,
		r,
		clusterName,
		nodePoolName,
		NodePoolStatusesReadyOrFailed,
		diags,
	)
	if !ok || diags.HasError() {
		return nil, false
	}
	status := string(result.Status.Phase)
	common.CheckResourceAvailableStatus(ctx, r, &status,
		[]string{string(kubernetesengine.NODEPOOLSTATUS_RUNNING),
			string(kubernetesengine.NODEPOOLSTATUS_RUNNING__SCHEDULING_DISABLE)},
		diags)
	if diags.HasError() {
		return nil, false
	}
	return result, true
}

func (r *nodePoolResource) getInstanceTypeFromFlavor(
	ctx context.Context,
	flavorId string,
	respDiags *diag.Diagnostics,
) (*bcs.InstanceType, bool) {
	flavorResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, respDiags,
		func() (*bcs.ResponseFlavorModel, *http.Response, error) {
			return r.kc.ApiClient.FlavorAPI.GetInstanceType(ctx, flavorId).XAuthToken(r.kc.XAuthToken).Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "GetInstanceType", err, respDiags)
		return nil, false
	}
	return flavorResp.Flavor.InstanceType.Get(), true
}

func (r *nodePoolResource) updateAutoScaling(
	ctx context.Context,
	plan *NodePoolResourceModel,
	autoscalingPlan NodePoolAutoscalingModel,
	diags *diag.Diagnostics,
) *kubernetesengine.KubernetesEngineV1ApiGetNodePoolModelNodePoolResponseModel {
	autoscalingEnabled := autoscalingPlan.IsAutoscalerEnable.ValueBool()
	scaling := kubernetesengine.NewNodePoolScalingResourceRequestModel(autoscalingEnabled)

	if autoscalingEnabled {
		scaling.SetAutoscalerDesiredNodeCount(autoscalingPlan.AutoscalerDesiredNodeCount.ValueInt32())
		scaling.SetAutoscalerMaxNodeCount(autoscalingPlan.AutoscalerMaxNodeCount.ValueInt32())
		scaling.SetAutoscalerMinNodeCount(autoscalingPlan.AutoscalerMinNodeCount.ValueInt32())
		scaling.SetAutoscalerScaleDownThreshold(autoscalingPlan.AutoscalerScaleDownThreshold.ValueFloat32())
		scaling.SetAutoscalerScaleDownUnneededTime(autoscalingPlan.AutoscalerScaleDownUnneededTime.ValueInt32())
		scaling.SetAutoscalerScaleDownUnreadyTime(autoscalingPlan.AutoscalerScaleDownUnreadyTime.ValueInt32())
	}

	updBody := kubernetesengine.NewUpdateKubernetesEngineClusterNodePoolScalingResourceRequestModel(*scaling)
	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, diags, func() (interface{}, *http.Response, error) {
		return r.kc.ApiClient.ScalingAPI.
			SetNodePoolResourceBasedAutoScaling(ctx, plan.ClusterName.ValueString(), plan.Name.ValueString()).
			XAuthToken(r.kc.XAuthToken).
			UpdateKubernetesEngineClusterNodePoolScalingResourceRequestModel(*updBody).
			Execute()
	})
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "SetNodePoolResourceBasedAutoScaling", err, diags)
		return nil
	}

	result, ok := r.checkNodePoolReadyAndGetResult(ctx, plan.ClusterName.ValueString(), plan.Name.ValueString(), diags)
	if !ok {
		return nil
	}

	return result
}

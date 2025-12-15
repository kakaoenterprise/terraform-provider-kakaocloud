// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package kubernetesengine

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-kakaocloud/internal/common"
	"terraform-provider-kakaocloud/internal/utils"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/kakaoenterprise/kc-sdk-go/services/kubernetesengine"
)

var (
	_ resource.ResourceWithConfigure      = &clusterResource{}
	_ resource.ResourceWithImportState    = &clusterResource{}
	_ resource.ResourceWithValidateConfig = &clusterResource{}
	_ resource.ResourceWithModifyPlan     = &clusterResource{}
)

func NewClusterResource() resource.Resource { return &clusterResource{} }

type clusterResource struct {
	kc *common.KakaoCloudClient
}

func (r *clusterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func (r *clusterResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kubernetes_engine_cluster"
}

func (r *clusterResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: utils.MergeResourceSchemaAttributes(
			clusterResourceSchemaAttributes,
			map[string]schema.Attribute{
				"timeouts": timeouts.AttributesAll(ctx),
			},
		),
	}
}

func (r *clusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan clusterResourceModel
	diags := req.Plan.Get(ctx, &plan)
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

	var ver OmtInfoModel
	diags = plan.Version.As(ctx, &ver, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := kubernetesengine.KubernetesEngineV1ApiCreateClusterModelClusterRequestModel{
		Name:          plan.Name.ValueString(),
		Version:       ver.MinorVersion.ValueString(),
		IsAllocateFip: plan.IsAllocateFip.ValueBool(),
	}

	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		createReq.SetDescription(plan.Description.ValueString())
	}

	var vpcInfo VpcInfoModel
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

	createReq.SetVpcInfo(reqVpcInfo)

	var targetNetwork targetNetworkModel
	diags = plan.Network.As(ctx, &targetNetwork, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	n := kubernetesengine.ClusterNetworkRequestModel{}
	n.SetCni(kubernetesengine.ClusterNetworkCNI(targetNetwork.Cni.ValueString()))
	if !targetNetwork.ServiceCidr.IsNull() && !targetNetwork.ServiceCidr.IsUnknown() {
		n.SetServiceCidr(targetNetwork.ServiceCidr.ValueString())
	}
	if !targetNetwork.PodCidr.IsNull() && !targetNetwork.PodCidr.IsUnknown() {
		n.SetPodCidr(targetNetwork.PodCidr.ValueString())
	}
	createReq.SetNetwork(n)

	body := kubernetesengine.CreateK8sClusterRequestModel{Cluster: createReq}

	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (interface{}, *http.Response, error) {
			return r.kc.ApiClient.ClustersAPI.
				CreateCluster(ctx).
				XAuthToken(r.kc.XAuthToken).CreateK8sClusterRequestModel(body).Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "CreateCluster", err, &resp.Diagnostics)
		return
	}

	result, ok := r.pollClusterUtilStatus(
		ctx,
		plan.Name.ValueString(),
		[]string{common.ClusterStatusProvisioned, common.ClusterStatusFailed, common.ClusterStatusDeleting},
		&resp.Diagnostics,
	)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	status := string(result.Status.Phase)
	common.CheckResourceAvailableStatus(ctx, r, &status, []string{common.ClusterStatusProvisioned}, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	ok = mapClusterBaseModel(ctx, &plan.ClusterBaseModel, result, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *clusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state clusterResourceModel
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

	respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*kubernetesengine.GetK8sClusterResponseModel, *http.Response, error) {
			return r.kc.ApiClient.ClustersAPI.
				GetCluster(ctx, state.Name.ValueString()).
				XAuthToken(r.kc.XAuthToken).Execute()
		},
	)

	if httpResp != nil && httpResp.StatusCode == 404 {
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "GetCluster", err, &resp.Diagnostics)
		return
	}

	result := respModel.Cluster
	ok := mapClusterBaseModel(ctx, &state.ClusterBaseModel, &result, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *clusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state clusterResourceModel
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

	result, ok := r.pollClusterUtilStatus(
		ctx,
		plan.Name.ValueString(),
		[]string{common.ClusterStatusProvisioned, common.ClusterStatusFailed, common.ClusterStatusDeleting},
		&resp.Diagnostics,
	)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	status := string(result.Status.Phase)
	common.CheckResourceAvailableStatus(ctx, r, &status, []string{common.ClusterStatusProvisioned}, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Description.Equal(state.Description) && !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		editReq := kubernetesengine.KubernetesEngineV1ApiUpdateClusterModelClusterRequestModel{}
		editReq.SetDescription(plan.Description.ValueString())

		body := *kubernetesengine.NewUpdateK8sClusterRequestModel(editReq)
		_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
			func() (interface{}, *http.Response, error) {
				return r.kc.ApiClient.ClustersAPI.UpdateCluster(ctx, plan.Name.ValueString()).
					XAuthToken(r.kc.XAuthToken).
					UpdateK8sClusterRequestModel(body).
					Execute()
			},
		)

		if err != nil {
			common.AddApiActionError(ctx, r, httpResp, "UpdateCluster", err, &resp.Diagnostics)
			return
		}
	}

	var planVer OmtInfoModel
	diags = plan.Version.As(ctx, &planVer, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var stateVer OmtInfoModel
	diags = state.Version.As(ctx, &stateVer, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !planVer.MinorVersion.Equal(stateVer.MinorVersion) {
		_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
			func() (interface{}, *http.Response, error) {
				return r.kc.ApiClient.ClustersAPI.UpgradeCluster(ctx, plan.Name.ValueString()).
					XAuthToken(r.kc.XAuthToken).
					Execute()
			},
		)
		if err != nil {
			common.AddApiActionError(ctx, r, httpResp, "UpgradeCluster", err, &resp.Diagnostics)
			return
		}

		time.Sleep(5 * time.Second)
	}

	result, ok = r.pollClusterUtilStatus(
		ctx,
		plan.Name.ValueString(),
		[]string{common.ClusterStatusProvisioned, common.ClusterStatusFailed, common.ClusterStatusDeleting},
		&resp.Diagnostics,
	)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	status = string(result.Status.Phase)
	common.CheckResourceAvailableStatus(ctx, r, &status, []string{common.ClusterStatusProvisioned}, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	ok = mapClusterBaseModel(ctx, &state.ClusterBaseModel, result, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	if !plan.Description.Equal(state.Description) && !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		state.Description = plan.Description
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *clusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state clusterResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.Name.IsNull() || state.Name.IsUnknown() {
		resp.State.RemoveResource(ctx)
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
			httpResp, err := r.kc.ApiClient.ClustersAPI.DeleteCluster(ctx, state.Name.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				Execute()
			return nil, httpResp, err
		},
	)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			return
		}
		common.AddApiActionError(ctx, r, httpResp, "DeleteCluster", err, &resp.Diagnostics)
		return
	}

	common.PollUntilDeletion(ctx, r, 2*time.Second, &resp.Diagnostics, func(ctx context.Context) (bool, *http.Response, error) {
		_, httpResp, err := r.kc.ApiClient.ClustersAPI.
			GetCluster(ctx, state.Name.ValueString()).
			XAuthToken(r.kc.XAuthToken).
			Execute()

		if httpResp != nil && httpResp.StatusCode == 404 {
			return true, httpResp, nil
		}
		return false, httpResp, err
	})
}

func (r *clusterResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *clusterResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config clusterResourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var targetNetwork targetNetworkModel
	diags = config.Network.As(ctx, &targetNetwork, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var podCidr, serviceCidr *string
	if !targetNetwork.PodCidr.IsNull() && !targetNetwork.PodCidr.IsUnknown() {
		podCidr = targetNetwork.PodCidr.ValueStringPointer()
	}
	if !targetNetwork.ServiceCidr.IsNull() && !targetNetwork.ServiceCidr.IsUnknown() {
		serviceCidr = targetNetwork.ServiceCidr.ValueStringPointer()
	}

	if podCidr == nil && serviceCidr == nil {
		return
	}

	privateRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
	}

	if podCidr != nil {
		common.CidrContainListValidator(*podCidr, privateRanges, "pod_cidr", "Private Network", &resp.Diagnostics)
	}
	if serviceCidr != nil {
		common.CidrContainListValidator(*serviceCidr, privateRanges, "service_cidr", "Private Network", &resp.Diagnostics)
	}
	if podCidr != nil && serviceCidr != nil {
		common.CidrOverlapValidator(*podCidr, *serviceCidr, "pod_cidr", "service_cidr", &resp.Diagnostics)
	}
}

func (r *clusterResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var plan, state *clusterResourceModel

	planDiags := req.Plan.Get(ctx, &plan)
	stateDiags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(planDiags...)
	resp.Diagnostics.Append(stateDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if req.Plan.Raw.IsNull() {
		return
	}

	if req.State.Raw.IsNull() && !req.Plan.Raw.IsNull() {
		return
	}

	if !req.State.Raw.IsNull() && !req.Plan.Raw.IsNull() {
		var planVer OmtInfoModel
		diags := plan.Version.As(ctx, &planVer, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		var stateVer OmtInfoModel
		diags = state.Version.As(ctx, &stateVer, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		if !planVer.MinorVersion.Equal(stateVer.MinorVersion) {
			common.MajorMinorVersionNotDecreasingValidator(planVer.MinorVersion.ValueString(), stateVer.MinorVersion.ValueString(), &resp.Diagnostics)
			if resp.Diagnostics.HasError() {
				return
			}

			if !state.IsUpgradable.ValueBool() {
				common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
					fmt.Sprintf("The version cannot be upgraded. current state version: '%v'", stateVer.MinorVersion.ValueString()))
				return
			}
			if !planVer.MinorVersion.Equal(stateVer.NextVersion) {
				common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
					fmt.Sprintf("Upgrade to the requested version is not supported. The available version for upgrade is '%v'.", stateVer.NextVersion.ValueString()),
				)
				return
			}
		}
	}
}

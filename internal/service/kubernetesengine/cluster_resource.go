// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package kubernetesengine

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-kakaocloud/internal/common"
	"terraform-provider-kakaocloud/internal/docs"
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
)

func NewClusterResource() resource.Resource { return &clusterResource{} }

type clusterResource struct {
	kc *common.KakaoCloudClient
}

func (r *clusterResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config clusterResourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *clusterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func (r *clusterResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kubernetes_engine_cluster"
}

func (r *clusterResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: docs.GetResourceDescription("KubernetesEngineCluster"),
		Attributes: utils.MergeResourceSchemaAttributes(
			clusterResourceSchemaAttributes,
			map[string]schema.Attribute{
				"timeouts": timeouts.AttributesAll(ctx),
			},
		),
	}
}

func (r *clusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan, config clusterResourceModel
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

	if !plan.Network.IsNull() && !plan.Network.IsUnknown() {
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
	}

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
		[]string{ClusterStatusProvisioned, ClusterStatusFailed},
		&resp.Diagnostics,
	)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	status := string(result.Status.Phase)
	common.CheckResourceAvailableStatus(ctx, r, &status, []string{ClusterStatusProvisioned}, &resp.Diagnostics)

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

	if !plan.Description.Equal(state.Description) {
		editReq := kubernetesengine.KubernetesEngineV1ApiUpdateClusterModelClusterRequestModel{}
		if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
			editReq.SetDescription(plan.Description.ValueString())
		}

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

	respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*kubernetesengine.GetK8sClusterResponseModel, *http.Response, error) {
			return r.kc.ApiClient.ClustersAPI.
				GetCluster(ctx, state.Name.ValueString()).
				XAuthToken(r.kc.XAuthToken).Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "GetCluster", err, &resp.Diagnostics)
		return
	}

	result := respModel.Cluster
	ok := mapClusterBaseModel(ctx, &state.ClusterBaseModel, &result, &resp.Diagnostics)
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

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package kubernetesengine

import (
	"context"

	. "terraform-provider-kakaocloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kakaoenterprise/kc-sdk-go/services/kubernetesengine"
)

func mapClusterBaseModel(
	ctx context.Context,
	base *ClusterBaseModel,
	clusterResult *kubernetesengine.KubernetesEngineV1ApiGetClusterModelClusterResponseModel,
	respDiags *diag.Diagnostics,
) bool {
	base.IsAllocateFip = types.BoolValue(clusterResult.IsAllocateFip)
	base.ApiVersion = types.StringValue(clusterResult.ApiVersion)
	base.CreatedAt = ConvertNullableTime(clusterResult.CreatedAt)
	base.Description = types.StringValue(clusterResult.Description)
	base.Id = types.StringValue(clusterResult.Id)
	base.Name = types.StringValue(clusterResult.Name)
	base.FailureMessage = ConvertNullableString(clusterResult.FailureMessage)
	base.IsUpgradable = types.BoolValue(clusterResult.IsUpgradable)

	networkObj, networkDiags := ConvertObjectFromModel(
		ctx,
		clusterResult.Network,
		clusterNetworkAttrTypes,
		func(src kubernetesengine.ClusterNetworkResponseModel) any {
			return ClusterNetworkModel{
				Cni:         types.StringValue(string(src.Cni)),
				PodCidr:     types.StringValue(src.PodCidr),
				ServiceCidr: types.StringValue(src.ServiceCidr),
			}
		},
	)

	respDiags.Append(networkDiags...)
	base.Network = networkObj
	if respDiags.HasError() {
		return false
	}

	cpeVal, cpeDiags := types.ObjectValueFrom(ctx, controlPlaneEndpointAttrTypes, ControlPlaneEndpointModel{
		Host: types.StringValue(clusterResult.ControlPlaneEndpoint.Host),
		Port: types.Int32Value(clusterResult.ControlPlaneEndpoint.Port),
	})
	respDiags.Append(cpeDiags...)
	base.ControlPlaneEndpoint = cpeVal
	if respDiags.HasError() {
		return false
	}

	creatorVal, creatorDiags := types.ObjectValueFrom(ctx, creatorInfoAttrTypes, CreatorInfoModel{
		Id:   types.StringValue(clusterResult.CreatorInfo.Id),
		Name: types.StringValue(clusterResult.CreatorInfo.Name),
	})
	respDiags.Append(creatorDiags...)
	base.CreatorInfo = creatorVal
	if respDiags.HasError() {
		return false
	}

	versionVal, versionDiags := types.ObjectValueFrom(ctx, omtInfoAttrTypes, OmtInfoModel{
		IsDeprecated: types.BoolValue(clusterResult.Version.IsDeprecated),
		Eol:          types.StringValue(clusterResult.Version.Eol),
		MinorVersion: types.StringValue(clusterResult.Version.MinorVersion),
		NextVersion:  types.StringValue(clusterResult.Version.NextVersion),
		PatchVersion: types.StringValue(clusterResult.Version.PatchVersion),
	})
	respDiags.Append(versionDiags...)
	base.Version = versionVal
	if respDiags.HasError() {
		return false
	}

	statusVal, statusDiags := types.ObjectValueFrom(ctx, statusAttrTypes, StatusModel{
		Phase: types.StringValue(string(clusterResult.Status.Phase)),
	})
	respDiags.Append(statusDiags...)
	base.Status = statusVal
	if respDiags.HasError() {
		return false
	}

	subnetVals := make([]attr.Value, 0, len(clusterResult.VpcInfo.Subnets))
	for _, s := range clusterResult.VpcInfo.Subnets {
		obj, objDiags := types.ObjectValue(
			subnetAttrTypes,
			map[string]attr.Value{
				"availability_zone": types.StringValue(string(s.AvailabilityZone)),
				"cidr_block":        types.StringValue(s.CidrBlock),
				"id":                types.StringValue(s.Id),
			},
		)
		respDiags.Append(objDiags...)
		subnetVals = append(subnetVals, obj)
	}

	subnetsSet, setDiags := types.SetValue(
		types.ObjectType{AttrTypes: subnetAttrTypes},
		subnetVals,
	)
	respDiags.Append(setDiags...)

	vpcVal, vpcDiags := types.ObjectValue(
		vpcInfoAttrTypes,
		map[string]attr.Value{
			"id":      types.StringValue(clusterResult.VpcInfo.Id),
			"subnets": subnetsSet,
		},
	)
	respDiags.Append(vpcDiags...)
	base.VpcInfo = vpcVal

	return true
}

func (d *clusterDataSource) mapCluster(
	ctx context.Context,
	model *clusterDataSourceModel,
	clusterResult *kubernetesengine.KubernetesEngineV1ApiGetClusterModelClusterResponseModel,
	respDiags *diag.Diagnostics,
) bool {
	mapClusterBaseModel(ctx, &model.ClusterBaseModel, clusterResult, respDiags)

	if respDiags.HasError() {
		return false
	}

	return true
}

func (d *clustersDataSource) mapClusters(
	ctx context.Context,
	base *ClusterBaseModel,
	clusterResult *kubernetesengine.KubernetesEngineV1ApiListClustersModelClusterResponseModel,
	respDiags *diag.Diagnostics,
) bool {
	base.IsAllocateFip = types.BoolValue(clusterResult.IsAllocateFip)
	base.ApiVersion = types.StringValue(clusterResult.ApiVersion)
	base.CreatedAt = ConvertNullableTime(clusterResult.CreatedAt)
	base.Description = types.StringValue(clusterResult.Description)
	base.Id = types.StringValue(clusterResult.Id)
	base.Name = types.StringValue(clusterResult.Name)
	base.FailureMessage = ConvertNullableString(clusterResult.FailureMessage)
	base.IsUpgradable = types.BoolValue(clusterResult.IsUpgradable)

	networkObj, networkDiags := ConvertObjectFromModel(
		ctx,
		clusterResult.Network,
		clusterNetworkAttrTypes,
		func(src kubernetesengine.ClusterNetworkResponseModel) any {
			return ClusterNetworkModel{
				PodCidr:     types.StringValue(src.PodCidr),
				ServiceCidr: types.StringValue(src.ServiceCidr),
			}
		},
	)

	respDiags.Append(networkDiags...)
	base.Network = networkObj

	if respDiags.HasError() {
		return false
	}

	cpeVal, cpeDiags := types.ObjectValueFrom(ctx, controlPlaneEndpointAttrTypes, ControlPlaneEndpointModel{
		Host: types.StringValue(clusterResult.ControlPlaneEndpoint.Host),
		Port: types.Int32Value(clusterResult.ControlPlaneEndpoint.Port),
	})
	respDiags.Append(cpeDiags...)
	base.ControlPlaneEndpoint = cpeVal
	if respDiags.HasError() {
		return false
	}

	creatorVal, creatorDiags := types.ObjectValueFrom(ctx, creatorInfoAttrTypes, CreatorInfoModel{
		Id:   types.StringValue(clusterResult.CreatorInfo.Id),
		Name: types.StringValue(clusterResult.CreatorInfo.Name),
	})
	respDiags.Append(creatorDiags...)
	base.CreatorInfo = creatorVal
	if respDiags.HasError() {
		return false
	}

	versionVal, versionDiags := types.ObjectValueFrom(ctx, omtInfoAttrTypes, OmtInfoModel{
		IsDeprecated: types.BoolValue(clusterResult.Version.IsDeprecated),
		Eol:          types.StringValue(clusterResult.Version.Eol),
		MinorVersion: types.StringValue(clusterResult.Version.MinorVersion),
		NextVersion:  types.StringValue(clusterResult.Version.NextVersion),
		PatchVersion: types.StringValue(clusterResult.Version.PatchVersion),
	})
	respDiags.Append(versionDiags...)
	base.Version = versionVal
	if respDiags.HasError() {
		return false
	}

	statusVal, statusDiags := types.ObjectValueFrom(ctx, statusAttrTypes, StatusModel{
		Phase: types.StringValue(string(clusterResult.Status.Phase)),
	})
	respDiags.Append(statusDiags...)
	base.Status = statusVal
	if respDiags.HasError() {
		return false
	}

	subnetVals := make([]attr.Value, 0, len(clusterResult.VpcInfo.Subnets))
	for _, s := range clusterResult.VpcInfo.Subnets {
		obj, objDiags := types.ObjectValue(
			subnetAttrTypes,
			map[string]attr.Value{
				"availability_zone": types.StringValue(string(s.AvailabilityZone)),
				"cidr_block":        types.StringValue(s.CidrBlock),
				"id":                types.StringValue(s.Id),
			},
		)
		respDiags.Append(objDiags...)
		subnetVals = append(subnetVals, obj)
	}

	subnetsSet, setDiags := types.SetValue(
		types.ObjectType{AttrTypes: subnetAttrTypes},
		subnetVals,
	)
	respDiags.Append(setDiags...)

	vpcVal, vpcDiags := types.ObjectValue(
		vpcInfoAttrTypes,
		map[string]attr.Value{
			"id":      types.StringValue(clusterResult.VpcInfo.Id),
			"subnets": subnetsSet,
		},
	)
	respDiags.Append(vpcDiags...)
	base.VpcInfo = vpcVal

	return true
}

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kubernetesengine

import (
	"time"

	. "terraform-provider-kakaocloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kakaoenterprise/kc-sdk-go/services/kubernetesengine"
)

func mapClusterNodeBaseModel(
	base *nodeBaseModel,
	result *kubernetesengine.KubernetesEngineV1ApiListClusterNodesModelNodeResponseModel,
	respDiags *diag.Diagnostics,
) bool {
	base.IsCordon = types.BoolValue(result.IsCordon)
	base.CreatedAt = types.StringValue(result.CreatedAt.Format(time.RFC3339))
	base.Flavor = types.StringValue(result.Flavor)
	base.Id = types.StringValue(result.Id)
	base.Ip = ConvertNullableString(result.Ip)
	base.Name = types.StringValue(result.Name)
	base.NodePoolName = types.StringValue(result.NodePoolName)
	base.SshKeyName = types.StringValue(result.SshKeyName)
	base.FailureMessage = ConvertNullableString(result.FailureMessage)
	base.UpdatedAt = types.StringValue(result.UpdatedAt.Format(time.RFC3339))
	base.Version = types.StringValue(result.Version)
	base.VolumeSize = types.Int32Value(result.VolumeSize)
	base.IsHyperThreading = types.BoolValue(result.IsHyperThreading)

	imgObj, imgDiags := types.ObjectValue(
		nodeImageInfoAttrTypes,
		map[string]attr.Value{
			"architecture":   types.StringValue(result.Image.Architecture),
			"is_gpu_type":    types.BoolValue(result.Image.IsGpuType),
			"id":             types.StringValue(result.Image.Id),
			"instance_type":  types.StringValue(result.Image.InstanceType),
			"kernel_version": types.StringValue(result.Image.KernelVersion),
			"key_package":    types.StringValue(result.Image.KeyPackage),
			"name":           types.StringValue(result.Image.Name),
			"os_distro":      types.StringValue(result.Image.OsDistro),
			"os_type":        types.StringValue(result.Image.OsType),
			"os_version":     types.StringValue(result.Image.OsVersion),
		},
	)
	respDiags.Append(imgDiags...)
	base.Image = imgObj

	statusObj, statusDiags := types.ObjectValue(
		nodeStatusAttrTypes,
		map[string]attr.Value{
			"phase": types.StringValue(string(result.Status.Phase)),
		},
	)
	respDiags.Append(statusDiags...)
	base.Status = statusObj

	subnetVals := make([]attr.Value, 0, len(result.VpcInfo.Subnets))
	for _, s := range result.VpcInfo.Subnets {
		sObj, sDiags := types.ObjectValue(
			subnetAttrTypes,
			map[string]attr.Value{
				"availability_zone": types.StringValue(string(s.AvailabilityZone)),
				"cidr_block":        types.StringValue(s.CidrBlock),
				"id":                types.StringValue(s.Id),
			},
		)
		respDiags.Append(sDiags...)
		subnetVals = append(subnetVals, sObj)
	}
	subnetsSet, setDiags := types.SetValue(
		types.ObjectType{AttrTypes: subnetAttrTypes},
		subnetVals,
	)
	respDiags.Append(setDiags...)
	if respDiags.HasError() {
		return false

	}
	vpcInfoObj, vpcDiags := types.ObjectValue(
		nodeVpcInfoAttrTypes,
		map[string]attr.Value{
			"id":      types.StringValue(result.VpcInfo.Id),
			"subnets": subnetsSet,
		},
	)
	respDiags.Append(vpcDiags...)
	base.VpcInfo = vpcInfoObj
	if respDiags.HasError() {
		return false
	}

	return true
}

func mapNodePoolNodeBaseModel(
	base *nodeBaseModel,
	result *kubernetesengine.KubernetesEngineV1ApiListNodePoolNodesModelNodeResponseModel,
	respDiags *diag.Diagnostics,
) bool {
	base.IsCordon = types.BoolValue(result.IsCordon)
	base.CreatedAt = types.StringValue(result.CreatedAt.Format(time.RFC3339))
	base.Flavor = types.StringValue(result.Flavor)
	base.Id = types.StringValue(result.Id)
	base.Ip = ConvertNullableString(result.Ip)
	base.Name = types.StringValue(result.Name)
	base.NodePoolName = types.StringValue(result.NodePoolName)
	base.SshKeyName = types.StringValue(result.SshKeyName)
	base.FailureMessage = ConvertNullableString(result.FailureMessage)
	base.UpdatedAt = types.StringValue(result.UpdatedAt.Format(time.RFC3339))
	base.Version = types.StringValue(result.Version)
	base.VolumeSize = types.Int32Value(result.VolumeSize)
	base.IsHyperThreading = types.BoolValue(result.IsHyperThreading)

	imgObj, imgDiags := types.ObjectValue(
		nodeImageInfoAttrTypes,
		map[string]attr.Value{
			"architecture":   types.StringValue(result.Image.Architecture),
			"is_gpu_type":    types.BoolValue(result.Image.IsGpuType),
			"id":             types.StringValue(result.Image.Id),
			"instance_type":  types.StringValue(result.Image.InstanceType),
			"kernel_version": types.StringValue(result.Image.KernelVersion),
			"key_package":    types.StringValue(result.Image.KeyPackage),
			"name":           types.StringValue(result.Image.Name),
			"os_distro":      types.StringValue(result.Image.OsDistro),
			"os_type":        types.StringValue(result.Image.OsType),
			"os_version":     types.StringValue(result.Image.OsVersion),
		},
	)
	respDiags.Append(imgDiags...)
	base.Image = imgObj

	statusObj, statusDiags := types.ObjectValue(
		nodeStatusAttrTypes,
		map[string]attr.Value{
			"phase": types.StringValue(string(result.Status.Phase)),
		},
	)
	respDiags.Append(statusDiags...)
	base.Status = statusObj

	subnetVals := make([]attr.Value, 0, len(result.VpcInfo.Subnets))
	for _, s := range result.VpcInfo.Subnets {
		sObj, sDiags := types.ObjectValue(
			subnetAttrTypes,
			map[string]attr.Value{
				"availability_zone": types.StringValue(string(s.AvailabilityZone)),
				"cidr_block":        types.StringValue(s.CidrBlock),
				"id":                types.StringValue(s.Id),
			},
		)
		respDiags.Append(sDiags...)
		subnetVals = append(subnetVals, sObj)
	}
	subnetsSet, setDiags := types.SetValue(
		types.ObjectType{AttrTypes: subnetAttrTypes},
		subnetVals,
	)
	respDiags.Append(setDiags...)
	if respDiags.HasError() {
		return false

	}
	vpcInfoObj, vpcDiags := types.ObjectValue(
		nodeVpcInfoAttrTypes,
		map[string]attr.Value{
			"id":      types.StringValue(result.VpcInfo.Id),
			"subnets": subnetsSet,
		},
	)
	respDiags.Append(vpcDiags...)
	base.VpcInfo = vpcInfoObj
	if respDiags.HasError() {
		return false
	}

	return true
}

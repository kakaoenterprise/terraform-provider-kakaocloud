// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kubernetesengine

import (
	"context"
	"time"

	. "terraform-provider-kakaocloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kakaoenterprise/kc-sdk-go/services/kubernetesengine"
)

func mapNodePoolFromResponse(
	ctx context.Context,
	dst *NodePoolBaseModel,
	src *kubernetesengine.KubernetesEngineV1ApiGetNodePoolModelNodePoolResponseModel,
	diags *diag.Diagnostics,
	userSGsHint ...[]string,
) bool {
	// basics
	dst.Id = types.StringValue(src.Id)
	dst.Name = types.StringValue(src.Name)
	dst.ClusterName = types.StringValue(src.ClusterName)
	dst.Description = types.StringValue(src.Description)
	dst.CreatedAt = types.StringValue(src.CreatedAt.Format(time.RFC3339))
	dst.FailureMessage = ConvertNullableString(src.FailureMessage)
	dst.NodeCount = types.Int32Value(src.NodeCount)
	dst.FlavorId = types.StringValue(src.FlavorId)
	dst.Flavor = types.StringValue(src.Flavor)
	dst.SshKeyName = types.StringValue(src.SshKeyName)
	dst.ImageId = types.StringValue(src.Image.Id)
	dst.IsGpu = types.BoolValue(src.IsGpu)
	dst.IsBareMetal = types.BoolValue(src.IsBareMetal)
	dst.IsUpgradable = types.BoolValue(src.IsUpgradable)
	dst.Version = types.StringValue(src.Version)
	dst.VolumeSize = types.Int32Value(src.VolumeSize)
	dst.IsCordon = types.BoolValue(src.IsCordon)
	dst.IsHyperThreading = types.BoolValue(src.IsHyperThreading)

	// status
	statusVal, statusDiag := types.ObjectValueFrom(ctx, nodePoolStatusAttrTypes, nodePoolStatusModel{
		Phase:            types.StringValue(string(src.Status.Phase)),
		AvailableNodes:   types.Int32Value(src.Status.AvailableNodes),
		UnavailableNodes: types.Int32Value(src.Status.UnavailableNodes),
	})
	diags.Append(statusDiag...)
	dst.Status = statusVal
	if diags.HasError() {
		return false
	}

	// image
	imageVal, imageDiag := types.ObjectValueFrom(ctx, imageInfoAttrTypes, imageInfoModel{
		Id:   types.StringValue(src.Image.Id),
		Name: types.StringValue(src.Image.Name),
	})
	diags.Append(imageDiag...)
	dst.Image = imageVal
	if diags.HasError() {
		return false
	}

	// labels set(object) to avoid ordering diffs
	lbls := make([]attr.Value, 0, len(src.Labels))
	for _, l := range src.Labels {
		obj, _ := types.ObjectValue(nodePoolLabelAttrTypes, map[string]attr.Value{
			"key":   types.StringValue(l.Key),
			"value": types.StringValue(l.Value),
		})
		lbls = append(lbls, obj)
	}
	labelsVal, lDiag := types.SetValue(types.ObjectType{AttrTypes: nodePoolLabelAttrTypes}, lbls)
	diags.Append(lDiag...)
	dst.Labels = labelsVal
	if diags.HasError() {
		return false
	}

	// taints set(object) to avoid ordering diffs
	tnts := make([]attr.Value, 0, len(src.Taints))
	for _, t := range src.Taints {
		obj, _ := types.ObjectValue(nodePoolTaintAttrTypes, map[string]attr.Value{
			"key":    types.StringValue(t.Key),
			"value":  types.StringValue(t.Value),
			"effect": types.StringValue(string(t.Effect)),
		})
		tnts = append(tnts, obj)
	}
	taintsVal, tDiag := types.SetValue(types.ObjectType{AttrTypes: nodePoolTaintAttrTypes}, tnts)
	diags.Append(tDiag...)
	dst.Taints = taintsVal
	if diags.HasError() {
		return false
	}

	apiSGs := src.SecurityGroups
	userSGs := []string{}
	defaultSGs := []string{}

	hintSet := map[string]struct{}{}
	if len(userSGsHint) > 0 && userSGsHint[0] != nil {
		for _, s := range userSGsHint[0] {
			hintSet[s] = struct{}{}
		}
	}

	for _, sg := range apiSGs {
		if _, ok := hintSet[sg]; ok {
			userSGs = append(userSGs, sg)
		} else {
			defaultSGs = append(defaultSGs, sg)
		}
	}

	if len(userSGs) > 0 {
		userSet, d1 := types.SetValueFrom(ctx, types.StringType, userSGs)
		diags.Append(d1...)
		dst.SecurityGroups = userSet
	} else {
		emptySet, d1 := types.SetValueFrom(ctx, types.StringType, []string{})
		diags.Append(d1...)
		dst.SecurityGroups = emptySet
	}

	defaultSet, d2 := types.SetValueFrom(ctx, types.StringType, defaultSGs)
	diags.Append(d2...)
	dst.DefaultSecurityGroups = defaultSet

	if diags.HasError() {
		return false
	}

	subnetObjs := make([]attr.Value, 0, len(src.VpcInfo.Subnets))
	for _, s := range src.VpcInfo.Subnets {
		obj, _ := types.ObjectValue(
			subnetAttrTypes,
			map[string]attr.Value{
				"availability_zone": types.StringValue(string(s.AvailabilityZone)),
				"cidr_block":        types.StringValue(s.CidrBlock),
				"id":                types.StringValue(s.Id),
			},
		)
		subnetObjs = append(subnetObjs, obj)
	}
	subnetsSet, setDiags := types.SetValue(
		types.ObjectType{AttrTypes: subnetAttrTypes},
		subnetObjs,
	)
	diags.Append(setDiags...)

	vpcVal, vpcDiags := types.ObjectValueFrom(ctx, vpcInfoAttrTypes, VpcInfoModel{
		Id:      types.StringValue(src.VpcInfo.Id),
		Subnets: subnetsSet,
	})
	diags.Append(vpcDiags...)
	dst.VpcInfo = vpcVal

	if diags.HasError() {
		return false
	}

	// autoscaling
	a := src.Autoscaling
	autoscalingVal, autoDiag := types.ObjectValueFrom(ctx, nodePoolAutoscalingAttrTypes, NodePoolAutoscalingModel{
		IsAutoscalerEnable:              types.BoolValue(a.IsAutoscalerEnable),
		AutoscalerDesiredNodeCount:      ConvertNullableInt32(a.AutoscalerDesiredNodeCount),
		AutoscalerMaxNodeCount:          ConvertNullableInt32(a.AutoscalerMaxNodeCount),
		AutoscalerMinNodeCount:          ConvertNullableInt32(a.AutoscalerMinNodeCount),
		AutoscalerScaleDownThreshold:    ConvertNullableFloat32(a.AutoscalerScaleDownThreshold),
		AutoscalerScaleDownUnneededTime: ConvertNullableInt32(a.AutoscalerScaleDownUnneededTime),
		AutoscalerScaleDownUnreadyTime:  ConvertNullableInt32(a.AutoscalerScaleDownUnreadyTime),
	})

	diags.Append(autoDiag...)
	dst.Autoscaling = autoscalingVal
	if diags.HasError() {
		return false
	}

	return true
}

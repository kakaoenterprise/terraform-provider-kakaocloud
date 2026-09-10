// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package kubernetesengine

import (
	"context"
	"fmt"
	"strings"
	"time"

	. "terraform-provider-kakaocloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kakaoenterprise/kc-sdk-go/v2/services/kubernetesengine"
)

func mapNodePoolBaseModel(
	ctx context.Context,
	dst *nodePoolBaseModel,
	src *kubernetesengine.NodePool,
	diags *diag.Diagnostics,
) bool {

	dst.Id = types.StringValue(src.Id)
	dst.Name = types.StringValue(src.Name)
	dst.ClusterName = types.StringValue(src.ClusterName)
	dst.Description = types.StringValue(src.Description)
	dst.CreatedAt = types.StringValue(src.CreatedAt.Format(time.RFC3339))
	dst.FailureMessage = ConvertNullableString(src.FailureMessage)
	dst.NodeCount = types.Int32Value(src.NodeCount)
	dst.FlavorId = types.StringValue(src.FlavorId)
	dst.SshKeyName = types.StringValue(src.SshKeyName)
	dst.IsGpu = types.BoolValue(src.IsGpu)
	dst.IsBareMetal = types.BoolValue(src.IsBareMetal)
	dst.IsUpgradable = types.BoolValue(src.IsUpgradable)
	dst.Version = types.StringValue(src.Version)
	dst.VolumeSize = types.Int32Value(src.VolumeSize)
	dst.IsCordon = types.BoolValue(src.IsCordon)
	dst.IsHyperThreading = types.BoolValue(src.IsHyperThreading)
	dst.UserData = ConvertNullableString(src.UserData)

	statusVal, statusDiag := buildNodePoolStatus(ctx, src.Status)
	diags.Append(statusDiag...)
	dst.Status = statusVal
	if diags.HasError() {
		return false
	}

	if len(src.Labels) > 0 {
		labelsVal, lDiag := buildLabelsSet(src.Labels)
		diags.Append(lDiag...)
		if diags.HasError() {
			return false
		}
		dst.Labels = labelsVal
	} else {
		dst.Labels = types.SetNull(types.ObjectType{AttrTypes: nodePoolLabelAttrTypes})
	}

	if len(src.Taints) > 0 {
		taintsVal, tDiag := buildTaintsSet(src.Taints)
		diags.Append(tDiag...)
		if diags.HasError() {
			return false
		}
		dst.Taints = taintsVal
	} else {
		dst.Taints = types.SetNull(types.ObjectType{AttrTypes: nodePoolTaintAttrTypes})
	}

	setVal, d1 := types.SetValueFrom(ctx, types.StringType, src.SecurityGroups)
	diags.Append(d1...)
	dst.SecurityGroups = setVal
	if diags.HasError() {
		return false
	}

	vpcVal, vpcDiags := buildVpcInfo(ctx, src.VpcInfo)
	diags.Append(vpcDiags...)
	dst.VpcInfo = vpcVal

	if diags.HasError() {
		return false
	}

	a := src.Autoscaling
	autoscalingVal, autoDiag := buildAutoscaling(ctx, a)

	diags.Append(autoDiag...)
	dst.Autoscaling = autoscalingVal
	if diags.HasError() {
		return false
	}

	return true
}

func mapNodePoolDataSource(
	ctx context.Context,
	dst *nodePoolDataSourceBaseModel,
	src *kubernetesengine.NodePool,
	diags *diag.Diagnostics,
) bool {
	ok := mapNodePoolBaseModel(ctx, &dst.nodePoolBaseModel, src, diags)
	if !ok {
		return false
	}

	dst.Flavor = types.StringValue(src.Flavor)

	imageVal, imageDiag := buildImageInfo(ctx, src.Image)
	diags.Append(imageDiag...)
	dst.Image = imageVal
	if diags.HasError() {
		return false
	}

	return true
}

func (r *nodePoolResource) mapNodePoolResource(
	ctx context.Context,
	dst *nodePoolResourceModel,
	src *kubernetesengine.NodePool,
	diags *diag.Diagnostics,
) bool {
	ok := mapNodePoolBaseModel(ctx, &dst.nodePoolBaseModel, src, diags)
	if !ok {
		return false
	}

	verParts := strings.Split(src.Version, ".")
	if len(verParts) >= 2 {
		dst.MinorVersion = types.StringValue(fmt.Sprintf("%s.%s", verParts[0], verParts[1]))
	} else {
		dst.MinorVersion = types.StringValue(src.Version)
	}

	return true
}

func buildNodePoolStatus(ctx context.Context, st kubernetesengine.NodePoolStatusInfo) (types.Object, diag.Diagnostics) {
	return types.ObjectValueFrom(ctx, nodePoolStatusAttrTypes, nodePoolStatusModel{
		Phase:            types.StringValue(string(st.Phase)),
		AvailableNodes:   types.Int32Value(st.AvailableNodes),
		UnavailableNodes: types.Int32Value(st.UnavailableNodes),
	})
}

func buildImageInfo(ctx context.Context, img kubernetesengine.ImageInfo) (types.Object, diag.Diagnostics) {
	return types.ObjectValueFrom(ctx, imageInfoAttrTypes, imageInfoModel{
		Id:            types.StringValue(img.Id),
		Name:          types.StringValue(img.Name),
		Architecture:  types.StringValue(img.Architecture),
		IsGpuType:     types.BoolValue(img.IsGpuType),
		InstanceType:  types.StringValue(img.InstanceType),
		KernelVersion: types.StringValue(img.KernelVersion),
		KeyPackage:    types.StringValue(img.KeyPackage),
		OsDistro:      types.StringValue(img.OsDistro),
		OsType:        types.StringValue(img.OsType),
		OsVersion:     types.StringValue(img.OsVersion),
	})
}

func buildLabelsSet(labels []kubernetesengine.LabelInfo) (types.Set, diag.Diagnostics) {
	lbls := make([]attr.Value, 0, len(labels))
	for _, l := range labels {
		obj, _ := types.ObjectValue(nodePoolLabelAttrTypes, map[string]attr.Value{
			"key":   types.StringValue(l.Key),
			"value": types.StringValue(l.Value),
		})
		lbls = append(lbls, obj)
	}
	set, d := types.SetValue(types.ObjectType{AttrTypes: nodePoolLabelAttrTypes}, lbls)
	return set, d
}

func buildTaintsSet(taints []kubernetesengine.TaintInfo) (types.Set, diag.Diagnostics) {
	tnts := make([]attr.Value, 0, len(taints))
	for _, t := range taints {
		obj, _ := types.ObjectValue(nodePoolTaintAttrTypes, map[string]attr.Value{
			"key":    types.StringValue(t.Key),
			"value":  types.StringValue(t.Value),
			"effect": types.StringValue(string(t.Effect)),
		})
		tnts = append(tnts, obj)
	}
	set, d := types.SetValue(types.ObjectType{AttrTypes: nodePoolTaintAttrTypes}, tnts)
	return set, d
}

func buildVpcInfo(ctx context.Context, vpc kubernetesengine.VpcInfo) (types.Object, diag.Diagnostics) {
	subnetObjs := make([]attr.Value, 0, len(vpc.Subnets))
	for _, s := range vpc.Subnets {
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
	v, vDiags := types.ObjectValueFrom(ctx, vpcInfoAttrTypes, VpcInfoModel{
		Id:      types.StringValue(vpc.Id),
		Subnets: subnetsSet,
	})
	var all diag.Diagnostics
	all.Append(setDiags...)
	all.Append(vDiags...)
	return v, all
}

func buildAutoscaling(ctx context.Context, a kubernetesengine.Autoscaling) (types.Object, diag.Diagnostics) {
	return types.ObjectValueFrom(ctx, nodePoolAutoscalingAttrTypes, nodePoolAutoscalingModel{
		IsAutoscalerEnable:              types.BoolValue(a.GetIsAutoscalerEnable()),
		AutoscalerDesiredNodeCount:      ConvertNullableInt32(a.AutoscalerDesiredNodeCount),
		AutoscalerMaxNodeCount:          ConvertNullableInt32(a.AutoscalerMaxNodeCount),
		AutoscalerMinNodeCount:          ConvertNullableInt32(a.AutoscalerMinNodeCount),
		AutoscalerScaleDownThreshold:    ConvertNullableFloat32(a.AutoscalerScaleDownThreshold),
		AutoscalerScaleDownUnneededTime: ConvertNullableInt32(a.AutoscalerScaleDownUnneededTime),
		AutoscalerScaleDownUnreadyTime:  ConvertNullableInt32(a.AutoscalerScaleDownUnreadyTime),
	})
}

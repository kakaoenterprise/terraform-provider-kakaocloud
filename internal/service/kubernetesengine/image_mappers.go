// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kubernetesengine

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kakaoenterprise/kc-sdk-go/services/kubernetesengine"
)

func (d *kubernetesImagesDataSource) mapImages(
	ctx context.Context,
	base *imageBaseModel,
	imageResult *kubernetesengine.ImageResponseModel,
	respDiags *diag.Diagnostics,
) bool {
	base.Architecture = types.StringValue(imageResult.Architecture)
	base.IsGpuType = types.BoolValue(imageResult.IsGpuType)
	base.Id = types.StringValue(imageResult.Id)
	base.InstanceType = types.StringValue(imageResult.InstanceType)
	base.K8sVersion = types.StringValue(imageResult.K8sVersion)
	base.KernelVersion = types.StringValue(imageResult.KernelVersion)
	base.KeyPackage = types.StringValue(imageResult.KeyPackage)
	base.Name = types.StringValue(imageResult.Name)
	base.OsDistro = types.StringValue(imageResult.OsDistro)
	base.OsType = types.StringValue(imageResult.OsType)
	base.OsVersion = types.StringValue(imageResult.OsVersion)

	return true
}

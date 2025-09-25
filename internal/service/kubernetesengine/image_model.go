// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kubernetesengine

import (
	"terraform-provider-kakaocloud/internal/common"

	datasourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type imageBaseModel struct {
	Architecture  types.String `tfsdk:"architecture"`
	IsGpuType     types.Bool   `tfsdk:"is_gpu_type"`
	Id            types.String `tfsdk:"id"`
	InstanceType  types.String `tfsdk:"instance_type"`
	K8sVersion    types.String `tfsdk:"k8s_version"`
	KernelVersion types.String `tfsdk:"kernel_version"`
	KeyPackage    types.String `tfsdk:"key_package"`
	Name          types.String `tfsdk:"name"`
	OsDistro      types.String `tfsdk:"os_distro"`
	OsType        types.String `tfsdk:"os_type"`
	OsVersion     types.String `tfsdk:"os_version"`
}

type imagesDataSourceModel struct {
	Filter   []common.FilterModel     `tfsdk:"filter"`
	Images   []imageBaseModel         `tfsdk:"images"`
	Timeouts datasourceTimeouts.Value `tfsdk:"timeouts"`
}

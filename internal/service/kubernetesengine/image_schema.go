// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package kubernetesengine

import (
	"terraform-provider-kakaocloud/internal/docs"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func getKubernetesImageDataSourceSchema() map[string]schema.Attribute {
	desc := docs.Kubernetesengine("ImageResponseModel")

	return map[string]schema.Attribute{
		"architecture": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("architecture"),
		},
		"is_gpu_type": schema.BoolAttribute{
			Computed:    true,
			Description: desc.String("is_gpu_type"),
		},
		"id": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("id"),
		},
		"instance_type": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("instance_type"),
		},
		"k8s_version": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("k8s_version"),
		},
		"kernel_version": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("kernel_version"),
		},
		"key_package": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("key_package"),
		},
		"name": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("name"),
		},
		"os_distro": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("os_distro"),
		},
		"os_type": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("os_type"),
		},
		"os_version": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("os_version"),
		},
	}
}

var kubernetesImageDataSourceSchemaAttributes = getKubernetesImageDataSourceSchema()

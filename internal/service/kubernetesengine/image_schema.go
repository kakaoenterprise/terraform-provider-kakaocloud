// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package kubernetesengine

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func getKubernetesImageDataSourceSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"architecture": schema.StringAttribute{
			Computed: true,
		},
		"is_gpu_type": schema.BoolAttribute{
			Computed: true,
		},
		"id": schema.StringAttribute{
			Computed: true,
		},
		"instance_type": schema.StringAttribute{
			Computed: true,
		},
		"k8s_version": schema.StringAttribute{
			Computed: true,
		},
		"kernel_version": schema.StringAttribute{
			Computed: true,
		},
		"key_package": schema.StringAttribute{
			Computed: true,
		},
		"name": schema.StringAttribute{
			Computed: true,
		},
		"os_distro": schema.StringAttribute{
			Computed: true,
		},
		"os_type": schema.StringAttribute{
			Computed: true,
		},
		"os_version": schema.StringAttribute{
			Computed: true,
		},
	}
}

var kubernetesImageDataSourceSchemaAttributes = getKubernetesImageDataSourceSchema()

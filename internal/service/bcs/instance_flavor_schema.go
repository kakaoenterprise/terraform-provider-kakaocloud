// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package bcs

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func getInstanceFlavorSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Computed: true,
		},
		"vcpus": schema.Int32Attribute{
			Computed: true,
		},
		"description": schema.StringAttribute{
			Computed: true,
		},
		"is_burstable": schema.BoolAttribute{
			Computed: true,
		},
		"architecture": schema.StringAttribute{
			Computed: true,
		},
		"manufacturer": schema.StringAttribute{
			Computed: true,
		},
		"group": schema.StringAttribute{
			Computed: true,
		},
		"instance_type": schema.StringAttribute{
			Computed: true,
		},
		"processor": schema.StringAttribute{
			Computed: true,
		},
		"memory_mb": schema.Int64Attribute{
			Computed: true,
		},
		"created_at": schema.StringAttribute{
			Computed: true,
		},
		"updated_at": schema.StringAttribute{
			Computed: true,
		},
		"availability_zone": schema.ListAttribute{
			Computed:    true,
			ElementType: types.StringType,
		},
		"available": schema.MapAttribute{
			ElementType: types.Int32Type,
			Computed:    true,
		},
		"instance_family": schema.StringAttribute{
			Computed: true,
		},
		"instance_size": schema.StringAttribute{
			Computed: true,
		},
		"disk_type": schema.StringAttribute{
			Computed: true,
		},
		"root_gb": schema.Int32Attribute{
			Computed: true,
		},
		"os_distro": schema.StringAttribute{
			Computed: true,
		},
		"hw_count": schema.Int32Attribute{
			Computed: true,
		},
		"hw_type": schema.StringAttribute{
			Computed: true,
		},
		"hw_name": schema.StringAttribute{
			Computed: true,
		},
		"maximum_network_interfaces": schema.Int32Attribute{
			Computed: true,
		},
		"is_hyper_threading_disabled": schema.BoolAttribute{
			Computed: true,
		},
		"is_hyper_threading_supported": schema.BoolAttribute{
			Computed: true,
		},
		"is_hyper_threading_disable_supported": schema.BoolAttribute{
			Computed: true,
		},
	}
}

var instanceFlavorDataSourceSchemaAttributes = getInstanceFlavorSchema()

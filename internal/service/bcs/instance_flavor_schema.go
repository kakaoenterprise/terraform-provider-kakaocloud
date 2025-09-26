// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package bcs

import (
	"terraform-provider-kakaocloud/internal/docs"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func getInstanceFlavorSchema() map[string]schema.Attribute {
	desc := docs.Bcs("bcs_instance__v1__api__get_instance_type__model__FlavorModel")
	descList := docs.Bcs("bcs_instance__v1__api__list_instances__model__InstanceFlavorModel")

	return map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("name"),
		},
		"vcpus": schema.Int32Attribute{
			Computed:    true,
			Description: desc.String("vcpus"),
		},
		"description": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("description"),
		},
		"is_burstable": schema.BoolAttribute{
			Computed:    true,
			Description: desc.String("is_burstable"),
		},
		"architecture": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("architecture"),
		},
		"manufacturer": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("manufacturer"),
		},
		"group": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("group"),
		},
		"instance_type": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("instance_type"),
		},
		"processor": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("processor"),
		},
		"memory_mb": schema.Int64Attribute{
			Computed:    true,
			Description: desc.String("memory_mb"),
		},
		"created_at": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("created_at"),
		},
		"updated_at": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("updated_at"),
		},
		"availability_zone": schema.ListAttribute{
			Computed:    true,
			ElementType: types.StringType,
			Description: desc.String("availability_zone"),
		},
		"available": schema.MapAttribute{
			ElementType: types.Int32Type,
			Computed:    true,
			Description: desc.String("available"),
		},
		"instance_family": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("instance_family"),
		},
		"instance_size": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("instance_size"),
		},
		"disk_type": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("disk_type"),
		},
		"root_gb": schema.Int32Attribute{
			Computed:    true,
			Description: descList.String("root_gb"),
		},
		"os_distro": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("os_distro"),
		},
		"hw_count": schema.Int32Attribute{
			Computed:    true,
			Description: desc.String("hw_count"),
		},
		"hw_type": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("hw_type"),
		},
		"hw_name": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("hw_name"),
		},
		"maximum_network_interfaces": schema.Int32Attribute{
			Computed:    true,
			Description: desc.String("maximum_network_interfaces"),
		},
		"is_hyper_threading_disabled": schema.BoolAttribute{
			Computed:    true,
			Description: desc.String("is_hyper_threading_disabled"),
		},
		"is_hyper_threading_supported": schema.BoolAttribute{
			Computed:    true,
			Description: desc.String("is_hyper_threading_supported"),
		},
		"is_hyper_threading_disable_supported": schema.BoolAttribute{
			Computed:    true,
			Description: desc.String("is_hyper_threading_disable_supported"),
		},
	}
}

var instanceFlavorDataSourceSchemaAttributes = getInstanceFlavorSchema()

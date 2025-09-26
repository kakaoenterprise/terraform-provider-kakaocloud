// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package bcs

import (
	"terraform-provider-kakaocloud/internal/docs"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func getInstanceDataSourceSchema() map[string]schema.Attribute {
	desc := docs.Bcs("bcs_instance__v1__api__get_instance__model__InstanceModel")

	return map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("name"),
		},
		"description": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("description"),
		},
		"metadata": schema.MapAttribute{
			ElementType: types.StringType,
			Computed:    true,
			Description: desc.String("metadata"),
		},
		"flavor": schema.SingleNestedAttribute{
			Computed:    true,
			Description: desc.String("flavor"),
			Attributes:  getInstanceFlavorSchemaAttributes(),
		},
		"addresses": schema.ListNestedAttribute{
			Computed:    true,
			Description: desc.String("addresses"),
			NestedObject: schema.NestedAttributeObject{
				Attributes: getInstanceAddressesSchemaAttributes(),
			},
		},
		"is_hyper_threading": schema.BoolAttribute{
			Computed:    true,
			Description: desc.String("is_hyper_threading"),
		},
		"is_hadoop": schema.BoolAttribute{
			Computed:    true,
			Description: desc.String("is_hadoop"),
		},
		"is_k8se": schema.BoolAttribute{
			Computed:    true,
			Description: desc.String("is_k8se"),
		},
		"image": schema.SingleNestedAttribute{
			Computed:    true,
			Description: desc.String("image"),
			Attributes:  getInstanceImageSchemaAttributes(),
		},
		"vm_state": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("vm_state"),
		},
		"task_state": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("task_state"),
		},
		"power_state": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("power_state"),
		},
		"status": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("status"),
		},
		"user_id": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("user_id"),
		},
		"project_id": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("project_id"),
		},
		"key_name": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("key_name"),
		},
		"hostname": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("hostname"),
		},
		"availability_zone": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("availability_zone"),
		},
		"attached_volumes": schema.ListNestedAttribute{
			Computed:    true,
			Description: desc.String("attached_volumes"),
			NestedObject: schema.NestedAttributeObject{
				Attributes: getInstanceAttachedVolumesSchemaAttributes(),
			},
		},
		"attached_volume_count": schema.Int64Attribute{
			Computed:    true,
			Description: desc.String("attached_volume_count"),
		},
		"security_groups": schema.SetNestedAttribute{
			Computed:    true,
			Description: desc.String("security_groups"),
			NestedObject: schema.NestedAttributeObject{
				Attributes: getInstanceSecurityGroupsSchemaAttributes(),
			},
		},
		"security_group_count": schema.Int64Attribute{
			Computed:    true,
			Description: desc.String("security_group_count"),
		},
		"instance_type": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("instance_type"),
		},
		"created_at": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("created_at"),
		},
		"updated_at": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("updated_at"),
		},
	}
}

func getInstanceFlavorSchemaAttributes() map[string]schema.Attribute {
	desc := docs.Bcs("bcs_instance__v1__api__get_instance__model__InstanceFlavorModel")

	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("id"),
		},
		"name": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("name"),
		},
		"group": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("group"),
		},
		"vcpus": schema.Int32Attribute{
			Computed:    true,
			Description: desc.String("vcpus"),
		},
		"is_burstable": schema.BoolAttribute{
			Computed:    true,
			Description: desc.String("is_burstable"),
		},
		"manufacturer": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("manufacturer"),
		},
		"memory_mb": schema.Int32Attribute{
			Computed:    true,
			Description: desc.String("memory_mb"),
		},
		"root_gb": schema.Int32Attribute{
			Computed:    true,
			Description: desc.String("root_gb"),
		},
		"disk_type": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("disk_type"),
		},
		"instance_family": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("instance_family"),
		},
		"os_distro": schema.ListAttribute{
			Computed:    true,
			ElementType: types.StringType,
			Description: desc.String("os_distro"),
		},
		"maximum_network_interfaces": schema.Int32Attribute{
			Computed:    true,
			Description: desc.String("maximum_network_interfaces"),
		},
		"hw_type": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("hw_type"),
		},
		"hw_count": schema.Int32Attribute{
			Computed:    true,
			Description: desc.String("hw_count"),
		},
		"is_hyper_threading_supported": schema.BoolAttribute{
			Computed:    true,
			Description: desc.String("is_hyper_threading_supported"),
		},
		"real_vcpus": schema.Int32Attribute{
			Computed:    true,
			Description: desc.String("real_vcpus"),
		},
	}
}

func getInstanceAddressesSchemaAttributes() map[string]schema.Attribute {
	desc := docs.Bcs("bcs_instance__v1__api__get_instance__model__InstanceAddressModel")

	return map[string]schema.Attribute{
		"private_ip": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("private_ip"),
		},
		"public_ip": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("public_ip"),
		},
		"network_interface_id": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("network_interface_id"),
		},
	}
}

func getInstanceImageSchemaAttributes() map[string]schema.Attribute {
	desc := docs.Bcs("bcs_instance__v1__api__get_instance__model__InstanceImageModel")

	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("id"),
		},
		"name": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("name"),
		},
		"description": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("description"),
		},
		"owner": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("owner"),
		},
		"is_windows": schema.BoolAttribute{
			Computed:    true,
			Description: desc.String("is_windows"),
		},
		"size": schema.Int64Attribute{
			Computed:    true,
			Description: desc.String("size"),
		},
		"status": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("status"),
		},
		"image_type": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("image_type"),
		},
		"disk_format": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("disk_format"),
		},
		"instance_type": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("instance_type"),
		},
		"member_status": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("member_status"),
		},
		"min_disk": schema.Int32Attribute{
			Computed:    true,
			Description: desc.String("min_disk"),
		},
		"min_memory": schema.Int32Attribute{
			Computed:    true,
			Description: desc.String("min_memory"),
		},
		"os_admin": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("os_admin"),
		},
		"os_distro": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("os_distro"),
		},
		"os_type": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("os_type"),
		},
		"os_architecture": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("os_architecture"),
		},
		"created_at": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("created_at"),
		},
		"updated_at": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("updated_at"),
		},
	}
}

func getInstanceAttachedVolumesSchemaAttributes() map[string]schema.Attribute {
	desc := docs.Bcs("bcs_instance__v1__api__get_instance__model__InstanceAttachedVolumeModel")

	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("id"),
		},
		"name": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("name"),
		},
		"status": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("status"),
		},
		"mount_point": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("mount_point"),
		},
		"type": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("type"),
		},
		"size": schema.Int32Attribute{
			Computed:    true,
			Description: desc.String("size"),
		},
		"is_delete_on_termination": schema.BoolAttribute{
			Computed:    true,
			Description: desc.String("is_delete_on_termination"),
		},
		"created_at": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("created_at"),
		},
		"is_root": schema.BoolAttribute{
			Computed:    true,
			Description: desc.String("is_root"),
		},
	}
}

func getInstanceSecurityGroupsSchemaAttributes() map[string]schema.Attribute {
	desc := docs.Bcs("bcs_instance__v1__api__get_instance__model__InstanceSecurityGroupModel")

	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("id"),
		},
		"name": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("name"),
		},
	}
}

var instanceDataSourceSchemaAttributes = getInstanceDataSourceSchema()
var instanceFlavorSchemaAttributes = getInstanceFlavorSchemaAttributes()
var instanceAddressesSchemaAttributes = getInstanceAddressesSchemaAttributes()
var instanceImageSchemaAttributes = getInstanceImageSchemaAttributes()
var instanceAttachedVolumesSchemaAttributes = getInstanceAttachedVolumesSchemaAttributes()
var instanceSecurityGroupsSchemaAttributes = getInstanceSecurityGroupsSchemaAttributes()

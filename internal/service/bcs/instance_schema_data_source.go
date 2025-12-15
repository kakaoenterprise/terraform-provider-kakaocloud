// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package bcs

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func getInstanceDataSourceSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Computed: true,
		},
		"description": schema.StringAttribute{
			Computed: true,
		},
		"metadata": schema.MapAttribute{
			ElementType: types.StringType,
			Computed:    true,
		},
		"flavor": schema.SingleNestedAttribute{
			Computed:   true,
			Attributes: getInstanceFlavorSchemaAttributes(),
		},
		"addresses": schema.ListNestedAttribute{
			Computed: true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: getInstanceAddressesSchemaAttributes(),
			},
		},
		"is_hyper_threading": schema.BoolAttribute{
			Computed: true,
		},
		"is_hadoop": schema.BoolAttribute{
			Computed: true,
		},
		"is_k8se": schema.BoolAttribute{
			Computed: true,
		},
		"image": schema.SingleNestedAttribute{
			Computed:   true,
			Attributes: getInstanceImageSchemaAttributes(),
		},
		"vm_state": schema.StringAttribute{
			Computed: true,
		},
		"task_state": schema.StringAttribute{
			Computed: true,
		},
		"power_state": schema.StringAttribute{
			Computed: true,
		},
		"status": schema.StringAttribute{
			Computed: true,
		},
		"user_id": schema.StringAttribute{
			Computed: true,
		},
		"project_id": schema.StringAttribute{
			Computed: true,
		},
		"key_name": schema.StringAttribute{
			Computed: true,
		},
		"hostname": schema.StringAttribute{
			Computed: true,
		},
		"availability_zone": schema.StringAttribute{
			Computed: true,
		},
		"attached_volumes": schema.ListNestedAttribute{
			Computed: true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: getInstanceAttachedVolumesSchemaAttributes(),
			},
		},
		"attached_volume_count": schema.Int64Attribute{
			Computed: true,
		},
		"security_groups": schema.SetNestedAttribute{
			Computed: true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: getInstanceSecurityGroupsSchemaAttributes(),
			},
		},
		"security_group_count": schema.Int64Attribute{
			Computed: true,
		},
		"instance_type": schema.StringAttribute{
			Computed: true,
		},
		"created_at": schema.StringAttribute{
			Computed: true,
		},
		"updated_at": schema.StringAttribute{
			Computed: true,
		},
	}
}

func getInstanceFlavorSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed: true,
		},
		"name": schema.StringAttribute{
			Computed: true,
		},
		"group": schema.StringAttribute{
			Computed: true,
		},
		"vcpus": schema.Int32Attribute{
			Computed: true,
		},
		"is_burstable": schema.BoolAttribute{
			Computed: true,
		},
		"manufacturer": schema.StringAttribute{
			Computed: true,
		},
		"memory_mb": schema.Int32Attribute{
			Computed: true,
		},
		"root_gb": schema.Int32Attribute{
			Computed: true,
		},
		"disk_type": schema.StringAttribute{
			Computed: true,
		},
		"instance_family": schema.StringAttribute{
			Computed: true,
		},
		"os_distro": schema.ListAttribute{
			Computed:    true,
			ElementType: types.StringType,
		},
		"maximum_network_interfaces": schema.Int32Attribute{
			Computed: true,
		},
		"hw_type": schema.StringAttribute{
			Computed: true,
		},
		"hw_count": schema.Int32Attribute{
			Computed: true,
		},
		"is_hyper_threading_supported": schema.BoolAttribute{
			Computed: true,
		},
		"real_vcpus": schema.Int32Attribute{
			Computed: true,
		},
	}
}

func getInstanceAddressesSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"private_ip": schema.StringAttribute{
			Computed: true,
		},
		"public_ip": schema.StringAttribute{
			Computed: true,
		},
		"network_interface_id": schema.StringAttribute{
			Computed: true,
		},
	}
}

func getInstanceImageSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed: true,
		},
		"name": schema.StringAttribute{
			Computed: true,
		},
		"description": schema.StringAttribute{
			Computed: true,
		},
		"owner": schema.StringAttribute{
			Computed: true,
		},
		"is_windows": schema.BoolAttribute{
			Computed: true,
		},
		"size": schema.Int64Attribute{
			Computed: true,
		},
		"status": schema.StringAttribute{
			Computed: true,
		},
		"image_type": schema.StringAttribute{
			Computed: true,
		},
		"disk_format": schema.StringAttribute{
			Computed: true,
		},
		"instance_type": schema.StringAttribute{
			Computed: true,
		},
		"member_status": schema.StringAttribute{
			Computed: true,
		},
		"min_disk": schema.Int32Attribute{
			Computed: true,
		},
		"min_memory": schema.Int32Attribute{
			Computed: true,
		},
		"os_admin": schema.StringAttribute{
			Computed: true,
		},
		"os_distro": schema.StringAttribute{
			Computed: true,
		},
		"os_type": schema.StringAttribute{
			Computed: true,
		},
		"os_architecture": schema.StringAttribute{
			Computed: true,
		},
		"created_at": schema.StringAttribute{
			Computed: true,
		},
		"updated_at": schema.StringAttribute{
			Computed: true,
		},
	}
}

func getInstanceAttachedVolumesSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed: true,
		},
		"name": schema.StringAttribute{
			Computed: true,
		},
		"status": schema.StringAttribute{
			Computed: true,
		},
		"mount_point": schema.StringAttribute{
			Computed: true,
		},
		"type": schema.StringAttribute{
			Computed: true,
		},
		"size": schema.Int32Attribute{
			Computed: true,
		},
		"is_delete_on_termination": schema.BoolAttribute{
			Computed: true,
		},
		"created_at": schema.StringAttribute{
			Computed: true,
		},
		"is_root": schema.BoolAttribute{
			Computed: true,
		},
	}
}

func getInstanceSecurityGroupsSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed: true,
		},
		"name": schema.StringAttribute{
			Computed: true,
		},
	}
}

var instanceDataSourceSchemaAttributes = getInstanceDataSourceSchema()

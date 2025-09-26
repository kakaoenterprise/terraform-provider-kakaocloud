// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package bcs

import (
	"terraform-provider-kakaocloud/internal/common"
	"terraform-provider-kakaocloud/internal/docs"

	"github.com/hashicorp/terraform-plugin-framework-nettypes/iptypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func getInstanceResourceSchema() map[string]schema.Attribute {
	desc := docs.Bcs("bcs_instance__v1__api__get_instance__model__InstanceModel")
	instanceDesc := docs.Bcs("CreateInstanceModel")

	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
			Description: desc.String("id"),
		},
		"name": schema.StringAttribute{
			Required:    true,
			Description: desc.String("name"),
			Validators:  common.NameValidator(63),
		},
		"description": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: desc.String("description"),
			Validators:  common.DescriptionValidator(),
		},
		"metadata": schema.MapAttribute{
			ElementType: types.StringType,
			Computed:    true,
			Description: desc.String("metadata"),
		},
		"flavor": schema.SingleNestedAttribute{
			Computed:    true,
			Description: desc.String("flavor"),
			Attributes:  getInstanceFlavorResourceSchemaAttributes(),
		},
		"addresses": schema.ListNestedAttribute{
			Computed:    true,
			Description: desc.String("addresses"),
			NestedObject: schema.NestedAttributeObject{
				Attributes: getInstanceAddressesResourceSchemaAttributes(),
			},
		},
		"is_hyper_threading": schema.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Description: desc.String("is_hyper_threading"),
			PlanModifiers: []planmodifier.Bool{
				boolplanmodifier.UseStateForUnknown(),
				boolplanmodifier.RequiresReplace(),
			},
		},
		"is_hadoop": schema.BoolAttribute{
			Computed:    true,
			Description: desc.String("is_hadoop"),
			PlanModifiers: []planmodifier.Bool{
				boolplanmodifier.UseStateForUnknown(),
			},
		},
		"is_k8se": schema.BoolAttribute{
			Computed:    true,
			Description: desc.String("is_k8se"),
			PlanModifiers: []planmodifier.Bool{
				boolplanmodifier.UseStateForUnknown(),
			},
		},
		"image": schema.SingleNestedAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.Object{
				objectplanmodifier.UseStateForUnknown(),
			},
			Description: desc.String("image"),
			Attributes:  getInstanceImageResourceSchemaAttributes(),
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
			Optional:    true,
			Computed:    true,
			Description: desc.String("status"),
			Validators: []validator.String{
				stringvalidator.OneOf(
					common.InstanceStatusActive,
					common.InstanceStatusStopped,
					common.InstanceStatusShelved,
				),
			},
		},
		"user_id": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("user_id"),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"project_id": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("project_id"),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"key_name": schema.StringAttribute{
			Optional:    true,
			Description: desc.String("key_name"),
			Validators:  common.NameValidator(250),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"hostname": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("hostname"),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"availability_zone": schema.StringAttribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
				stringplanmodifier.RequiresReplace(),
			},
			Description: desc.String("availability_zone"),
		},
		"attached_volumes": schema.ListNestedAttribute{
			Computed:    true,
			Description: desc.String("attached_volumes"),
			NestedObject: schema.NestedAttributeObject{
				Attributes: getInstanceAttachedVolumesResourceSchemaAttributes(),
			},
		},
		"attached_volume_count": schema.Int64Attribute{
			Computed:    true,
			Description: desc.String("attached_volume_count"),
		},
		"security_groups": schema.SetNestedAttribute{
			Optional:    true,
			Computed:    true,
			Description: desc.String("security_groups"),
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
			},
			NestedObject: schema.NestedAttributeObject{
				Attributes: getInstanceSecurityGroupsResourceSchemaAttributes(),
			},
		},
		"security_group_count": schema.Int64Attribute{
			Computed:    true,
			Description: desc.String("security_group_count"),
		},
		"instance_type": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("instance_type"),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"created_at": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("created_at"),
		},
		"updated_at": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("updated_at"),
		},

		"image_id": schema.StringAttribute{
			Required:   true,
			Validators: common.UuidValidator(),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
			Description: instanceDesc.String("image_id"),
		},
		"flavor_id": schema.StringAttribute{
			Required:    true,
			Validators:  common.UuidValidator(),
			Description: instanceDesc.String("flavor_id"),
		},
		"subnets": schema.ListNestedAttribute{
			Required:    true,
			Description: instanceDesc.String("subnets"),
			NestedObject: schema.NestedAttributeObject{
				Attributes: getInstanceSubnetsResourceSchemaAttributes(),
			},
		},
		"volumes": schema.ListNestedAttribute{
			Required:    true,
			Description: instanceDesc.String("volumes"),
			NestedObject: schema.NestedAttributeObject{
				Attributes: getInstanceVolumesResourceSchemaAttributes(),
			},
		},
		"user_data": schema.StringAttribute{
			Optional:    true,
			WriteOnly:   true,
			Description: instanceDesc.String("user_data"),
		},
		"is_bonding": schema.BoolAttribute{
			Optional: true,
			PlanModifiers: []planmodifier.Bool{
				boolplanmodifier.RequiresReplace(),
			},
			Description: instanceDesc.String("is_bonding"),
		},
	}
}

func getInstanceFlavorResourceSchemaAttributes() map[string]schema.Attribute {
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

func getInstanceAddressesResourceSchemaAttributes() map[string]schema.Attribute {
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

func getInstanceImageResourceSchemaAttributes() map[string]schema.Attribute {
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

func getInstanceAttachedVolumesResourceSchemaAttributes() map[string]schema.Attribute {
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

func getInstanceSecurityGroupsResourceSchemaAttributes() map[string]schema.Attribute {
	desc := docs.Bcs("bcs_instance__v1__api__get_instance__model__InstanceSecurityGroupModel")

	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("id"),
		},
		"name": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: desc.String("name"),
		},
	}
}

func getInstanceSubnetsResourceSchemaAttributes() map[string]schema.Attribute {
	desc := docs.Bcs("CreateInstanceSubnetModel")

	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Required:    true,
			Validators:  common.UuidValidator(),
			Description: desc.String("id"),
		},
		"network_interface_id": schema.StringAttribute{
			Optional:   true,
			Computed:   true,
			Validators: common.UuidValidator(),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
			Description: desc.String("network_interface_id"),
		},
		"private_ip": schema.StringAttribute{
			Optional:    true,
			CustomType:  iptypes.IPAddressType{},
			Description: desc.String("private_ip"),
		},
	}
}

func getInstanceVolumesResourceSchemaAttributes() map[string]schema.Attribute {
	desc := docs.Bcs("CreateInstanceVolumeModel")

	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: desc.String("uuid"),
			Validators:  common.UuidValidator(),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"is_delete_on_termination": schema.BoolAttribute{
			Optional:    true,
			Description: desc.String("is_delete_on_termination"),
		},
		"size": schema.Int32Attribute{
			Optional:    true,
			Description: desc.String("size"),
			Validators:  common.VolumeSizeValidator(),
			PlanModifiers: []planmodifier.Int32{
				common.PreventShrinkModifier[int32]{
					TypeName:        "Volume Size",
					DescriptionText: "Prevents reducing the volume size",
				},
			},
		},
		"image_id": schema.StringAttribute{
			Optional:    true,
			Description: docs.Description("volume", "CreateVolumeModel", "image_id"),
			Validators:  common.UuidValidator(),
		},
		"type_id": schema.StringAttribute{
			Optional:    true,
			Description: desc.String("type_id"),
			Validators:  common.UuidValidator(),
		},
		"encryption_secret_id": schema.StringAttribute{
			Optional:    true,
			Description: desc.String("encryption_secret_id"),
			Validators:  common.UuidValidator(),
		},
	}
}

var instanceResourceSchemaAttributes = getInstanceResourceSchema()

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package bcs

import (
	"terraform-provider-kakaocloud/internal/common"

	"github.com/hashicorp/terraform-plugin-framework-nettypes/iptypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
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
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"name": schema.StringAttribute{
			Required:   true,
			Validators: common.NameValidator(63),
		},
		"description": schema.StringAttribute{
			Optional:   true,
			Computed:   true,
			Validators: common.DescriptionValidator(),
		},
		"metadata": schema.MapAttribute{
			ElementType: types.StringType,
			Computed:    true,
		},
		"flavor": schema.SingleNestedAttribute{
			Computed:   true,
			Attributes: getInstanceFlavorResourceSchemaAttributes(),
		},
		"addresses": schema.ListNestedAttribute{
			Computed: true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: getInstanceAddressesResourceSchemaAttributes(),
			},
		},
		"is_hyper_threading": schema.BoolAttribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.Bool{
				boolplanmodifier.UseStateForUnknown(),
				boolplanmodifier.RequiresReplace(),
			},
		},
		"is_hadoop": schema.BoolAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.Bool{
				boolplanmodifier.UseStateForUnknown(),
			},
		},
		"is_k8se": schema.BoolAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.Bool{
				boolplanmodifier.UseStateForUnknown(),
			},
		},
		"image": schema.SingleNestedAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.Object{
				objectplanmodifier.UseStateForUnknown(),
			},
			Attributes: getInstanceImageResourceSchemaAttributes(),
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
			Optional: true,
			Computed: true,
			Validators: []validator.String{
				stringvalidator.OneOf(
					common.InstanceStatusActive,
					common.InstanceStatusStopped,
					common.InstanceStatusShelved,
				),
			},
		},
		"user_id": schema.StringAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"project_id": schema.StringAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"key_name": schema.StringAttribute{
			Optional:   true,
			Validators: common.NameValidator(250),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"hostname": schema.StringAttribute{
			Computed: true,
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
		},
		"attached_volumes": schema.ListNestedAttribute{
			Computed: true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: getInstanceAttachedVolumesResourceSchemaAttributes(),
			},
		},
		"attached_volume_count": schema.Int64Attribute{
			Computed: true,
		},
		"security_groups": schema.SetNestedAttribute{
			Computed: true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: getInstanceSecurityGroupsResourceSchemaAttributes(),
			},
		},
		"security_group_count": schema.Int64Attribute{
			Computed: true,
		},
		"instance_type": schema.StringAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"created_at": schema.StringAttribute{
			Computed: true,
		},
		"updated_at": schema.StringAttribute{
			Computed: true,
		},

		"image_id": schema.StringAttribute{
			Required:   true,
			Validators: common.UuidValidator(),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"flavor_id": schema.StringAttribute{
			Required:   true,
			Validators: common.UuidValidator(),
		},
		"subnets": schema.ListNestedAttribute{
			Required: true,
			Validators: []validator.List{
				listvalidator.SizeAtLeast(1),
			},
			NestedObject: schema.NestedAttributeObject{
				Attributes: getInstanceSubnetsResourceSchemaAttributes(),
			},
		},
		"initial_security_groups": schema.SetNestedAttribute{
			Optional: true,
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
			},
			NestedObject: schema.NestedAttributeObject{
				Attributes: getInstanceInitialSecurityGroupsResourceSchemaAttributes(),
			},
		},
		"volumes": schema.ListNestedAttribute{
			Optional: true,
			Validators: []validator.List{
				listvalidator.SizeAtLeast(1),
			},
			NestedObject: schema.NestedAttributeObject{
				Attributes: getInstanceVolumesResourceSchemaAttributes(),
			},
		},
		"user_data": schema.StringAttribute{
			Optional:  true,
			WriteOnly: true,
		},
		"is_bonding": schema.BoolAttribute{
			Optional: true,
			PlanModifiers: []planmodifier.Bool{
				boolplanmodifier.RequiresReplace(),
			},
		},
	}
}

func getInstanceFlavorResourceSchemaAttributes() map[string]schema.Attribute {
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

func getInstanceAddressesResourceSchemaAttributes() map[string]schema.Attribute {
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

func getInstanceImageResourceSchemaAttributes() map[string]schema.Attribute {
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

func getInstanceAttachedVolumesResourceSchemaAttributes() map[string]schema.Attribute {
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

func getInstanceSecurityGroupsResourceSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed: true,
		},
		"name": schema.StringAttribute{
			Optional: true,
			Computed: true,
		},
	}
}

func getInstanceInitialSecurityGroupsResourceSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Required: true,
		},
	}
}

func getInstanceSubnetsResourceSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Required:   true,
			Validators: common.UuidValidator(),
		},
		"network_interface_id": schema.StringAttribute{
			Optional:   true,
			Computed:   true,
			Validators: common.UuidValidator(),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"private_ip": schema.StringAttribute{
			Optional:   true,
			CustomType: iptypes.IPAddressType{},
		},
	}
}

func getInstanceVolumesResourceSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Optional:   true,
			Computed:   true,
			Validators: common.UuidValidator(),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"is_delete_on_termination": schema.BoolAttribute{
			Optional: true,
		},
		"size": schema.Int32Attribute{
			Optional:   true,
			Validators: common.VolumeSizeValidator(),
			PlanModifiers: []planmodifier.Int32{
				common.PreventShrinkModifier[int32]{
					TypeName:        "Volume Size",
					DescriptionText: "Prevents reducing the volume size",
				},
			},
		},
		"image_id": schema.StringAttribute{
			Optional:   true,
			Validators: common.UuidValidator(),
		},
		"type_id": schema.StringAttribute{
			Optional:   true,
			Validators: common.UuidValidator(),
		},
		"encryption_secret_id": schema.StringAttribute{
			Optional:   true,
			Validators: common.UuidValidator(),
		},
	}
}

var instanceResourceSchemaAttributes = getInstanceResourceSchema()

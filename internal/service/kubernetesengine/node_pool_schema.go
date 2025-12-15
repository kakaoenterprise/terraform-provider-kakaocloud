// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package kubernetesengine

import (
	"terraform-provider-kakaocloud/internal/common"

	"github.com/hashicorp/terraform-plugin-framework-validators/float32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kakaoenterprise/kc-sdk-go/services/kubernetesengine"
)

func getNodePoolDataSourceSchema() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"id": dschema.StringAttribute{
			Computed: true,
		},
		"description": dschema.StringAttribute{
			Computed: true,
		},
		"flavor_id": dschema.StringAttribute{
			Computed: true,
		},
		"volume_size": dschema.Int32Attribute{
			Computed: true,
		},
		"node_count": dschema.Int32Attribute{
			Computed: true,
		},
		"ssh_key_name": dschema.StringAttribute{
			Computed: true,
		},
		"is_hyper_threading": dschema.BoolAttribute{
			Computed: true,
		},
		"security_groups": dschema.SetAttribute{
			Computed:    true,
			ElementType: types.StringType,
		},
		"labels": dschema.SetNestedAttribute{
			Computed: true,
			NestedObject: dschema.NestedAttributeObject{
				Attributes: getNodePoolLabelDataSourceSchemaAttributes(),
			},
		},
		"taints": dschema.SetNestedAttribute{
			Computed: true,
			NestedObject: dschema.NestedAttributeObject{
				Attributes: getNodePoolTaintDataSourceSchemaAttributes(),
			},
		},
		"user_data": rschema.StringAttribute{
			Computed: true,
		},
		"vpc_info": dschema.SingleNestedAttribute{
			Computed:   true,
			Attributes: getNodePoolVpcInfoDataSourceSchemaAttributes(),
		},
		"created_at": dschema.StringAttribute{
			Computed: true,
		},
		"failure_message": dschema.StringAttribute{
			Computed: true,
		},
		"is_gpu": dschema.BoolAttribute{
			Computed: true,
		},
		"is_bare_metal": dschema.BoolAttribute{
			Computed: true,
		},
		"is_upgradable": dschema.BoolAttribute{
			Computed: true,
		},
		"flavor": dschema.StringAttribute{
			Computed: true,
		},
		"status": dschema.SingleNestedAttribute{
			Computed:   true,
			Attributes: getNodePoolStatusDataSourceSchemaAttributes(),
		},
		"image": dschema.SingleNestedAttribute{
			Computed:   true,
			Attributes: getNodePoolImageDataSourceSchemaAttributes(),
		},
		"version": dschema.StringAttribute{
			Computed: true,
		},
		"is_cordon": dschema.BoolAttribute{
			Computed: true,
		},
		"autoscaling": dschema.SingleNestedAttribute{
			Computed:   true,
			Attributes: getNodePoolAutoscalingDataSourceSchemaAttributes(),
		},
	}
}

func getNodePoolLabelDataSourceSchemaAttributes() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"key": dschema.StringAttribute{
			Computed: true,
		},
		"value": dschema.StringAttribute{
			Computed: true,
		},
	}
}

func getNodePoolTaintDataSourceSchemaAttributes() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"key": dschema.StringAttribute{
			Computed: true,
		},
		"value": dschema.StringAttribute{
			Computed: true,
		},
		"effect": dschema.StringAttribute{
			Computed: true,
		},
	}
}

func getNodePoolVpcInfoDataSourceSchemaAttributes() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"id": dschema.StringAttribute{
			Computed: true,
		},
		"subnets": dschema.SetNestedAttribute{
			Computed: true,
			NestedObject: dschema.NestedAttributeObject{
				Attributes: getNodePoolSubnetDataSourceSchemaAttributes(),
			},
		},
	}
}

func getNodePoolSubnetDataSourceSchemaAttributes() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"availability_zone": dschema.StringAttribute{
			Computed: true,
		},
		"cidr_block": dschema.StringAttribute{
			Computed: true,
		},
		"id": dschema.StringAttribute{
			Computed: true,
		},
	}
}

func getNodePoolStatusDataSourceSchemaAttributes() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"phase": dschema.StringAttribute{
			Computed: true,
		},
		"available_nodes": dschema.Int32Attribute{
			Computed: true,
		},
		"unavailable_nodes": dschema.Int32Attribute{
			Computed: true,
		},
	}
}

func getNodePoolImageDataSourceSchemaAttributes() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"id": dschema.StringAttribute{
			Computed: true,
		},
		"name": dschema.StringAttribute{
			Computed: true,
		},
		"architecture": dschema.StringAttribute{
			Computed: true,
		},
		"is_gpu_type": dschema.BoolAttribute{
			Computed: true,
		},
		"instance_type": dschema.StringAttribute{
			Computed: true,
		},
		"kernel_version": dschema.StringAttribute{
			Computed: true,
		},
		"key_package": dschema.StringAttribute{
			Computed: true,
		},
		"os_distro": dschema.StringAttribute{
			Computed: true,
		},
		"os_type": dschema.StringAttribute{
			Computed: true,
		},
		"os_version": dschema.StringAttribute{
			Computed: true,
		},
	}
}

func getNodePoolAutoscalingDataSourceSchemaAttributes() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"is_autoscaler_enable": dschema.BoolAttribute{
			Computed: true,
		},
		"autoscaler_desired_node_count": dschema.Int32Attribute{
			Computed: true,
		},
		"autoscaler_max_node_count": dschema.Int32Attribute{
			Computed: true,
		},
		"autoscaler_min_node_count": dschema.Int32Attribute{
			Computed: true,
		},
		"autoscaler_scale_down_unneeded_time": dschema.Int32Attribute{
			Computed: true,
		},
		"autoscaler_scale_down_unready_time": dschema.Int32Attribute{
			Computed: true,
		},
		"autoscaler_scale_down_threshold": dschema.Float32Attribute{
			Computed: true,
		},
	}
}

func getNodePoolResourceSchema() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			Computed: true,
		},
		"cluster_name": rschema.StringAttribute{
			Required:   true,
			Validators: common.NameValidator(20),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"name": rschema.StringAttribute{
			Required:   true,
			Validators: common.NameValidator(20),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"description": rschema.StringAttribute{
			Optional:   true,
			Computed:   true,
			Validators: common.DescriptionValidatorWithMaxLength(60),
		},
		"flavor_id": rschema.StringAttribute{
			Required:   true,
			Validators: common.UuidValidator(),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"volume_size": rschema.Int32Attribute{
			Optional: true,
			Computed: true,
			Validators: []validator.Int32{
				int32validator.Between(30, 5120),
			},
			PlanModifiers: []planmodifier.Int32{
				int32planmodifier.UseStateForUnknown(),
				int32planmodifier.RequiresReplace(),
			},
		},
		"node_count": rschema.Int32Attribute{
			Computed: true,
		},
		"ssh_key_name": rschema.StringAttribute{
			Required: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"is_hyper_threading": rschema.BoolAttribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.Bool{
				boolplanmodifier.UseStateForUnknown(),
				boolplanmodifier.RequiresReplace(),
			},
		},
		"security_groups": rschema.SetAttribute{
			Computed:    true,
			ElementType: types.StringType,
		},
		"labels": rschema.SetNestedAttribute{
			Optional: true,
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
			},
			NestedObject: rschema.NestedAttributeObject{
				Attributes: getNodePoolLabelResourceSchemaAttributes(),
			},
		},
		"taints": rschema.SetNestedAttribute{
			Optional: true,
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
			},
			NestedObject: rschema.NestedAttributeObject{
				Attributes: getNodePoolTaintResourceSchemaAttributes(),
			},
			PlanModifiers: []planmodifier.Set{
				setplanmodifier.RequiresReplace(),
			},
		},
		"user_data": rschema.StringAttribute{
			Optional: true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
		"vpc_info": rschema.SingleNestedAttribute{
			Required:      true,
			PlanModifiers: []planmodifier.Object{objectplanmodifier.RequiresReplace()},
			Attributes:    getNodePoolVpcInfoResourceSchemaAttributes(),
		},
		"autoscaling": rschema.SingleNestedAttribute{
			Optional:   true,
			Computed:   true,
			Attributes: getNodePoolAutoscalingResourceSchemaAttributes(),
			Default: objectdefault.StaticValue(
				types.ObjectValueMust(
					map[string]attr.Type{
						"is_autoscaler_enable":                types.BoolType,
						"autoscaler_desired_node_count":       types.Int32Type,
						"autoscaler_max_node_count":           types.Int32Type,
						"autoscaler_min_node_count":           types.Int32Type,
						"autoscaler_scale_down_unneeded_time": types.Int32Type,
						"autoscaler_scale_down_unready_time":  types.Int32Type,
						"autoscaler_scale_down_threshold":     types.Float32Type,
					},
					map[string]attr.Value{
						"is_autoscaler_enable":                types.BoolValue(false),
						"autoscaler_desired_node_count":       types.Int32Null(),
						"autoscaler_max_node_count":           types.Int32Null(),
						"autoscaler_min_node_count":           types.Int32Null(),
						"autoscaler_scale_down_unneeded_time": types.Int32Null(),
						"autoscaler_scale_down_unready_time":  types.Int32Null(),
						"autoscaler_scale_down_threshold":     types.Float32Null(),
					},
				),
			),
		},

		"image_id": rschema.StringAttribute{
			Required:   true,
			Validators: common.UuidValidator(),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"request_security_groups": rschema.SetAttribute{
			Optional:    true,
			ElementType: types.StringType,
			Validators: []validator.Set{
				setvalidator.ValueStringsAre(common.UuidValidator()...),
			},
		},
		"request_node_count": rschema.Int32Attribute{
			Optional: true,
			Validators: []validator.Int32{
				int32validator.Between(0, 100),
			},
		},
		"minor_version": rschema.StringAttribute{
			Optional:   true,
			Computed:   true,
			Validators: common.MajorMinorVersionValidator(),
		},

		"created_at": rschema.StringAttribute{
			Computed:      true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"failure_message": rschema.StringAttribute{
			Computed: true,
		},
		"is_gpu": rschema.BoolAttribute{
			Computed:      true,
			PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
		},
		"is_bare_metal": rschema.BoolAttribute{
			Computed:      true,
			PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
		},
		"is_upgradable": rschema.BoolAttribute{
			Computed: true,
		},
		"flavor": rschema.StringAttribute{
			Computed: true,
		},
		"status": rschema.SingleNestedAttribute{
			Computed: true,

			Attributes: getNodePoolStatusResourceSchemaAttributes(),
		},
		"image": rschema.SingleNestedAttribute{
			Computed:   true,
			Attributes: getNodePoolImageResourceSchemaAttributes(),
		},
		"version": rschema.StringAttribute{
			Computed: true,
		},
		"is_cordon": rschema.BoolAttribute{
			Computed: true,
		},
	}
}

func getNodePoolLabelResourceSchemaAttributes() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"key": rschema.StringAttribute{
			Required:   true,
			Validators: common.K8sKeyValidator(60),
		},
		"value": rschema.StringAttribute{
			Required:   true,
			Validators: common.K8sValueValidator(63),
		},
	}
}

func getNodePoolTaintResourceSchemaAttributes() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"key": rschema.StringAttribute{
			Required:   true,
			Validators: common.K8sKeyValidator(253),
		},
		"value": rschema.StringAttribute{
			Required:   true,
			Validators: common.K8sValueValidator(63),
		},
		"effect": rschema.StringAttribute{
			Required: true,
			Validators: []validator.String{
				stringvalidator.OneOf(
					string(kubernetesengine.NODEPOOLTAINTEFFECT_NO_EXECUTE),
					string(kubernetesengine.NODEPOOLTAINTEFFECT_NO_SCHEDULE),
					string(kubernetesengine.NODEPOOLTAINTEFFECT_PREFER_NO_SCHEDULE),
				),
			},
		},
	}
}

func getNodePoolVpcInfoResourceSchemaAttributes() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			Required:   true,
			Validators: common.UuidValidator(),
		},
		"subnets": rschema.SetNestedAttribute{
			Required: true,
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
			},
			NestedObject: rschema.NestedAttributeObject{
				Attributes: getNodePoolSubnetResourceSchemaAttributes(),
			},
		},
	}
}

func getNodePoolSubnetResourceSchemaAttributes() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			Required:   true,
			Validators: common.UuidValidator(),
		},
		"availability_zone": rschema.StringAttribute{
			Computed: true,
		},
		"cidr_block": rschema.StringAttribute{
			Computed: true,
		},
	}
}

func getNodePoolStatusResourceSchemaAttributes() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"phase": rschema.StringAttribute{
			Computed: true,
		},
		"available_nodes": rschema.Int32Attribute{
			Computed: true,
		},
		"unavailable_nodes": rschema.Int32Attribute{
			Computed: true,
		},
	}
}

func getNodePoolImageResourceSchemaAttributes() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			Computed: true,
		},
		"name": rschema.StringAttribute{
			Computed: true,
		},
		"architecture": rschema.StringAttribute{
			Computed: true,
		},
		"is_gpu_type": rschema.BoolAttribute{
			Computed: true,
		},
		"instance_type": rschema.StringAttribute{
			Computed: true,
		},
		"kernel_version": rschema.StringAttribute{
			Computed: true,
		},
		"key_package": rschema.StringAttribute{
			Computed: true,
		},
		"os_distro": rschema.StringAttribute{
			Computed: true,
		},
		"os_type": rschema.StringAttribute{
			Computed: true,
		},
		"os_version": rschema.StringAttribute{
			Computed: true,
		},
	}
}

func getNodePoolAutoscalingResourceSchemaAttributes() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"is_autoscaler_enable": rschema.BoolAttribute{
			Required: true,
		},
		"autoscaler_desired_node_count": rschema.Int32Attribute{
			Optional: true,
			Validators: []validator.Int32{
				int32validator.Between(0, 100),
			},
		},
		"autoscaler_max_node_count": rschema.Int32Attribute{
			Optional: true,
			Validators: []validator.Int32{
				int32validator.Between(0, 100),
			},
		},
		"autoscaler_min_node_count": rschema.Int32Attribute{
			Optional: true,
			Validators: []validator.Int32{
				int32validator.Between(0, 100),
			},
		},
		"autoscaler_scale_down_unneeded_time": rschema.Int32Attribute{
			Optional: true,
			Validators: []validator.Int32{
				int32validator.Between(1, 86400),
			},
		},
		"autoscaler_scale_down_unready_time": rschema.Int32Attribute{
			Optional: true,
			Validators: []validator.Int32{
				int32validator.Between(1, 86400),
			},
		},
		"autoscaler_scale_down_threshold": rschema.Float32Attribute{
			Optional: true,
			Validators: []validator.Float32{
				float32validator.Between(0.01, 1.0),
			},
		},
	}
}

var nodePoolDataSourceAttributes = getNodePoolDataSourceSchema()
var nodePoolResourceAttributes = getNodePoolResourceSchema()

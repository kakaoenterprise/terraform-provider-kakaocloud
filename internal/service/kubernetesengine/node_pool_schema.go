// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package kubernetesengine

import (
	"terraform-provider-kakaocloud/internal/common"
	"terraform-provider-kakaocloud/internal/docs"

	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func getNodePoolDataSourceSchema() map[string]dschema.Attribute {
	desc := docs.Kubernetesengine("kubernetes_engine__v1__api__get_node_pool__model__NodePoolResponseModel")

	return map[string]dschema.Attribute{
		"id": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("id"),
		},
		"cluster_name": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("cluster_name"),
		},
		"description": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("description"),
		},
		"flavor_id": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("flavor_id"),
		},
		"volume_size": dschema.Int32Attribute{
			Computed:    true,
			Description: desc.String("volume_size"),
		},
		"node_count": dschema.Int32Attribute{
			Computed:    true,
			Description: desc.String("node_count"),
		},
		"ssh_key_name": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("ssh_key_name"),
		},
		"is_hyper_threading": dschema.BoolAttribute{
			Computed:    true,
			Description: desc.String("is_hyper_threading"),
		},
		"security_groups": dschema.SetAttribute{
			Computed:    true,
			ElementType: types.StringType,
			Description: desc.String("security_groups"),
		},
		"labels": dschema.SetNestedAttribute{
			Computed:    true,
			Description: desc.String("labels"),
			NestedObject: dschema.NestedAttributeObject{
				Attributes: getNodePoolLabelDataSourceSchemaAttributes(),
			},
		},
		"taints": dschema.SetNestedAttribute{
			Computed:    true,
			Description: desc.String("taints"),
			NestedObject: dschema.NestedAttributeObject{
				Attributes: getNodePoolTaintDataSourceSchemaAttributes(),
			},
		},
		"user_data": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("user_data"),
		},
		"vpc_info": dschema.SingleNestedAttribute{
			Computed:    true,
			Description: desc.String("vpc_info"),
			Attributes:  getNodePoolVpcInfoDataSourceSchemaAttributes(),
		},
		"created_at": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("created_at"),
		},
		"failure_message": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("failure_message"),
		},
		"is_gpu": dschema.BoolAttribute{
			Computed:    true,
			Description: desc.String("is_gpu"),
		},
		"is_bare_metal": dschema.BoolAttribute{
			Computed:    true,
			Description: desc.String("is_bare_metal"),
		},
		"is_upgradable": dschema.BoolAttribute{
			Computed:    true,
			Description: desc.String("is_upgradable"),
		},
		"flavor": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("flavor"),
		},
		"status": dschema.SingleNestedAttribute{
			Computed:    true,
			Description: desc.String("status"),
			Attributes:  getNodePoolStatusDataSourceSchemaAttributes(),
		},
		"image": dschema.SingleNestedAttribute{
			Computed:    true,
			Description: desc.String("image"),
			Attributes:  getNodePoolImageDataSourceSchemaAttributes(),
		},
		"version": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("version"),
		},
		"is_cordon": dschema.BoolAttribute{
			Computed:    true,
			Description: desc.String("is_cordon"),
		},
		"autoscaling": dschema.SingleNestedAttribute{
			Computed:    true,
			Description: desc.String("autoscaling"),
			Attributes:  getNodePoolAutoscalingDataSourceSchemaAttributes(),
		},
	}
}

func getNodePoolLabelDataSourceSchemaAttributes() map[string]dschema.Attribute {
	desc := docs.Kubernetesengine("LabelInfoResponseModel")

	return map[string]dschema.Attribute{
		"key": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("key"),
		},
		"value": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("value"),
		},
	}
}

func getNodePoolTaintDataSourceSchemaAttributes() map[string]dschema.Attribute {
	desc := docs.Kubernetesengine("TaintInfoResponseModel")

	return map[string]dschema.Attribute{
		"key": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("key"),
		},
		"value": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("value"),
		},
		"effect": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("effect"),
		},
	}
}

func getNodePoolVpcInfoDataSourceSchemaAttributes() map[string]dschema.Attribute {
	desc := docs.Kubernetesengine("kubernetes_engine__v1__api__get_node_pool__model__VpcInfoResponseModel")

	return map[string]dschema.Attribute{
		"id": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("id"),
		},
		"subnets": dschema.SetNestedAttribute{
			Computed:    true,
			Description: desc.String("subnets"),
			NestedObject: dschema.NestedAttributeObject{
				Attributes: getNodePoolSubnetDataSourceSchemaAttributes(),
			},
		},
	}
}

func getNodePoolSubnetDataSourceSchemaAttributes() map[string]dschema.Attribute {
	desc := docs.Kubernetesengine("kubernetes_engine__v1__api__get_node_pool__model__SubnetResponseModel")

	return map[string]dschema.Attribute{
		"availability_zone": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("availability_zone"),
		},
		"cidr_block": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("cidr_block"),
		},
		"id": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("id"),
		},
	}
}

func getNodePoolStatusDataSourceSchemaAttributes() map[string]dschema.Attribute {
	desc := docs.Kubernetesengine("kubernetes_engine__v1__api__get_node_pool__model__StatusInfoResponseModel")

	return map[string]dschema.Attribute{
		"phase": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("phase"),
		},
		"available_nodes": dschema.Int32Attribute{
			Computed:    true,
			Description: desc.String("available_nodes"),
		},
		"unavailable_nodes": dschema.Int32Attribute{
			Computed:    true,
			Description: desc.String("unavailable_nodes"),
		},
	}
}

func getNodePoolImageDataSourceSchemaAttributes() map[string]dschema.Attribute {
	desc := docs.Kubernetesengine("ImageResponseModel")

	return map[string]dschema.Attribute{
		"id": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("id"),
		},
		"name": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("name"),
		},
		"architecture": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("architecture"),
		},
		"is_gpu_type": dschema.BoolAttribute{
			Computed:    true,
			Description: desc.String("is_gpu_type"),
		},
		"instance_type": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("instance_type"),
		},
		"kernel_version": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("kernel_version"),
		},
		"key_package": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("key_package"),
		},
		"os_distro": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("os_distro"),
		},
		"os_type": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("os_type"),
		},
		"os_version": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("os_version"),
		},
	}
}

func getNodePoolAutoscalingDataSourceSchemaAttributes() map[string]dschema.Attribute {
	desc := docs.Kubernetesengine("NodePoolScalingResourceRequestModel")

	return map[string]dschema.Attribute{
		"is_autoscaler_enable": dschema.BoolAttribute{
			Computed:    true,
			Description: desc.String("is_autoscaler_enable"),
		},
		"autoscaler_desired_node_count": dschema.Int32Attribute{
			Computed:    true,
			Description: desc.String("autoscaler_desired_node_count"),
		},
		"autoscaler_max_node_count": dschema.Int32Attribute{
			Computed:    true,
			Description: desc.String("autoscaler_max_node_count"),
		},
		"autoscaler_min_node_count": dschema.Int32Attribute{
			Computed:    true,
			Description: desc.String("autoscaler_min_node_count"),
		},
		"autoscaler_scale_down_unneeded_time": dschema.Int32Attribute{
			Computed:    true,
			Description: desc.String("autoscaler_scale_down_unneeded_time"),
		},
		"autoscaler_scale_down_unready_time": dschema.Int32Attribute{
			Computed:    true,
			Description: desc.String("autoscaler_scale_down_unready_time"),
		},
		"autoscaler_scale_down_threshold": dschema.Float32Attribute{
			Computed:    true,
			Description: desc.String("autoscaler_scale_down_threshold"),
		},
	}
}

func getNodePoolResourceSchema() map[string]rschema.Attribute {
	desc := docs.Kubernetesengine("kubernetes_engine__v1__api__get_node_pool__model__NodePoolResponseModel")
	createDesc := docs.Kubernetesengine("kubernetes_engine__v1__api__create_node_pool__model__NodePoolRequestModel")

	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("id"),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"cluster_name": rschema.StringAttribute{
			Required:    true,
			Description: desc.String("cluster_name"),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
				stringplanmodifier.RequiresReplace(),
			},
		},
		"name": rschema.StringAttribute{
			Required:    true,
			Description: createDesc.String("name"),
			Validators:  common.NameValidator(200),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
				stringplanmodifier.RequiresReplace(),
			},
		},
		"description": rschema.StringAttribute{
			Optional:      true,
			Computed:      true,
			Validators:    common.DescriptionValidator(),
			Default:       stringdefault.StaticString(""),
			Description:   createDesc.String("description"),
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"flavor_id": rschema.StringAttribute{
			Required:    true,
			Description: createDesc.String("flavor_id"),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
				stringplanmodifier.RequiresReplace(),
			},
		},
		"volume_size": rschema.Int32Attribute{
			Optional:    true,
			Computed:    true,
			Description: createDesc.String("volume_size"),
			PlanModifiers: []planmodifier.Int32{
				int32planmodifier.UseStateForUnknown(),
				int32planmodifier.RequiresReplace(),
			},
		},
		"node_count": rschema.Int32Attribute{
			Optional:      true,
			Computed:      true,
			Description:   createDesc.String("node_count"),
			PlanModifiers: []planmodifier.Int32{int32planmodifier.UseStateForUnknown()},
		},
		"ssh_key_name": rschema.StringAttribute{
			Required:    true,
			Description: createDesc.String("ssh_key_name"),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
				stringplanmodifier.RequiresReplace(),
			},
		},
		"image_id": rschema.StringAttribute{
			Required:    true,
			Description: createDesc.String("image_id"),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
				stringplanmodifier.RequiresReplace(),
			},
		},
		"is_hyper_threading": rschema.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Description: createDesc.String("is_hyper_threading"),
			PlanModifiers: []planmodifier.Bool{
				boolplanmodifier.UseStateForUnknown(),
				boolplanmodifier.RequiresReplace(),
			},
		},
		"security_groups": rschema.SetAttribute{
			Computed:    true,
			ElementType: types.StringType,
			Description: createDesc.String("security_groups"),
		},
		"request_security_groups": rschema.SetAttribute{
			Optional:    true,
			ElementType: types.StringType,
			Description: createDesc.String("security_groups"),
		},
		"labels": rschema.SetNestedAttribute{
			Optional:      true,
			Computed:      true,
			Description:   createDesc.String("labels"),
			PlanModifiers: []planmodifier.Set{setplanmodifier.UseStateForUnknown()},
			NestedObject: rschema.NestedAttributeObject{
				Attributes: getNodePoolLabelResourceSchemaAttributes(),
			},
		},
		"taints": rschema.SetNestedAttribute{
			Optional:      true,
			Computed:      true,
			Description:   createDesc.String("taints"),
			PlanModifiers: []planmodifier.Set{setplanmodifier.UseStateForUnknown()},
			NestedObject: rschema.NestedAttributeObject{
				Attributes: getNodePoolTaintResourceSchemaAttributes(),
			},
		},
		"user_data": rschema.StringAttribute{
			Optional:    true,
			Description: createDesc.String("user_data"),
		},
		"vpc_info": rschema.SingleNestedAttribute{
			Required:      true,
			Description:   createDesc.String("vpc_info"),
			PlanModifiers: []planmodifier.Object{objectplanmodifier.RequiresReplace()},
			Attributes:    getNodePoolVpcInfoResourceSchemaAttributes(),
		},

		"created_at": rschema.StringAttribute{
			Computed:      true,
			Description:   desc.String("created_at"),
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"failure_message": rschema.StringAttribute{
			Computed:      true,
			Description:   desc.String("failure_message"),
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"is_gpu": rschema.BoolAttribute{
			Computed:      true,
			Description:   desc.String("is_gpu"),
			PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
		},
		"is_bare_metal": rschema.BoolAttribute{
			Computed:      true,
			Description:   desc.String("is_bare_metal"),
			PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
		},
		"is_upgradable": rschema.BoolAttribute{
			Computed:      true,
			Description:   desc.String("is_upgradable"),
			PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
		},
		"flavor": rschema.StringAttribute{
			Computed:      true,
			Description:   desc.String("flavor"),
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"status": rschema.SingleNestedAttribute{
			Computed:    true,
			Description: desc.String("status"),

			Attributes: getNodePoolStatusResourceSchemaAttributes(),
		},
		"image": rschema.SingleNestedAttribute{
			Computed:      true,
			Description:   desc.String("image"),
			PlanModifiers: []planmodifier.Object{objectplanmodifier.UseStateForUnknown()},
			Attributes:    getNodePoolImageResourceSchemaAttributes(),
		},
		"version": rschema.StringAttribute{
			Computed:      true,
			Description:   desc.String("version"),
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"is_cordon": rschema.BoolAttribute{
			Computed:      true,
			Description:   desc.String("is_cordon"),
			PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
		},
		"autoscaling": rschema.SingleNestedAttribute{
			Optional:      true,
			Computed:      true,
			Description:   desc.String("autoscaling"),
			PlanModifiers: []planmodifier.Object{objectplanmodifier.UseStateForUnknown()},
			Attributes:    getNodePoolAutoscalingResourceSchemaAttributes(),
		},
	}
}

func getNodePoolLabelResourceSchemaAttributes() map[string]rschema.Attribute {
	desc := docs.Kubernetesengine("LabelRequestModel")

	return map[string]rschema.Attribute{
		"key": rschema.StringAttribute{
			Required:    true,
			Description: desc.String("key"),
		},
		"value": rschema.StringAttribute{
			Required:    true,
			Description: desc.String("value"),
		},
	}
}

func getNodePoolTaintResourceSchemaAttributes() map[string]rschema.Attribute {
	desc := docs.Kubernetesengine("TaintRequestModel")

	return map[string]rschema.Attribute{
		"key": rschema.StringAttribute{
			Required:    true,
			Description: desc.String("key"),
		},
		"value": rschema.StringAttribute{
			Required:    true,
			Description: desc.String("value"),
		},
		"effect": rschema.StringAttribute{
			Required:    true,
			Description: desc.String("effect"),
		},
	}
}

func getNodePoolVpcInfoResourceSchemaAttributes() map[string]rschema.Attribute {
	desc := docs.Kubernetesengine("VpcInfoRequestModel")

	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			Required:    true,
			Description: desc.String("id"),
		},
		"subnets": rschema.SetNestedAttribute{
			Required:    true,
			Description: desc.String("subnets"),
			NestedObject: rschema.NestedAttributeObject{
				Attributes: getNodePoolSubnetResourceSchemaAttributes(),
			},
		},
	}
}

func getNodePoolSubnetResourceSchemaAttributes() map[string]rschema.Attribute {
	desc := docs.Kubernetesengine("kubernetes_engine__v1__api__get_node_pool__model__SubnetResponseModel")

	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			Required:    true,
			Description: desc.String("id"),
		},
		"availability_zone": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("availability_zone"),
		},
		"cidr_block": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("cidr_block"),
		},
	}
}

func getNodePoolStatusResourceSchemaAttributes() map[string]rschema.Attribute {
	desc := docs.Kubernetesengine("kubernetes_engine__v1__api__get_node_pool__model__StatusInfoResponseModel")

	return map[string]rschema.Attribute{
		"phase": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("phase"),
		},
		"available_nodes": rschema.Int32Attribute{
			Computed:    true,
			Description: desc.String("available_nodes"),
		},
		"unavailable_nodes": rschema.Int32Attribute{
			Computed:    true,
			Description: desc.String("unavailable_nodes"),
		},
	}
}

func getNodePoolImageResourceSchemaAttributes() map[string]rschema.Attribute {
	desc := docs.Kubernetesengine("ImageResponseModel")

	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("id"),
		},
		"name": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("name"),
		},
		"architecture": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("architecture"),
		},
		"is_gpu_type": rschema.BoolAttribute{
			Computed:    true,
			Description: desc.String("is_gpu_type"),
		},
		"instance_type": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("instance_type"),
		},
		"kernel_version": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("kernel_version"),
		},
		"key_package": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("key_package"),
		},
		"os_distro": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("os_distro"),
		},
		"os_type": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("os_type"),
		},
		"os_version": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("os_version"),
		},
	}
}

func getNodePoolAutoscalingResourceSchemaAttributes() map[string]rschema.Attribute {
	desc := docs.Kubernetesengine("NodePoolScalingResourceRequestModel")

	return map[string]rschema.Attribute{
		"is_autoscaler_enable": rschema.BoolAttribute{
			Optional:    true,
			Description: desc.String("is_autoscaler_enable"),
		},
		"autoscaler_desired_node_count": rschema.Int32Attribute{
			Optional:    true,
			Description: desc.String("autoscaler_desired_node_count"),
		},
		"autoscaler_max_node_count": rschema.Int32Attribute{
			Optional:    true,
			Description: desc.String("autoscaler_max_node_count"),
		},
		"autoscaler_min_node_count": rschema.Int32Attribute{
			Optional:    true,
			Description: desc.String("autoscaler_min_node_count"),
		},
		"autoscaler_scale_down_unneeded_time": rschema.Int32Attribute{
			Optional:    true,
			Description: desc.String("autoscaler_scale_down_unneeded_time"),
		},
		"autoscaler_scale_down_unready_time": rschema.Int32Attribute{
			Optional:    true,
			Description: desc.String("autoscaler_scale_down_unready_time"),
		},
		"autoscaler_scale_down_threshold": rschema.Float32Attribute{
			Optional:    true,
			Description: desc.String("autoscaler_scale_down_threshold"),
		},
	}
}

var nodePoolDataSourceAttributes = getNodePoolDataSourceSchema()
var nodePoolResourceAttributes = getNodePoolResourceSchema()

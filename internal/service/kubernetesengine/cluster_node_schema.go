// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package kubernetesengine

import (
	"terraform-provider-kakaocloud/internal/docs"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func getNodeDataSourceSchema() map[string]schema.Attribute {
	desc := docs.Kubernetesengine("kubernetes_engine__v1__api__get_cluster_node__model__NodeResponseModel")

	return map[string]schema.Attribute{
		"is_cordon": schema.BoolAttribute{
			Computed:    true,
			Description: desc.String("is_cordon"),
		},
		"created_at": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("created_at"),
		},
		"flavor": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("flavor"),
		},
		"id": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("id"),
		},
		"ip": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("ip"),
		},
		"name": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("name"),
		},
		"node_pool_name": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("node_pool_name"),
		},
		"ssh_key_name": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("ssh_key_name"),
		},
		"failure_message": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("failure_message"),
		},
		"updated_at": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("updated_at"),
		},
		"version": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("version"),
		},
		"volume_size": schema.Int32Attribute{
			Computed:    true,
			Description: desc.String("volume_size"),
		},
		"is_hyper_threading": schema.BoolAttribute{
			Computed:    true,
			Description: desc.String("is_hyper_threading"),
		},

		"image": schema.SingleNestedAttribute{
			Computed:    true,
			Description: desc.String("image"),
			Attributes:  getNodeImageSchemaAttributes(),
		},

		"status": schema.SingleNestedAttribute{
			Computed:    true,
			Description: desc.String("status"),
			Attributes:  getNodeStatusSchemaAttributes(),
		},

		"vpc_info": schema.SingleNestedAttribute{
			Computed:    true,
			Description: desc.String("vpc_info"),
			Attributes:  getNodeVpcInfoSchemaAttributes(),
		},
	}
}

func getNodeImageSchemaAttributes() map[string]schema.Attribute {
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

func getNodeStatusSchemaAttributes() map[string]schema.Attribute {
	desc := docs.Kubernetesengine("kubernetes_engine__v1__api__get_cluster_node__model__StatusInfoResponseModel")

	return map[string]schema.Attribute{
		"phase": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("phase"),
		},
	}
}

func getNodeVpcInfoSchemaAttributes() map[string]schema.Attribute {
	desc := docs.Kubernetesengine("kubernetes_engine__v1__api__get_cluster_node__model__VpcInfoResponseModel")

	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("id"),
		},
		"subnets": schema.SetNestedAttribute{
			Computed:    true,
			Description: desc.String("subnets"),
			NestedObject: schema.NestedAttributeObject{
				Attributes: getNodeSubnetSchemaAttributes(),
			},
		},
	}
}

func getNodeSubnetSchemaAttributes() map[string]schema.Attribute {
	desc := docs.Kubernetesengine("kubernetes_engine__v1__api__get_cluster_node__model__SubnetResponseModel")

	return map[string]schema.Attribute{
		"availability_zone": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("availability_zone"),
		},
		"cidr_block": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("cidr_block"),
		},
		"id": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("id"),
		},
	}
}

func getNodeResourceSchema() map[string]rschema.Attribute {
	deleteDesc := docs.Kubernetesengine("kubernetes_engine__v1__api__delete_cluster_nodes__model__ClusterRequestModel")
	cordonDesc := docs.Kubernetesengine("kubernetes_engine__v1__api__set_cluster_nodes_cordon__model__ClusterRequestModel")
	nodeDesc := docs.Kubernetesengine("kubernetes_engine__v1__api__get_cluster_node__model__NodeResponseModel")

	return map[string]rschema.Attribute{
		"cluster_name": rschema.StringAttribute{
			Required:    true,
			Description: docs.ParameterDescription("kubernetesengine", "delete_cluster_nodes", "path_cluster_name"),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"node_names": rschema.SetAttribute{
			ElementType: types.StringType,
			Required:    true,
			Description: deleteDesc.String("node_names"),
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
			},
			PlanModifiers: []planmodifier.Set{
				setplanmodifier.RequiresReplace(),
			},
		},
		"is_remove": rschema.BoolAttribute{
			Optional:      true,
			Description:   deleteDesc.String("is_remove"),
			PlanModifiers: []planmodifier.Bool{boolplanmodifier.RequiresReplace()},
		},
		"is_cordon": rschema.BoolAttribute{
			Optional:      true,
			Description:   cordonDesc.String("is_cordon"),
			PlanModifiers: []planmodifier.Bool{boolplanmodifier.RequiresReplace()},
		},
		"id": rschema.StringAttribute{
			Computed:    true,
			Description: nodeDesc.String("id"),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
	}
}

var nodeDataSourceSchemaAttributes = getNodeDataSourceSchema()
var nodeResourceSchemaAttributes = getNodeResourceSchema()

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kubernetesengine

import (
	"regexp"
	"terraform-provider-kakaocloud/internal/common"
	"terraform-provider-kakaocloud/internal/docs"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/kakaoenterprise/kc-sdk-go/services/kubernetesengine"
)

func getClusterDataSourceSchema() map[string]dschema.Attribute {
	desc := docs.Kubernetesengine("kubernetes_engine__v1__api__get_cluster__model__ClusterResponseModel")

	return map[string]dschema.Attribute{
		"is_allocate_fip": dschema.BoolAttribute{
			Computed:    true,
			Description: desc.String("is_allocate_fip"),
		},
		"api_version": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("api_version"),
		},
		"network": dschema.SingleNestedAttribute{
			Computed:    true,
			Description: desc.String("network"),
			Attributes:  getClusterNetworkDataSourceSchemaAttributes(),
		},
		"control_plane_endpoint": dschema.SingleNestedAttribute{
			Computed:    true,
			Description: desc.String("control_plane_endpoint"),
			Attributes:  getClusterControlPlaneEndpointDataSourceSchemaAttributes(),
		},
		"created_at": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("created_at"),
		},
		"creator_info": dschema.SingleNestedAttribute{
			Computed:    true,
			Description: desc.String("creator_info"),
			Attributes:  getClusterCreatorInfoDataSourceSchemaAttributes(),
		},
		"version": dschema.SingleNestedAttribute{
			Computed:    true,
			Description: desc.String("version"),
			Attributes:  getClusterVersionDataSourceSchemaAttributes(),
		},
		"description": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("description"),
		},
		"id": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("id"),
		},
		"status": dschema.SingleNestedAttribute{
			Computed:    true,
			Description: desc.String("status"),
			Attributes:  getClusterStatusDataSourceSchemaAttributes(),
		},
		"failure_message": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("failure_message"),
		},
		"is_upgradable": dschema.BoolAttribute{
			Computed:    true,
			Description: desc.String("is_upgradable"),
		},
		"vpc_info": dschema.SingleNestedAttribute{
			Computed:    true,
			Description: desc.String("vpc_info"),
			Attributes:  getClusterVpcInfoDataSourceSchemaAttributes(),
		},
	}
}

func getClusterNetworkDataSourceSchemaAttributes() map[string]dschema.Attribute {
	desc := docs.Kubernetesengine("ClusterNetworkRequestModel")

	return map[string]dschema.Attribute{
		"cni": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("cni"),
		},
		"pod_cidr": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("pod_cidr"),
		},
		"service_cidr": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("service_cidr"),
		},
	}
}

func getClusterControlPlaneEndpointDataSourceSchemaAttributes() map[string]dschema.Attribute {
	desc := docs.Kubernetesengine("ControlPlaneEndpointResponseModel")

	return map[string]dschema.Attribute{
		"host": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("host"),
		},
		"port": dschema.Int32Attribute{
			Computed:    true,
			Description: desc.String("port"),
		},
	}
}

func getClusterCreatorInfoDataSourceSchemaAttributes() map[string]dschema.Attribute {
	desc := docs.Kubernetesengine("CreatorInfoResponseModel")

	return map[string]dschema.Attribute{
		"id": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("id"),
		},
		"name": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("name"),
		},
	}
}

func getClusterVersionDataSourceSchemaAttributes() map[string]dschema.Attribute {
	desc := docs.Kubernetesengine("CurrentOMTResponseModel")

	return map[string]dschema.Attribute{
		"is_deprecated": dschema.BoolAttribute{
			Computed:    true,
			Description: desc.String("is_deprecated"),
		},
		"eol": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("eol"),
		},
		"minor_version": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("minor_version"),
		},
		"next_version": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("next_version"),
		},
		"patch_version": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("patch_version"),
		},
	}
}

func getClusterStatusDataSourceSchemaAttributes() map[string]dschema.Attribute {
	desc := docs.Kubernetesengine("kubernetes_engine__v1__api__get_cluster__model__StatusResponseModel")

	return map[string]dschema.Attribute{
		"phase": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("phase"),
		},
	}
}

func getClusterVpcInfoDataSourceSchemaAttributes() map[string]dschema.Attribute {
	desc := docs.Kubernetesengine("kubernetes_engine__v1__api__get_cluster__model__VPCInfoResponseModel")

	return map[string]dschema.Attribute{
		"id": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("id"),
		},
		"subnets": dschema.SetNestedAttribute{
			Computed:    true,
			Description: desc.String("subnets"),
			NestedObject: dschema.NestedAttributeObject{
				Attributes: getClusterSubnetDataSourceSchemaAttributes(),
			},
		},
	}
}

func getClusterSubnetDataSourceSchemaAttributes() map[string]dschema.Attribute {
	desc := docs.Kubernetesengine("kubernetes_engine__v1__api__get_cluster__model__SubnetResponseModel")

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

func getClusterResourceSchema() map[string]rschema.Attribute {
	desc := docs.Kubernetesengine("kubernetes_engine__v1__api__get_cluster__model__ClusterResponseModel")
	createDesc := docs.Kubernetesengine("kubernetes_engine__v1__api__create_cluster__model__ClusterRequestModel")

	return map[string]rschema.Attribute{
		"is_allocate_fip": rschema.BoolAttribute{
			Required:    true,
			Description: createDesc.String("is_allocate_fip"),
			PlanModifiers: []planmodifier.Bool{
				boolplanmodifier.RequiresReplace(),
			},
		},
		"api_version": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("api_version"),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"network": rschema.SingleNestedAttribute{
			Required:    true,
			Description: desc.String("network"),
			PlanModifiers: []planmodifier.Object{
				objectplanmodifier.RequiresReplace(),
			},
			Attributes: getClusterNetworkResourceSchemaAttributes(),
		},
		"control_plane_endpoint": rschema.SingleNestedAttribute{
			Computed:    true,
			Description: desc.String("control_plane_endpoint"),
			PlanModifiers: []planmodifier.Object{
				objectplanmodifier.UseStateForUnknown(),
			},
			Attributes: getClusterControlPlaneEndpointResourceSchemaAttributes(),
		},
		"created_at": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("created_at"),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"creator_info": rschema.SingleNestedAttribute{
			Computed:    true,
			Description: desc.String("creator_info"),
			PlanModifiers: []planmodifier.Object{
				objectplanmodifier.UseStateForUnknown(),
			},
			Attributes: getClusterCreatorInfoResourceSchemaAttributes(),
		},
		"version": rschema.SingleNestedAttribute{
			Required:    true,
			Description: desc.String("version"),
			Attributes:  getClusterVersionResourceSchemaAttributes(),
		},
		"description": rschema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: createDesc.String("description"),
			Validators: []validator.String{
				stringvalidator.LengthBetween(0, 60),
			},
		},
		"id": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("id"),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"name": rschema.StringAttribute{
			Required:    true,
			Description: createDesc.String("name"),
			Validators: []validator.String{
				stringvalidator.LengthBetween(4, 20),
			},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"failure_message": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("failure_message"),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"is_upgradable": rschema.BoolAttribute{
			Computed:    true,
			Description: desc.String("is_upgradable"),
			PlanModifiers: []planmodifier.Bool{
				boolplanmodifier.UseStateForUnknown(),
			},
		},
		"status": rschema.SingleNestedAttribute{
			Computed:    true,
			Description: desc.String("status"),
			PlanModifiers: []planmodifier.Object{
				objectplanmodifier.UseStateForUnknown(),
			},
			Attributes: getClusterStatusResourceSchemaAttributes(),
		},
		"vpc_info": rschema.SingleNestedAttribute{
			Required:    true,
			Description: desc.String("vpc_info"),
			PlanModifiers: []planmodifier.Object{
				objectplanmodifier.RequiresReplace(),
			},
			Attributes: getClusterVpcInfoResourceSchemaAttributes(),
		},
	}
}

func getClusterNetworkResourceSchemaAttributes() map[string]rschema.Attribute {
	desc := docs.Kubernetesengine("ClusterNetworkRequestModel")

	return map[string]rschema.Attribute{
		"cni": rschema.StringAttribute{
			Required:    true,
			Description: desc.String("cni"),
			Validators: []validator.String{
				stringvalidator.OneOf(
					string(kubernetesengine.CLUSTERNETWORKCNI_CILIUM),
					string(kubernetesengine.CLUSTERNETWORKCNI_CALICO),
				),
			},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"pod_cidr": rschema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: desc.String("pod_cidr"),
			Validators: []validator.String{
				common.NewCIDRPrefixLengthValidator(16, 24),
			},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"service_cidr": rschema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: desc.String("service_cidr"),
			Validators: []validator.String{
				common.NewCIDRPrefixLengthValidator(12, 28),
			},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
				stringplanmodifier.UseStateForUnknown(),
			},
		},
	}
}

func getClusterControlPlaneEndpointResourceSchemaAttributes() map[string]rschema.Attribute {
	desc := docs.Kubernetesengine("ControlPlaneEndpointResponseModel")

	return map[string]rschema.Attribute{
		"host": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("host"),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"port": rschema.Int32Attribute{
			Computed:    true,
			Description: desc.String("port"),
			PlanModifiers: []planmodifier.Int32{
				int32planmodifier.UseStateForUnknown(),
			},
		},
	}
}

func getClusterCreatorInfoResourceSchemaAttributes() map[string]rschema.Attribute {
	desc := docs.Kubernetesengine("CreatorInfoResponseModel")

	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("id"),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"name": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("name"),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
	}
}

func getClusterVersionResourceSchemaAttributes() map[string]rschema.Attribute {
	desc := docs.Kubernetesengine("CurrentOMTResponseModel")

	return map[string]rschema.Attribute{
		"is_deprecated": rschema.BoolAttribute{
			Computed:    true,
			Description: desc.String("is_deprecated"),
			PlanModifiers: []planmodifier.Bool{
				boolplanmodifier.UseStateForUnknown(),
			},
		},
		"eol": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("eol"),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"minor_version": rschema.StringAttribute{
			Required:    true,
			Description: desc.String("minor_version"),
			Validators: []validator.String{
				stringvalidator.RegexMatches(
					regexp.MustCompile(`^\d+\.\d{2}$`),
					"Only major.minor version format is allowed (e.g. 1.30)",
				),
			},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"next_version": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("next_version"),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"patch_version": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("patch_version"),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
	}
}

func getClusterStatusResourceSchemaAttributes() map[string]rschema.Attribute {
	desc := docs.Kubernetesengine("kubernetes_engine__v1__api__get_cluster__model__StatusResponseModel")

	return map[string]rschema.Attribute{
		"phase": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("phase"),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
	}
}

func getClusterVpcInfoResourceSchemaAttributes() map[string]rschema.Attribute {
	desc := docs.Kubernetesengine("kubernetes_engine__v1__api__get_cluster__model__VPCInfoResponseModel")

	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			Required:    true,
			Description: desc.String("id"),
			Validators:  common.UuidValidator(),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"subnets": rschema.SetNestedAttribute{
			Required:    true,
			Description: desc.String("subnets"),
			PlanModifiers: []planmodifier.Set{
				setplanmodifier.RequiresReplace(),
			},
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
			},
			NestedObject: rschema.NestedAttributeObject{
				Attributes: getClusterSubnetResourceSchemaAttributes(),
			},
		},
	}
}

func getClusterSubnetResourceSchemaAttributes() map[string]rschema.Attribute {
	desc := docs.Kubernetesengine("kubernetes_engine__v1__api__get_cluster__model__SubnetResponseModel")

	return map[string]rschema.Attribute{
		"availability_zone": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("availability_zone"),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"cidr_block": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("cidr_block"),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"id": rschema.StringAttribute{
			Required:    true,
			Description: desc.String("id"),
			Validators:  common.UuidValidator(),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
	}
}

var clusterDataSourceAttributes = getClusterDataSourceSchema()
var clusterResourceSchemaAttributes = getClusterResourceSchema()

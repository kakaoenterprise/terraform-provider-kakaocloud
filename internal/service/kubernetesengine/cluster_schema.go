// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package kubernetesengine

import (
	"terraform-provider-kakaocloud/internal/common"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/kakaoenterprise/kc-sdk-go/services/kubernetesengine"
)

func getClusterDataSourceSchema() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"is_allocate_fip": dschema.BoolAttribute{
			Computed: true,
		},
		"api_version": dschema.StringAttribute{
			Computed: true,
		},
		"network": dschema.SingleNestedAttribute{
			Computed:   true,
			Attributes: getClusterNetworkDataSourceSchemaAttributes(),
		},
		"control_plane_endpoint": dschema.SingleNestedAttribute{
			Computed:   true,
			Attributes: getClusterControlPlaneEndpointDataSourceSchemaAttributes(),
		},
		"created_at": dschema.StringAttribute{
			Computed: true,
		},
		"creator_info": dschema.SingleNestedAttribute{
			Computed:   true,
			Attributes: getClusterCreatorInfoDataSourceSchemaAttributes(),
		},
		"version": dschema.SingleNestedAttribute{
			Computed:   true,
			Attributes: getClusterVersionDataSourceSchemaAttributes(),
		},
		"description": dschema.StringAttribute{
			Computed: true,
		},
		"id": dschema.StringAttribute{
			Computed: true,
		},
		"status": dschema.SingleNestedAttribute{
			Computed:   true,
			Attributes: getClusterStatusDataSourceSchemaAttributes(),
		},
		"failure_message": dschema.StringAttribute{
			Computed: true,
		},
		"is_upgradable": dschema.BoolAttribute{
			Computed: true,
		},
		"vpc_info": dschema.SingleNestedAttribute{
			Computed:   true,
			Attributes: getClusterVpcInfoDataSourceSchemaAttributes(),
		},
	}
}

func getClusterNetworkDataSourceSchemaAttributes() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"cni": dschema.StringAttribute{
			Computed: true,
		},
		"pod_cidr": dschema.StringAttribute{
			Computed: true,
		},
		"service_cidr": dschema.StringAttribute{
			Computed: true,
		},
	}
}

func getClusterControlPlaneEndpointDataSourceSchemaAttributes() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"host": dschema.StringAttribute{
			Computed: true,
		},
		"port": dschema.Int32Attribute{
			Computed: true,
		},
	}
}

func getClusterCreatorInfoDataSourceSchemaAttributes() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"id": dschema.StringAttribute{
			Computed: true,
		},
		"name": dschema.StringAttribute{
			Computed: true,
		},
	}
}

func getClusterVersionDataSourceSchemaAttributes() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"is_deprecated": dschema.BoolAttribute{
			Computed: true,
		},
		"eol": dschema.StringAttribute{
			Computed: true,
		},
		"minor_version": dschema.StringAttribute{
			Computed: true,
		},
		"next_version": dschema.StringAttribute{
			Computed: true,
		},
		"patch_version": dschema.StringAttribute{
			Computed: true,
		},
	}
}

func getClusterStatusDataSourceSchemaAttributes() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"phase": dschema.StringAttribute{
			Computed: true,
		},
	}
}

func getClusterVpcInfoDataSourceSchemaAttributes() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"id": dschema.StringAttribute{
			Computed: true,
		},
		"subnets": dschema.SetNestedAttribute{
			Computed: true,
			NestedObject: dschema.NestedAttributeObject{
				Attributes: getClusterSubnetDataSourceSchemaAttributes(),
			},
		},
	}
}

func getClusterSubnetDataSourceSchemaAttributes() map[string]dschema.Attribute {
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

func getClusterResourceSchema() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"is_allocate_fip": rschema.BoolAttribute{
			Required: true,
			PlanModifiers: []planmodifier.Bool{
				boolplanmodifier.RequiresReplace(),
			},
		},
		"api_version": rschema.StringAttribute{
			Computed: true,
		},
		"network": rschema.SingleNestedAttribute{
			Required: true,
			PlanModifiers: []planmodifier.Object{
				objectplanmodifier.RequiresReplace(),
			},
			Attributes: getClusterNetworkResourceSchemaAttributes(),
		},
		"control_plane_endpoint": rschema.SingleNestedAttribute{
			Computed:   true,
			Attributes: getClusterControlPlaneEndpointResourceSchemaAttributes(),
		},
		"created_at": rschema.StringAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"creator_info": rschema.SingleNestedAttribute{
			Computed:   true,
			Attributes: getClusterCreatorInfoResourceSchemaAttributes(),
		},
		"version": rschema.SingleNestedAttribute{
			Required:   true,
			Attributes: getClusterVersionResourceSchemaAttributes(),
		},
		"description": rschema.StringAttribute{
			Optional:   true,
			Computed:   true,
			Validators: common.DescriptionValidatorWithMaxLength(60),
		},
		"id": rschema.StringAttribute{
			Computed: true,
		},
		"name": rschema.StringAttribute{
			Required:   true,
			Validators: common.NameValidator(20),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"failure_message": rschema.StringAttribute{
			Computed: true,
		},
		"is_upgradable": rschema.BoolAttribute{
			Computed: true,
		},
		"status": rschema.SingleNestedAttribute{
			Computed:   true,
			Attributes: getClusterStatusResourceSchemaAttributes(),
		},
		"vpc_info": rschema.SingleNestedAttribute{
			Required: true,
			PlanModifiers: []planmodifier.Object{
				objectplanmodifier.RequiresReplace(),
			},
			Attributes: getClusterVpcInfoResourceSchemaAttributes(),
		},
	}
}

func getClusterNetworkResourceSchemaAttributes() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"cni": rschema.StringAttribute{
			Required: true,
			Validators: []validator.String{
				stringvalidator.OneOf(
					string(kubernetesengine.CLUSTERNETWORKCNI_CILIUM),
					string(kubernetesengine.CLUSTERNETWORKCNI_CALICO),
				),
			},
		},
		"pod_cidr": rschema.StringAttribute{
			Optional: true,
			Computed: true,
			Validators: []validator.String{
				common.NewCIDRPrefixLengthValidator(16, 24),
			},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"service_cidr": rschema.StringAttribute{
			Optional: true,
			Computed: true,
			Validators: []validator.String{
				common.NewCIDRPrefixLengthValidator(12, 28),
			},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
	}
}

func getClusterControlPlaneEndpointResourceSchemaAttributes() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"host": rschema.StringAttribute{
			Computed: true,
		},
		"port": rschema.Int32Attribute{
			Computed: true,
		},
	}
}

func getClusterCreatorInfoResourceSchemaAttributes() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			Computed: true,
		},
		"name": rschema.StringAttribute{
			Computed: true,
		},
	}
}

func getClusterVersionResourceSchemaAttributes() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"is_deprecated": rschema.BoolAttribute{
			Computed: true,
		},
		"eol": rschema.StringAttribute{
			Computed: true,
		},
		"minor_version": rschema.StringAttribute{
			Required:   true,
			Validators: common.MajorMinorVersionValidator(),
		},
		"next_version": rschema.StringAttribute{
			Computed: true,
		},
		"patch_version": rschema.StringAttribute{
			Computed: true,
		},
	}
}

func getClusterStatusResourceSchemaAttributes() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"phase": rschema.StringAttribute{
			Computed: true,
		},
	}
}

func getClusterVpcInfoResourceSchemaAttributes() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			Required:   true,
			Validators: common.UuidValidator(),
		},
		"subnets": rschema.SetNestedAttribute{
			Required: true,
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(2),
			},
			NestedObject: rschema.NestedAttributeObject{
				Attributes: getClusterSubnetResourceSchemaAttributes(),
			},
		},
	}
}

func getClusterSubnetResourceSchemaAttributes() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"availability_zone": rschema.StringAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"cidr_block": rschema.StringAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"id": rschema.StringAttribute{
			Required:   true,
			Validators: common.UuidValidator(),
		},
	}
}

var clusterDataSourceAttributes = getClusterDataSourceSchema()
var clusterResourceSchemaAttributes = getClusterResourceSchema()

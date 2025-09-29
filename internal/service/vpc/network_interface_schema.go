// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package vpc

import (
	"terraform-provider-kakaocloud/internal/common"
	"terraform-provider-kakaocloud/internal/docs"

	"github.com/hashicorp/terraform-plugin-framework-nettypes/iptypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func getAllowedAddressPairResourceSchema() map[string]rschema.Attribute {
	desc := docs.Vpc("bns_vpc__v1__api__get_network_interface__model__AllowedAddressPairModel")

	return map[string]rschema.Attribute{
		"mac_address": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("mac_address"),
		},
		"ip_address": rschema.StringAttribute{
			Required:    true,
			Description: desc.String("ip_address"),
			Validators: []validator.String{
				common.IpOrCIDRValidator{},
			},
		},
	}
}

func getSecurityGroupResourceSchema() map[string]rschema.Attribute {
	desc := docs.Vpc("SecurityGroupModel")

	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			Required:    true,
			Description: desc.String("id"),
			Validators:  common.UuidValidator(),
		},
		"name": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("name"),
		},
	}
}

func getNetworkInterfaceResourceSchema() map[string]rschema.Attribute {
	desc := docs.Vpc("bns_vpc__v1__api__get_network_interface__model__NetworkInterfaceModel")

	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("id"),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"name": rschema.StringAttribute{
			Required:    true,
			Description: desc.String("name"),
			Validators:  common.NameValidator(63),
		},
		"status": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("status"),
		},
		"description": rschema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: desc.String("description"),
			Validators:  common.DescriptionValidator(),
		},
		"project_id": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("project_id"),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"vpc_id": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("vpc_id"),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"subnet_id": rschema.StringAttribute{
			Required:    true,
			Description: desc.String("subnet_id"),
			Validators:  common.UuidValidator(),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"mac_address": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("mac_address"),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"device_id": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("device_id"),
		},
		"device_owner": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("device_owner"),
		},
		"project_name": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("project_name"),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"secondary_ips": rschema.ListAttribute{
			ElementType: types.StringType,
			Computed:    true,
			Description: desc.String("secondary_ips"),
		},
		"public_ip": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("public_ip"),
		},
		"private_ip": rschema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: desc.String("private_ip"),
			CustomType:  iptypes.IPAddressType{},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
				stringplanmodifier.RequiresReplace(),
			},
		},
		"is_network_interface_security_enabled": rschema.BoolAttribute{
			Computed:    true,
			Description: desc.String("is_network_interface_security_enabled"),
			PlanModifiers: []planmodifier.Bool{
				boolplanmodifier.UseStateForUnknown(),
			},
		},
		"allowed_address_pairs": rschema.SetNestedAttribute{
			Optional:    true,
			Computed:    true,
			Description: desc.String("allowed_address_pairs"),
			NestedObject: rschema.NestedAttributeObject{
				Attributes: allowedAddressPairResourceSchema,
			},
		},
		"security_groups": rschema.SetNestedAttribute{
			Required:    true,
			Description: desc.String("security_groups"),
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
			},
			NestedObject: rschema.NestedAttributeObject{
				Attributes: securityGroupResourceSchema,
			},
		},
		"created_at": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("created_at"),
		},
		"updated_at": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("updated_at"),
		},
	}
}

func getAllowedAddressPairDataSourceSchema() map[string]dschema.Attribute {
	desc := docs.Vpc("bns_vpc__v1__api__get_network_interface__model__AllowedAddressPairModel")

	return map[string]dschema.Attribute{
		"mac_address": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("mac_address"),
		},
		"ip_address": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("ip_address"),
		},
	}
}

func getSecurityGroupDataSourceSchema() map[string]dschema.Attribute {
	desc := docs.Vpc("SecurityGroupModel")

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

func getNetworkInterfaceDataSourceBaseSchema() map[string]dschema.Attribute {
	desc := docs.Vpc("bns_vpc__v1__api__get_network_interface__model__NetworkInterfaceModel")

	return map[string]dschema.Attribute{
		"name": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("name"),
		},
		"status": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("status"),
		},
		"description": dschema.StringAttribute{
			Optional:    true,
			Description: desc.String("description"),
		},
		"project_id": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("project_id"),
		},
		"vpc_id": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("vpc_id"),
		},
		"subnet_id": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("subnet_id"),
		},
		"mac_address": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("mac_address"),
		},
		"device_id": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("device_id"),
		},
		"device_owner": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("device_owner"),
		},
		"project_name": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("project_name"),
		},
		"secondary_ips": dschema.ListAttribute{
			ElementType: types.StringType,
			Computed:    true,
			Description: desc.String("secondary_ips"),
		},
		"public_ip": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("public_ip"),
		},
		"private_ip": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("private_ip"),
			CustomType:  iptypes.IPAddressType{},
		},
		"is_network_interface_security_enabled": dschema.BoolAttribute{
			Computed:    true,
			Description: desc.String("is_network_interface_security_enabled"),
		},
		"allowed_address_pairs": dschema.SetNestedAttribute{
			Computed:    true,
			Description: desc.String("allowed_address_pairs"),
			NestedObject: dschema.NestedAttributeObject{
				Attributes: allowedAddressPairDataSourceSchema,
			},
		},
		"security_groups": dschema.SetNestedAttribute{
			Computed:    true,
			Description: desc.String("security_groups"),
			NestedObject: dschema.NestedAttributeObject{
				Attributes: securityGroupDataSourceSchema,
			},
		},
		"created_at": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("created_at"),
		},
		"updated_at": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("updated_at"),
		},
	}
}

var allowedAddressPairDataSourceSchema = getAllowedAddressPairDataSourceSchema()
var securityGroupDataSourceSchema = getSecurityGroupDataSourceSchema()
var networkInterfaceDataSourceBaseSchema = getNetworkInterfaceDataSourceBaseSchema()

var allowedAddressPairResourceSchema = getAllowedAddressPairResourceSchema()
var securityGroupResourceSchema = getSecurityGroupResourceSchema()
var networkInterfaceResourceSchema = getNetworkInterfaceResourceSchema()

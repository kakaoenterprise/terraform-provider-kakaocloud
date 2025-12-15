// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package vpc

import (
	"terraform-provider-kakaocloud/internal/common"

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
	return map[string]rschema.Attribute{
		"mac_address": rschema.StringAttribute{
			Computed: true,
		},
		"ip_address": rschema.StringAttribute{
			Required: true,
			Validators: []validator.String{
				common.IpOrCIDRValidator{},
			},
		},
	}
}

func getSecurityGroupResourceSchema() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			Required:   true,
			Validators: common.UuidValidator(),
		},
		"name": rschema.StringAttribute{
			Computed: true,
		},
	}
}

func getNetworkInterfaceResourceSchema() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"name": rschema.StringAttribute{
			Required:   true,
			Validators: common.NameValidator(63),
		},
		"status": rschema.StringAttribute{
			Computed: true,
		},
		"description": rschema.StringAttribute{
			Optional:   true,
			Computed:   true,
			Validators: common.DescriptionValidator(),
		},
		"project_id": rschema.StringAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"vpc_id": rschema.StringAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"subnet_id": rschema.StringAttribute{
			Required:   true,
			Validators: common.UuidValidator(),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"mac_address": rschema.StringAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"device_id": rschema.StringAttribute{
			Computed: true,
		},
		"device_owner": rschema.StringAttribute{
			Computed: true,
		},
		"project_name": rschema.StringAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"secondary_ips": rschema.ListAttribute{
			ElementType: types.StringType,
			Computed:    true,
		},
		"public_ip": rschema.StringAttribute{
			Computed: true,
		},
		"private_ip": rschema.StringAttribute{
			Optional:   true,
			Computed:   true,
			CustomType: iptypes.IPAddressType{},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
				stringplanmodifier.RequiresReplace(),
			},
		},
		"is_network_interface_security_enabled": rschema.BoolAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.Bool{
				boolplanmodifier.UseStateForUnknown(),
			},
		},
		"allowed_address_pairs": rschema.SetNestedAttribute{
			Optional: true,
			Computed: true,
			NestedObject: rschema.NestedAttributeObject{
				Attributes: allowedAddressPairResourceSchema,
			},
		},
		"security_groups": rschema.SetNestedAttribute{
			Required: true,
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
			},
			NestedObject: rschema.NestedAttributeObject{
				Attributes: securityGroupResourceSchema,
			},
		},
		"created_at": rschema.StringAttribute{
			Computed: true,
		},
		"updated_at": rschema.StringAttribute{
			Computed: true,
		},
	}
}

func getAllowedAddressPairDataSourceSchema() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"mac_address": dschema.StringAttribute{
			Computed: true,
		},
		"ip_address": dschema.StringAttribute{
			Computed: true,
		},
	}
}

func getSecurityGroupDataSourceSchema() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"id": dschema.StringAttribute{
			Computed: true,
		},
		"name": dschema.StringAttribute{
			Computed: true,
		},
	}
}

func getNetworkInterfaceDataSourceBaseSchema() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"name": dschema.StringAttribute{
			Computed: true,
		},
		"status": dschema.StringAttribute{
			Computed: true,
		},
		"description": dschema.StringAttribute{
			Optional: true,
		},
		"project_id": dschema.StringAttribute{
			Computed: true,
		},
		"vpc_id": dschema.StringAttribute{
			Computed: true,
		},
		"subnet_id": dschema.StringAttribute{
			Computed: true,
		},
		"mac_address": dschema.StringAttribute{
			Computed: true,
		},
		"device_id": dschema.StringAttribute{
			Computed: true,
		},
		"device_owner": dschema.StringAttribute{
			Computed: true,
		},
		"project_name": dschema.StringAttribute{
			Computed: true,
		},
		"secondary_ips": dschema.ListAttribute{
			ElementType: types.StringType,
			Computed:    true,
		},
		"public_ip": dschema.StringAttribute{
			Computed: true,
		},
		"private_ip": dschema.StringAttribute{
			Computed:   true,
			CustomType: iptypes.IPAddressType{},
		},
		"is_network_interface_security_enabled": dschema.BoolAttribute{
			Computed: true,
		},
		"allowed_address_pairs": dschema.SetNestedAttribute{
			Computed: true,
			NestedObject: dschema.NestedAttributeObject{
				Attributes: allowedAddressPairDataSourceSchema,
			},
		},
		"security_groups": dschema.SetNestedAttribute{
			Computed: true,
			NestedObject: dschema.NestedAttributeObject{
				Attributes: securityGroupDataSourceSchema,
			},
		},
		"created_at": dschema.StringAttribute{
			Computed: true,
		},
		"updated_at": dschema.StringAttribute{
			Computed: true,
		},
	}
}

var allowedAddressPairDataSourceSchema = getAllowedAddressPairDataSourceSchema()
var securityGroupDataSourceSchema = getSecurityGroupDataSourceSchema()
var networkInterfaceDataSourceBaseSchema = getNetworkInterfaceDataSourceBaseSchema()

var allowedAddressPairResourceSchema = getAllowedAddressPairResourceSchema()
var securityGroupResourceSchema = getSecurityGroupResourceSchema()
var networkInterfaceResourceSchema = getNetworkInterfaceResourceSchema()

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package vpc

import (
	"terraform-provider-kakaocloud/internal/common"

	"github.com/hashicorp/terraform-plugin-framework-nettypes/iptypes"
	datasourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	resourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type networkInterfaceBaseModel struct {
	Id                                types.String      `tfsdk:"id"`
	Name                              types.String      `tfsdk:"name"`
	Status                            types.String      `tfsdk:"status"`
	Description                       types.String      `tfsdk:"description"`
	ProjectId                         types.String      `tfsdk:"project_id"`
	VpcId                             types.String      `tfsdk:"vpc_id"`
	SubnetId                          types.String      `tfsdk:"subnet_id"`
	MacAddress                        types.String      `tfsdk:"mac_address"`
	DeviceId                          types.String      `tfsdk:"device_id"`
	DeviceOwner                       types.String      `tfsdk:"device_owner"`
	ProjectName                       types.String      `tfsdk:"project_name"`
	SecondaryIps                      types.List        `tfsdk:"secondary_ips"`
	PublicIp                          types.String      `tfsdk:"public_ip"`
	PrivateIp                         iptypes.IPAddress `tfsdk:"private_ip"`
	IsNetworkInterfaceSecurityEnabled types.Bool        `tfsdk:"is_network_interface_security_enabled"`
	AllowedAddressPairs               types.Set         `tfsdk:"allowed_address_pairs"`
	SecurityGroups                    types.Set         `tfsdk:"security_groups"`
	CreatedAt                         types.String      `tfsdk:"created_at"`
	UpdatedAt                         types.String      `tfsdk:"updated_at"`
}

type allowedAddressPairModel struct {
	MacAddress types.String `tfsdk:"mac_address"`
	IpAddress  types.String `tfsdk:"ip_address"`
}

var allowedAddressPairAttrType = map[string]attr.Type{
	"mac_address": types.StringType,
	"ip_address":  types.StringType,
}

type securityGroupModel struct {
	Id   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

var securityGroupAttrType = map[string]attr.Type{
	"id":   types.StringType,
	"name": types.StringType,
}

type networkInterfaceResourceModel struct {
	networkInterfaceBaseModel
	Timeouts resourceTimeouts.Value `tfsdk:"timeouts"`
}

type networkInterfaceDataSourceModel struct {
	networkInterfaceBaseModel
	Timeouts datasourceTimeouts.Value `tfsdk:"timeouts"`
}

type networkInterfacesDataSourceModel struct {
	Filter            []common.FilterModel        `tfsdk:"filter"`
	NetworkInterfaces []networkInterfaceBaseModel `tfsdk:"network_interfaces"`
	Timeouts          datasourceTimeouts.Value    `tfsdk:"timeouts"`
}

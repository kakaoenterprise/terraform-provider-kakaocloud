// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package network

import (
	"terraform-provider-kakaocloud/internal/common"

	datasourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	resourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type publicIpBaseModel struct {
	Id              types.String `tfsdk:"id"`
	Status          types.String `tfsdk:"status"`
	Description     types.String `tfsdk:"description"`
	ProjectId       types.String `tfsdk:"project_id"`
	PublicIp        types.String `tfsdk:"public_ip"`
	PrivateIp       types.String `tfsdk:"private_ip"`
	RelatedResource types.Object `tfsdk:"related_resource"`
	CreatedAt       types.String `tfsdk:"created_at"`
	UpdatedAt       types.String `tfsdk:"updated_at"`
}

type resourceModel struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Status      types.String `tfsdk:"status"`
	Type        types.String `tfsdk:"type"`
	DeviceId    types.String `tfsdk:"device_id"`
	DeviceOwner types.String `tfsdk:"device_owner"`
	DeviceType  types.String `tfsdk:"device_type"`
	SubnetId    types.String `tfsdk:"subnet_id"`
	SubnetName  types.String `tfsdk:"subnet_name"`
	SubnetCIDR  types.String `tfsdk:"subnet_cidr"`
	VpcId       types.String `tfsdk:"vpc_id"`
	VpcName     types.String `tfsdk:"vpc_name"`
}

var relatedResourceAttrType = map[string]attr.Type{
	"id":           types.StringType,
	"name":         types.StringType,
	"status":       types.StringType,
	"type":         types.StringType,
	"device_id":    types.StringType,
	"device_owner": types.StringType,
	"device_type":  types.StringType,
	"subnet_id":    types.StringType,
	"subnet_name":  types.StringType,
	"subnet_cidr":  types.StringType,
	"vpc_id":       types.StringType,
	"vpc_name":     types.StringType,
}

type publicIpResourceModel struct {
	publicIpBaseModel
	Timeouts resourceTimeouts.Value `tfsdk:"timeouts"`
}

type publicIpDataSourceModel struct {
	publicIpBaseModel
	Timeouts datasourceTimeouts.Value `tfsdk:"timeouts"`
}

type publicIpsDataSourceModel struct {
	Filter    []common.FilterModel     `tfsdk:"filter"`
	PublicIps []publicIpBaseModel      `tfsdk:"public_ips"`
	Timeouts  datasourceTimeouts.Value `tfsdk:"timeouts"`
}

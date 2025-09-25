// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package loadbalancer

import (
	"terraform-provider-kakaocloud/internal/common"

	datasourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	resourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type beyondLoadBalancerBaseModel struct {
	Id                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	Description        types.String `tfsdk:"description"`
	ProviderName       types.String `tfsdk:"provider_name"`
	Scheme             types.String `tfsdk:"scheme"`
	ProjectId          types.String `tfsdk:"project_id"`
	DnsName            types.String `tfsdk:"dns_name"`
	TypeId             types.String `tfsdk:"type_id"`
	CreatedAt          types.String `tfsdk:"created_at"`
	UpdatedAt          types.String `tfsdk:"updated_at"`
	ProvisioningStatus types.String `tfsdk:"provisioning_status"`
	OperatingStatus    types.String `tfsdk:"operating_status"`
	VpcId              types.String `tfsdk:"vpc_id"`
	Type               types.String `tfsdk:"type"`
	VpcName            types.String `tfsdk:"vpc_name"`
	VpcCidrBlock       types.String `tfsdk:"vpc_cidr_block"`
	AvailabilityZones  types.List   `tfsdk:"availability_zones"`
	LoadBalancers      types.List   `tfsdk:"load_balancers"`
}

type blbLoadBalancerModel struct {
	Id                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	Description        types.String `tfsdk:"description"`
	Type               types.String `tfsdk:"type"`
	ProvisioningStatus types.String `tfsdk:"provisioning_status"`
	OperatingStatus    types.String `tfsdk:"operating_status"`
	AvailabilityZone   types.String `tfsdk:"availability_zone"`
	TypeId             types.String `tfsdk:"type_id"`
	SubnetId           types.String `tfsdk:"subnet_id"`
	SubnetName         types.String `tfsdk:"subnet_name"`
	SubnetCidrBlock    types.String `tfsdk:"subnet_cidr_block"`
	CreatedAt          types.String `tfsdk:"created_at"`
	UpdatedAt          types.String `tfsdk:"updated_at"`
}

type beyondLoadBalancerDataSourceModel struct {
	beyondLoadBalancerBaseModel
	Timeouts datasourceTimeouts.Value `tfsdk:"timeouts"`
}

type beyondLoadBalancerResourceModel struct {
	beyondLoadBalancerBaseModel
	AttachedLoadBalancers types.Set              `tfsdk:"attached_load_balancers"`
	Timeouts              resourceTimeouts.Value `tfsdk:"timeouts"`
}

type beyondLoadBalancersDataSourceModel struct {
	Filter              []common.FilterModel          `tfsdk:"filter"`
	BeyondLoadBalancers []beyondLoadBalancerBaseModel `tfsdk:"beyond_load_balancers"`
	Timeouts            datasourceTimeouts.Value      `tfsdk:"timeouts"`
}

var blbLoadBalancerAttrType = map[string]attr.Type{
	"id":                  types.StringType,
	"name":                types.StringType,
	"description":         types.StringType,
	"type":                types.StringType,
	"provisioning_status": types.StringType,
	"operating_status":    types.StringType,
	"availability_zone":   types.StringType,
	"type_id":             types.StringType,
	"subnet_id":           types.StringType,
	"subnet_name":         types.StringType,
	"subnet_cidr_block":   types.StringType,
	"created_at":          types.StringType,
	"updated_at":          types.StringType,
}

type attachedLoadBalancerModel struct {
	Id               types.String `tfsdk:"id"`
	AvailabilityZone types.String `tfsdk:"availability_zone"`
}

var attachedLoadBalancerAttrType = map[string]attr.Type{
	"id":                types.StringType,
	"availability_zone": types.StringType,
}

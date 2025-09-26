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

type loadBalancerTargetGroupMemberSubnetModel struct {
	Id               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	CidrBlock        types.String `tfsdk:"cidr_block"`
	AvailabilityZone types.String `tfsdk:"availability_zone"`
	HealthCheckIps   types.List   `tfsdk:"health_check_ips"`
}

type loadBalancerTargetGroupMemberSecurityGroupModel struct {
	Id   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type loadBalancerTargetGroupMemberBaseModel struct {
	Id                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	Address            types.String `tfsdk:"address"`
	ProtocolPort       types.Int64  `tfsdk:"protocol_port"`
	SubnetId           types.String `tfsdk:"subnet_id"`
	Weight             types.Int64  `tfsdk:"weight"`
	MonitorPort        types.Int64  `tfsdk:"monitor_port"`
	OperatingStatus    types.String `tfsdk:"operating_status"`
	ProvisioningStatus types.String `tfsdk:"provisioning_status"`
	IsBackup           types.Bool   `tfsdk:"is_backup"`
	ProjectId          types.String `tfsdk:"project_id"`
	CreatedAt          types.String `tfsdk:"created_at"`
	UpdatedAt          types.String `tfsdk:"updated_at"`
	TargetGroupId      types.String `tfsdk:"target_group_id"`

	NetworkInterfaceId types.String `tfsdk:"network_interface_id"`
	InstanceId         types.String `tfsdk:"instance_id"`
	InstanceName       types.String `tfsdk:"instance_name"`
	VpcId              types.String `tfsdk:"vpc_id"`
	Subnet             types.Object `tfsdk:"subnet"`
	SecurityGroups     types.List   `tfsdk:"security_groups"`
}

type loadBalancerTargetGroupMemberResourceModel struct {
	loadBalancerTargetGroupMemberBaseModel
	Timeouts resourceTimeouts.Value `tfsdk:"timeouts"`
}

type loadBalancerTargetGroupMemberListDataSourceModel struct {
	TargetGroupId types.String                                   `tfsdk:"target_group_id"`
	Filter        []common.FilterModel                           `tfsdk:"filter"`
	Members       []loadBalancerTargetGroupMemberListMemberModel `tfsdk:"members"`
	Timeouts      datasourceTimeouts.Value                       `tfsdk:"timeouts"`
}

type loadBalancerTargetGroupMemberListMemberModel struct {
	loadBalancerTargetGroupMemberBaseModel
}

var loadBalancerTargetGroupMemberSubnetAttrType = map[string]attr.Type{
	"id":                types.StringType,
	"name":              types.StringType,
	"cidr_block":        types.StringType,
	"availability_zone": types.StringType,
	"health_check_ips":  types.ListType{ElemType: types.StringType},
}

var loadBalancerTargetGroupMemberSecurityGroupAttrType = map[string]attr.Type{
	"id":   types.StringType,
	"name": types.StringType,
}

var loadBalancerTargetGroupMemberAttrType = map[string]attr.Type{
	"id":                   types.StringType,
	"target_group_id":      types.StringType,
	"name":                 types.StringType,
	"address":              types.StringType,
	"protocol_port":        types.Int64Type,
	"subnet_id":            types.StringType,
	"weight":               types.Int64Type,
	"monitor_port":         types.Int64Type,
	"operating_status":     types.StringType,
	"provisioning_status":  types.StringType,
	"is_backup":            types.BoolType,
	"project_id":           types.StringType,
	"created_at":           types.StringType,
	"updated_at":           types.StringType,
	"network_interface_id": types.StringType,
	"instance_id":          types.StringType,
	"instance_name":        types.StringType,
	"vpc_id":               types.StringType,
	"subnet":               types.ObjectType{AttrTypes: loadBalancerTargetGroupMemberSubnetAttrType},
	"security_groups":      types.ListType{ElemType: types.ObjectType{AttrTypes: loadBalancerTargetGroupMemberSecurityGroupAttrType}},
}

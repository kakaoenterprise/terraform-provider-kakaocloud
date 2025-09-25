// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package loadbalancer

import (
	"terraform-provider-kakaocloud/internal/common"

	datasourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	resourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// accessLogModel defines the access log configuration structure.
type accessLogModel struct {
	Bucket    types.String `tfsdk:"bucket"`
	AccessKey types.String `tfsdk:"access_key"`
	SecretKey types.String `tfsdk:"secret_key"`
}

// loadBalancerBaseModel defines the common attributes for a load balancer.
type loadBalancerBaseModel struct {
	Id                        types.String `tfsdk:"id"`
	Name                      types.String `tfsdk:"name"`
	Description               types.String `tfsdk:"description"`
	Type                      types.String `tfsdk:"type"`
	ListenerIds               types.List   `tfsdk:"listener_ids"`
	ProjectId                 types.String `tfsdk:"project_id"`
	ProvisioningStatus        types.String `tfsdk:"provisioning_status"`
	OperatingStatus           types.String `tfsdk:"operating_status"`
	CreatedAt                 types.String `tfsdk:"created_at"`
	UpdatedAt                 types.String `tfsdk:"updated_at"`
	AvailabilityZone          types.String `tfsdk:"availability_zone"`
	AccessLogs                types.Object `tfsdk:"access_logs"`
	BeyondLoadBalancerId      types.String `tfsdk:"beyond_load_balancer_id"`
	BeyondLoadBalancerName    types.String `tfsdk:"beyond_load_balancer_name"`
	BeyondLoadBalancerDnsName types.String `tfsdk:"beyond_load_balancer_dns_name"`
	TargetGroupCount          types.Int64  `tfsdk:"target_group_count"`
	ListenerCount             types.Int64  `tfsdk:"listener_count"`
	PrivateVip                types.String `tfsdk:"private_vip"`
	PublicVip                 types.String `tfsdk:"public_vip"`
	SubnetName                types.String `tfsdk:"subnet_name"`
	SubnetCidrBlock           types.String `tfsdk:"subnet_cidr_block"`
	VpcId                     types.String `tfsdk:"vpc_id"`
	VpcName                   types.String `tfsdk:"vpc_name"`
	SubnetId                  types.String `tfsdk:"subnet_id"`
}

// loadBalancerDataSourceBaseModel defines the data source model with string access_logs
type loadBalancerDataSourceBaseModel struct {
	Id                        types.String `tfsdk:"id"`
	Name                      types.String `tfsdk:"name"`
	Description               types.String `tfsdk:"description"`
	Type                      types.String `tfsdk:"type"`
	ListenerIds               types.List   `tfsdk:"listener_ids"`
	ProjectId                 types.String `tfsdk:"project_id"`
	ProvisioningStatus        types.String `tfsdk:"provisioning_status"`
	OperatingStatus           types.String `tfsdk:"operating_status"`
	CreatedAt                 types.String `tfsdk:"created_at"`
	UpdatedAt                 types.String `tfsdk:"updated_at"`
	AvailabilityZone          types.String `tfsdk:"availability_zone"`
	AccessLogs                types.String `tfsdk:"access_logs"` // String for data sources
	BeyondLoadBalancerId      types.String `tfsdk:"beyond_load_balancer_id"`
	BeyondLoadBalancerName    types.String `tfsdk:"beyond_load_balancer_name"`
	BeyondLoadBalancerDnsName types.String `tfsdk:"beyond_load_balancer_dns_name"`
	TargetGroupCount          types.Int64  `tfsdk:"target_group_count"`
	ListenerCount             types.Int64  `tfsdk:"listener_count"`
	PrivateVip                types.String `tfsdk:"private_vip"`
	PublicVip                 types.String `tfsdk:"public_vip"`
	SubnetName                types.String `tfsdk:"subnet_name"`
	SubnetCidrBlock           types.String `tfsdk:"subnet_cidr_block"`
	VpcId                     types.String `tfsdk:"vpc_id"`
	VpcName                   types.String `tfsdk:"vpc_name"`
	SubnetId                  types.String `tfsdk:"subnet_id"`
}

// loadBalancerDataSourceModel maps the data source schema to a Go struct.
type loadBalancerDataSourceModel struct {
	loadBalancerDataSourceBaseModel
	Timeouts datasourceTimeouts.Value `tfsdk:"timeouts"`
}

// loadBalancersDataSourceModel maps the plural data source schema to a Go struct.
type loadBalancersDataSourceModel struct {
	Filter        []common.FilterModel              `tfsdk:"filter"`
	LoadBalancers []loadBalancerDataSourceBaseModel `tfsdk:"load_balancers"`
	Timeouts      datasourceTimeouts.Value          `tfsdk:"timeouts"`
}

// loadBalancerResourceModel maps the resource schema to a Go struct.
type loadBalancerResourceModel struct {
	loadBalancerBaseModel
	FlavorId types.String           `tfsdk:"flavor_id"`
	Timeouts resourceTimeouts.Value `tfsdk:"timeouts"`
}

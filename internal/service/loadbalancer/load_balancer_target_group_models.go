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

// loadBalancerTargetGroupBaseModel represents the base model for target group (embedded in resource and data source models)
type loadBalancerTargetGroupBaseModel struct {
	// Core fields (from all operations)
	Id                    types.String `tfsdk:"id"`
	Name                  types.String `tfsdk:"name"`
	Description           types.String `tfsdk:"description"`
	Protocol              types.String `tfsdk:"protocol"`
	LoadBalancerAlgorithm types.String `tfsdk:"load_balancer_algorithm"`
	LoadBalancerId        types.String `tfsdk:"load_balancer_id"`

	// Status fields (from Create/Get responses)
	ProvisioningStatus types.String `tfsdk:"provisioning_status"`
	OperatingStatus    types.String `tfsdk:"operating_status"`
	ProjectId          types.String `tfsdk:"project_id"`
	CreatedAt          types.String `tfsdk:"created_at"`
	UpdatedAt          types.String `tfsdk:"updated_at"`

	// Load balancer details (from Get response only)
	LoadBalancerName               types.String `tfsdk:"load_balancer_name"`
	LoadBalancerProvisioningStatus types.String `tfsdk:"load_balancer_provisioning_status"`
	LoadBalancerType               types.String `tfsdk:"load_balancer_type"`

	// Network details (from Get response only)
	SubnetId         types.String `tfsdk:"subnet_id"`
	SubnetName       types.String `tfsdk:"subnet_name"`
	VpcId            types.String `tfsdk:"vpc_id"`
	VpcName          types.String `tfsdk:"vpc_name"`
	AvailabilityZone types.String `tfsdk:"availability_zone"`

	// Member count (from Get response only)
	MemberCount types.Int64 `tfsdk:"member_count"`

	// Nested objects
	HealthMonitor      types.Object `tfsdk:"health_monitor"`
	SessionPersistence types.Object `tfsdk:"session_persistence"`
	Listeners          types.List   `tfsdk:"listeners"`
}

// loadBalancerTargetGroupResourceModel represents the resource model
type loadBalancerTargetGroupResourceModel struct {
	loadBalancerTargetGroupBaseModel
	ListenerId    types.String           `tfsdk:"listener_id"`
	LoadBalancers types.List             `tfsdk:"load_balancers"`
	Timeouts      resourceTimeouts.Value `tfsdk:"timeouts"`
}

// loadBalancerTargetGroupDataSourceModel represents the single data source model
type loadBalancerTargetGroupDataSourceModel struct {
	loadBalancerTargetGroupBaseModel
	Timeouts datasourceTimeouts.Value `tfsdk:"timeouts"`
}

// loadBalancerTargetGroupListDataSourceModel represents the list data source model
type loadBalancerTargetGroupListDataSourceModel struct {
	Filter       []common.FilterModel               `tfsdk:"filter"`
	TargetGroups []loadBalancerTargetGroupBaseModel `tfsdk:"target_groups"`
	Timeouts     datasourceTimeouts.Value           `tfsdk:"timeouts"`
}

// loadBalancerTargetGroupHealthMonitorModel represents the health monitor configuration
// Based on LoadBalancerPool.HealthMonitor from Get response
type loadBalancerTargetGroupHealthMonitorModel struct {
	Id                 types.String `tfsdk:"id"`
	Type               types.String `tfsdk:"type"`
	Delay              types.Int64  `tfsdk:"delay"`
	Timeout            types.Int64  `tfsdk:"timeout"`
	FallThreshold      types.Int64  `tfsdk:"fall_threshold"` // Maps to fallThreshold
	RiseThreshold      types.Int64  `tfsdk:"rise_threshold"` // Maps to riseThreshold
	HttpMethod         types.String `tfsdk:"http_method"`
	HttpVersion        types.String `tfsdk:"http_version"` // Float converted to string
	ExpectedCodes      types.String `tfsdk:"expected_codes"`
	UrlPath            types.String `tfsdk:"url_path"`
	OperatingStatus    types.String `tfsdk:"operating_status"`
	ProvisioningStatus types.String `tfsdk:"provisioning_status"`
	ProjectId          types.String `tfsdk:"project_id"`
}

// loadBalancerTargetGroupSessionPersistenceModel represents the session persistence configuration
// Based on SessionPersistence domain model
type loadBalancerTargetGroupSessionPersistenceModel struct {
	Type                   types.String `tfsdk:"type"`
	CookieName             types.String `tfsdk:"cookie_name"`
	PersistenceTimeout     types.Int64  `tfsdk:"persistence_timeout"`
	PersistenceGranularity types.String `tfsdk:"persistence_granularity"`
}

// loadBalancerTargetGroupLoadBalancerModel represents load balancer reference
type loadBalancerTargetGroupLoadBalancerModel struct {
	Id types.String `tfsdk:"id"`
}

type loadBalancerTargetGroupMemberModel struct {
	Id                 types.String `tfsdk:"id"`
	Address            types.String `tfsdk:"address"`
	ProtocolPort       types.Int64  `tfsdk:"protocol_port"`
	Weight             types.Int64  `tfsdk:"weight"`
	MonitorPort        types.Int64  `tfsdk:"monitor_port"`
	IsBackup           types.Bool   `tfsdk:"is_backup"`
	SubnetId           types.String `tfsdk:"subnet_id"`
	TargetGroupId      types.String `tfsdk:"target_group_id"`
	Name               types.String `tfsdk:"name"`
	OperatingStatus    types.String `tfsdk:"operating_status"`
	ProvisioningStatus types.String `tfsdk:"provisioning_status"`
	ProjectId          types.String `tfsdk:"project_id"`
	CreatedAt          types.String `tfsdk:"created_at"`
	UpdatedAt          types.String `tfsdk:"updated_at"`
}

type loadBalancerTargetGroupListenerModel struct {
	Id types.String `tfsdk:"id"`
}

type loadBalancerTargetGroupListListenerModel struct {
	Id           types.String `tfsdk:"id"`
	Protocol     types.String `tfsdk:"protocol"`
	ProtocolPort types.Int64  `tfsdk:"protocol_port"`
}

// Attribute type definitions for complex objects
var loadBalancerTargetGroupHealthMonitorAttrType = map[string]attr.Type{
	"id":                  types.StringType,
	"type":                types.StringType,
	"delay":               types.Int64Type,
	"timeout":             types.Int64Type,
	"fall_threshold":      types.Int64Type,
	"rise_threshold":      types.Int64Type,
	"http_method":         types.StringType,
	"http_version":        types.StringType,
	"expected_codes":      types.StringType,
	"url_path":            types.StringType,
	"operating_status":    types.StringType,
	"provisioning_status": types.StringType,
	"project_id":          types.StringType,
}

var loadBalancerTargetGroupSessionPersistenceAttrType = map[string]attr.Type{
	"type":                    types.StringType,
	"cookie_name":             types.StringType,
	"persistence_timeout":     types.Int64Type,
	"persistence_granularity": types.StringType,
}

var loadBalancerTargetGroupLoadBalancerAttrType = map[string]attr.Type{
	"id": types.StringType,
}

var loadBalancerTargetGroupListenerAttrType = map[string]attr.Type{
	"id": types.StringType,
}

var loadBalancerTargetGroupListListenerAttrType = map[string]attr.Type{
	"id":            types.StringType,
	"protocol":      types.StringType,
	"protocol_port": types.Int64Type,
}

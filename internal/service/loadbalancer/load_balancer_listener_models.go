// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package loadbalancer

import (
	"terraform-provider-kakaocloud/internal/common"

	datasourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	resourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type loadBalancerListenerBaseModel struct {
	Id                     types.String `tfsdk:"id"`
	Name                   types.String `tfsdk:"name"`
	Description            types.String `tfsdk:"description"`
	Protocol               types.String `tfsdk:"protocol"`
	IsEnabled              types.Bool   `tfsdk:"is_enabled"`
	TlsCiphers             types.String `tfsdk:"tls_ciphers"`
	TlsVersions            types.List   `tfsdk:"tls_versions"`
	AlpnProtocols          types.List   `tfsdk:"alpn_protocols"`
	ProjectId              types.String `tfsdk:"project_id"`
	ProtocolPort           types.Int64  `tfsdk:"protocol_port"`
	ConnectionLimit        types.Int64  `tfsdk:"connection_limit"`
	LoadBalancerId         types.String `tfsdk:"load_balancer_id"`
	TlsCertificateId       types.String `tfsdk:"tls_certificate_id"`
	ProvisioningStatus     types.String `tfsdk:"provisioning_status"`
	OperatingStatus        types.String `tfsdk:"operating_status"`
	InsertHeaders          types.Object `tfsdk:"insert_headers"`
	CreatedAt              types.String `tfsdk:"created_at"`
	UpdatedAt              types.String `tfsdk:"updated_at"`
	TimeoutClientData      types.Int64  `tfsdk:"timeout_client_data"`
	DefaultTargetGroupName types.String `tfsdk:"default_target_group_name"`
	DefaultTargetGroupId   types.String `tfsdk:"default_target_group_id"`
	LoadBalancerType       types.String `tfsdk:"load_balancer_type"`
	Secrets                types.List   `tfsdk:"secrets"`
	L7Policies             types.List   `tfsdk:"l7_policies"`
}

type loadBalancerListenerDataSourceModel struct {
	loadBalancerListenerBaseModel
	Timeouts datasourceTimeouts.Value `tfsdk:"timeouts"`
}

type loadBalancerListenerResourceModel struct {
	loadBalancerListenerBaseModel
	TargetGroupId          types.String `tfsdk:"target_group_id"`
	SniContainerRefs       types.List   `tfsdk:"sni_container_refs"`
	DefaultTlsContainerRef types.String `tfsdk:"default_tls_container_ref"`
	TlsMinVersion          types.String `tfsdk:"tls_min_version"`

	Timeouts resourceTimeouts.Value `tfsdk:"timeouts"`
}

type loadBalancerListenersDataSourceModel struct {
	Filter    []common.FilterModel            `tfsdk:"filter"`
	Listeners []loadBalancerListenerBaseModel `tfsdk:"listeners"`
	Timeouts  datasourceTimeouts.Value        `tfsdk:"timeouts"`
}

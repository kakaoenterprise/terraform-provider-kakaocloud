// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package loadbalancer

import (
	"terraform-provider-kakaocloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kakaoenterprise/kc-sdk-go/v2/services/loadbalancer"
)

func mapLoadBalancerTargetGroupMemberToCreateRequest(
	model *loadBalancerTargetGroupMemberResourceModel,
) *loadbalancer.AddTarget {
	createReq := loadbalancer.NewAddTarget(
		model.Address.ValueString(),
		model.ProtocolPort.ValueInt32(),
		model.SubnetId.ValueString(),
	)

	if !model.Name.IsNull() {
		createReq.SetName(model.Name.ValueString())
	}

	if !model.Weight.IsNull() && !model.Weight.IsUnknown() {
		createReq.SetWeight(model.Weight.ValueInt32())
	}

	if !model.MonitorPort.IsNull() && !model.MonitorPort.IsUnknown() {
		createReq.SetMonitorPort(model.MonitorPort.ValueInt32())
	}

	return createReq
}

func mapLoadBalancerTargetGroupMemberFromGetResponse(
	model *loadBalancerTargetGroupMemberResourceModel,
	s *loadbalancer.PoolMember,
) bool {
	model.Id = utils.ConvertNullableString(s.Id)
	model.Name = utils.ConvertNullableString(s.Name)
	model.Address = utils.ConvertNullableString(s.IpAddress)
	model.ProjectId = utils.ConvertNullableString(s.ProjectId)
	model.CreatedAt = utils.ConvertNullableTime(s.CreatedAt)
	model.UpdatedAt = utils.ConvertNullableTime(s.UpdatedAt)
	model.OperatingStatus = types.StringValue(string(s.OperatingStatus))
	model.ProvisioningStatus = utils.ConvertNullableString(s.ProvisioningStatus)

	if s.ProtocolPort.IsSet() && s.ProtocolPort.Get() != nil {
		model.ProtocolPort = types.Int32Value(*s.ProtocolPort.Get())
	} else {
		model.ProtocolPort = types.Int32Null()
	}

	model.SubnetId = utils.ConvertNullableString(s.Subnet.Id)

	if s.Weight.IsSet() && s.Weight.Get() != nil {
		model.Weight = types.Int32Value(*s.Weight.Get())
	} else {
		model.Weight = types.Int32Null()
	}
	model.IsBackup = types.BoolValue(false)

	if s.MonitorPort.IsSet() && s.MonitorPort.Get() != nil {
		model.MonitorPort = types.Int32Value(*s.MonitorPort.Get())
	} else {
		model.MonitorPort = types.Int32Null()
	}

	model.NetworkInterfaceId = utils.ConvertNullableString(s.NetworkInterfaceId)
	model.InstanceId = utils.ConvertNullableString(s.InstanceId)
	model.InstanceName = utils.ConvertNullableString(s.InstanceName)
	model.VpcId = utils.ConvertNullableString(s.VpcId)

	var healthCheckIps types.List
	if len(s.Subnet.HealthCheckIps) > 0 {
		healthCheckIpValues := make([]attr.Value, len(s.Subnet.HealthCheckIps))
		for i, ip := range s.Subnet.HealthCheckIps {
			healthCheckIpValues[i] = types.StringValue(ip)
		}
		healthCheckIps = types.ListValueMust(types.StringType, healthCheckIpValues)
	} else {
		healthCheckIps = types.ListNull(types.StringType)
	}

	var availabilityZone types.String
	if s.Subnet.AvailabilityZone.IsSet() && s.Subnet.AvailabilityZone.Get() != nil {
		availabilityZone = types.StringValue(string(*s.Subnet.AvailabilityZone.Get()))
	} else {
		availabilityZone = types.StringNull()
	}

	model.Subnet = types.ObjectValueMust(loadBalancerTargetGroupMemberSubnetAttrType, map[string]attr.Value{
		"id":                utils.ConvertNullableString(s.Subnet.Id),
		"name":              utils.ConvertNullableString(s.Subnet.Name),
		"cidr_block":        utils.ConvertNullableString(s.Subnet.CidrBlock),
		"availability_zone": availabilityZone,
		"health_check_ips":  healthCheckIps,
	})

	if len(s.SecurityGroups) > 0 {
		securityGroupValues := make([]attr.Value, len(s.SecurityGroups))
		for i, sg := range s.SecurityGroups {
			securityGroupValues[i] = types.ObjectValueMust(loadBalancerTargetGroupMemberSecurityGroupAttrType, map[string]attr.Value{
				"id":   types.StringValue(sg.Id),
				"name": types.StringValue(sg.Name),
			})
		}
		model.SecurityGroups = types.ListValueMust(types.ObjectType{AttrTypes: loadBalancerTargetGroupMemberSecurityGroupAttrType}, securityGroupValues)
	} else {
		model.SecurityGroups = types.ListNull(types.ObjectType{AttrTypes: loadBalancerTargetGroupMemberSecurityGroupAttrType})
	}

	return true
}

func mapLoadBalancerTargetGroupMemberListFromGetResponse(
	model *loadBalancerTargetGroupMemberListDataSourceModel,
	src *loadbalancer.ListTargetsInTargetGroupResponse,
	diags *diag.Diagnostics,
) bool {
	if src.Members == nil {
		model.Members = []loadBalancerTargetGroupMemberBaseModel{}
		return true
	}

	members := make([]loadBalancerTargetGroupMemberBaseModel, 0, len(src.Members))
	for _, member := range src.Members {
		var memberModel loadBalancerTargetGroupMemberResourceModel
		if mapLoadBalancerTargetGroupMemberFromGetResponse(&memberModel, &member) {
			members = append(members, memberModel.loadBalancerTargetGroupMemberBaseModel)
		}
	}

	model.Members = members
	return !diags.HasError()
}

func mapLoadBalancerTargetGroupMembersToBatchRequest(
	model *loadBalancerTargetGroupMemberListResourceModel,
) *loadbalancer.UpdateTargetsRequest {
	membersReq := make([]loadbalancer.UpdateTargets, 0, len(model.Members))

	for _, member := range model.Members {
		memberReq := loadbalancer.NewUpdateTargets(
			member.Address.ValueString(),
			member.ProtocolPort.ValueInt32(),
			member.SubnetId.ValueString(),
		)

		if !member.Name.IsNull() && !member.Name.IsUnknown() {
			memberReq.SetName(member.Name.ValueString())
		}
		if !member.Weight.IsNull() && !member.Weight.IsUnknown() {
			memberReq.SetWeight(member.Weight.ValueInt32())
		}
		if !member.MonitorPort.IsNull() && !member.MonitorPort.IsUnknown() {
			memberReq.SetMonitorPort(member.MonitorPort.ValueInt32())
		}

		membersReq = append(membersReq, *memberReq)
	}

	return loadbalancer.NewUpdateTargetsRequest(membersReq)
}

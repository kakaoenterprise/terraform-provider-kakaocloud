// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package loadbalancer

import (
	"context"
	"terraform-provider-kakaocloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kakaoenterprise/kc-sdk-go/services/loadbalancer"
)

func mapMemberBaseFields(
	ctx context.Context,
	model *loadBalancerTargetGroupMemberBaseModel,
	src interface{},
	diags *diag.Diagnostics,
) bool {
	switch s := src.(type) {
	case *loadbalancer.BnsLoadBalancerV1ApiAddTargetModelResponseTargetGroupMemberModel:

		model.Id = types.StringValue(s.Member.Id)
		model.Name = utils.ConvertNullableString(s.Member.Name)
		model.Address = types.StringValue(s.Member.Address)
		model.ProtocolPort = types.Int64Value(int64(s.Member.ProtocolPort))
		model.SubnetId = types.StringValue(s.Member.SubnetId)
		model.Weight = types.Int64Value(int64(s.Member.Weight))
		model.IsBackup = types.BoolValue(s.Member.IsBackup)
		model.ProjectId = types.StringValue(s.Member.ProjectId)
		model.CreatedAt = types.StringValue(s.Member.CreatedAt.Format("2006-01-02T15:04:05Z"))
		model.UpdatedAt = utils.ConvertNullableTime(s.Member.UpdatedAt)
		model.OperatingStatus = types.StringValue(string(s.Member.OperatingStatus))
		model.ProvisioningStatus = types.StringValue(string(s.Member.ProvisioningStatus))

		if s.Member.MonitorPort.IsSet() && s.Member.MonitorPort.Get() != nil {
			model.MonitorPort = types.Int64Value(int64(*s.Member.MonitorPort.Get()))
		} else {
			model.MonitorPort = types.Int64Null()
		}

		model.Subnet = types.ObjectNull(loadBalancerTargetGroupMemberSubnetAttrType)

		model.NetworkInterfaceId = types.StringNull()
		model.InstanceId = types.StringNull()
		model.InstanceName = types.StringNull()
		model.VpcId = types.StringNull()

		model.SecurityGroups = types.ListNull(types.ObjectType{AttrTypes: loadBalancerTargetGroupMemberSecurityGroupAttrType})

	case *loadbalancer.BnsLoadBalancerV1ApiListTargetsInTargetGroupModelTargetGroupMemberModel:

		model.Id = types.StringValue(s.Id)
		model.Name = utils.ConvertNullableString(s.Name)
		model.Address = utils.ConvertNullableString(s.IpAddress)
		model.ProjectId = utils.ConvertNullableString(s.ProjectId)
		model.CreatedAt = utils.ConvertNullableTime(s.CreatedAt)
		model.UpdatedAt = utils.ConvertNullableTime(s.UpdatedAt)
		model.OperatingStatus = types.StringValue(string(s.OperatingStatus))
		model.ProvisioningStatus = utils.ConvertNullableString(s.ProvisioningStatus)

		if s.ProtocolPort.IsSet() && s.ProtocolPort.Get() != nil {
			model.ProtocolPort = types.Int64Value(int64(*s.ProtocolPort.Get()))
		} else {
			model.ProtocolPort = types.Int64Null()
		}

		model.SubnetId = types.StringValue(s.Subnet.Id)

		if s.Weight.IsSet() && s.Weight.Get() != nil {
			model.Weight = types.Int64Value(int64(*s.Weight.Get()))
		} else {
			model.Weight = types.Int64Null()
		}

		if s.MonitorPort.IsSet() && s.MonitorPort.Get() != nil {
			model.MonitorPort = types.Int64Value(int64(*s.MonitorPort.Get()))
		} else {
			model.MonitorPort = types.Int64Null()
		}

		model.IsBackup = types.BoolValue(false)

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
			"id":                types.StringValue(s.Subnet.Id),
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

	case *loadbalancer.BnsLoadBalancerV1ApiUpdateTargetModelResponseTargetGroupMemberModel:

		model.Id = types.StringValue(s.Member.Id)
		model.Name = utils.ConvertNullableString(s.Member.Name)
		model.Address = types.StringValue(s.Member.Address)
		model.ProtocolPort = types.Int64Value(int64(s.Member.ProtocolPort))
		model.SubnetId = types.StringValue(s.Member.SubnetId)
		model.Weight = types.Int64Value(int64(s.Member.Weight))
		model.IsBackup = types.BoolValue(s.Member.IsBackup)
		model.ProjectId = types.StringValue(s.Member.ProjectId)
		model.CreatedAt = types.StringValue(s.Member.CreatedAt.Format("2006-01-02T15:04:05Z"))
		model.UpdatedAt = utils.ConvertNullableTime(s.Member.UpdatedAt)
		model.OperatingStatus = types.StringValue(string(s.Member.OperatingStatus))
		model.ProvisioningStatus = types.StringValue(string(s.Member.ProvisioningStatus))

		if s.Member.MonitorPort.IsSet() && s.Member.MonitorPort.Get() != nil {
			model.MonitorPort = types.Int64Value(int64(*s.Member.MonitorPort.Get()))
		} else {
			model.MonitorPort = types.Int64Null()
		}

		model.Subnet = types.ObjectNull(loadBalancerTargetGroupMemberSubnetAttrType)

		model.NetworkInterfaceId = types.StringNull()
		model.InstanceId = types.StringNull()
		model.InstanceName = types.StringNull()
		model.VpcId = types.StringNull()

		model.SecurityGroups = types.ListNull(types.ObjectType{AttrTypes: loadBalancerTargetGroupMemberSecurityGroupAttrType})

	default:
		diags.AddError("Unknown source type", "Unsupported source type for member mapping")
		return false
	}

	return true
}

func mapLoadBalancerTargetGroupMemberToCreateRequest(
	ctx context.Context,
	model *loadBalancerTargetGroupMemberResourceModel,
	diags *diag.Diagnostics,
) *loadbalancer.CreateTargetGroupMember {
	createReq := loadbalancer.NewCreateTargetGroupMember(
		model.Address.ValueString(),
		int32(model.ProtocolPort.ValueInt64()),
		model.SubnetId.ValueString(),
	)

	if !model.Name.IsNull() && !model.Name.IsUnknown() {
		createReq.SetName(model.Name.ValueString())
	}

	if !model.Weight.IsNull() && !model.Weight.IsUnknown() {
		createReq.SetWeight(int32(model.Weight.ValueInt64()))
	}

	if !model.MonitorPort.IsNull() && !model.MonitorPort.IsUnknown() {
		createReq.SetMonitorPort(int32(model.MonitorPort.ValueInt64()))
	}

	return createReq
}

func mapLoadBalancerTargetGroupMemberFromGetResponse(
	ctx context.Context,
	model *loadBalancerTargetGroupMemberResourceModel,
	src *loadbalancer.BnsLoadBalancerV1ApiListTargetsInTargetGroupModelTargetGroupMemberModel,
	diags *diag.Diagnostics,
) bool {
	return mapMemberBaseFields(ctx, &model.loadBalancerTargetGroupMemberBaseModel, src, diags)
}

func mapLoadBalancerTargetGroupMemberToUpdateRequest(
	ctx context.Context,
	model *loadBalancerTargetGroupMemberResourceModel,
	diags *diag.Diagnostics,
) *loadbalancer.BnsLoadBalancerV1ApiUpdateTargetModelEditTargetGroupMember {
	updateReq := loadbalancer.NewBnsLoadBalancerV1ApiUpdateTargetModelEditTargetGroupMember()

	if !model.Name.IsNull() && !model.Name.IsUnknown() {
		updateReq.SetName(model.Name.ValueString())
	}

	if !model.Weight.IsNull() && !model.Weight.IsUnknown() {
		updateReq.SetWeight(int32(model.Weight.ValueInt64()))
	}

	if !model.MonitorPort.IsNull() && !model.MonitorPort.IsUnknown() {
		updateReq.SetMonitorPort(int32(model.MonitorPort.ValueInt64()))
	}

	return updateReq
}

func mapLoadBalancerTargetGroupMemberFromUpdateResponse(
	ctx context.Context,
	model *loadBalancerTargetGroupMemberResourceModel,
	src *loadbalancer.BnsLoadBalancerV1ApiUpdateTargetModelResponseTargetGroupMemberModel,
	diags *diag.Diagnostics,
) bool {
	return mapMemberBaseFields(ctx, &model.loadBalancerTargetGroupMemberBaseModel, src, diags)
}

func mapLoadBalancerTargetGroupMemberListFromGetResponse(
	ctx context.Context,
	model *loadBalancerTargetGroupMemberListDataSourceModel,
	src *loadbalancer.TargetGroupMemberListModel,
	diags *diag.Diagnostics,
) bool {
	if src.Members == nil {
		model.Members = []loadBalancerTargetGroupMemberListMemberModel{}
		return true
	}

	members := make([]loadBalancerTargetGroupMemberListMemberModel, 0, len(src.Members))
	for _, member := range src.Members {
		var memberModel loadBalancerTargetGroupMemberResourceModel
		if mapLoadBalancerTargetGroupMemberFromGetResponse(ctx, &memberModel, &member, diags) {
			members = append(members, loadBalancerTargetGroupMemberListMemberModel{
				loadBalancerTargetGroupMemberBaseModel: memberModel.loadBalancerTargetGroupMemberBaseModel,
			})

			members[len(members)-1].TargetGroupId = model.TargetGroupId
		}
	}

	model.Members = members
	return !diags.HasError()
}

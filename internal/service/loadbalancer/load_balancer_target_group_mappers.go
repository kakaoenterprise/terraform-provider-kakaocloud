// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package loadbalancer

import (
	"context"
	"fmt"
	"terraform-provider-kakaocloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/kakaoenterprise/kc-sdk-go/v2/services/loadbalancer"
)

func mapLoadBalancerTargetGroupSessionPersistenceFromResponse(
	ctx context.Context,
	src *loadbalancer.SessionPersistence,
) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	if src == nil {
		return types.ObjectNull(loadBalancerTargetGroupSessionPersistenceAttrType), diags
	}

	sessionPersistence := loadBalancerTargetGroupSessionPersistence{
		Type:                   types.StringValue(string(src.Type)),
		CookieName:             utils.ConvertNullableString(src.CookieName),
		PersistenceTimeout:     types.Int64Value(int64(src.PersistenceTimeout)),
		PersistenceGranularity: utils.ConvertNullableString(src.PersistenceGranularity),
	}

	obj, objDiags := types.ObjectValueFrom(ctx, loadBalancerTargetGroupSessionPersistenceAttrType, sessionPersistence)
	diags.Append(objDiags...)
	return obj, diags
}

func mapLoadBalancerTargetGroupToCreateRequest(
	ctx context.Context,
	model *loadBalancerTargetGroupResourceModel,
	diags *diag.Diagnostics,
) *loadbalancer.CreateTargetGroup {

	protocol := loadbalancer.LoadBalancerPoolProtocol(model.Protocol.ValueString())
	algorithm := loadbalancer.LoadBalancerPoolAlgorithm(model.LoadBalancerAlgorithm.ValueString())

	createReq := loadbalancer.NewCreateTargetGroup(
		algorithm,
		model.Name.ValueString(),
		protocol,
	)
	createReq.SetLoadBalancerId(model.LoadBalancerId.ValueString())

	if !model.ListenerId.IsNull() {
		createReq.SetListenerId(model.ListenerId.ValueString())
	}

	if !model.Description.IsNull() && !model.Description.IsUnknown() {
		createReq.SetDescription(model.Description.ValueString())
	}

	if !model.LoadBalancerAlpnProtocols.IsNull() && !model.LoadBalancerAlpnProtocols.IsUnknown() {
		var vals []string
		diags.Append(model.LoadBalancerAlpnProtocols.ElementsAs(ctx, &vals, false)...)
		if diags.HasError() {
			return nil
		}

		if len(vals) == 1 {
			p := loadbalancer.LoadBalancerAlpnProtocol(vals[0])
			createReq.SetAlpnProtocols(p)
		}
	}

	if !model.SessionPersistence.IsNull() {
		var sessionPersistence loadBalancerTargetGroupSessionPersistence
		diags.Append(model.SessionPersistence.As(ctx, &sessionPersistence, basetypes.ObjectAsOptions{})...)
		if !diags.HasError() {
			sessionPersistenceReq := loadbalancer.NewSessionPersistenceRequest(
				loadbalancer.LoadBalancerSessionPersistenceType(sessionPersistence.Type.ValueString()),
				int32(sessionPersistence.PersistenceTimeout.ValueInt64()),
			)

			if !sessionPersistence.CookieName.IsNull() {
				sessionPersistenceReq.CookieName.Set(new(sessionPersistence.CookieName.ValueString()))
			}
			if !sessionPersistence.PersistenceGranularity.IsNull() {
				sessionPersistenceReq.PersistenceGranularity.Set(new(sessionPersistence.PersistenceGranularity.ValueString()))
			}

			createReq.SetSessionPersistence(*sessionPersistenceReq)
		}
	}

	return createReq
}

func mapLoadBalancerTargetGroupFromGetResponse(
	ctx context.Context,
	model *loadBalancerTargetGroupBaseModel,
	src *loadbalancer.LoadBalancerPool,
	diags *diag.Diagnostics,
) bool {

	model.Id = types.StringValue(src.Id)
	model.Name = utils.ConvertNullableString(src.Name)
	model.Description = utils.ConvertNullableString(src.Description)
	model.Protocol = utils.ConvertNullableString(src.Protocol)
	model.LoadBalancerAlgorithm = utils.ConvertNullableString(src.LoadBalancerAlgorithm)
	model.SubnetId = utils.ConvertNullableString(src.SubnetId)
	model.VpcId = utils.ConvertNullableString(src.VpcId)
	model.AvailabilityZone = utils.ConvertNullableString(src.AvailabilityZone)
	model.ProvisioningStatus = utils.ConvertNullableString(src.ProvisioningStatus)
	model.OperatingStatus = utils.ConvertNullableString(src.OperatingStatus)
	model.ProjectId = utils.ConvertNullableString(src.ProjectId)
	model.CreatedAt = utils.ConvertNullableTime(src.CreatedAt)
	model.UpdatedAt = utils.ConvertNullableTime(src.UpdatedAt)
	model.LoadBalancerId = utils.ConvertNullableString(src.LoadBalancerId)
	model.LoadBalancerName = utils.ConvertNullableString(src.LoadBalancerName)
	model.LoadBalancerProvisioningStatus = utils.ConvertNullableString(src.LoadBalancerProvisioningStatus)
	model.LoadBalancerType = utils.ConvertNullableString(src.LoadBalancerType)
	model.SubnetName = utils.ConvertNullableString(src.SubnetName)
	model.VpcName = utils.ConvertNullableString(src.VpcName)
	model.MemberCount = utils.ConvertNullableInt32ToInt64(src.MemberCount)
	if len(src.AlpnProtocols) > 0 {
		alpnProtocols := make([]string, 0, len(src.AlpnProtocols))
		for _, protocol := range src.AlpnProtocols {
			if protocol != nil {
				alpnProtocols = append(alpnProtocols, string(*protocol))
			}
		}
		model.LoadBalancerAlpnProtocols = utils.ConvertNullableStringList(alpnProtocols)
	} else {
		model.LoadBalancerAlpnProtocols = types.ListNull(types.StringType)
	}

	healthMonitor, healthMonitorDiags := utils.ConvertObjectFromModel(ctx, src.HealthMonitor, loadBalancerTargetGroupHealthMonitorAttrType, func(healthMonitor loadbalancer.TargetGroupHealthMonitor) any {
		return loadBalancerTargetGroupHealthMonitorModel{
			Id:            utils.ConvertNullableString(healthMonitor.Id),
			Type:          utils.ConvertNullableString(healthMonitor.Type),
			Delay:         utils.ConvertNullableInt32ToInt64(healthMonitor.Delay),
			Timeout:       utils.ConvertNullableInt32ToInt64(healthMonitor.Timeout),
			FallThreshold: utils.ConvertNullableInt32ToInt64(healthMonitor.FallThreshold),
			RiseThreshold: utils.ConvertNullableInt32ToInt64(healthMonitor.RiseThreshold),
			HttpMethod:    utils.ConvertNullableString(healthMonitor.HttpMethod),
			HttpVersion: func() types.String {
				if healthMonitor.HttpVersion.IsSet() && healthMonitor.HttpVersion.Get() != nil {
					return types.StringValue(fmt.Sprintf("%.1f", *healthMonitor.HttpVersion.Get()))
				}
				return types.StringNull()
			}(),
			ExpectedCodes:      utils.ConvertNullableString(healthMonitor.ExpectedCodes),
			UrlPath:            utils.ConvertNullableString(healthMonitor.UrlPath),
			OperatingStatus:    utils.ConvertNullableString(healthMonitor.OperatingStatus),
			ProvisioningStatus: utils.ConvertNullableString(healthMonitor.ProvisioningStatus),
			ProjectId:          utils.ConvertNullableString(healthMonitor.ProjectId),
		}
	})
	diags.Append(healthMonitorDiags...)
	model.HealthMonitor = healthMonitor

	if src.SessionPersistence.IsSet() {
		sessionPersistence, sessionPersistenceDiags := mapLoadBalancerTargetGroupSessionPersistenceFromResponse(ctx, src.SessionPersistence.Get())
		diags.Append(sessionPersistenceDiags...)
		model.SessionPersistence = sessionPersistence
	} else {
		model.SessionPersistence = types.ObjectNull(loadBalancerTargetGroupSessionPersistenceAttrType)
	}

	if len(src.Listeners) > 0 {
		listeners := make([]loadBalancerTargetGroupListenerModel, 0, len(src.Listeners))
		for _, listener := range src.Listeners {
			listeners = append(listeners, loadBalancerTargetGroupListenerModel{
				Id:           utils.ConvertNullableString(listener.Id),
				Protocol:     utils.ConvertNullableString(listener.Protocol),
				ProtocolPort: utils.ConvertNullableInt32ToInt64(listener.ProtocolPort),
			})
		}

		list, listDiags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: loadBalancerTargetGroupListenerAttrType}, listeners)
		diags.Append(listDiags...)
		model.Listeners = list
	} else {
		model.Listeners = types.ListNull(types.ObjectType{AttrTypes: loadBalancerTargetGroupListenerAttrType})
	}

	return !diags.HasError()
}

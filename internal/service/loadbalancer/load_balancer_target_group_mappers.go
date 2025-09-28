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
	"github.com/kakaoenterprise/kc-sdk-go/services/loadbalancer"
)

func mapLoadBalancerTargetGroupFromCreateResponse(
	ctx context.Context,
	model *loadBalancerTargetGroupResourceModel,
	src *loadbalancer.BnsLoadBalancerV1ApiCreateTargetGroupModelTargetGroupModel,
	diags *diag.Diagnostics,
) bool {
	originalSessionPersistence := model.SessionPersistence

	model.Id = types.StringValue(src.Id)
	model.Name = utils.ConvertNullableString(src.Name)

	model.Description = utils.ConvertNullableStringWithEmptyToNull(src.Description)
	model.Protocol = utils.ConvertNullableString(src.Protocol)
	model.LoadBalancerAlgorithm = utils.ConvertNullableString(src.LoadBalancerAlgorithm)
	model.ProvisioningStatus = utils.ConvertNullableString(src.ProvisioningStatus)
	model.OperatingStatus = utils.ConvertNullableString(src.OperatingStatus)
	model.ProjectId = utils.ConvertNullableString(src.ProjectId)
	model.CreatedAt = utils.ConvertNullableTime(src.CreatedAt)
	model.UpdatedAt = utils.ConvertNullableTime(src.UpdatedAt)

	if len(src.LoadBalancers) > 0 {
		lb := src.LoadBalancers[0]
		model.LoadBalancerId = types.StringValue(lb.Id)
	}

	loadBalancers := make([]loadBalancerTargetGroupLoadBalancerModel, 0, len(src.LoadBalancers))
	for _, lb := range src.LoadBalancers {
		loadBalancers = append(loadBalancers, loadBalancerTargetGroupLoadBalancerModel{
			Id: types.StringValue(lb.Id),
		})
	}

	list, listDiags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: loadBalancerTargetGroupLoadBalancerAttrType}, loadBalancers)
	diags.Append(listDiags...)
	model.LoadBalancers = list

	if len(src.Listeners) > 0 {
		listener := src.Listeners[0]
		model.ListenerId = types.StringValue(listener.Id)

		listeners := make([]loadBalancerTargetGroupListenerModel, 0, len(src.Listeners))
		for _, listener := range src.Listeners {
			listeners = append(listeners, loadBalancerTargetGroupListenerModel{
				Id: types.StringValue(listener.Id),
			})
		}

		list, listDiags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: loadBalancerTargetGroupListenerAttrType}, listeners)
		diags.Append(listDiags...)
		model.Listeners = list
	} else {
		model.ListenerId = types.StringNull()
		model.Listeners = types.ListNull(types.ObjectType{AttrTypes: loadBalancerTargetGroupListenerAttrType})
	}

	if src.SessionPersistence.IsSet() && src.SessionPersistence.Get() != nil {
		sessionPersistence := src.SessionPersistence.Get()
		sessionPersistenceModel := loadBalancerTargetGroupSessionPersistenceModel{
			Type:                   types.StringValue(sessionPersistence.Type),
			CookieName:             utils.ConvertNullableString(sessionPersistence.CookieName),
			PersistenceTimeout:     types.Int64Value(int64(sessionPersistence.PersistenceTimeout)),
			PersistenceGranularity: utils.ConvertNullableString(sessionPersistence.PersistenceGranularity),
		}

		obj, objDiags := types.ObjectValueFrom(ctx, loadBalancerTargetGroupSessionPersistenceAttrType, sessionPersistenceModel)
		diags.Append(objDiags...)
		model.SessionPersistence = obj
	} else {

		if !originalSessionPersistence.IsNull() && !originalSessionPersistence.IsUnknown() {
			model.SessionPersistence = originalSessionPersistence
		} else {
			model.SessionPersistence = types.ObjectNull(loadBalancerTargetGroupSessionPersistenceAttrType)
		}
	}

	model.HealthMonitor = types.ObjectNull(loadBalancerTargetGroupHealthMonitorAttrType)

	model.MemberCount = types.Int64Value(0)

	return !diags.HasError()
}

func mapLoadBalancerTargetGroupFromUpdateResponse(
	ctx context.Context,
	model *loadBalancerTargetGroupResourceModel,
	src *loadbalancer.BnsLoadBalancerV1ApiUpdateTargetGroupModelTargetGroupModel,
	previousState *loadBalancerTargetGroupResourceModel,
	diags *diag.Diagnostics,
) bool {

	model.Id = types.StringValue(src.Id)
	model.Name = utils.ConvertNullableString(src.Name)

	model.Description = utils.ConvertNullableStringWithEmptyToNull(src.Description)
	model.Protocol = utils.ConvertNullableString(src.Protocol)
	model.LoadBalancerAlgorithm = utils.ConvertNullableString(src.LoadBalancerAlgorithm)
	model.ProvisioningStatus = utils.ConvertNullableString(src.ProvisioningStatus)
	model.OperatingStatus = utils.ConvertNullableString(src.OperatingStatus)
	model.ProjectId = utils.ConvertNullableString(src.ProjectId)
	model.CreatedAt = utils.ConvertNullableTime(src.CreatedAt)
	model.UpdatedAt = utils.ConvertNullableTime(src.UpdatedAt)

	if len(src.LoadBalancers) > 0 {
		lb := src.LoadBalancers[0]
		model.LoadBalancerId = types.StringValue(lb.Id)

		loadBalancers := make([]loadBalancerTargetGroupLoadBalancerModel, 0, len(src.LoadBalancers))
		for _, lb := range src.LoadBalancers {
			loadBalancers = append(loadBalancers, loadBalancerTargetGroupLoadBalancerModel{
				Id: types.StringValue(lb.Id),
			})
		}

		list, listDiags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: loadBalancerTargetGroupLoadBalancerAttrType}, loadBalancers)
		diags.Append(listDiags...)
		model.LoadBalancers = list
	}

	if len(src.Listeners) > 0 {
		listener := src.Listeners[0]
		model.ListenerId = types.StringValue(listener.Id)

		listeners := make([]loadBalancerTargetGroupListenerModel, 0, len(src.Listeners))
		for _, listener := range src.Listeners {
			listeners = append(listeners, loadBalancerTargetGroupListenerModel{
				Id: types.StringValue(listener.Id),
			})
		}

		list, listDiags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: loadBalancerTargetGroupListenerAttrType}, listeners)
		diags.Append(listDiags...)
		model.Listeners = list
	} else {
		model.ListenerId = types.StringNull()
		model.Listeners = types.ListNull(types.ObjectType{AttrTypes: loadBalancerTargetGroupListenerAttrType})
	}

	if src.SessionPersistence.IsSet() && src.SessionPersistence.Get() != nil {
		sessionPersistence := src.SessionPersistence.Get()
		sessionPersistenceModel := loadBalancerTargetGroupSessionPersistenceModel{
			Type:                   types.StringValue(sessionPersistence.Type),
			CookieName:             utils.ConvertNullableString(sessionPersistence.CookieName),
			PersistenceTimeout:     types.Int64Value(int64(sessionPersistence.PersistenceTimeout)),
			PersistenceGranularity: utils.ConvertNullableString(sessionPersistence.PersistenceGranularity),
		}

		obj, objDiags := types.ObjectValueFrom(ctx, loadBalancerTargetGroupSessionPersistenceAttrType, sessionPersistenceModel)
		diags.Append(objDiags...)
		model.SessionPersistence = obj
	} else {

		model.SessionPersistence = types.ObjectNull(loadBalancerTargetGroupSessionPersistenceAttrType)
	}

	if previousState != nil && !previousState.HealthMonitor.IsNull() && !previousState.HealthMonitor.IsUnknown() {
		model.HealthMonitor = previousState.HealthMonitor
	} else {
		model.HealthMonitor = types.ObjectNull(loadBalancerTargetGroupHealthMonitorAttrType)
	}

	if previousState != nil && !previousState.MemberCount.IsNull() && !previousState.MemberCount.IsUnknown() {
		model.MemberCount = previousState.MemberCount
	} else {
		model.MemberCount = types.Int64Value(0)
	}

	return true
}

func mapLoadBalancerTargetGroupSingleFromListResponse(
	ctx context.Context,
	model *loadBalancerTargetGroupBaseModel,
	src *loadbalancer.BnsLoadBalancerV1ApiListTargetGroupsModelTargetGroupModel,
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

	if src.HealthMonitor.IsSet() && src.HealthMonitor.Get() != nil {
		healthMonitor, healthMonitorDiags := mapLoadBalancerTargetGroupHealthMonitorFromResponse(ctx, src.HealthMonitor.Get())
		diags.Append(healthMonitorDiags...)
		model.HealthMonitor = healthMonitor
	} else {
		model.HealthMonitor = types.ObjectNull(loadBalancerTargetGroupHealthMonitorAttrType)
	}

	if src.SessionPersistence.IsSet() && src.SessionPersistence.Get() != nil {
		sessionPersistence, sessionPersistenceDiags := mapLoadBalancerTargetGroupSessionPersistenceFromResponse(ctx, src.SessionPersistence.Get())
		diags.Append(sessionPersistenceDiags...)
		model.SessionPersistence = sessionPersistence
	} else {
		model.SessionPersistence = types.ObjectNull(loadBalancerTargetGroupSessionPersistenceAttrType)
	}

	if len(src.Listeners) > 0 {
		listeners := make([]loadBalancerTargetGroupListListenerModel, 0, len(src.Listeners))
		for _, listener := range src.Listeners {
			listeners = append(listeners, loadBalancerTargetGroupListListenerModel{
				Id:           types.StringValue(listener.Id),
				Protocol:     utils.ConvertNullableString(listener.Protocol),
				ProtocolPort: utils.ConvertNullableInt32ToInt64(listener.ProtocolPort),
			})
		}

		list, listDiags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: loadBalancerTargetGroupListListenerAttrType}, listeners)
		diags.Append(listDiags...)
		model.Listeners = list
	} else {
		model.Listeners = types.ListNull(types.ObjectType{AttrTypes: loadBalancerTargetGroupListListenerAttrType})
	}

	return !diags.HasError()
}

func mapLoadBalancerTargetGroupHealthMonitorFromResponse(
	ctx context.Context,
	src interface{},
) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	if src == nil {
		return types.ObjectNull(loadBalancerTargetGroupHealthMonitorAttrType), diags
	}

	var healthMonitor loadBalancerTargetGroupHealthMonitorModel

	switch s := src.(type) {
	case *loadbalancer.BnsLoadBalancerV1ApiListTargetGroupsModelHealthMonitorModel:
		healthMonitor = loadBalancerTargetGroupHealthMonitorModel{
			Id:            types.StringValue(s.Id),
			Type:          utils.ConvertNullableString(s.Type),
			Delay:         utils.ConvertNullableInt32ToInt64(s.Delay),
			Timeout:       utils.ConvertNullableInt32ToInt64(s.Timeout),
			FallThreshold: utils.ConvertNullableInt32ToInt64(s.FallThreshold),
			RiseThreshold: utils.ConvertNullableInt32ToInt64(s.RiseThreshold),
			HttpMethod:    utils.ConvertNullableString(s.HttpMethod),
			HttpVersion: func() types.String {
				if s.HttpVersion.IsSet() && s.HttpVersion.Get() != nil {
					return types.StringValue(fmt.Sprintf("%.1f", *s.HttpVersion.Get()))
				}
				return types.StringNull()
			}(),
			ExpectedCodes:      utils.ConvertNullableString(s.ExpectedCodes),
			UrlPath:            utils.ConvertNullableString(s.UrlPath),
			OperatingStatus:    utils.ConvertNullableString(s.OperatingStatus),
			ProvisioningStatus: utils.ConvertNullableString(s.ProvisioningStatus),
			ProjectId:          utils.ConvertNullableString(s.ProjectId),
		}
	case *loadbalancer.BnsLoadBalancerV1ApiGetTargetGroupModelHealthMonitorModel:
		healthMonitor = loadBalancerTargetGroupHealthMonitorModel{
			Id:            types.StringValue(s.Id),
			Type:          utils.ConvertNullableString(s.Type),
			Delay:         utils.ConvertNullableInt32ToInt64(s.Delay),
			Timeout:       utils.ConvertNullableInt32ToInt64(s.Timeout),
			FallThreshold: utils.ConvertNullableInt32ToInt64(s.FallThreshold),
			RiseThreshold: utils.ConvertNullableInt32ToInt64(s.RiseThreshold),
			HttpMethod:    utils.ConvertNullableString(s.HttpMethod),
			HttpVersion: func() types.String {
				if s.HttpVersion.IsSet() && s.HttpVersion.Get() != nil {
					return types.StringValue(fmt.Sprintf("%.1f", *s.HttpVersion.Get()))
				}
				return types.StringNull()
			}(),
			ExpectedCodes:      utils.ConvertNullableString(s.ExpectedCodes),
			UrlPath:            utils.ConvertNullableString(s.UrlPath),
			OperatingStatus:    utils.ConvertNullableString(s.OperatingStatus),
			ProvisioningStatus: utils.ConvertNullableString(s.ProvisioningStatus),
			ProjectId:          utils.ConvertNullableString(s.ProjectId),
		}
	default:
		diags.AddError("Invalid health monitor type", fmt.Sprintf("Unsupported health monitor type: %T", src))
		return types.ObjectNull(loadBalancerTargetGroupHealthMonitorAttrType), diags
	}

	obj, objDiags := types.ObjectValueFrom(ctx, loadBalancerTargetGroupHealthMonitorAttrType, healthMonitor)
	diags.Append(objDiags...)
	return obj, diags
}

func mapLoadBalancerTargetGroupSessionPersistenceFromResponse(
	ctx context.Context,
	src *loadbalancer.SessionPersistenceModel,
) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	if src == nil {
		return types.ObjectNull(loadBalancerTargetGroupSessionPersistenceAttrType), diags
	}

	sessionPersistence := loadBalancerTargetGroupSessionPersistenceModel{
		Type:                   types.StringValue(src.Type),
		CookieName:             utils.ConvertNullableString(src.CookieName),
		PersistenceTimeout:     types.Int64Value(int64(src.PersistenceTimeout)),
		PersistenceGranularity: utils.ConvertNullableString(src.PersistenceGranularity),
	}

	obj, objDiags := types.ObjectValueFrom(ctx, loadBalancerTargetGroupSessionPersistenceAttrType, sessionPersistence)
	diags.Append(objDiags...)
	return obj, diags
}

func mapLoadBalancerTargetGroupFromGetResponse(
	ctx context.Context,
	model *loadBalancerTargetGroupResourceModel,
	src *loadbalancer.BnsLoadBalancerV1ApiGetTargetGroupModelTargetGroupModel,
	diags *diag.Diagnostics,
) bool {

	originalHealthMonitor := model.HealthMonitor
	originalSessionPersistence := model.SessionPersistence
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

	if len(src.Listeners) > 0 {
		listener := src.Listeners[0]
		model.ListenerId = types.StringValue(listener.Id)
	} else {
		model.ListenerId = types.StringNull()
	}

	if src.HealthMonitor.IsSet() && src.HealthMonitor.Get() != nil {
		healthMonitor, healthMonitorDiags := mapLoadBalancerTargetGroupHealthMonitorFromResponse(ctx, src.HealthMonitor.Get())
		diags.Append(healthMonitorDiags...)
		model.HealthMonitor = healthMonitor
	} else {

		if !originalHealthMonitor.IsNull() && !originalHealthMonitor.IsUnknown() {

			model.HealthMonitor = preserveHealthMonitorUserConfig(ctx, originalHealthMonitor)
		} else {
			model.HealthMonitor = types.ObjectNull(loadBalancerTargetGroupHealthMonitorAttrType)
		}
	}

	if src.SessionPersistence.IsSet() && src.SessionPersistence.Get() != nil {
		sessionPersistence, sessionPersistenceDiags := mapLoadBalancerTargetGroupSessionPersistenceFromResponse(ctx, src.SessionPersistence.Get())
		diags.Append(sessionPersistenceDiags...)
		model.SessionPersistence = sessionPersistence
	} else {

		if !originalSessionPersistence.IsNull() && !originalSessionPersistence.IsUnknown() {

			model.SessionPersistence = preserveSessionPersistenceUserConfig(ctx, originalSessionPersistence)
		} else {
			model.SessionPersistence = types.ObjectNull(loadBalancerTargetGroupSessionPersistenceAttrType)
		}
	}

	if len(src.Listeners) > 0 {
		listeners := make([]loadBalancerTargetGroupListenerModel, 0, len(src.Listeners))
		for _, listener := range src.Listeners {
			listeners = append(listeners, loadBalancerTargetGroupListenerModel{
				Id: types.StringValue(listener.Id),
			})
		}

		list, listDiags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: loadBalancerTargetGroupListenerAttrType}, listeners)
		diags.Append(listDiags...)
		model.Listeners = list
	} else {
		model.Listeners = types.ListNull(types.ObjectType{AttrTypes: loadBalancerTargetGroupListenerAttrType})
	}

	model.LoadBalancers = types.ListNull(types.ObjectType{AttrTypes: loadBalancerTargetGroupLoadBalancerAttrType})

	return !diags.HasError()
}

func preserveHealthMonitorUserConfig(ctx context.Context, original types.Object) types.Object {
	if original.IsNull() || original.IsUnknown() {
		return types.ObjectNull(loadBalancerTargetGroupHealthMonitorAttrType)
	}

	var originalModel loadBalancerTargetGroupHealthMonitorModel
	diags := original.As(ctx, &originalModel, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return types.ObjectNull(loadBalancerTargetGroupHealthMonitorAttrType)
	}

	preservedModel := loadBalancerTargetGroupHealthMonitorModel{
		Id:            types.StringNull(),
		Type:          originalModel.Type,
		Delay:         originalModel.Delay,
		Timeout:       originalModel.Timeout,
		FallThreshold: originalModel.FallThreshold,
		RiseThreshold: originalModel.RiseThreshold,
		HttpMethod:    originalModel.HttpMethod,
		HttpVersion:   originalModel.HttpVersion,
		UrlPath:       originalModel.UrlPath,
		ExpectedCodes: originalModel.ExpectedCodes,
	}

	result, diags := types.ObjectValueFrom(ctx, loadBalancerTargetGroupHealthMonitorAttrType, preservedModel)
	if diags.HasError() {
		return types.ObjectNull(loadBalancerTargetGroupHealthMonitorAttrType)
	}

	return result
}

func preserveSessionPersistenceUserConfig(ctx context.Context, original types.Object) types.Object {
	if original.IsNull() || original.IsUnknown() {
		return types.ObjectNull(loadBalancerTargetGroupSessionPersistenceAttrType)
	}

	var originalModel loadBalancerTargetGroupSessionPersistenceModel
	diags := original.As(ctx, &originalModel, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return types.ObjectNull(loadBalancerTargetGroupSessionPersistenceAttrType)
	}

	preservedModel := loadBalancerTargetGroupSessionPersistenceModel{
		Type:                   originalModel.Type,
		CookieName:             originalModel.CookieName,
		PersistenceTimeout:     originalModel.PersistenceTimeout,
		PersistenceGranularity: originalModel.PersistenceGranularity,
	}

	result, diags := types.ObjectValueFrom(ctx, loadBalancerTargetGroupSessionPersistenceAttrType, preservedModel)
	if diags.HasError() {
		return types.ObjectNull(loadBalancerTargetGroupSessionPersistenceAttrType)
	}

	return result
}

func mapLoadBalancerTargetGroupToCreateRequest(
	ctx context.Context,
	model *loadBalancerTargetGroupResourceModel,
	diags *diag.Diagnostics,
) *loadbalancer.CreateTargetGroup {

	protocol := loadbalancer.TargetGroupProtocol(model.Protocol.ValueString())
	algorithm := loadbalancer.TargetGroupAlgorithm(model.LoadBalancerAlgorithm.ValueString())

	createReq := loadbalancer.NewCreateTargetGroup(
		algorithm,
		model.LoadBalancerId.ValueString(),
		model.Name.ValueString(),
		protocol,
	)

	if !model.ListenerId.IsNull() && !model.ListenerId.IsUnknown() {
		createReq.SetListenerId(model.ListenerId.ValueString())
	}

	if !model.Description.IsNull() && !model.Description.IsUnknown() {
		createReq.SetDescription(model.Description.ValueString())
	}

	if !model.SessionPersistence.IsNull() && !model.SessionPersistence.IsUnknown() {
		var sessionPersistence loadBalancerTargetGroupSessionPersistenceModel
		diags.Append(model.SessionPersistence.As(ctx, &sessionPersistence, basetypes.ObjectAsOptions{})...)
		if !diags.HasError() {
			sessionPersistenceReq := loadbalancer.SessionPersistenceModel{
				Type: sessionPersistence.Type.ValueString(),
			}

			if !sessionPersistence.CookieName.IsNull() && !sessionPersistence.CookieName.IsUnknown() {
				sessionPersistenceReq.CookieName = *loadbalancer.NewNullableString(loadbalancer.PtrString(sessionPersistence.CookieName.ValueString()))
			}
			if !sessionPersistence.PersistenceTimeout.IsNull() && !sessionPersistence.PersistenceTimeout.IsUnknown() {
				sessionPersistenceReq.PersistenceTimeout = int32(sessionPersistence.PersistenceTimeout.ValueInt64())
			}
			if !sessionPersistence.PersistenceGranularity.IsNull() && !sessionPersistence.PersistenceGranularity.IsUnknown() {
				sessionPersistenceReq.PersistenceGranularity = *loadbalancer.NewNullableString(loadbalancer.PtrString(sessionPersistence.PersistenceGranularity.ValueString()))
			}

			createReq.SetSessionPersistence(sessionPersistenceReq)
		}
	}

	return createReq
}

func mapLoadBalancerTargetGroupToUpdateRequest(
	ctx context.Context,
	model *loadBalancerTargetGroupResourceModel,
	diags *diag.Diagnostics,
) *loadbalancer.EditTargetGroup {
	updateReq := loadbalancer.NewEditTargetGroup()

	if !model.Name.IsNull() && !model.Name.IsUnknown() {
		updateReq.SetName(model.Name.ValueString())
	}

	if !model.Description.IsNull() && !model.Description.IsUnknown() {
		updateReq.SetDescription(model.Description.ValueString())
	}

	algorithm := loadbalancer.TargetGroupAlgorithm(model.LoadBalancerAlgorithm.ValueString())
	updateReq.SetLoadBalancerAlgorithm(algorithm)

	if model.SessionPersistence.IsNull() || model.SessionPersistence.IsUnknown() {

		updateReq.SetSessionPersistenceNil()
	} else {

		var sessionPersistence loadBalancerTargetGroupSessionPersistenceModel
		diags.Append(model.SessionPersistence.As(ctx, &sessionPersistence, basetypes.ObjectAsOptions{})...)
		if !diags.HasError() {
			sessionPersistenceReq := loadbalancer.SessionPersistenceModel{
				Type: sessionPersistence.Type.ValueString(),
			}

			if !sessionPersistence.CookieName.IsNull() && !sessionPersistence.CookieName.IsUnknown() {
				sessionPersistenceReq.CookieName = *loadbalancer.NewNullableString(loadbalancer.PtrString(sessionPersistence.CookieName.ValueString()))
			}
			if !sessionPersistence.PersistenceTimeout.IsNull() && !sessionPersistence.PersistenceTimeout.IsUnknown() {
				sessionPersistenceReq.PersistenceTimeout = int32(sessionPersistence.PersistenceTimeout.ValueInt64())
			}
			if !sessionPersistence.PersistenceGranularity.IsNull() && !sessionPersistence.PersistenceGranularity.IsUnknown() {
				sessionPersistenceReq.PersistenceGranularity = *loadbalancer.NewNullableString(loadbalancer.PtrString(sessionPersistence.PersistenceGranularity.ValueString()))
			}

			updateReq.SetSessionPersistence(sessionPersistenceReq)
		}
	}

	return updateReq
}

func mapLoadBalancerTargetGroupSingleFromGetResponse(
	ctx context.Context,
	model *loadBalancerTargetGroupBaseModel,
	src *loadbalancer.BnsLoadBalancerV1ApiGetTargetGroupModelTargetGroupModel,
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

	healthMonitor, healthMonitorDiags := utils.ConvertObjectFromModel(ctx, src.HealthMonitor, loadBalancerTargetGroupHealthMonitorAttrType, func(healthMonitor loadbalancer.BnsLoadBalancerV1ApiGetTargetGroupModelHealthMonitorModel) any {
		return loadBalancerTargetGroupHealthMonitorModel{
			Id:            types.StringValue(healthMonitor.Id),
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
			}(), ExpectedCodes: utils.ConvertNullableString(healthMonitor.ExpectedCodes),
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
		listeners := make([]loadBalancerTargetGroupListListenerModel, 0, len(src.Listeners))
		for _, listener := range src.Listeners {
			listeners = append(listeners, loadBalancerTargetGroupListListenerModel{
				Id:           types.StringValue(listener.Id),
				Protocol:     utils.ConvertNullableString(listener.Protocol),
				ProtocolPort: utils.ConvertNullableInt32ToInt64(listener.ProtocolPort),
			})
		}

		list, listDiags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: loadBalancerTargetGroupListListenerAttrType}, listeners)
		diags.Append(listDiags...)
		model.Listeners = list
	} else {
		model.Listeners = types.ListNull(types.ObjectType{AttrTypes: loadBalancerTargetGroupListListenerAttrType})
	}

	return !diags.HasError()
}

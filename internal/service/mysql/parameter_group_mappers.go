// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package mysql

import (
	"context"

	"terraform-provider-kakaocloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	mysqlsdk "github.com/kakaoenterprise/kc-sdk-go/services/mysql"
)

func toMysqlParameterModelFromParameter(src mysqlsdk.ParameterResponseModel) (mysqlParameterModel, bool) {
	return mysqlParameterModel{
		Key:                   types.StringValue(src.Key),
		Value:                 utils.ConvertNullableString(src.Value),
		DefaultParameterValue: types.StringNull(),
		ParameterType:         types.StringValue(string(src.ParameterType)),
		DataType:              types.StringValue(string(src.DataType)),
		ValidationValueFormat: utils.ConvertNullableString(src.ValidationValueFormat),
		IsEditable:            types.BoolValue(src.IsEditable),
		IsRequired:            types.BoolValue(src.IsRequired),
	}, true
}

func toMysqlParameterModelFromDetail(src mysqlsdk.DataDetailParametersResponseModel) (mysqlParameterModel, bool) {
	return mysqlParameterModel{
		Key:                   types.StringValue(src.Key),
		Value:                 utils.ConvertNullableString(src.Value),
		DefaultParameterValue: utils.ConvertNullableString(src.DefaultParameterValue),
		ParameterType:         utils.ConvertNullableString(src.ParameterType),
		DataType:              types.StringValue(string(src.DataType)),
		ValidationValueFormat: utils.ConvertNullableString(src.ValidationValueFormat),
		IsEditable:            types.BoolValue(src.IsEditable),
		IsRequired:            types.BoolValue(src.IsRequired),
	}, true
}

func customParameterGroupParametersValue(ctx context.Context, src []mysqlsdk.DataDetailParametersResponseModel, respDiags *diag.Diagnostics) (types.List, bool) {
	value, diags := utils.ConvertListFromModel(ctx, src, mysqlParameterAttrTypes, func(item mysqlsdk.DataDetailParametersResponseModel) any {
		mapped, _ := toMysqlParameterModelFromDetail(item)
		return mapped
	})
	respDiags.Append(diags...)
	if respDiags.HasError() {
		return types.ListNull(types.ObjectType{AttrTypes: mysqlParameterAttrTypes}), false
	}
	return value, true
}

func defaultParameterGroupParametersValue(ctx context.Context, src []mysqlsdk.ParameterResponseModel, respDiags *diag.Diagnostics) (types.List, bool) {
	value, diags := utils.ConvertListFromModel(ctx, src, mysqlParameterAttrTypes, func(item mysqlsdk.ParameterResponseModel) any {
		mapped, _ := toMysqlParameterModelFromParameter(item)
		return mapped
	})
	respDiags.Append(diags...)
	if respDiags.HasError() {
		return types.ListNull(types.ObjectType{AttrTypes: mysqlParameterAttrTypes}), false
	}
	return value, true
}

func toCustomParameterGroupListModel(src mysqlsdk.MysqlV1ApiListMysqlCustomParameterGroupsModelCustomParameterGroupResponseModel) (customParameterGroupListModel, bool) {
	return customParameterGroupListModel{
		Id:                      utils.ConvertNullableString(src.Id),
		Name:                    utils.ConvertNullableString(src.Name),
		Description:             utils.ConvertNullableString(src.Description),
		EngineVersion:           utils.ConvertNullableString(src.EngineVersion),
		DefaultParameterGroupId: utils.ConvertNullableString(src.DefaultParameterGroupId),
		ExistErrorSync:          utils.ConvertNullableBool(src.ExistErrorSync),
		InstanceGroupCount:      utils.ConvertNullableInt32(src.InstanceGroupCount),
		IsRollbackPossible:      utils.ConvertNullableBool(src.IsRollbackPossible),
	}, true
}

func toDefaultParameterGroupListModel(src mysqlsdk.MysqlV1ApiListMysqlDefaultParameterGroupsModelDefaultParameterGroupResponseModel) (defaultParameterGroupListModel, bool) {
	return defaultParameterGroupListModel{
		Id:                         types.StringValue(src.Id),
		Name:                       types.StringValue(src.Name),
		Description:                types.StringValue(src.Description),
		EngineVersion:              types.StringValue(src.EngineVersion),
		ExistErrorSync:             utils.ConvertNullableBool(src.ExistErrorSync),
		ExistEngineVersionMismatch: utils.ConvertNullableBool(src.ExistEngineVersionMismatch),
		InstanceGroupCount:         utils.ConvertNullableInt32(src.InstanceGroupCount),
	}, true
}

func toCustomParameterGroupSingleModel(ctx context.Context, src mysqlsdk.MysqlV1ApiGetMysqlCustomParameterGroupModelCustomParameterGroupResponseModel, respDiags *diag.Diagnostics) (customParameterGroupSingleModel, bool) {
	parameters, ok := customParameterGroupParametersValue(ctx, src.Parameters, respDiags)
	if !ok {
		return customParameterGroupSingleModel{}, false
	}

	return customParameterGroupSingleModel{
		Id:                      utils.ConvertNullableString(src.Id),
		Name:                    utils.ConvertNullableString(src.Name),
		Description:             utils.ConvertNullableString(src.Description),
		EngineVersion:           utils.ConvertNullableString(src.EngineVersion),
		DefaultParameterGroupId: utils.ConvertNullableString(src.DefaultParameterGroupId),
		ExistErrorSync:          utils.ConvertNullableBool(src.ExistErrorSync),
		InstanceGroupCount:      types.Int32Value(src.InstanceGroupCount),
		IsRollbackPossible:      utils.ConvertNullableBool(src.IsRollbackPossible),
		Parameters:              parameters,
	}, true
}

func toDefaultParameterGroupSingleModel(ctx context.Context, src mysqlsdk.MysqlV1ApiGetMysqlDefaultParameterGroupModelDefaultParameterGroupResponseModel, respDiags *diag.Diagnostics) (defaultParameterGroupSingleModel, bool) {
	parameters, ok := defaultParameterGroupParametersValue(ctx, src.Parameters, respDiags)
	if !ok {
		return defaultParameterGroupSingleModel{}, false
	}

	return defaultParameterGroupSingleModel{
		Id:                         types.StringValue(src.Id),
		Name:                       types.StringValue(src.Name),
		Description:                types.StringValue(src.Description),
		EngineVersion:              types.StringValue(src.EngineVersion),
		ExistErrorSync:             utils.ConvertNullableBool(src.ExistErrorSync),
		ExistEngineVersionMismatch: utils.ConvertNullableBool(src.ExistEngineVersionMismatch),
		InstanceGroupCount:         utils.ConvertNullableInt32(src.InstanceGroupCount),
		Parameters:                 parameters,
	}, true
}

func toCustomParameterGroupEventModel(src mysqlsdk.CustomParameterGroupEventResponseModel) mysqlParameterGroupEventModel {
	return mysqlParameterGroupEventModel{
		CreatedAt:   types.StringValue(src.CreatedAt),
		Description: types.StringValue(src.Description),
		Name:        types.StringValue(src.Name),
	}
}

func toDefaultParameterGroupEventModel(src mysqlsdk.DefaultParameterGroupEventResponseModel) mysqlParameterGroupEventModel {
	return mysqlParameterGroupEventModel{
		CreatedAt:   utils.ConvertNullableString(src.CreatedAt),
		Description: utils.ConvertNullableString(src.Description),
		Name:        utils.ConvertNullableString(src.Name),
	}
}

func toParameterGroupInstanceGroupModel(src mysqlsdk.DefaultParameterGroupInstanceGroupResponseModel) mysqlParameterGroupInstanceGroupModel {
	return mysqlParameterGroupInstanceGroupModel{
		Id:                   utils.ConvertNullableString(src.Id),
		Name:                 utils.ConvertNullableString(src.Name),
		Status:               utils.ConvertNullableString(src.Status),
		EngineVersion:        utils.ConvertNullableString(src.EngineVersion),
		FlavorId:             utils.ConvertNullableString(src.FlavorId),
		ParameterGroupStatus: utils.ConvertNullableString(src.ParameterGroupStatus),
		InstanceGroupType:    utils.ConvertNullableString(src.InstanceGroupType),
		IsMultiAz:            utils.ConvertNullableBool(src.IsMultiAz),
	}
}

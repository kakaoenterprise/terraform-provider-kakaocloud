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

func toBackupModelFromList(ctx context.Context, src mysqlsdk.MysqlV1ApiListMysqlBackupsModelMySQLBackupResponseModel, respDiags *diag.Diagnostics) (backupModel, bool) {
	extraInfo, ok := backupExtraInfoFromListValue(ctx, src.ExtraInfo, respDiags)
	if !ok {
		return backupModel{}, false
	}

	return backupModel{
		Id:                utils.ConvertNullableString(src.Id),
		Name:              utils.ConvertNullableString(src.Name),
		CreatedAt:         utils.ConvertNullableString(src.CreatedAt),
		CreatorName:       utils.ConvertNullableString(src.CreatorName),
		Description:       utils.ConvertNullableString(src.Description),
		DiskSize:          utils.ConvertNullableInt32(src.DiskSize),
		ExpireAt:          utils.ConvertNullableString(src.ExpireAt),
		ExpiryDuration:    utils.ConvertNullableInt32(src.ExpiryDuration),
		ExtraInfo:         extraInfo,
		InstanceGroupId:   utils.ConvertNullableString(src.InstanceGroupId),
		InstanceGroupName: utils.ConvertNullableString(src.InstanceGroupName),
		ProjectId:         utils.ConvertNullableString(src.ProjectId),
		Size:              utils.ConvertNullableInt64(src.Size),
		Status:            utils.ConvertNullableString(src.Status),
		Type:              utils.ConvertNullableString(src.Type),
		StartedAt:         utils.ConvertNullableString(src.StartedAt),
		UpdatedAt:         utils.ConvertNullableString(src.UpdatedAt),
		EngineVersion:     utils.ConvertNullableString(src.EngineVersion),
	}, true
}

func toBackupModelFromGet(ctx context.Context, src mysqlsdk.MysqlV1ApiGetMysqlBackupModelMySQLBackupResponseModel, respDiags *diag.Diagnostics) (backupModel, bool) {
	extraInfo, ok := backupExtraInfoFromGetValue(ctx, src.ExtraInfo, respDiags)
	if !ok {
		return backupModel{}, false
	}

	return backupModel{
		Id:                utils.ConvertNullableString(src.Id),
		Name:              utils.ConvertNullableString(src.Name),
		CreatedAt:         utils.ConvertNullableString(src.CreatedAt),
		CreatorName:       utils.ConvertNullableString(src.CreatorName),
		Description:       utils.ConvertNullableString(src.Description),
		DiskSize:          utils.ConvertNullableInt32(src.DiskSize),
		ExpireAt:          utils.ConvertNullableString(src.ExpireAt),
		ExpiryDuration:    utils.ConvertNullableInt32(src.ExpiryDuration),
		ExtraInfo:         extraInfo,
		InstanceGroupId:   utils.ConvertNullableString(src.InstanceGroupId),
		InstanceGroupName: utils.ConvertNullableString(src.InstanceGroupName),
		ProjectId:         utils.ConvertNullableString(src.ProjectId),
		Size:              utils.ConvertNullableInt64(src.Size),
		Status:            utils.ConvertNullableString(src.Status),
		Type:              utils.ConvertNullableString(src.Type),
		StartedAt:         utils.ConvertNullableString(src.StartedAt),
		UpdatedAt:         utils.ConvertNullableString(src.UpdatedAt),
		EngineVersion:     utils.ConvertNullableString(src.EngineVersion),
	}, true
}

func backupExtraInfoFromListValue(ctx context.Context, src mysqlsdk.NullableMysqlV1ApiListMysqlBackupsModelExtraInfoResponseModel, respDiags *diag.Diagnostics) (types.Object, bool) {
	extraInfo, diags := utils.ConvertObjectFromModel(ctx, src, backupExtraInfoAttrTypes, func(info mysqlsdk.MysqlV1ApiListMysqlBackupsModelExtraInfoResponseModel) any {
		return backupExtraInfoModel{
			UseCaseSensitiveTableNames: utils.ConvertNullableBool(info.UseCaseSensitiveTableNames),
		}
	})
	respDiags.Append(diags...)
	if respDiags.HasError() {
		return types.ObjectNull(backupExtraInfoAttrTypes), false
	}
	return extraInfo, true
}

func backupExtraInfoFromGetValue(
	ctx context.Context,
	src mysqlsdk.NullableMysqlV1ApiGetMysqlBackupModelExtraInfoResponseModel,
	respDiags *diag.Diagnostics,
) (types.Object, bool) {
	extraInfo, diags := utils.ConvertObjectFromModel(ctx, src, backupExtraInfoAttrTypes, func(info mysqlsdk.MysqlV1ApiGetMysqlBackupModelExtraInfoResponseModel) any {
		return backupExtraInfoModel{
			UseCaseSensitiveTableNames: utils.ConvertNullableBool(info.UseCaseSensitiveTableNames),
		}
	})
	respDiags.Append(diags...)
	if respDiags.HasError() {
		return types.ObjectNull(backupExtraInfoAttrTypes), false
	}
	return extraInfo, true
}

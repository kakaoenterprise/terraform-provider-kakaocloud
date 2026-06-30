// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package mysql

import (
	"context"
	"net/http"
	"terraform-provider-kakaocloud/internal/common"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	mysqlsdk "github.com/kakaoenterprise/kc-sdk-go/services/mysql"
)

func (r *instanceGroupResource) applyBackupSchedule(
	ctx context.Context,
	plan instanceGroupBackupScheduleModel,
	respDiags *diag.Diagnostics,
) bool {
	backupSchedule := mysqlsdk.NewMysqlV1ApiUpdateMysqlInstanceGroupBackupScheduleModelBackupScheduleRequestModel(plan.Enabled.ValueBool())
	if !plan.Type.IsNull() && !plan.Type.IsUnknown() && plan.Type.ValueString() != "" {
		backupSchedule.SetType(mysqlsdk.BackupScheduleType(plan.Type.ValueString()))
	}
	if !plan.StartTime.IsNull() && !plan.StartTime.IsUnknown() && plan.StartTime.ValueString() != "" {
		backupSchedule.SetStartTime(plan.StartTime.ValueString())
	}
	if !plan.ExpiryDuration.IsNull() && !plan.ExpiryDuration.IsUnknown() && plan.ExpiryDuration.ValueInt32() > 0 {
		backupSchedule.SetExpiryDuration(plan.ExpiryDuration.ValueInt32())
	}

	request := mysqlsdk.NewBodyUpdateMysqlInstanceGroupBackupSchedule(*backupSchedule)

	_, httpResp, err := common.ExecuteWithRetryAndAuth[struct{}](ctx, r.kc, respDiags,
		func() (struct{}, *http.Response, error) {
			httpResp, err := r.kc.ApiClient.MySQLInstanceGroupsBackupSchedulesAPI.
				UpdateMysqlInstanceGroupBackupSchedule(ctx, plan.InstanceGroupId.ValueString(), plan.BackupScheduleId.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				BodyUpdateMysqlInstanceGroupBackupSchedule(*request).
				Execute()
			return struct{}{}, httpResp, err
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "UpdateMysqlInstanceGroupBackupSchedule", err, respDiags)
		return false
	}

	return true
}

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package mysql

import (
	"context"
	"net/http"
	"terraform-provider-kakaocloud/internal/common"
	"terraform-provider-kakaocloud/internal/utils"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	mysqlsdk "github.com/kakaoenterprise/kc-sdk-go/services/mysql"
)

func (r *instanceGroupResource) applyParameterGroup(
	ctx context.Context,
	plan instanceGroupParameterGroupModel,
	respDiags *diag.Diagnostics,
) bool {
	requestGroup := mysqlsdk.NewMysqlV1ApiApplyMysqlParameterGroupModelParameterGroupRequestModel(
		plan.ParameterGroupId.ValueString(),
		mysqlsdk.ParameterGroupType(plan.ParameterGroupType.ValueString()),
	)
	request := mysqlsdk.NewBodyApplyMysqlParameterGroup(*requestGroup)

	_, httpResp, err := common.ExecuteWithRetryAndAuth[struct{}](ctx, r.kc, respDiags,
		func() (struct{}, *http.Response, error) {
			httpResp, err := r.kc.ApiClient.MySQLInstanceGroupsAPI.
				ApplyMysqlParameterGroup(ctx, plan.InstanceGroupId.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				BodyApplyMysqlParameterGroup(*request).
				Execute()
			return struct{}{}, httpResp, err
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "ApplyMysqlParameterGroup", err, respDiags)
		return false
	}

	return true
}

func (r *instanceGroupResource) retryParameterGroupSync(
	ctx context.Context,
	instanceGroupID string,
	respDiags *diag.Diagnostics,
) bool {
	_, httpResp, err := common.ExecuteWithRetryAndAuth[struct{}](ctx, r.kc, respDiags,
		func() (struct{}, *http.Response, error) {
			httpResp, err := r.kc.ApiClient.MySQLInstanceGroupsParameterGroupsAPI.
				RetryMysqlParameterGroupSync(ctx, instanceGroupID).
				XAuthToken(r.kc.XAuthToken).
				Execute()
			return struct{}{}, httpResp, err
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "RetryMysqlParameterGroupSync", err, respDiags)
		return false
	}

	return true
}

func (r *instanceGroupResource) readParameterGroupState(
	ctx context.Context,
	instanceGroupId string,
	respDiags *diag.Diagnostics,
) (instanceGroupParameterGroupModel, bool, bool) {
	result, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, respDiags,
		func() (*mysqlsdk.GetMySQLInstanceGroupResponseModel, *http.Response, error) {
			return r.kc.ApiClient.MySQLInstanceGroupsAPI.
				GetMysqlInstanceGroup(ctx, instanceGroupId).
				XAuthToken(r.kc.XAuthToken).
				Execute()
		},
	)
	if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
		return instanceGroupParameterGroupModel{}, false, true
	}
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "GetMysqlInstanceGroup", err, respDiags)
		return instanceGroupParameterGroupModel{}, false, false
	}

	return instanceGroupParameterGroupModel{
		Id:                      types.StringValue(instanceGroupId),
		InstanceGroupId:         types.StringValue(instanceGroupId),
		ParameterGroupId:        types.StringValue(result.InstanceGroup.ParameterGroup.Id),
		ParameterGroupType:      types.StringValue(string(result.InstanceGroup.ParameterGroup.Type)),
		ApplyStatus:             utils.ConvertNullableString(result.InstanceGroup.ParameterGroup.ApplyStatus),
		EngineVersion:           types.StringValue(result.InstanceGroup.ParameterGroup.EngineVersion),
		IsEngineVersionMismatch: types.BoolValue(result.InstanceGroup.ParameterGroup.IsEngineVersionMismatch),
	}, true, true
}

func (r *instanceGroupResource) pollParameterGroupUntilApplied(
	ctx context.Context,
	plan instanceGroupParameterGroupModel,
	respDiags *diag.Diagnostics,
) (instanceGroupParameterGroupModel, bool, bool) {
	ticker := time.NewTicker(mysqlPollInterval)
	defer ticker.Stop()

	for {
		state, found, ok := r.readParameterGroupState(ctx, plan.InstanceGroupId.ValueString(), respDiags)
		if !ok || !found {
			return state, found, ok
		}

		if state.ParameterGroupId.ValueString() == plan.ParameterGroupId.ValueString() &&
			state.ParameterGroupType.ValueString() == plan.ParameterGroupType.ValueString() {
			switch state.ApplyStatus.ValueString() {
			case string(mysqlsdk.PARAMETERGROUPSTATUS_PENDING), string(mysqlsdk.PARAMETERGROUPSTATUS_APPLYING):
			case string(mysqlsdk.PARAMETERGROUPSTATUS_ERROR_SYNC):
				return state, true, true
			default:
				return state, true, true
			}
		}

		select {
		case <-ctx.Done():
			common.AddGeneralError(ctx, r, respDiags, "context deadline exceeded")
			return instanceGroupParameterGroupModel{}, false, false
		case <-ticker.C:
		}
	}
}

func parameterGroupSyncNeedsRetry(state instanceGroupParameterGroupModel) bool {
	return state.ApplyStatus.ValueString() == string(mysqlsdk.PARAMETERGROUPSTATUS_ERROR_SYNC)
}

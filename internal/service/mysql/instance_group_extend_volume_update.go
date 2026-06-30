// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package mysql

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-kakaocloud/internal/common"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	mysqlsdk "github.com/kakaoenterprise/kc-sdk-go/services/mysql"
)

func (r *instanceGroupResource) extendVolume(
	ctx context.Context,
	plan instanceGroupExtendVolumeModel,
	respDiags *diag.Diagnostics,
) bool {
	request := mysqlsdk.NewBodyExtendMysqlInstanceGroupVolume(
		*mysqlsdk.NewMysqlV1ApiExtendMysqlInstanceGroupVolumeModelInstanceGroupRequestModel(
			plan.LogDiskSize.ValueInt32(),
			plan.DataDiskSize.ValueInt32(),
		),
	)

	_, httpResp, err := common.ExecuteWithRetryAndAuth[struct{}](ctx, r.kc, respDiags,
		func() (struct{}, *http.Response, error) {
			httpResp, err := r.kc.ApiClient.MySQLInstanceGroupsAPI.
				ExtendMysqlInstanceGroupVolume(ctx, plan.InstanceGroupId.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				BodyExtendMysqlInstanceGroupVolume(*request).
				Execute()
			return struct{}{}, httpResp, err
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "ExtendMysqlInstanceGroupVolume", err, respDiags)
		return false
	}

	return true
}

func (r *instanceGroupResource) readExtendVolumeState(
	ctx context.Context,
	current instanceGroupExtendVolumeModel,
	respDiags *diag.Diagnostics,
) (instanceGroupExtendVolumeModel, bool, bool) {
	result, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, respDiags,
		func() (*mysqlsdk.GetMySQLInstanceGroupResponseModel, *http.Response, error) {
			return r.kc.ApiClient.MySQLInstanceGroupsAPI.
				GetMysqlInstanceGroup(ctx, current.InstanceGroupId.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				Execute()
		},
	)
	if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
		return instanceGroupExtendVolumeModel{}, false, true
	}
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "GetMysqlInstanceGroup", err, respDiags)
		return instanceGroupExtendVolumeModel{}, false, false
	}

	return instanceGroupExtendVolumeModel{
		Id:                  types.StringValue(current.InstanceGroupId.ValueString()),
		InstanceGroupId:     current.InstanceGroupId,
		LogDiskSize:         types.Int32Value(result.InstanceGroup.SpecContent.LogDiskSize),
		DataDiskSize:        types.Int32Value(result.InstanceGroup.SpecContent.DataDiskSize),
		InstanceGroupStatus: types.StringValue(string(result.InstanceGroup.Status)),
	}, true, true
}

func (r *instanceGroupResource) pollExtendVolumeUntilStable(
	ctx context.Context,
	plan instanceGroupExtendVolumeModel,
	respDiags *diag.Diagnostics,
) (instanceGroupExtendVolumeModel, bool, bool) {
	ticker := time.NewTicker(mysqlPollInterval)
	defer ticker.Stop()

	for {
		state, found, ok := r.readExtendVolumeState(ctx, plan, respDiags)
		if !ok || !found {
			return state, found, ok
		}

		if state.LogDiskSize.ValueInt32() >= plan.LogDiskSize.ValueInt32() &&
			state.DataDiskSize.ValueInt32() >= plan.DataDiskSize.ValueInt32() &&
			state.InstanceGroupStatus.ValueString() == string(mysqlsdk.INSTANCEGROUPSTATUS_AVAILABLE) {
			return state, true, true
		}

		switch state.InstanceGroupStatus.ValueString() {
		case string(mysqlsdk.INSTANCEGROUPSTATUS_ERROR), string(mysqlsdk.INSTANCEGROUPSTATUS_TERMINATED):
			common.AddGeneralError(ctx, r, respDiags, fmt.Sprintf("extend volume finished with unexpected instance group status %q", state.InstanceGroupStatus.ValueString()))
			return state, true, false
		}

		select {
		case <-ctx.Done():
			common.AddGeneralError(ctx, r, respDiags, "context deadline exceeded")
			return instanceGroupExtendVolumeModel{}, false, false
		case <-ticker.C:
		}
	}
}

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package mysql

import (
	"context"
	"net/http"
	"slices"
	"terraform-provider-kakaocloud/internal/common"
	"terraform-provider-kakaocloud/internal/utils"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	mysqlsdk "github.com/kakaoenterprise/kc-sdk-go/services/mysql"
)

func (r *instanceGroupResource) applySecurityGroups(
	ctx context.Context,
	plan instanceGroupSecurityGroupsModel,
	respDiags *diag.Diagnostics,
) bool {
	var securityGroupIds []string
	respDiags.Append(plan.SecurityGroupIds.ElementsAs(ctx, &securityGroupIds, false)...)
	if respDiags.HasError() {
		return false
	}

	requestGroup := mysqlsdk.NewMysqlV1ApiUpdateMysqlSecurityGroupsModelInstanceGroupRequestModel(securityGroupIds)
	request := mysqlsdk.NewBodyUpdateMysqlSecurityGroups(*requestGroup)

	_, httpResp, err := common.ExecuteWithRetryAndAuth[struct{}](ctx, r.kc, respDiags,
		func() (struct{}, *http.Response, error) {
			httpResp, err := r.kc.ApiClient.MySQLInstanceGroupsAPI.
				UpdateMysqlSecurityGroups(ctx, plan.InstanceGroupId.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				BodyUpdateMysqlSecurityGroups(*request).
				Execute()
			return struct{}{}, httpResp, err
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "UpdateMysqlSecurityGroups", err, respDiags)
		return false
	}

	return true
}

func (r *instanceGroupResource) readSecurityGroupsState(
	ctx context.Context,
	instanceGroupId string,
	respDiags *diag.Diagnostics,
) (instanceGroupSecurityGroupsModel, bool, bool) {
	result, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, respDiags,
		func() (*mysqlsdk.GetMySQLInstanceGroupResponseModel, *http.Response, error) {
			return r.kc.ApiClient.MySQLInstanceGroupsAPI.
				GetMysqlInstanceGroup(ctx, instanceGroupId).
				XAuthToken(r.kc.XAuthToken).
				Execute()
		},
	)
	if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
		return instanceGroupSecurityGroupsModel{}, false, true
	}
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "GetMysqlInstanceGroup", err, respDiags)
		return instanceGroupSecurityGroupsModel{}, false, false
	}

	var securityGroupIds []string
	if networkInfo := result.InstanceGroup.NetworkInfo.Get(); networkInfo != nil {
		securityGroupIds = slices.Clone(networkInfo.SecurityGroupIds)
	}
	slices.Sort(securityGroupIds)

	setValue, diags := utils.SetFromStrings(ctx, securityGroupIds)
	respDiags.Append(diags...)
	if respDiags.HasError() {
		return instanceGroupSecurityGroupsModel{}, false, false
	}

	return instanceGroupSecurityGroupsModel{
		Id:               types.StringValue(instanceGroupId),
		InstanceGroupId:  types.StringValue(instanceGroupId),
		SecurityGroupIds: setValue,
	}, true, true
}

func (r *instanceGroupResource) pollSecurityGroupsUntilApplied(
	ctx context.Context,
	desired instanceGroupSecurityGroupsModel,
	respDiags *diag.Diagnostics,
) (instanceGroupSecurityGroupsModel, bool, bool) {
	ticker := time.NewTicker(mysqlPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			common.AddGeneralError(ctx, r, respDiags, "context deadline exceeded")
			return instanceGroupSecurityGroupsModel{}, false, false
		case <-ticker.C:
			state, found, ok := r.readSecurityGroupsState(ctx, desired.InstanceGroupId.ValueString(), respDiags)
			if !ok {
				return instanceGroupSecurityGroupsModel{}, false, false
			}
			if !found {
				return instanceGroupSecurityGroupsModel{}, false, true
			}
			if desired.SecurityGroupIds.Equal(state.SecurityGroupIds) {
				return state, true, true
			}
		}
	}
}

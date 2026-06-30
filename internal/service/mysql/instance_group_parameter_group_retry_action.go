// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package mysql

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-kakaocloud/internal/common"
	"terraform-provider-kakaocloud/internal/utils"
	"time"

	resourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/action"
	actionschema "github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	mysqlsdk "github.com/kakaoenterprise/kc-sdk-go/services/mysql"
)

var _ action.ActionWithConfigure = &instanceGroupParameterGroupRetryAction{}

func NewInstanceGroupParameterGroupRetryAction() action.Action {
	return &instanceGroupParameterGroupRetryAction{}
}

type instanceGroupParameterGroupRetryAction struct{ mysqlActionBase }

type instanceGroupParameterGroupRetryActionModel struct {
	InstanceGroupId types.String `tfsdk:"instance_group_id"`
}

type instanceGroupParameterGroupRetryModel struct {
	Id                      types.String           `tfsdk:"id"`
	InstanceGroupId         types.String           `tfsdk:"instance_group_id"`
	ParameterGroupId        types.String           `tfsdk:"parameter_group_id"`
	ParameterGroupType      types.String           `tfsdk:"parameter_group_type"`
	ApplyStatus             types.String           `tfsdk:"apply_status"`
	EngineVersion           types.String           `tfsdk:"engine_version"`
	IsEngineVersionMismatch types.Bool             `tfsdk:"is_engine_version_mismatch"`
	Timeouts                resourceTimeouts.Value `tfsdk:"timeouts"`
}

func (a *instanceGroupParameterGroupRetryAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mysql_instance_group_parameter_group_retry"
}

func (a *instanceGroupParameterGroupRetryAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = actionschema.Schema{
		Attributes: map[string]actionschema.Attribute{
			"instance_group_id": mysqlActionInstanceGroupIDAttribute(),
		},
	}
}

func (a *instanceGroupParameterGroupRetryAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	a.configure(req, resp)
}

func (a *instanceGroupParameterGroupRetryAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var config instanceGroupParameterGroupRetryActionModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, common.DefaultCreateTimeout)
	defer cancel()

	if !a.retryParameterGroupSync(ctx, config.InstanceGroupId.ValueString(), &resp.Diagnostics) {
		return
	}
	stopProgress := common.StartActionProgress(ctx, resp.SendProgress, fmt.Sprintf("Waiting for MySQL instance group %s parameter group sync to complete", config.InstanceGroupId.ValueString()))
	defer stopProgress()
	a.pollParameterGroupSyncUntilStable(ctx, config.InstanceGroupId.ValueString(), resourceTimeouts.Value{}, &resp.Diagnostics)
}

func (a *instanceGroupParameterGroupRetryAction) retryParameterGroupSync(ctx context.Context, instanceGroupID string, respDiags *diag.Diagnostics) bool {
	_, httpResp, err := common.ExecuteWithRetryAndAuth[struct{}](ctx, a.kc, respDiags,
		func() (struct{}, *http.Response, error) {
			httpResp, err := a.kc.ApiClient.MySQLInstanceGroupsParameterGroupsAPI.
				RetryMysqlParameterGroupSync(ctx, instanceGroupID).
				XAuthToken(a.kc.XAuthToken).
				Execute()
			return struct{}{}, httpResp, err
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, a, httpResp, "RetryMysqlParameterGroupSync", err, respDiags)
		return false
	}

	return true
}

func (a *instanceGroupParameterGroupRetryAction) readParameterGroupRetryState(ctx context.Context, instanceGroupID string, timeouts resourceTimeouts.Value, respDiags *diag.Diagnostics) (instanceGroupParameterGroupRetryModel, bool, bool) {
	result, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, a.kc, respDiags,
		func() (*mysqlsdk.GetMySQLInstanceGroupResponseModel, *http.Response, error) {
			return a.kc.ApiClient.MySQLInstanceGroupsAPI.
				GetMysqlInstanceGroup(ctx, instanceGroupID).
				XAuthToken(a.kc.XAuthToken).
				Execute()
		},
	)
	if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
		return instanceGroupParameterGroupRetryModel{}, false, true
	}
	if err != nil {
		common.AddApiActionError(ctx, a, httpResp, "GetMysqlInstanceGroup", err, respDiags)
		return instanceGroupParameterGroupRetryModel{}, false, false
	}

	return instanceGroupParameterGroupRetryModel{
		Id:                      types.StringValue(instanceGroupID),
		InstanceGroupId:         types.StringValue(instanceGroupID),
		ParameterGroupId:        types.StringValue(result.InstanceGroup.ParameterGroup.Id),
		ParameterGroupType:      types.StringValue(string(result.InstanceGroup.ParameterGroup.Type)),
		ApplyStatus:             utils.ConvertNullableString(result.InstanceGroup.ParameterGroup.ApplyStatus),
		EngineVersion:           types.StringValue(result.InstanceGroup.ParameterGroup.EngineVersion),
		IsEngineVersionMismatch: types.BoolValue(result.InstanceGroup.ParameterGroup.IsEngineVersionMismatch),
		Timeouts:                timeouts,
	}, true, true
}

func (a *instanceGroupParameterGroupRetryAction) pollParameterGroupSyncUntilStable(ctx context.Context, instanceGroupID string, timeouts resourceTimeouts.Value, respDiags *diag.Diagnostics) (instanceGroupParameterGroupRetryModel, bool, bool) {
	ticker := time.NewTicker(mysqlPollInterval)
	defer ticker.Stop()

	for {
		state, found, ok := a.readParameterGroupRetryState(ctx, instanceGroupID, timeouts, respDiags)
		if !ok || !found {
			return state, found, ok
		}

		switch state.ApplyStatus.ValueString() {
		case string(mysqlsdk.PARAMETERGROUPSTATUS_PENDING), string(mysqlsdk.PARAMETERGROUPSTATUS_APPLYING):
		case string(mysqlsdk.PARAMETERGROUPSTATUS_ERROR_SYNC):
			respDiags.AddError(
				"invoke action: kakaocloud_mysql_instance_group_parameter_group_retry",
				"parameter group sync is still in ERROR-SYNC status after retry",
			)
			return state, true, false
		default:
			return state, true, true
		}

		select {
		case <-ctx.Done():
			respDiags.AddError(
				"invoke action: kakaocloud_mysql_instance_group_parameter_group_retry",
				"context deadline exceeded",
			)
			return instanceGroupParameterGroupRetryModel{}, false, false
		case <-ticker.C:
		}
	}
}

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package mysql

import (
	"context"
	"fmt"
	"net/http"
	"slices"
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

var _ action.ActionWithConfigure = &instanceGroupSwitchoverAction{}

func NewInstanceGroupSwitchoverAction() action.Action { return &instanceGroupSwitchoverAction{} }

type instanceGroupSwitchoverAction struct{ mysqlActionBase }

type instanceGroupSwitchoverActionModel struct {
	InstanceGroupId types.String `tfsdk:"instance_group_id"`
}

type instanceGroupSwitchoverModel struct {
	Id                 types.String           `tfsdk:"id"`
	InstanceGroupId    types.String           `tfsdk:"instance_group_id"`
	IsMultiAz          types.Bool             `tfsdk:"is_multi_az"`
	Status             types.String           `tfsdk:"status"`
	PrimaryInstanceId  types.String           `tfsdk:"primary_instance_id"`
	StandbyInstanceIds types.Set              `tfsdk:"standby_instance_ids"`
	Timeouts           resourceTimeouts.Value `tfsdk:"timeouts"`
}

func (a *instanceGroupSwitchoverAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mysql_instance_group_switchover"
}

func (a *instanceGroupSwitchoverAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = actionschema.Schema{
		Attributes: map[string]actionschema.Attribute{
			"instance_group_id": mysqlActionInstanceGroupIDAttribute(),
		},
	}
}

func (a *instanceGroupSwitchoverAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	a.configure(req, resp)
}

func (a *instanceGroupSwitchoverAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var config instanceGroupSwitchoverActionModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, common.DefaultCreateTimeout)
	defer cancel()

	beforeState, found, ok := a.readSwitchoverState(ctx, config.InstanceGroupId.ValueString(), resourceTimeouts.Value{}, &resp.Diagnostics)
	if !ok {
		return
	}
	if !found {
		common.AddGeneralError(ctx, a, &resp.Diagnostics, "instance group was not found before switchover")
		return
	}
	if !beforeState.IsMultiAz.ValueBool() {
		common.AddGeneralError(ctx, a, &resp.Diagnostics, "mysql_instance_group_switchover is supported only for multi-AZ instance groups")
		return
	}
	if beforeState.PrimaryInstanceId.IsNull() || beforeState.PrimaryInstanceId.ValueString() == "" {
		common.AddGeneralError(ctx, a, &resp.Diagnostics, "instance group primary instance information is missing before switchover")
		return
	}
	if !a.switchover(ctx, config.InstanceGroupId.ValueString(), &resp.Diagnostics) {
		return
	}
	stopProgress := common.StartActionProgress(ctx, resp.SendProgress, fmt.Sprintf("Waiting for MySQL instance group %s to become available after switchover", config.InstanceGroupId.ValueString()))
	defer stopProgress()
	a.pollSwitchoverUntilStable(ctx, beforeState, beforeState.PrimaryInstanceId.ValueString(), &resp.Diagnostics)
}

func (a *instanceGroupSwitchoverAction) switchover(ctx context.Context, instanceGroupID string, respDiags *diag.Diagnostics) bool {
	_, httpResp, err := common.ExecuteWithRetryAndAuth[struct{}](ctx, a.kc, respDiags,
		func() (struct{}, *http.Response, error) {
			httpResp, err := a.kc.ApiClient.MySQLInstanceGroupsAPI.
				SwitchoverMysqlInstanceGroup(ctx, instanceGroupID).
				XAuthToken(a.kc.XAuthToken).
				Execute()
			return struct{}{}, httpResp, err
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, a, httpResp, "SwitchoverMysqlInstanceGroup", err, respDiags)
		return false
	}

	return true
}

func (a *instanceGroupSwitchoverAction) readSwitchoverState(ctx context.Context, instanceGroupID string, timeouts resourceTimeouts.Value, respDiags *diag.Diagnostics) (instanceGroupSwitchoverModel, bool, bool) {
	result, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, a.kc, respDiags,
		func() (*mysqlsdk.GetMySQLInstanceGroupResponseModel, *http.Response, error) {
			return a.kc.ApiClient.MySQLInstanceGroupsAPI.
				GetMysqlInstanceGroup(ctx, instanceGroupID).
				XAuthToken(a.kc.XAuthToken).
				Execute()
		},
	)
	if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
		return instanceGroupSwitchoverModel{}, false, true
	}
	if err != nil {
		common.AddApiActionError(ctx, a, httpResp, "GetMysqlInstanceGroup", err, respDiags)
		return instanceGroupSwitchoverModel{}, false, false
	}

	instances, hasInstances := result.InstanceGroup.GetInstancesOk()
	if !hasInstances || instances == nil {
		common.AddGeneralError(ctx, a, respDiags, "instance group instance information is missing")
		return instanceGroupSwitchoverModel{}, false, false
	}
	standbyInstanceIDs := make([]string, 0, len(instances.Standby))
	for _, standby := range instances.Standby {
		if instanceID, ok := utils.GetNullableStringValue(standby.InstanceId); ok && instanceID != "" {
			standbyInstanceIDs = append(standbyInstanceIDs, instanceID)
		}
	}
	slices.Sort(standbyInstanceIDs)

	standbySet, diags := utils.SetFromStrings(ctx, standbyInstanceIDs)
	respDiags.Append(diags...)
	if respDiags.HasError() {
		return instanceGroupSwitchoverModel{}, false, false
	}
	primaryInstanceID := types.StringNull()
	if primary, _ := instances.GetPrimaryOk(); primary != nil {
		if instanceID, ok := utils.GetNullableStringValue(primary.InstanceId); ok && instanceID != "" {
			primaryInstanceID = types.StringValue(instanceID)
		}
	}

	return instanceGroupSwitchoverModel{
		Id:                 types.StringValue(instanceGroupID),
		InstanceGroupId:    types.StringValue(instanceGroupID),
		IsMultiAz:          types.BoolValue(result.InstanceGroup.IsMultiAz),
		Status:             types.StringValue(string(result.InstanceGroup.Status)),
		PrimaryInstanceId:  primaryInstanceID,
		StandbyInstanceIds: standbySet,
		Timeouts:           timeouts,
	}, true, true
}

func (a *instanceGroupSwitchoverAction) pollSwitchoverUntilStable(ctx context.Context, plan instanceGroupSwitchoverModel, previousPrimaryInstanceID string, respDiags *diag.Diagnostics) (instanceGroupSwitchoverModel, bool, bool) {
	ticker := time.NewTicker(mysqlPollInterval)
	defer ticker.Stop()

	for {
		state, found, ok := a.readSwitchoverState(ctx, plan.InstanceGroupId.ValueString(), plan.Timeouts, respDiags)
		if !ok || !found {
			return state, found, ok
		}

		switch state.Status.ValueString() {
		case string(mysqlsdk.INSTANCEGROUPSTATUS_SWITCHING):
		default:
			if !state.PrimaryInstanceId.IsNull() && (previousPrimaryInstanceID == "" || state.PrimaryInstanceId.ValueString() != previousPrimaryInstanceID) {
				return state, true, true
			}
		}

		select {
		case <-ctx.Done():
			common.AddGeneralError(ctx, a, respDiags, "context deadline exceeded")
			return instanceGroupSwitchoverModel{}, false, false
		case <-ticker.C:
		}
	}
}

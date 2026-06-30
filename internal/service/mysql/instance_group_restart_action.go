// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package mysql

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"terraform-provider-kakaocloud/internal/common"
	"terraform-provider-kakaocloud/internal/utils"
	"time"

	resourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/action"
	actionschema "github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	mysqlsdk "github.com/kakaoenterprise/kc-sdk-go/services/mysql"
)

var _ action.ActionWithConfigure = &instanceGroupRestartAction{}

func NewInstanceGroupRestartAction() action.Action { return &instanceGroupRestartAction{} }

type instanceGroupRestartAction struct{ mysqlActionBase }

type instanceGroupRestartActionModel struct {
	InstanceGroupId types.String `tfsdk:"instance_group_id"`
	InstanceIds     types.Set    `tfsdk:"instance_ids"`
}

type instanceGroupRestartModel struct {
	Id                  types.String           `tfsdk:"id"`
	InstanceGroupId     types.String           `tfsdk:"instance_group_id"`
	InstanceIds         types.Set              `tfsdk:"instance_ids"`
	InstanceGroupStatus types.String           `tfsdk:"instance_group_status"`
	Timeouts            resourceTimeouts.Value `tfsdk:"timeouts"`
}

func (a *instanceGroupRestartAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mysql_instance_group_restart"
}

func (a *instanceGroupRestartAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = actionschema.Schema{
		Attributes: map[string]actionschema.Attribute{
			"instance_group_id": mysqlActionInstanceGroupIDAttribute(),
			"instance_ids": actionschema.SetAttribute{
				Required:    true,
				ElementType: types.StringType,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
					setvalidator.ValueStringsAre(common.UuidValidator()...),
				},
			},
		},
	}
}

func (a *instanceGroupRestartAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	a.configure(req, resp)
}

func (a *instanceGroupRestartAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var config instanceGroupRestartActionModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, common.DefaultCreateTimeout)
	defer cancel()

	instanceIDs, ok := mysqlActionInstanceIDsFromSet(ctx, config.InstanceIds, &resp.Diagnostics)
	if !ok {
		return
	}
	tflog.Info(ctx, "invoking MySQL restart action", map[string]any{
		"instance_group_id": config.InstanceGroupId.ValueString(),
		"instance_ids":      instanceIDs,
	})
	if !a.restartInstances(ctx, config.InstanceGroupId.ValueString(), instanceIDs, &resp.Diagnostics) {
		return
	}
	stopProgress := common.StartActionProgress(ctx, resp.SendProgress, fmt.Sprintf("Waiting for MySQL instance group %s to become available after restart", config.InstanceGroupId.ValueString()))
	defer stopProgress()
	a.pollRestartUntilStable(ctx, config.InstanceGroupId.ValueString(), instanceIDs, resourceTimeouts.Value{}, &resp.Diagnostics)
}

func (a *instanceGroupRestartAction) restartInstances(ctx context.Context, instanceGroupID string, instanceIDs []string, respDiags *diag.Diagnostics) bool {
	requestGroup := mysqlsdk.NewMysqlV1ApiRestartMysqlInstancesModelInstanceGroupRequestModel(instanceIDs)
	request := mysqlsdk.NewBodyRestartMysqlInstances(*requestGroup)

	_, httpResp, err := common.ExecuteWithRetryAndAuth[struct{}](ctx, a.kc, respDiags,
		func() (struct{}, *http.Response, error) {
			httpResp, err := a.kc.ApiClient.MySQLInstanceGroupsAPI.
				RestartMysqlInstances(ctx, instanceGroupID).
				XAuthToken(a.kc.XAuthToken).
				BodyRestartMysqlInstances(*request).
				Execute()
			return struct{}{}, httpResp, err
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, a, httpResp, "RestartMysqlInstances", err, respDiags)
		return false
	}

	return true
}

func (a *instanceGroupRestartAction) readRestartState(ctx context.Context, instanceGroupID string, instanceIDs []string, timeouts resourceTimeouts.Value, respDiags *diag.Diagnostics) (instanceGroupRestartModel, bool, bool) {
	groupResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, a.kc, respDiags,
		func() (*mysqlsdk.GetMySQLInstanceGroupResponseModel, *http.Response, error) {
			return a.kc.ApiClient.MySQLInstanceGroupsAPI.
				GetMysqlInstanceGroup(ctx, instanceGroupID).
				XAuthToken(a.kc.XAuthToken).
				Execute()
		},
	)
	if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
		return instanceGroupRestartModel{}, false, true
	}
	if err != nil {
		common.AddApiActionError(ctx, a, httpResp, "GetMysqlInstanceGroup", err, respDiags)
		return instanceGroupRestartModel{}, false, false
	}

	instances, hasInstances := groupResp.InstanceGroup.GetInstancesOk()
	if !hasInstances || instances == nil {
		common.AddGeneralError(ctx, a, respDiags, "instance group instance information is missing")
		return instanceGroupRestartModel{}, false, false
	}
	exists := make(map[string]struct{}, 1+len(instances.Standby))
	primary, _ := instances.GetPrimaryOk()
	if primary != nil {
		if instanceID, ok := utils.GetNullableStringValue(primary.InstanceId); ok && instanceID != "" {
			exists[instanceID] = struct{}{}
		}
	}
	for _, instance := range instances.Standby {
		if instanceID, ok := utils.GetNullableStringValue(instance.InstanceId); ok && instanceID != "" {
			exists[instanceID] = struct{}{}
		}
	}
	if len(exists) > 0 {
		for _, instanceID := range instanceIDs {
			if _, ok := exists[instanceID]; !ok {
				common.AddGeneralError(ctx, a, respDiags, fmt.Sprintf("instance_id %q was not found in the instance group", instanceID))
				return instanceGroupRestartModel{}, false, false
			}
		}
	}

	instanceIDSet, diags := utils.SetFromStrings(ctx, instanceIDs)
	respDiags.Append(diags...)
	if respDiags.HasError() {
		return instanceGroupRestartModel{}, false, false
	}

	return instanceGroupRestartModel{
		Id:                  types.StringValue(instanceGroupID + "/" + strings.Join(instanceIDs, ",")),
		InstanceGroupId:     types.StringValue(instanceGroupID),
		InstanceIds:         instanceIDSet,
		InstanceGroupStatus: types.StringValue(string(groupResp.InstanceGroup.Status)),
		Timeouts:            timeouts,
	}, true, true
}

func (a *instanceGroupRestartAction) pollRestartUntilStable(ctx context.Context, instanceGroupID string, instanceIDs []string, timeouts resourceTimeouts.Value, respDiags *diag.Diagnostics) (instanceGroupRestartModel, bool, bool) {
	ticker := time.NewTicker(mysqlPollInterval)
	defer ticker.Stop()

	for {
		state, found, ok := a.readRestartState(ctx, instanceGroupID, instanceIDs, timeouts, respDiags)
		if !ok || !found {
			return state, found, ok
		}

		switch state.InstanceGroupStatus.ValueString() {
		case string(mysqlsdk.INSTANCEGROUPSTATUS_REBOOTING), string(mysqlsdk.INSTANCEGROUPSTATUS_MODIFYING):
		case string(mysqlsdk.INSTANCEGROUPSTATUS_ERROR), string(mysqlsdk.INSTANCEGROUPSTATUS_TERMINATED):
			common.AddGeneralError(ctx, a, respDiags, fmt.Sprintf("instance restart finished with unexpected instance group status %q", state.InstanceGroupStatus.ValueString()))
			return state, true, false
		default:
			return state, true, true
		}

		select {
		case <-ctx.Done():
			common.AddGeneralError(ctx, a, respDiags, "context deadline exceeded")
			return instanceGroupRestartModel{}, false, false
		case <-ticker.C:
		}
	}
}

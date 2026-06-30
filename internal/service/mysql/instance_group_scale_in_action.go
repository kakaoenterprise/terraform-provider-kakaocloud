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

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/action"
	actionschema "github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	mysqlsdk "github.com/kakaoenterprise/kc-sdk-go/services/mysql"
)

var _ action.ActionWithConfigure = &instanceGroupScaleInAction{}

func NewInstanceGroupScaleInAction() action.Action { return &instanceGroupScaleInAction{} }

type instanceGroupScaleInAction struct{ mysqlActionBase }

type instanceGroupScaleInActionModel struct {
	InstanceGroupId types.String `tfsdk:"instance_group_id"`
	InstanceIds     types.Set    `tfsdk:"instance_ids"`
}

func (a *instanceGroupScaleInAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mysql_instance_group_scale_in"
}

func (a *instanceGroupScaleInAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
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

func (a *instanceGroupScaleInAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	a.configure(req, resp)
}

func (a *instanceGroupScaleInAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var config instanceGroupScaleInActionModel
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
	if !a.validateScaleInTargets(ctx, config.InstanceGroupId.ValueString(), instanceIDs, &resp.Diagnostics) {
		return
	}

	tflog.Info(ctx, "invoking MySQL scale-in action", map[string]any{
		"instance_group_id": config.InstanceGroupId.ValueString(),
		"instance_ids":      instanceIDs,
	})
	if !a.scaleIn(ctx, config.InstanceGroupId.ValueString(), instanceIDs, &resp.Diagnostics) {
		return
	}
	stopProgress := common.StartActionProgress(ctx, resp.SendProgress, fmt.Sprintf("Waiting for MySQL instance group %s to become available after scale in", config.InstanceGroupId.ValueString()))
	defer stopProgress()
	a.pollScaleInUntilApplied(ctx, config.InstanceGroupId.ValueString(), instanceIDs, &resp.Diagnostics)
}

func (a *instanceGroupScaleInAction) scaleIn(ctx context.Context, instanceGroupID string, instanceIDs []string, respDiags *diag.Diagnostics) bool {
	request := mysqlsdk.NewBodyScaleInMysqlInstanceGroup(*mysqlsdk.NewMysqlV1ApiScaleInMysqlInstanceGroupModelInstanceGroupRequestModel(instanceIDs))

	_, httpResp, err := common.ExecuteWithRetryAndAuth[struct{}](ctx, a.kc, respDiags,
		func() (struct{}, *http.Response, error) {
			httpResp, err := a.kc.ApiClient.MySQLInstanceGroupsAPI.
				ScaleInMysqlInstanceGroup(ctx, instanceGroupID).
				XAuthToken(a.kc.XAuthToken).
				BodyScaleInMysqlInstanceGroup(*request).
				Execute()
			return struct{}{}, httpResp, err
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, a, httpResp, "ScaleInMysqlInstanceGroup", err, respDiags)
		return false
	}
	return true
}

func (a *instanceGroupScaleInAction) validateScaleInTargets(ctx context.Context, instanceGroupID string, instanceIDs []string, respDiags *diag.Diagnostics) bool {
	result, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, a.kc, respDiags,
		func() (*mysqlsdk.GetMySQLInstanceGroupResponseModel, *http.Response, error) {
			return a.kc.ApiClient.MySQLInstanceGroupsAPI.
				GetMysqlInstanceGroup(ctx, instanceGroupID).
				XAuthToken(a.kc.XAuthToken).
				Execute()
		},
	)
	if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
		common.AddGeneralError(ctx, a, respDiags, "instance group was not found before scale in")
		return false
	}
	if err != nil {
		common.AddApiActionError(ctx, a, httpResp, "GetMysqlInstanceGroup", err, respDiags)
		return false
	}

	instances, hasInstances := result.InstanceGroup.GetInstancesOk()
	if !hasInstances || instances == nil {
		common.AddGeneralError(ctx, a, respDiags, "instance group instance information is missing")
		return false
	}

	standbyIDs := map[string]struct{}{}
	for _, standby := range instances.Standby {
		if instanceID, ok := utils.GetNullableStringValue(standby.InstanceId); ok && instanceID != "" {
			standbyIDs[instanceID] = struct{}{}
		}
	}
	primaryInstanceID := ""
	if primary, _ := instances.GetPrimaryOk(); primary != nil {
		if instanceID, ok := utils.GetNullableStringValue(primary.InstanceId); ok {
			primaryInstanceID = instanceID
		}
	}
	for _, instanceID := range instanceIDs {
		if instanceID == primaryInstanceID {
			common.AddGeneralError(ctx, a, respDiags, fmt.Sprintf("instance_id %q is the primary instance and cannot be used for scale in", instanceID))
			return false
		}
		if _, ok := standbyIDs[instanceID]; !ok {
			common.AddGeneralError(ctx, a, respDiags, fmt.Sprintf("instance_id %q was not found in the instance group standby instances", instanceID))
			return false
		}
	}
	return true
}

func (a *instanceGroupScaleInAction) pollScaleInUntilApplied(
	ctx context.Context,
	instanceGroupID string,
	targetInstanceIDs []string,
	respDiags *diag.Diagnostics,
) bool {
	targets := make(map[string]struct{}, len(targetInstanceIDs))
	for _, instanceID := range targetInstanceIDs {
		targets[instanceID] = struct{}{}
	}

	ticker := time.NewTicker(mysqlPollInterval)
	defer ticker.Stop()

	for {
		status, remaining, ok := a.readScaleInStatus(ctx, instanceGroupID, targets, respDiags)
		if !ok {
			return false
		}

		switch status {
		case string(mysqlsdk.INSTANCEGROUPSTATUS_ERROR), string(mysqlsdk.INSTANCEGROUPSTATUS_TERMINATED):
			common.AddGeneralError(ctx, a, respDiags, fmt.Sprintf("scale in finished with unexpected instance group status %q", status))
			return false
		case string(mysqlsdk.INSTANCEGROUPSTATUS_AVAILABLE), string(mysqlsdk.INSTANCEGROUPSTATUS_PRIMARY_AVAILABLE):
			if !remaining {
				return true
			}
		}

		select {
		case <-ctx.Done():
			common.AddGeneralError(ctx, a, respDiags, "context deadline exceeded")
			return false
		case <-ticker.C:
		}
	}
}

func (a *instanceGroupScaleInAction) readScaleInStatus(
	ctx context.Context,
	instanceGroupID string,
	targets map[string]struct{},
	respDiags *diag.Diagnostics,
) (string, bool, bool) {
	result, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, a.kc, respDiags,
		func() (*mysqlsdk.GetMySQLInstanceGroupResponseModel, *http.Response, error) {
			return a.kc.ApiClient.MySQLInstanceGroupsAPI.
				GetMysqlInstanceGroup(ctx, instanceGroupID).
				XAuthToken(a.kc.XAuthToken).
				Execute()
		},
	)
	if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
		common.AddGeneralError(ctx, a, respDiags, "instance group was not found after scale in")
		return "", false, false
	}
	if err != nil {
		common.AddApiActionError(ctx, a, httpResp, "GetMysqlInstanceGroup", err, respDiags)
		return "", false, false
	}

	remaining := false
	if instances, ok := result.InstanceGroup.GetInstancesOk(); ok && instances != nil {
		for _, standby := range instances.Standby {
			instanceID, hasInstanceID := utils.GetNullableStringValue(standby.InstanceId)
			if !hasInstanceID {
				continue
			}
			if _, ok := targets[instanceID]; ok {
				remaining = true
				break
			}
		}
	}
	return string(result.InstanceGroup.Status), remaining, true
}

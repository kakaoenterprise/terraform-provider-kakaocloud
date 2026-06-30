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
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	mysqlsdk "github.com/kakaoenterprise/kc-sdk-go/services/mysql"
)

var (
	_ resource.ResourceWithConfigure      = &instanceGroupResource{}
	_ resource.ResourceWithImportState    = &instanceGroupResource{}
	_ resource.ResourceWithValidateConfig = &instanceGroupResource{}
	_ resource.ResourceWithModifyPlan     = &instanceGroupResource{}
)

func NewInstanceGroupResource() resource.Resource { return &instanceGroupResource{} }

type instanceGroupResource struct {
	kc *common.KakaoCloudClient
}

func (r *instanceGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mysql_instance_group"
}

func (r *instanceGroupResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: utils.MergeAttributes[schema.Attribute](
			instanceGroupResourceSchemaAttributes,
			map[string]schema.Attribute{
				"timeouts": resourceTimeouts.AttributesAll(ctx),
			},
		),
	}
}

func (r *instanceGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*common.KakaoCloudClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *common.KakaoCloudClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.kc = client
}

func (r *instanceGroupResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config instanceGroupResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.validateInstanceGroupConfig(ctx, config, &resp.Diagnostics)
}

func (r *instanceGroupResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}

	if req.State.Raw.IsNull() {
		var config instanceGroupResourceModel
		resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
		if resp.Diagnostics.HasError() {
			return
		}
		r.validateInstanceGroupCreateConfig(ctx, config, &resp.Diagnostics)
		return
	}

	var plan instanceGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	var state instanceGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	stateNetworkInfo, planNetworkInfo, ok := instanceGroupNetworkInfoForUpdate(ctx, state, plan, &resp.Diagnostics)
	if !ok {
		return
	}
	stateSpec, planSpec, ok := instanceGroupSpecContentForUpdate(ctx, state, plan, &resp.Diagnostics)
	if !ok {
		return
	}
	if primarySubnetIDChanged(ctx, stateNetworkInfo, planNetworkInfo, &resp.Diagnostics) {
		resp.RequiresReplace.Append(path.Root("desired_network_info").AtName("primary_subnet_info").AtName("subnet_id"))
	}
	if standbySubnetIDsReplaced(ctx, stateNetworkInfo, planNetworkInfo, &resp.Diagnostics) {
		resp.RequiresReplace.Append(path.Root("desired_network_info").AtName("standby_subnet_info"))
	}
	if resp.Diagnostics.HasError() {
		return
	}

	stateReplicas, ok := topologyReplicaCounts(ctx, stateNetworkInfo, &resp.Diagnostics)
	if !ok {
		return
	}
	planReplicas, ok := topologyReplicaCounts(ctx, planNetworkInfo, &resp.Diagnostics)
	if !ok {
		return
	}

	standbyPortChanged := !sameOptionalInt32Value(stateSpec.StandbyPort, planSpec.StandbyPort)
	isCurrentlyHA := totalReplicaCount(stateReplicas) > 1
	willBeHA := totalReplicaCount(planReplicas) > 1

	if standbyPortChanged && isCurrentlyHA && willBeHA {
		resp.RequiresReplace.Append(path.Root("spec_content").AtName("standby_port"))
	}
}

func (r *instanceGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan instanceGroupResourceModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	var config instanceGroupResourceModel
	diags = req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	timeout, diags := plan.Timeouts.Create(ctx, common.DefaultCreateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	createReq, ok := r.buildCreateRequest(ctx, plan, config, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	createResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*mysqlsdk.CreateMySQLInstanceGroupResponseModel, *http.Response, error) {
			return r.kc.ApiClient.MySQLInstanceGroupsAPI.
				CreateMysqlInstanceGroup(ctx).
				XAuthToken(r.kc.XAuthToken).
				BodyCreateMysqlInstanceGroup(createReq).
				Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "CreateMysqlInstanceGroup", err, &resp.Diagnostics)
		return
	}

	plan.Id = types.StringValue(createResp.InstanceGroup.Id)

	result, ok := r.pollInstanceGroupUntilStatus(
		ctx,
		plan.Id.ValueString(),
		[]string{
			string(mysqlsdk.INSTANCEGROUPSTATUS_AVAILABLE),
			string(mysqlsdk.INSTANCEGROUPSTATUS_PRIMARY_AVAILABLE),
			string(mysqlsdk.INSTANCEGROUPSTATUS_ERROR),
		},
		&resp.Diagnostics,
	)
	if !ok || resp.Diagnostics.HasError() {
		return
	}
	expectedStandbyCount, ok := expectedStandbyInstanceCount(ctx, plan.DesiredNetworkInfo, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	result, ok = r.pollInstanceGroupUntilTopologyReady(ctx, result.Id, expectedStandbyCount, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	common.CheckResourceAvailableStatus(
		ctx,
		r,
		stringPtr(string(result.Status)),
		[]string{
			string(mysqlsdk.INSTANCEGROUPSTATUS_AVAILABLE),
			string(mysqlsdk.INSTANCEGROUPSTATUS_PRIMARY_AVAILABLE),
		},
		&resp.Diagnostics,
	)
	if resp.Diagnostics.HasError() {
		return
	}

	state, mapDiags, ok := toInstanceGroupResourceModel(
		ctx,
		*result,
		plan.DesiredNetworkInfo,
		plan.Source,
		plan.Timeouts,
	)
	resp.Diagnostics.Append(mapDiags...)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *instanceGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state instanceGroupResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	timeout, diags := state.Timeouts.Read(ctx, common.DefaultReadTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	refreshed, found, ok := r.readInstanceGroupState(ctx, state.Id.ValueString(), state, &resp.Diagnostics)
	if !found {
		resp.State.RemoveResource(ctx)
		return
	}
	if !ok {
		return
	}

	diags = resp.State.Set(ctx, &refreshed)
	resp.Diagnostics.Append(diags...)
}

func (r *instanceGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan instanceGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state instanceGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	timeout, diags := plan.Timeouts.Update(ctx, common.DefaultUpdateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if !r.updateManagedSubresources(ctx, state, plan, &resp.Diagnostics) {
		return
	}

	refreshed, found, ok := r.readInstanceGroupState(ctx, state.Id.ValueString(), plan, &resp.Diagnostics)
	if !found {
		resp.State.RemoveResource(ctx)
		return
	}
	if !ok {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &refreshed)...)
}

func (r *instanceGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state instanceGroupResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	timeout, diags := state.Timeouts.Delete(ctx, common.DefaultDeleteTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if !r.deleteInstanceGroup(ctx, state.Id.ValueString(), &resp.Diagnostics) {
		return
	}
}

func (r *instanceGroupResource) deleteInstanceGroup(ctx context.Context, id string, respDiags *diag.Diagnostics) bool {
	_, httpResp, err := common.ExecuteWithRetryAndAuth[struct{}](ctx, r.kc, respDiags,
		func() (struct{}, *http.Response, error) {
			httpResp, err := r.kc.ApiClient.MySQLInstanceGroupsAPI.
				DeleteMysqlInstanceGroup(ctx, id).
				KeepAutoBackup(true).
				XAuthToken(r.kc.XAuthToken).
				Execute()
			return struct{}{}, httpResp, err
		},
	)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			return true
		}
		if httpResp != nil && httpResp.StatusCode == http.StatusConflict && strings.Contains(err.Error(), "already been deleted") {
			return true
		}
		common.AddApiActionError(ctx, r, httpResp, "DeleteMysqlInstanceGroup", err, respDiags)
		return false
	}

	return r.pollInstanceGroupUntilDeleted(ctx, id, respDiags)
}

func (r *instanceGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID == "" {
		common.AddImportFormatError(ctx, r, &resp.Diagnostics, "Expected import ID in the format: instance_group_id")
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

func (r *instanceGroupResource) readInstanceGroupState(
	ctx context.Context,
	id string,
	prev instanceGroupResourceModel,
	respDiags *diag.Diagnostics,
) (instanceGroupResourceModel, bool, bool) {
	result, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, respDiags,
		func() (*mysqlsdk.GetMySQLInstanceGroupResponseModel, *http.Response, error) {
			return r.kc.ApiClient.MySQLInstanceGroupsAPI.
				GetMysqlInstanceGroup(ctx, id).
				XAuthToken(r.kc.XAuthToken).
				Execute()
		},
	)
	if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
		var zero instanceGroupResourceModel
		return zero, false, true
	}
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "GetMysqlInstanceGroup", err, respDiags)
		var zero instanceGroupResourceModel
		return zero, false, false
	}

	model, mapDiags, ok := toInstanceGroupResourceModel(ctx, result.InstanceGroup, prev.DesiredNetworkInfo, prev.Source, prev.Timeouts)
	respDiags.Append(mapDiags...)
	if !ok {
		var zero instanceGroupResourceModel
		return zero, false, false
	}

	return model, true, true
}

func (r *instanceGroupResource) updateManagedSubresources(
	ctx context.Context,
	state instanceGroupResourceModel,
	plan instanceGroupResourceModel,
	respDiags *diag.Diagnostics,
) bool {
	if !r.updateSecurityGroupsIfChanged(ctx, state, plan, respDiags) {
		return false
	}
	if !r.updateTopologyIfChanged(ctx, state, plan, respDiags) {
		return false
	}
	if !r.extendVolumeIfChanged(ctx, state, plan, respDiags) {
		return false
	}
	if !r.updateBackupScheduleIfChanged(ctx, state, plan, respDiags) {
		return false
	}
	if !r.updateParameterGroupIfChanged(ctx, state, plan, respDiags) {
		return false
	}

	return true
}

func (r *instanceGroupResource) updateSecurityGroupsIfChanged(
	ctx context.Context,
	state instanceGroupResourceModel,
	plan instanceGroupResourceModel,
	respDiags *diag.Diagnostics,
) bool {
	stateNetworkInfo, planNetworkInfo, ok := instanceGroupNetworkInfoForUpdate(ctx, state, plan, respDiags)
	if !ok {
		return false
	}
	if stateNetworkInfo.SecurityGroupIds.Equal(planNetworkInfo.SecurityGroupIds) {
		return true
	}

	if !r.waitInstanceGroupAvailableForManagedUpdate(ctx, plan.Id.ValueString(), "security groups update", respDiags) {
		return false
	}
	securityPlan := instanceGroupSecurityGroupsModel{
		InstanceGroupId:  plan.Id,
		SecurityGroupIds: planNetworkInfo.SecurityGroupIds,
	}
	if !r.applySecurityGroups(ctx, securityPlan, respDiags) {
		return false
	}
	_, found, ok := r.pollSecurityGroupsUntilApplied(ctx, securityPlan, respDiags)
	if !ok {
		return false
	}
	if !found {
		common.AddGeneralError(ctx, r, respDiags, "instance group was not found after applying security groups")
		return false
	}
	if !r.waitInstanceGroupAvailableForManagedUpdate(ctx, plan.Id.ValueString(), "security groups update", respDiags) {
		return false
	}

	return true
}

func (r *instanceGroupResource) extendVolumeIfChanged(
	ctx context.Context,
	state instanceGroupResourceModel,
	plan instanceGroupResourceModel,
	respDiags *diag.Diagnostics,
) bool {
	stateSpec, planSpec, ok := instanceGroupSpecContentForUpdate(ctx, state, plan, respDiags)
	if !ok {
		return false
	}
	if planSpec.LogDiskSize.ValueInt32() <= stateSpec.LogDiskSize.ValueInt32() &&
		planSpec.DataDiskSize.ValueInt32() <= stateSpec.DataDiskSize.ValueInt32() {
		return true
	}

	if !r.waitInstanceGroupAvailableForManagedUpdate(ctx, plan.Id.ValueString(), "extend volume", respDiags) {
		return false
	}
	extendPlan := instanceGroupExtendVolumeModel{
		InstanceGroupId: plan.Id,
		LogDiskSize:     planSpec.LogDiskSize,
		DataDiskSize:    planSpec.DataDiskSize,
	}
	if !r.extendVolume(ctx, extendPlan, respDiags) {
		return false
	}
	_, found, ok := r.pollExtendVolumeUntilStable(ctx, extendPlan, respDiags)
	if !ok {
		return false
	}
	if !found {
		common.AddGeneralError(ctx, r, respDiags, "instance group was not found after extending volume")
		return false
	}

	return true
}

func (r *instanceGroupResource) updateBackupScheduleIfChanged(
	ctx context.Context,
	state instanceGroupResourceModel,
	plan instanceGroupResourceModel,
	respDiags *diag.Diagnostics,
) bool {
	stateBackupSchedule, planBackupSchedule, ok := instanceGroupBackupScheduleForUpdate(ctx, state, plan, respDiags)
	if !ok {
		return false
	}
	if sameBoolValue(stateBackupSchedule.Enabled, planBackupSchedule.Enabled) &&
		sameOptionalStringValue(stateBackupSchedule.Type, planBackupSchedule.Type) &&
		sameOptionalStringValue(stateBackupSchedule.StartTime, planBackupSchedule.StartTime) &&
		sameOptionalInt32Value(stateBackupSchedule.ExpiryDuration, planBackupSchedule.ExpiryDuration) {
		return true
	}

	if !r.waitInstanceGroupAvailableForManagedUpdate(ctx, plan.Id.ValueString(), "backup schedule update", respDiags) {
		return false
	}
	backupScheduleID := stateBackupSchedule.Id
	if !planBackupSchedule.Id.IsNull() && !planBackupSchedule.Id.IsUnknown() && planBackupSchedule.Id.ValueString() != "" {
		backupScheduleID = planBackupSchedule.Id
	}
	backupPlan := instanceGroupBackupScheduleModel{
		InstanceGroupId:  plan.Id,
		BackupScheduleId: backupScheduleID,
		Type:             planBackupSchedule.Type,
		StartTime:        planBackupSchedule.StartTime,
		ExpiryDuration:   planBackupSchedule.ExpiryDuration,
		Enabled:          planBackupSchedule.Enabled,
	}
	if !r.applyBackupSchedule(ctx, backupPlan, respDiags) {
		return false
	}
	if !r.waitInstanceGroupAvailableForManagedUpdate(ctx, plan.Id.ValueString(), "backup schedule update", respDiags) {
		return false
	}

	return true
}

func (r *instanceGroupResource) updateParameterGroupIfChanged(
	ctx context.Context,
	state instanceGroupResourceModel,
	plan instanceGroupResourceModel,
	respDiags *diag.Diagnostics,
) bool {
	stateParameterGroup, planParameterGroup, ok := instanceGroupParameterGroupForUpdate(ctx, state, plan, respDiags)
	if !ok {
		return false
	}
	if sameStringValue(stateParameterGroup.Id, planParameterGroup.Id) &&
		sameStringValue(stateParameterGroup.Type, planParameterGroup.Type) {
		return true
	}

	if !r.waitInstanceGroupAvailableForManagedUpdate(ctx, plan.Id.ValueString(), "parameter group update", respDiags) {
		return false
	}
	parameterGroupPlan := instanceGroupParameterGroupModel{
		InstanceGroupId:    plan.Id,
		ParameterGroupId:   planParameterGroup.Id,
		ParameterGroupType: planParameterGroup.Type,
	}
	if !r.applyParameterGroup(ctx, parameterGroupPlan, respDiags) {
		return false
	}
	applied, found, ok := r.pollParameterGroupUntilApplied(ctx, parameterGroupPlan, respDiags)
	if !ok {
		return false
	}
	if !found {
		common.AddGeneralError(ctx, r, respDiags, "instance group was not found after applying parameter group")
		return false
	}
	if parameterGroupSyncNeedsRetry(applied) {
		if !r.retryParameterGroupSync(ctx, plan.Id.ValueString(), respDiags) {
			return false
		}
		retried, found, ok := r.pollParameterGroupUntilApplied(ctx, parameterGroupPlan, respDiags)
		if !ok {
			return false
		}
		if !found {
			common.AddGeneralError(ctx, r, respDiags, "instance group was not found after retrying parameter group sync")
			return false
		}
		if parameterGroupSyncNeedsRetry(retried) {
			common.AddGeneralError(ctx, r, respDiags, "parameter group sync is still in ERROR-SYNC status after retry")
			return false
		}
	}

	return true
}

func (r *instanceGroupResource) waitInstanceGroupAvailableForManagedUpdate(
	ctx context.Context,
	instanceGroupId string,
	operation string,
	respDiags *diag.Diagnostics,
) bool {
	tflog.Info(ctx, fmt.Sprintf("waiting for MySQL instance group to become AVAILABLE before %s", operation))
	_, ok := r.pollInstanceGroupUntilStatus(
		ctx,
		instanceGroupId,
		[]string{
			string(mysqlsdk.INSTANCEGROUPSTATUS_AVAILABLE),
			string(mysqlsdk.INSTANCEGROUPSTATUS_PRIMARY_AVAILABLE),
		},
		respDiags,
	)
	return ok
}

func instanceGroupNetworkInfoForUpdate(
	ctx context.Context,
	state instanceGroupResourceModel,
	plan instanceGroupResourceModel,
	respDiags *diag.Diagnostics,
) (instanceGroupResourceDesiredNetworkInfoModel, instanceGroupResourceDesiredNetworkInfoModel, bool) {
	var stateNetworkInfo instanceGroupResourceDesiredNetworkInfoModel
	respDiags.Append(state.DesiredNetworkInfo.As(ctx, &stateNetworkInfo, basetypes.ObjectAsOptions{})...)
	var planNetworkInfo instanceGroupResourceDesiredNetworkInfoModel
	respDiags.Append(plan.DesiredNetworkInfo.As(ctx, &planNetworkInfo, basetypes.ObjectAsOptions{})...)
	return stateNetworkInfo, planNetworkInfo, !respDiags.HasError()
}

func instanceGroupSpecContentForUpdate(
	ctx context.Context,
	state instanceGroupResourceModel,
	plan instanceGroupResourceModel,
	respDiags *diag.Diagnostics,
) (instanceGroupSpecContentResourceModel, instanceGroupSpecContentResourceModel, bool) {
	var stateSpec instanceGroupSpecContentResourceModel
	respDiags.Append(state.SpecContent.As(ctx, &stateSpec, basetypes.ObjectAsOptions{})...)
	var planSpec instanceGroupSpecContentResourceModel
	respDiags.Append(plan.SpecContent.As(ctx, &planSpec, basetypes.ObjectAsOptions{})...)
	return stateSpec, planSpec, !respDiags.HasError()
}

func instanceGroupBackupScheduleForUpdate(
	ctx context.Context,
	state instanceGroupResourceModel,
	plan instanceGroupResourceModel,
	respDiags *diag.Diagnostics,
) (backupScheduleModel, backupScheduleModel, bool) {
	var stateBackupSchedule backupScheduleModel
	respDiags.Append(state.BackupSchedule.As(ctx, &stateBackupSchedule, basetypes.ObjectAsOptions{})...)
	var planBackupSchedule backupScheduleModel
	respDiags.Append(plan.BackupSchedule.As(ctx, &planBackupSchedule, basetypes.ObjectAsOptions{})...)
	return stateBackupSchedule, planBackupSchedule, !respDiags.HasError()
}

func instanceGroupParameterGroupForUpdate(
	ctx context.Context,
	state instanceGroupResourceModel,
	plan instanceGroupResourceModel,
	respDiags *diag.Diagnostics,
) (parameterGroupModel, parameterGroupModel, bool) {
	var stateParameterGroup parameterGroupModel
	respDiags.Append(state.ParameterGroup.As(ctx, &stateParameterGroup, basetypes.ObjectAsOptions{})...)
	var planParameterGroup parameterGroupModel
	respDiags.Append(plan.ParameterGroup.As(ctx, &planParameterGroup, basetypes.ObjectAsOptions{})...)
	return stateParameterGroup, planParameterGroup, !respDiags.HasError()
}

func sameStringValue(a, b types.String) bool {
	if a.IsNull() || a.IsUnknown() || b.IsNull() || b.IsUnknown() {
		return a.IsNull() == b.IsNull() && a.IsUnknown() == b.IsUnknown()
	}
	return a.ValueString() == b.ValueString()
}

func sameOptionalStringValue(a, b types.String) bool {
	if a.IsUnknown() || b.IsUnknown() {
		return true
	}
	return optionalStringValue(a) == optionalStringValue(b)
}

func optionalStringValue(value types.String) string {
	if value.IsNull() {
		return ""
	}
	return value.ValueString()
}

func sameOptionalInt32Value(a, b types.Int32) bool {
	if a.IsUnknown() || b.IsUnknown() {
		return true
	}
	return optionalInt32Value(a) == optionalInt32Value(b)
}

func optionalInt32Value(value types.Int32) int32 {
	if value.IsNull() {
		return 0
	}
	return value.ValueInt32()
}

func sameBoolValue(a, b types.Bool) bool {
	if a.IsNull() || a.IsUnknown() || b.IsNull() || b.IsUnknown() {
		return a.IsNull() == b.IsNull() && a.IsUnknown() == b.IsUnknown()
	}
	return a.ValueBool() == b.ValueBool()
}

func optionalSubnetInfoModels(ctx context.Context, value types.Set, respDiags *diag.Diagnostics) ([]instanceGroupResourceDesiredSubnetInfoModel, bool) {
	if value.IsNull() || value.IsUnknown() {
		return nil, true
	}

	var subnets []instanceGroupResourceDesiredSubnetInfoModel
	diags := value.ElementsAs(ctx, &subnets, false)
	respDiags.Append(diags...)
	if diags.HasError() {
		return nil, false
	}
	return subnets, true
}

func (r *instanceGroupResource) validateInstanceGroupConfig(
	ctx context.Context,
	config instanceGroupResourceModel,
	respDiags *diag.Diagnostics,
) {
	networkInfo, hasNetworkInfo := networkInfoConfigModel(ctx, config.DesiredNetworkInfo, respDiags)
	specContent, hasSpecContent := specContentConfigModel(ctx, config.SpecContent, respDiags)
	backupSchedule, hasBackupSchedule := backupScheduleConfigModel(ctx, config.BackupSchedule, respDiags)
	source, hasSource := restoreSourceConfigModel(ctx, config.Source, respDiags)
	extraInfo, hasExtraInfo := extraInfoConfigModel(ctx, config.ExtraInfo, respDiags)

	if hasNetworkInfo && hasSpecContent {
		r.validateInstanceGroupTopologyConfig(ctx, networkInfo, specContent, respDiags)
		r.validateInstanceGroupSubnetCountConfig(ctx, networkInfo, respDiags)
	}
	if hasSpecContent {
		r.validateInstanceGroupPortConfig(specContent, respDiags)
	}
	if hasBackupSchedule {
		r.validateInstanceGroupBackupScheduleConfig(backupSchedule, respDiags)
	}
	if hasSource {
		r.validateInstanceGroupSourceConfig(source, respDiags)
	}
	if hasSource && hasExtraInfo {
		r.validateRestoreSourceExtraInfoConfig(extraInfo, respDiags)
	}
}

func (r *instanceGroupResource) validateInstanceGroupCreateConfig(
	ctx context.Context,
	config instanceGroupResourceModel,
	respDiags *diag.Diagnostics,
) {
	specContent, hasSpecContent := specContentConfigModel(ctx, config.SpecContent, respDiags)
	if !hasSpecContent {
		return
	}

	if specContent.DatabaseUserPassword.IsNull() {
		respDiags.AddAttributeError(
			path.Root("spec_content").AtName("database_user_password"),
			"Missing database user password",
			"database_user_password must be set when creating a MySQL instance group.",
		)
	}
}

func (r *instanceGroupResource) validateInstanceGroupTopologyConfig(
	ctx context.Context,
	networkInfo instanceGroupResourceDesiredNetworkInfoModel,
	specContent instanceGroupSpecContentResourceModel,
	respDiags *diag.Diagnostics,
) {
	topology, ok := buildInstanceGroupTopologyValidationModel(ctx, networkInfo, respDiags)
	if !ok {
		return
	}

	r.validateInstanceGroupReplicaCount(topology, respDiags)
	r.validateInstanceGroupStandbyPort(topology, specContent, respDiags)
	r.validateInstanceGroupSubnetIDUniqueness(topology, respDiags)
}

func (r *instanceGroupResource) validateInstanceGroupSubnetCountConfig(
	ctx context.Context,
	networkInfo instanceGroupResourceDesiredNetworkInfoModel,
	respDiags *diag.Diagnostics,
) {
	allowedAZCount, ok := r.mysqlAvailabilityZoneCount()
	if !ok || allowedAZCount == 0 {
		return
	}
	subnetCount, ok := desiredSubnetBlockCount(ctx, networkInfo, respDiags)
	if !ok {
		return
	}
	if subnetCount <= allowedAZCount {
		return
	}
	respDiags.AddAttributeError(
		path.Root("desired_network_info"),
		"Too many MySQL subnets",
		fmt.Sprintf("The number of subnet blocks in desired_network_info must not exceed the number of availability zones available for MySQL. Configured %d subnet blocks, but MySQL is available in %d availability zones.", subnetCount, allowedAZCount),
	)
}

func (r *instanceGroupResource) mysqlAvailabilityZoneCount() (int, bool) {
	if r.kc == nil {
		return 0, false
	}
	if azPolicy, ok := r.kc.ServiceAzPolicy[common.ServiceMySQL]; ok {
		return len(azPolicy), true
	}
	return 0, false
}

func desiredSubnetBlockCount(
	ctx context.Context,
	networkInfo instanceGroupResourceDesiredNetworkInfoModel,
	respDiags *diag.Diagnostics,
) (int, bool) {
	primarySubnet, primarySubnetKnown, ok := primarySubnetInfoModel(ctx, networkInfo.PrimarySubnetInfo, respDiags)
	if !ok {
		return 0, false
	}
	if !primarySubnetKnown || !subnetIDKnown(primarySubnet) {
		return 0, false
	}

	standbySubnets, ok := optionalSubnetInfoModels(ctx, networkInfo.StandbySubnetInfo, respDiags)
	if !ok {
		return 0, false
	}
	return 1 + len(standbySubnets), true
}

type instanceGroupTopologyValidationModel struct {
	TotalReplicas    int32
	ReplicasKnown    bool
	HasStandbySubnet bool
	SubnetIDs        []string
}

func buildInstanceGroupTopologyValidationModel(
	ctx context.Context,
	networkInfo instanceGroupResourceDesiredNetworkInfoModel,
	respDiags *diag.Diagnostics,
) (instanceGroupTopologyValidationModel, bool) {
	topology := instanceGroupTopologyValidationModel{ReplicasKnown: true}

	primarySubnet, primarySubnetKnown, ok := primarySubnetInfoModel(ctx, networkInfo.PrimarySubnetInfo, respDiags)
	if !ok {
		return instanceGroupTopologyValidationModel{}, false
	}
	if !primarySubnetKnown {
		topology.ReplicasKnown = false
	} else {
		replicas, replicasKnown := subnetReplicas(primarySubnet)
		topology.ReplicasKnown = topology.ReplicasKnown && replicasKnown
		topology.TotalReplicas += replicas
		if subnetIDKnown(primarySubnet) {
			topology.SubnetIDs = append(topology.SubnetIDs, primarySubnet.SubnetId.ValueString())
		}
	}

	standbySubnets, ok := optionalSubnetInfoModels(ctx, networkInfo.StandbySubnetInfo, respDiags)
	if !ok {
		return instanceGroupTopologyValidationModel{}, false
	}
	topology.HasStandbySubnet = len(standbySubnets) > 0

	for _, subnet := range standbySubnets {
		replicas, known := subnetReplicas(subnet)
		topology.ReplicasKnown = topology.ReplicasKnown && known
		topology.TotalReplicas += replicas
		if subnetIDKnown(subnet) {
			topology.SubnetIDs = append(topology.SubnetIDs, subnet.SubnetId.ValueString())
		}
	}
	return topology, true
}

func primarySubnetInfoModel(
	ctx context.Context,
	value types.Object,
	respDiags *diag.Diagnostics,
) (instanceGroupResourceDesiredSubnetInfoModel, bool, bool) {
	if value.IsNull() || value.IsUnknown() {
		return instanceGroupResourceDesiredSubnetInfoModel{}, false, true
	}

	var primarySubnet instanceGroupResourceDesiredSubnetInfoModel
	diags := value.As(ctx, &primarySubnet, basetypes.ObjectAsOptions{})
	respDiags.Append(diags...)
	if diags.HasError() {
		return instanceGroupResourceDesiredSubnetInfoModel{}, false, false
	}
	return primarySubnet, true, true
}

func subnetReplicas(subnet instanceGroupResourceDesiredSubnetInfoModel) (int32, bool) {
	if subnet.Replicas.IsNull() || subnet.Replicas.IsUnknown() {
		return 0, false
	}
	return subnet.Replicas.ValueInt32(), true
}

func subnetIDKnown(subnet instanceGroupResourceDesiredSubnetInfoModel) bool {
	return !subnet.SubnetId.IsNull() && !subnet.SubnetId.IsUnknown()
}

func (r *instanceGroupResource) validateInstanceGroupReplicaCount(
	topology instanceGroupTopologyValidationModel,
	respDiags *diag.Diagnostics,
) {
	if !topology.ReplicasKnown {
		return
	}
	if topology.TotalReplicas > 6 {
		respDiags.AddAttributeError(
			path.Root("desired_network_info"),
			"Invalid replica count",
			"Total replicas across primary_subnet_info and standby_subnet_info must be between 1 and 6.",
		)
	}
}

func (r *instanceGroupResource) validateInstanceGroupSubnetIDUniqueness(
	topology instanceGroupTopologyValidationModel,
	respDiags *diag.Diagnostics,
) {
	seen := make(map[string]struct{}, len(topology.SubnetIDs))
	for _, subnetID := range topology.SubnetIDs {
		if _, ok := seen[subnetID]; ok {
			respDiags.AddAttributeError(
				path.Root("desired_network_info"),
				"Duplicate subnet ID",
				"subnet_id values across primary_subnet_info and standby_subnet_info must be unique.",
			)
			return
		}
		seen[subnetID] = struct{}{}
	}
}

func (r *instanceGroupResource) validateInstanceGroupStandbyPort(
	topology instanceGroupTopologyValidationModel,
	specContent instanceGroupSpecContentResourceModel,
	respDiags *diag.Diagnostics,
) {
	hasStandbyPort := !specContent.StandbyPort.IsNull() && !specContent.StandbyPort.IsUnknown()
	if !topology.ReplicasKnown {
		if topology.HasStandbySubnet && !hasStandbyPort && !specContent.StandbyPort.IsUnknown() {
			respDiags.AddAttributeError(
				path.Root("spec_content").AtName("standby_port"),
				"Missing standby port",
				"standby_port must be set when the instance group is configured as HA.",
			)
		}
		return
	}

	isHARequest := topology.TotalReplicas > 1
	if specContent.StandbyPort.IsUnknown() {

		if !isHARequest {
			respDiags.AddAttributeError(
				path.Root("spec_content").AtName("standby_port"),
				"Unexpected standby port",
				"standby_port must not be set when the instance group is configured as single.",
			)
		}
		return
	}
	if isHARequest && !hasStandbyPort {
		respDiags.AddAttributeError(
			path.Root("spec_content").AtName("standby_port"),
			"Missing standby port",
			"standby_port must be set when the instance group is configured as HA.",
		)
	}

	if !isHARequest && hasStandbyPort {
		respDiags.AddAttributeError(
			path.Root("spec_content").AtName("standby_port"),
			"Unexpected standby port",
			"standby_port must not be set when the instance group is configured as single.",
		)
	}
}

func (r *instanceGroupResource) validateInstanceGroupPortConfig(
	specContent instanceGroupSpecContentResourceModel,
	respDiags *diag.Diagnostics,
) {
	hasStandbyPort := !specContent.StandbyPort.IsNull() && !specContent.StandbyPort.IsUnknown()
	if !specContent.PrimaryPort.IsNull() && !specContent.PrimaryPort.IsUnknown() && hasStandbyPort && specContent.PrimaryPort.ValueInt32() == specContent.StandbyPort.ValueInt32() {
		respDiags.AddAttributeError(
			path.Root("spec_content").AtName("standby_port"),
			"Invalid standby port",
			"Primary port and standby port must be different.",
		)
	}
}

func (r *instanceGroupResource) validateInstanceGroupBackupScheduleConfig(
	backupSchedule backupScheduleModel,
	respDiags *diag.Diagnostics,
) {
	if backupSchedule.Enabled.IsNull() || backupSchedule.Enabled.IsUnknown() {
		return
	}
	if backupSchedule.Enabled.ValueBool() {

		if !backupSchedule.Type.IsUnknown() && (backupSchedule.Type.IsNull() || backupSchedule.Type.ValueString() == "") {
			respDiags.AddAttributeError(
				path.Root("backup_schedule").AtName("type"),
				"Missing backup schedule type",
				"type must be set when backup_schedule.enabled is true.",
			)
		}
		if !backupSchedule.StartTime.IsUnknown() && (backupSchedule.StartTime.IsNull() || backupSchedule.StartTime.ValueString() == "") {
			respDiags.AddAttributeError(
				path.Root("backup_schedule").AtName("start_time"),
				"Missing backup start time",
				"start_time must be set when backup_schedule.enabled is true.",
			)
		}
		if !backupSchedule.ExpiryDuration.IsUnknown() && backupSchedule.ExpiryDuration.IsNull() {
			respDiags.AddAttributeError(
				path.Root("backup_schedule").AtName("expiry_duration"),
				"Missing backup expiry duration",
				"expiry_duration must be set when backup_schedule.enabled is true.",
			)
		}
		return
	}

	if !backupSchedule.Type.IsNull() && !backupSchedule.Type.IsUnknown() && backupSchedule.Type.ValueString() != "" {
		respDiags.AddAttributeError(
			path.Root("backup_schedule").AtName("type"),
			"Unexpected backup schedule type",
			"type must not be set when backup_schedule.enabled is false.",
		)
	}
	if !backupSchedule.StartTime.IsNull() && !backupSchedule.StartTime.IsUnknown() && backupSchedule.StartTime.ValueString() != "" {
		respDiags.AddAttributeError(
			path.Root("backup_schedule").AtName("start_time"),
			"Unexpected backup start time",
			"start_time must not be set when backup_schedule.enabled is false.",
		)
	}
	if !backupSchedule.ExpiryDuration.IsNull() && !backupSchedule.ExpiryDuration.IsUnknown() {
		respDiags.AddAttributeError(
			path.Root("backup_schedule").AtName("expiry_duration"),
			"Unexpected backup expiry duration",
			"expiry_duration must not be set when backup_schedule.enabled is false.",
		)
	}
}

func (r *instanceGroupResource) validateInstanceGroupSourceConfig(
	source instanceGroupRestoreSourceResourceModel,
	respDiags *diag.Diagnostics,
) {
	if source.Type.IsNull() || source.Type.IsUnknown() {
		return
	}

	switch source.Type.ValueString() {
	case string(mysqlsdk.RESTORESOURCETYPE_BACKUP):
		if !source.Time.IsNull() && !source.Time.IsUnknown() && source.Time.ValueString() != "" {
			respDiags.AddAttributeError(
				path.Root("source").AtName("time"),
				"Unexpected restore time",
				"time must not be set when source.type is BACKUP.",
			)
		}
	case string(mysqlsdk.RESTORESOURCETYPE_INSTANCE_GROUP):
		if source.Time.IsUnknown() {

			return
		}
		if source.Time.IsNull() || source.Time.ValueString() == "" {
			respDiags.AddAttributeError(
				path.Root("source").AtName("time"),
				"Missing restore time",
				"time must be set when source.type is INSTANCE_GROUP.",
			)
			return
		}
		if _, err := time.Parse(time.RFC3339, source.Time.ValueString()); err != nil {
			respDiags.AddAttributeError(
				path.Root("source").AtName("time"),
				"Invalid restore time",
				"time must be in RFC3339 format. The provider normalizes it to a 3-digit milliseconds timestamp before sending the request.",
			)
		}
	}
}

func networkInfoConfigModel(ctx context.Context, value types.Object, respDiags *diag.Diagnostics) (instanceGroupResourceDesiredNetworkInfoModel, bool) {
	if value.IsNull() || value.IsUnknown() {
		return instanceGroupResourceDesiredNetworkInfoModel{}, false
	}
	var model instanceGroupResourceDesiredNetworkInfoModel
	diags := value.As(ctx, &model, basetypes.ObjectAsOptions{})
	respDiags.Append(diags...)
	return model, !diags.HasError()
}

func specContentConfigModel(ctx context.Context, value types.Object, respDiags *diag.Diagnostics) (instanceGroupSpecContentResourceModel, bool) {
	if value.IsNull() || value.IsUnknown() {
		return instanceGroupSpecContentResourceModel{}, false
	}
	var model instanceGroupSpecContentResourceModel
	diags := value.As(ctx, &model, basetypes.ObjectAsOptions{})
	respDiags.Append(diags...)
	return model, !diags.HasError()
}

func backupScheduleConfigModel(ctx context.Context, value types.Object, respDiags *diag.Diagnostics) (backupScheduleModel, bool) {
	if value.IsNull() || value.IsUnknown() {
		return backupScheduleModel{}, false
	}
	var model backupScheduleModel
	diags := value.As(ctx, &model, basetypes.ObjectAsOptions{})
	respDiags.Append(diags...)
	return model, !diags.HasError()
}

func restoreSourceConfigModel(ctx context.Context, value types.Object, respDiags *diag.Diagnostics) (instanceGroupRestoreSourceResourceModel, bool) {
	if value.IsNull() || value.IsUnknown() {
		return instanceGroupRestoreSourceResourceModel{}, false
	}
	var model instanceGroupRestoreSourceResourceModel
	diags := value.As(ctx, &model, basetypes.ObjectAsOptions{})
	respDiags.Append(diags...)
	return model, !diags.HasError()
}

func extraInfoConfigModel(ctx context.Context, value types.Object, respDiags *diag.Diagnostics) (extraInfoModel, bool) {
	if value.IsNull() || value.IsUnknown() {
		return extraInfoModel{}, false
	}
	var model extraInfoModel
	diags := value.As(ctx, &model, basetypes.ObjectAsOptions{})
	respDiags.Append(diags...)
	return model, !diags.HasError()
}

func (r *instanceGroupResource) validateRestoreSourceExtraInfoConfig(
	extraInfo extraInfoModel,
	respDiags *diag.Diagnostics,
) {
	if extraInfo.UseCaseSensitiveTableNames.IsNull() || extraInfo.UseCaseSensitiveTableNames.IsUnknown() {
		return
	}
	respDiags.AddAttributeError(
		path.Root("extra_info").AtName("use_case_sensitive_table_names"),
		"Unexpected restore source extra info",
		"use_case_sensitive_table_names must not be set when source is set. The restore source value is used by the MySQL API.",
	)
}

func (r *instanceGroupResource) buildCreateRequest(
	ctx context.Context,
	plan instanceGroupResourceModel,
	config instanceGroupResourceModel,
	respDiags *diag.Diagnostics,
) (mysqlsdk.BodyCreateMysqlInstanceGroup, bool) {
	var networkInfo instanceGroupResourceDesiredNetworkInfoModel
	respDiags.Append(plan.DesiredNetworkInfo.As(ctx, &networkInfo, basetypes.ObjectAsOptions{})...)
	if respDiags.HasError() {
		return mysqlsdk.BodyCreateMysqlInstanceGroup{}, false
	}

	var specContent instanceGroupSpecContentResourceModel
	respDiags.Append(plan.SpecContent.As(ctx, &specContent, basetypes.ObjectAsOptions{})...)
	if respDiags.HasError() {
		return mysqlsdk.BodyCreateMysqlInstanceGroup{}, false
	}
	var configSpecContent instanceGroupSpecContentResourceModel
	respDiags.Append(config.SpecContent.As(ctx, &configSpecContent, basetypes.ObjectAsOptions{})...)
	if respDiags.HasError() {
		return mysqlsdk.BodyCreateMysqlInstanceGroup{}, false
	}

	var backupSchedule backupScheduleModel
	respDiags.Append(plan.BackupSchedule.As(ctx, &backupSchedule, basetypes.ObjectAsOptions{})...)
	if respDiags.HasError() {
		return mysqlsdk.BodyCreateMysqlInstanceGroup{}, false
	}

	var parameterGroup parameterGroupModel
	respDiags.Append(plan.ParameterGroup.As(ctx, &parameterGroup, basetypes.ObjectAsOptions{})...)
	if respDiags.HasError() {
		return mysqlsdk.BodyCreateMysqlInstanceGroup{}, false
	}

	configExtraInfo, hasConfigExtraInfo := extraInfoConfigModel(ctx, config.ExtraInfo, respDiags)
	if respDiags.HasError() {
		return mysqlsdk.BodyCreateMysqlInstanceGroup{}, false
	}

	var source instanceGroupRestoreSourceResourceModel
	hasSource := !plan.Source.IsNull() && !plan.Source.IsUnknown()
	if hasSource {
		respDiags.Append(plan.Source.As(ctx, &source, basetypes.ObjectAsOptions{})...)
		if respDiags.HasError() {
			return mysqlsdk.BodyCreateMysqlInstanceGroup{}, false
		}
	}

	var securityGroupIds []string
	respDiags.Append(networkInfo.SecurityGroupIds.ElementsAs(ctx, &securityGroupIds, false)...)
	if respDiags.HasError() {
		return mysqlsdk.BodyCreateMysqlInstanceGroup{}, false
	}

	var primarySubnet instanceGroupResourceDesiredSubnetInfoModel
	respDiags.Append(networkInfo.PrimarySubnetInfo.As(ctx, &primarySubnet, basetypes.ObjectAsOptions{})...)
	if respDiags.HasError() {
		return mysqlsdk.BodyCreateMysqlInstanceGroup{}, false
	}

	standbySubnets, ok := optionalSubnetInfoModels(ctx, networkInfo.StandbySubnetInfo, respDiags)
	if !ok {
		return mysqlsdk.BodyCreateMysqlInstanceGroup{}, false
	}

	primarySubnetReq := *mysqlsdk.NewMysqlV1ApiCreateMysqlInstanceGroupModelSubnetInfoRequestModel(
		primarySubnet.Replicas.ValueInt32(),
		primarySubnet.SubnetId.ValueString(),
	)

	standbySubnetReqs := make([]mysqlsdk.MysqlV1ApiCreateMysqlInstanceGroupModelSubnetInfoRequestModel, 0, len(standbySubnets))
	for _, subnet := range standbySubnets {
		standbySubnetReqs = append(standbySubnetReqs, *mysqlsdk.NewMysqlV1ApiCreateMysqlInstanceGroupModelSubnetInfoRequestModel(
			subnet.Replicas.ValueInt32(),
			subnet.SubnetId.ValueString(),
		))
	}

	networkReq := mysqlsdk.NewNetworkInfoRequestModel(
		securityGroupIds,
		primarySubnetReq,
	)
	if len(standbySubnetReqs) > 0 {
		networkReq.SetStandbySubnetInfo(standbySubnetReqs)
	}

	specReq := mysqlsdk.SpecContentRequestModel{
		DatabaseUserName:     specContent.DatabaseUserName.ValueString(),
		DatabaseUserPassword: configSpecContent.DatabaseUserPassword.ValueString(),
		EngineVersion:        specContent.EngineVersion.ValueString(),
		FlavorId:             specContent.FlavorId.ValueString(),
		LogDiskSize:          specContent.LogDiskSize.ValueInt32(),
		DataDiskSize:         specContent.DataDiskSize.ValueInt32(),
	}
	specReq.SetPrimaryPort(specContent.PrimaryPort.ValueInt32())
	if !specContent.StandbyPort.IsNull() && !specContent.StandbyPort.IsUnknown() {
		specReq.SetStandbyPort(specContent.StandbyPort.ValueInt32())
	}

	backupReq := mysqlsdk.NewMysqlV1ApiCreateMysqlInstanceGroupModelBackupScheduleRequestModel(
		backupSchedule.Enabled.ValueBool(),
	)
	if !backupSchedule.Type.IsNull() && !backupSchedule.Type.IsUnknown() && backupSchedule.Type.ValueString() != "" {
		backupReq.SetType(mysqlsdk.BackupScheduleType(backupSchedule.Type.ValueString()))
	}
	if !backupSchedule.StartTime.IsNull() && !backupSchedule.StartTime.IsUnknown() && backupSchedule.StartTime.ValueString() != "" {
		backupReq.SetStartTime(backupSchedule.StartTime.ValueString())
	}
	if !backupSchedule.ExpiryDuration.IsNull() && !backupSchedule.ExpiryDuration.IsUnknown() {
		backupReq.SetExpiryDuration(backupSchedule.ExpiryDuration.ValueInt32())
	}

	parameterReq := *mysqlsdk.NewMysqlV1ApiCreateMysqlInstanceGroupModelParameterGroupRequestModel(
		mysqlsdk.ParameterGroupType(parameterGroup.Type.ValueString()),
		parameterGroup.Id.ValueString(),
	)

	instanceGroupReq := mysqlsdk.NewMysqlV1ApiCreateMysqlInstanceGroupModelInstanceGroupRequestModel(
		plan.Name.ValueString(),
		*networkReq,
		specReq,
		*backupReq,
		parameterReq,
	)
	if !hasSource && hasConfigExtraInfo && !configExtraInfo.UseCaseSensitiveTableNames.IsNull() && !configExtraInfo.UseCaseSensitiveTableNames.IsUnknown() {
		extraReq := mysqlsdk.NewExtraInfoRequestModel()
		extraReq.SetUseCaseSensitiveTableNames(configExtraInfo.UseCaseSensitiveTableNames.ValueBool())
		instanceGroupReq.SetExtraInfo(*extraReq)
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() && plan.Description.ValueString() != "" {
		instanceGroupReq.SetDescription(plan.Description.ValueString())
	}
	if hasSource {
		sourceReq := mysqlsdk.NewRestoreSourceRequestModelWithDefaults()
		sourceReq.SetType(mysqlsdk.RestoreSourceType(source.Type.ValueString()))
		sourceReq.SetId(source.Id.ValueString())
		if !source.Time.IsNull() && !source.Time.IsUnknown() && source.Time.ValueString() != "" {
			restoreTime, err := time.Parse(time.RFC3339, source.Time.ValueString())
			if err != nil {
				respDiags.AddAttributeError(
					path.Root("source").AtName("time"),
					"Invalid restore time",
					"time must be in RFC3339 format. The provider normalizes it to a 3-digit milliseconds timestamp before sending the request.",
				)
				return mysqlsdk.BodyCreateMysqlInstanceGroup{}, false
			}
			sourceReq.SetTime(time.Date(
				restoreTime.Year(),
				restoreTime.Month(),
				restoreTime.Day(),
				restoreTime.Hour(),
				restoreTime.Minute(),
				restoreTime.Second(),
				(restoreTime.Nanosecond()/int(time.Millisecond))*int(time.Millisecond),
				restoreTime.Location(),
			).UTC().Format("2006-01-02T15:04:05.000Z"))
		}
		instanceGroupReq.SetSource(*sourceReq)
	}

	return *mysqlsdk.NewBodyCreateMysqlInstanceGroup(*instanceGroupReq), true
}

func (r *instanceGroupResource) pollInstanceGroupUntilStatus(
	ctx context.Context,
	id string,
	targetStatuses []string,
	respDiags *diag.Diagnostics,
) (*mysqlsdk.MysqlV1ApiGetMysqlInstanceGroupModelInstanceGroupResponseModel, bool) {

	return r.pollInstanceGroupUntilReady(ctx, id, respDiags, func(resp mysqlsdk.MysqlV1ApiGetMysqlInstanceGroupModelInstanceGroupResponseModel) bool {
		status := string(resp.Status)
		for _, targetStatus := range targetStatuses {
			if status == targetStatus {
				return true
			}
		}
		return false
	})
}

func expectedStandbyInstanceCount(
	ctx context.Context,
	desiredNetworkInfo types.Object,
	respDiags *diag.Diagnostics,
) (int, bool) {
	if desiredNetworkInfo.IsNull() || desiredNetworkInfo.IsUnknown() {
		return 0, true
	}

	var networkInfo instanceGroupResourceDesiredNetworkInfoModel
	respDiags.Append(desiredNetworkInfo.As(ctx, &networkInfo, basetypes.ObjectAsOptions{})...)
	if respDiags.HasError() {
		return 0, false
	}

	replicas, ok := topologyReplicaCounts(ctx, networkInfo, respDiags)
	if !ok {
		return 0, false
	}

	totalReplicas := totalReplicaCount(replicas)
	if totalReplicas <= 1 {
		return 0, true
	}
	return int(totalReplicas - 1), true
}

func (r *instanceGroupResource) pollInstanceGroupUntilTopologyReady(
	ctx context.Context,
	id string,
	expectedStandbyCount int,
	respDiags *diag.Diagnostics,
) (*mysqlsdk.MysqlV1ApiGetMysqlInstanceGroupModelInstanceGroupResponseModel, bool) {

	return r.pollInstanceGroupUntilReady(ctx, id, respDiags, func(resp mysqlsdk.MysqlV1ApiGetMysqlInstanceGroupModelInstanceGroupResponseModel) bool {
		if expectedStandbyCount > 0 && r.failIfStandbyInstancesInTerminalStatus(ctx, id, respDiags) {
			return false
		}
		if !instanceGroupTopologyReady(resp, expectedStandbyCount) {
			return false
		}
		return true
	})
}

func instanceGroupTopologyReady(resp mysqlsdk.MysqlV1ApiGetMysqlInstanceGroupModelInstanceGroupResponseModel, expectedStandbyCount int) bool {
	instances, ok := resp.GetInstancesOk()
	if !ok || instances == nil {
		return false
	}
	primary, ok := instances.GetPrimaryOk()
	if !ok || primary == nil {
		return false
	}
	instanceID, ok := utils.GetNullableStringValue(primary.InstanceId)
	if !ok || instanceID == "" {
		return false
	}
	if expectedStandbyCount == 0 {
		return true
	}

	readyStandbyCount := 0
	for _, standby := range instances.Standby {
		instanceID, ok := utils.GetNullableStringValue(standby.InstanceId)
		if ok && instanceID != "" {
			readyStandbyCount++
		}
	}
	return readyStandbyCount >= expectedStandbyCount
}

func (r *instanceGroupResource) pollInstanceGroupUntilReady(
	ctx context.Context,
	id string,
	respDiags *diag.Diagnostics,
	ready func(mysqlsdk.MysqlV1ApiGetMysqlInstanceGroupModelInstanceGroupResponseModel) bool,
) (*mysqlsdk.MysqlV1ApiGetMysqlInstanceGroupModelInstanceGroupResponseModel, bool) {
	ticker := time.NewTicker(mysqlPollInterval)
	defer ticker.Stop()

	start := time.Now()

	for {
		resp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, respDiags,
			func() (*mysqlsdk.GetMySQLInstanceGroupResponseModel, *http.Response, error) {
				return r.kc.ApiClient.MySQLInstanceGroupsAPI.
					GetMysqlInstanceGroup(ctx, id).
					XAuthToken(r.kc.XAuthToken).
					Execute()
			},
		)
		if err != nil {
			if isTransientInstanceGroupPollingError(err, httpResp) {
				tflog.Warn(ctx, "MySQL instance group polling got a transient backend error. Retrying until timeout...")
			} else {
				common.AddApiActionError(ctx, r, httpResp, "PollForCondition", err, respDiags)
				return nil, false
			}
		} else if ready(resp.InstanceGroup) {
			return &resp.InstanceGroup, true
		} else if respDiags.HasError() {
			return nil, false
		} else if r.failIfAllInstancesInError(ctx, id, respDiags) {
			return nil, false
		} else {
			elapsed := time.Since(start).Round(time.Second)
			tflog.Info(ctx, fmt.Sprintf("%s... [%s elapsed]", resp.InstanceGroup.Status, elapsed))
		}

		select {
		case <-ctx.Done():
			common.AddGeneralError(ctx, r, respDiags, "context deadline exceeded")
			return nil, false
		case <-ticker.C:
		}
	}
}

func (r *instanceGroupResource) failIfAllInstancesInError(
	ctx context.Context,
	id string,
	respDiags *diag.Diagnostics,
) bool {
	instancesResp, _, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, respDiags,
		func() (*mysqlsdk.GetMySQLInstanceGroupInstancesResponseModel, *http.Response, error) {
			return r.kc.ApiClient.MySQLInstanceGroupsAPI.
				ListMysqlInstances(ctx, id).
				XAuthToken(r.kc.XAuthToken).
				Execute()
		},
	)
	if err != nil {
		tflog.Warn(ctx, "failed to check MySQL instance statuses while polling instance group. Continuing polling...")
		return false
	}
	if !allMySQLInstancesInError(instancesResp.Instances) {
		return false
	}

	common.AddGeneralError(ctx, r, respDiags, fmt.Sprintf(
		"all MySQL instances are in error status while waiting for instance group %s: %s",
		id,
		strings.Join(formatMySQLInstanceStatuses(instancesResp.Instances), ", "),
	))
	return true
}

func (r *instanceGroupResource) failIfStandbyInstancesInTerminalStatus(
	ctx context.Context,
	id string,
	respDiags *diag.Diagnostics,
) bool {
	instancesResp, _, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, respDiags,
		func() (*mysqlsdk.GetMySQLInstanceGroupInstancesResponseModel, *http.Response, error) {
			return r.kc.ApiClient.MySQLInstanceGroupsAPI.
				ListMysqlInstances(ctx, id).
				XAuthToken(r.kc.XAuthToken).
				Execute()
		},
	)
	if err != nil {
		tflog.Warn(ctx, "failed to check MySQL standby instance statuses while polling instance group. Continuing polling...")
		return false
	}

	failedStandbys := terminalMySQLStandbyInstances(instancesResp.Instances)
	if len(failedStandbys) == 0 {
		return false
	}

	common.AddGeneralError(ctx, r, respDiags, fmt.Sprintf(
		"MySQL standby instances entered terminal status while waiting for instance group %s: %s",
		id,
		strings.Join(formatMySQLInstanceStatuses(failedStandbys), ", "),
	))
	return true
}

func terminalMySQLStandbyInstances(instances []mysqlsdk.InstanceResponseModel) []mysqlsdk.InstanceResponseModel {
	failed := make([]mysqlsdk.InstanceResponseModel, 0)
	for _, instance := range instances {
		if instance.Role != mysqlsdk.INSTANCEROLE_STANDBY {
			continue
		}
		if isTerminalMySQLInstanceStatus(instance.Status) {
			failed = append(failed, instance)
		}
	}
	return failed
}

func isTerminalMySQLInstanceStatus(status mysqlsdk.InstanceStatus) bool {
	switch status {
	case mysqlsdk.INSTANCESTATUS_ERROR,
		mysqlsdk.INSTANCESTATUS_STORAGE_FULL_ERROR,
		mysqlsdk.INSTANCESTATUS_TERMINATED:
		return true
	default:
		return false
	}
}

func allMySQLInstancesInError(instances []mysqlsdk.InstanceResponseModel) bool {
	if len(instances) == 0 {
		return false
	}

	for _, instance := range instances {
		switch instance.Status {
		case mysqlsdk.INSTANCESTATUS_ERROR, mysqlsdk.INSTANCESTATUS_STORAGE_FULL_ERROR:
		default:
			return false
		}
	}
	return true
}

func formatMySQLInstanceStatuses(instances []mysqlsdk.InstanceResponseModel) []string {
	statuses := make([]string, 0, len(instances))
	for _, instance := range instances {
		statuses = append(statuses, fmt.Sprintf("%s=%s", instance.Id, instance.Status))
	}
	return statuses
}

func stringPtr(v string) *string {
	return &v
}

func (r *instanceGroupResource) pollInstanceGroupUntilDeleted(
	ctx context.Context,
	id string,
	respDiags *diag.Diagnostics,
) bool {
	ticker := time.NewTicker(mysqlPollInterval)
	defer ticker.Stop()

	for {
		_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, respDiags,
			func() (*mysqlsdk.GetMySQLInstanceGroupResponseModel, *http.Response, error) {
				return r.kc.ApiClient.MySQLInstanceGroupsAPI.
					GetMysqlInstanceGroup(ctx, id).
					XAuthToken(r.kc.XAuthToken).
					Execute()
			},
		)
		if err != nil {
			logDeletionPollFailure(ctx, httpResp, err)

			if commonErrIsTransient(err) {
				tflog.Warn(ctx, fmt.Sprintf("Transient polling error (%v). Retrying...", err))
			} else if isTransientDeletePollingDecodeError(err) {
				tflog.Warn(ctx, "MySQL instance group deletion polling got a transient decode error. Retrying until timeout...")
			} else if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
				return true
			} else if httpResp != nil && httpResp.StatusCode == http.StatusInternalServerError {
				tflog.Warn(ctx, "MySQL instance group deletion polling got a transient 500 error. Retrying until timeout...")
			} else {
				common.AddApiActionError(ctx, r, httpResp, "PollForDeletion", err, respDiags)
				return false
			}
		}

		select {
		case <-ctx.Done():
			common.AddGeneralError(ctx, r, respDiags, "Context cancelled while waiting for deletion")
			return false
		case <-ticker.C:
		}
	}
}

func logDeletionPollFailure(ctx context.Context, httpResp *http.Response, err error) {
	fields := map[string]interface{}{
		"error": err.Error(),
	}

	if httpResp != nil {
		fields["status_code"] = httpResp.StatusCode
		fields["status"] = httpResp.Status
		fields["content_length"] = httpResp.ContentLength

		if contentType := httpResp.Header.Get("Content-Type"); contentType != "" {
			fields["content_type"] = contentType
		}

		if httpResp.Request != nil && httpResp.Request.URL != nil {
			fields["request_url"] = httpResp.Request.URL.String()
			fields["request_method"] = httpResp.Request.Method
		}
	}

	tflog.Warn(ctx, "MySQL instance group deletion polling failed", fields)
}

func commonErrIsTransient(err error) bool {
	if err == nil {
		return false
	}

	msg := err.Error()
	return strings.Contains(msg, "GOAWAY") ||
		strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "broken pipe") ||
		strings.Contains(msg, "EOF")
}

func isTransientDeletePollingDecodeError(err error) bool {
	if err == nil {
		return false
	}

	msg := err.Error()
	return strings.Contains(msg, "no value given for required property id") ||
		strings.Contains(msg, "no value given for required property instance_group")
}

func isTransientInstanceGroupPollingError(err error, resp *http.Response) bool {
	if resp != nil && resp.StatusCode == http.StatusInternalServerError {
		return true
	}

	return resp != nil && resp.StatusCode == http.StatusBadRequest && strings.Contains(err.Error(), "haStatus")
}

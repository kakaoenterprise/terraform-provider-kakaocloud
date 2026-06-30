// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package mysql

import (
	"cmp"
	"context"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"terraform-provider-kakaocloud/internal/common"
	"terraform-provider-kakaocloud/internal/utils"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	mysqlsdk "github.com/kakaoenterprise/kc-sdk-go/services/mysql"
)

type instanceGroupScaleOutSubnetInfo struct {
	Replicas int32
	SubnetId string
}

type scaleInTargetInstance struct {
	ID                 string
	SubnetID           string
	Status             mysqlsdk.InstanceStatus
	AvailabilityStatus string
	CreatedAt          string
}

const (
	mysqlAvailabilityStatusWarningSync = "WarningSync"
	mysqlAvailabilityStatusDelaying    = "Delaying"
	mysqlAvailabilityStatusAvailable   = "Available"
)

func (r *instanceGroupResource) updateTopologyIfChanged(
	ctx context.Context,
	state instanceGroupResourceModel,
	plan instanceGroupResourceModel,
	respDiags *diag.Diagnostics,
) bool {
	stateNetworkInfo, planNetworkInfo, ok := instanceGroupNetworkInfoForUpdate(ctx, state, plan, respDiags)
	if !ok {
		return false
	}
	_, planSpec, ok := instanceGroupSpecContentForUpdate(ctx, state, plan, respDiags)
	if !ok {
		return false
	}

	stateReplicas, ok := topologyReplicaCounts(ctx, stateNetworkInfo, respDiags)
	if !ok {
		return false
	}
	planReplicas, ok := topologyReplicaCounts(ctx, planNetworkInfo, respDiags)
	if !ok {
		return false
	}

	if !topologyReplicaCountsChanged(stateReplicas, planReplicas) {
		return true
	}

	currentReplicas, ok := r.waitCurrentTopologyReplicaCounts(ctx, plan.Id.ValueString(), respDiags)
	if !ok {
		return false
	}

	scaleInCounts, scaleOutSubnets := topologyReplicaDiffForUpdate(stateReplicas, currentReplicas, planReplicas)
	if len(scaleInCounts) == 0 && len(scaleOutSubnets) == 0 {
		return true
	}

	if len(scaleInCounts) > 0 {
		instanceIDs, ok := r.selectScaleInTargetInstanceIDs(ctx, plan.Id.ValueString(), scaleInCounts, respDiags)
		if !ok {
			return false
		}
		if len(instanceIDs) > 0 {
			if !r.scaleIn(ctx, plan.Id.ValueString(), instanceIDs, respDiags) {
				return false
			}
			if !r.pollScaleInUntilApplied(ctx, plan.Id.ValueString(), instanceIDs, respDiags) {
				return false
			}
		}
	}

	replicasAfterScaleIn := topologyReplicaCountsAfterScaleIn(currentReplicas, scaleInCounts)
	if len(scaleOutSubnets) > 0 {
		if !r.scaleOut(ctx, plan.Id.ValueString(), scaleOutStandbyPort(replicasAfterScaleIn, planSpec.StandbyPort), scaleOutSubnets, respDiags) {
			return false
		}
		if !r.waitInstanceGroupAvailableAfterTopologyUpdate(ctx, plan.Id.ValueString(), respDiags) {
			return false
		}
	}

	return true
}

func (r *instanceGroupResource) waitCurrentTopologyReplicaCounts(
	ctx context.Context,
	instanceGroupID string,
	respDiags *diag.Diagnostics,
) (map[string]int32, bool) {
	result, ok := r.pollInstanceGroupUntilStatus(
		ctx,
		instanceGroupID,
		[]string{
			string(mysqlsdk.INSTANCEGROUPSTATUS_AVAILABLE),
			string(mysqlsdk.INSTANCEGROUPSTATUS_PRIMARY_AVAILABLE),
		},
		respDiags,
	)
	if !ok {
		return nil, false
	}

	return topologyReplicaCountsFromRemoteNetworkInfo(result.NetworkInfo), true
}

func scaleOutStandbyPort(stateReplicas map[string]int32, planStandbyPort types.Int32) types.Int32 {

	if totalReplicaCount(stateReplicas) > 1 {
		return types.Int32Null()
	}
	return planStandbyPort
}

func totalReplicaCount(replicasBySubnetID map[string]int32) int32 {
	var total int32
	for _, replicas := range replicasBySubnetID {
		total += replicas
	}
	return total
}

func primarySubnetIDChanged(
	ctx context.Context,
	stateNetworkInfo instanceGroupResourceDesiredNetworkInfoModel,
	planNetworkInfo instanceGroupResourceDesiredNetworkInfoModel,
	respDiags *diag.Diagnostics,
) bool {
	stateSubnetID, ok := primarySubnetID(ctx, stateNetworkInfo.PrimarySubnetInfo, respDiags)
	if !ok {
		return false
	}
	planSubnetID, ok := primarySubnetID(ctx, planNetworkInfo.PrimarySubnetInfo, respDiags)
	if !ok {
		return false
	}
	return !sameStringValue(stateSubnetID, planSubnetID)
}

func primarySubnetID(
	ctx context.Context,
	value types.Object,
	respDiags *diag.Diagnostics,
) (types.String, bool) {
	if value.IsNull() || value.IsUnknown() {
		return types.StringNull(), true
	}

	var subnet instanceGroupResourceDesiredSubnetInfoModel
	respDiags.Append(value.As(ctx, &subnet, basetypes.ObjectAsOptions{})...)
	if respDiags.HasError() {
		return types.StringNull(), false
	}
	return subnet.SubnetId, true
}

func standbySubnetIDsReplaced(
	ctx context.Context,
	stateNetworkInfo instanceGroupResourceDesiredNetworkInfoModel,
	planNetworkInfo instanceGroupResourceDesiredNetworkInfoModel,
	respDiags *diag.Diagnostics,
) bool {
	stateSubnetIDs, ok := standbySubnetIDSet(ctx, stateNetworkInfo.StandbySubnetInfo, respDiags)
	if !ok {
		return false
	}
	planSubnetIDs, ok := standbySubnetIDSet(ctx, planNetworkInfo.StandbySubnetInfo, respDiags)
	if !ok {
		return false
	}

	hasRemoved := false
	for subnetID := range stateSubnetIDs {
		if _, ok := planSubnetIDs[subnetID]; !ok {
			hasRemoved = true
			break
		}
	}
	hasAdded := false
	for subnetID := range planSubnetIDs {
		if _, ok := stateSubnetIDs[subnetID]; !ok {
			hasAdded = true
			break
		}
	}
	return hasRemoved && hasAdded
}

func standbySubnetIDSet(
	ctx context.Context,
	value types.Set,
	respDiags *diag.Diagnostics,
) (map[string]struct{}, bool) {
	subnets, ok := standbySubnetInfoModels(ctx, value, respDiags)
	if !ok {
		return nil, false
	}
	subnetIDs := make(map[string]struct{}, len(subnets))
	for _, subnet := range subnets {
		if subnet.SubnetId.IsNull() || subnet.SubnetId.IsUnknown() {
			return nil, false
		}
		subnetIDs[subnet.SubnetId.ValueString()] = struct{}{}
	}
	return subnetIDs, true
}

func topologyReplicaCounts(
	ctx context.Context,
	networkInfo instanceGroupResourceDesiredNetworkInfoModel,
	respDiags *diag.Diagnostics,
) (map[string]int32, bool) {
	replicasBySubnetID := map[string]int32{}

	var primarySubnet instanceGroupResourceDesiredSubnetInfoModel
	if !networkInfo.PrimarySubnetInfo.IsNull() && !networkInfo.PrimarySubnetInfo.IsUnknown() {
		respDiags.Append(networkInfo.PrimarySubnetInfo.As(ctx, &primarySubnet, basetypes.ObjectAsOptions{})...)
		if respDiags.HasError() {
			return nil, false
		}
		addSubnetReplicaCount(replicasBySubnetID, primarySubnet)
	}

	standbySubnets, ok := standbySubnetInfoModels(ctx, networkInfo.StandbySubnetInfo, respDiags)
	if !ok {
		return nil, false
	}
	for _, subnet := range standbySubnets {
		addSubnetReplicaCount(replicasBySubnetID, subnet)
	}
	return replicasBySubnetID, true
}

func topologyReplicaCountsFromRemoteNetworkInfo(networkInfo mysqlsdk.NullableNetworkInfoResponseModel) map[string]int32 {
	replicasBySubnetID := map[string]int32{}
	if !networkInfo.IsSet() || networkInfo.Get() == nil {
		return replicasBySubnetID
	}

	networkInfoModel := networkInfo.Get()
	if primarySubnet, ok := networkInfoModel.GetPrimarySubnetInfoOk(); ok && primarySubnet != nil {
		addRemoteSubnetReplicaCount(replicasBySubnetID, *primarySubnet)
	}
	for _, subnet := range networkInfoModel.StandbySubnetInfo {
		addRemoteSubnetReplicaCount(replicasBySubnetID, subnet)
	}
	return replicasBySubnetID
}

func addRemoteSubnetReplicaCount(replicasBySubnetID map[string]int32, subnet mysqlsdk.SubnetInfoResponseModel) {
	if subnet.SubnetId == "" {
		return
	}
	replicasBySubnetID[subnet.SubnetId] += subnet.Replicas
}

func standbySubnetInfoModels(
	ctx context.Context,
	value types.Set,
	respDiags *diag.Diagnostics,
) ([]instanceGroupResourceDesiredSubnetInfoModel, bool) {
	if value.IsNull() || value.IsUnknown() {
		return nil, true
	}
	var subnets []instanceGroupResourceDesiredSubnetInfoModel
	respDiags.Append(value.ElementsAs(ctx, &subnets, false)...)
	if respDiags.HasError() {
		return nil, false
	}
	return subnets, true
}

func addSubnetReplicaCount(replicasBySubnetID map[string]int32, subnet instanceGroupResourceDesiredSubnetInfoModel) {
	if subnet.SubnetId.IsNull() || subnet.SubnetId.IsUnknown() || subnet.Replicas.IsNull() || subnet.Replicas.IsUnknown() {
		return
	}
	replicasBySubnetID[subnet.SubnetId.ValueString()] += subnet.Replicas.ValueInt32()
}

func topologyReplicaCountsChanged(stateReplicas map[string]int32, planReplicas map[string]int32) bool {
	scaleInCounts, scaleOutSubnets := subnetReplicaDiff(stateReplicas, planReplicas)
	return len(scaleInCounts) > 0 || len(scaleOutSubnets) > 0
}

func topologyReplicaDiffForUpdate(
	stateReplicas map[string]int32,
	currentReplicas map[string]int32,
	planReplicas map[string]int32,
) (map[string]int32, []instanceGroupScaleOutSubnetInfo) {
	if !topologyReplicaCountsChanged(stateReplicas, planReplicas) {
		return nil, nil
	}
	return subnetReplicaDiff(currentReplicas, planReplicas)
}

func subnetReplicaDiff(
	stateReplicas map[string]int32,
	planReplicas map[string]int32,
) (map[string]int32, []instanceGroupScaleOutSubnetInfo) {
	subnetIDs := make([]string, 0, len(stateReplicas)+len(planReplicas))
	seen := map[string]struct{}{}
	for subnetID := range stateReplicas {
		subnetIDs = append(subnetIDs, subnetID)
		seen[subnetID] = struct{}{}
	}
	for subnetID := range planReplicas {
		if _, ok := seen[subnetID]; !ok {
			subnetIDs = append(subnetIDs, subnetID)
		}
	}
	slices.Sort(subnetIDs)

	scaleInCounts := map[string]int32{}
	scaleOutSubnets := []instanceGroupScaleOutSubnetInfo{}
	for _, subnetID := range subnetIDs {
		current := stateReplicas[subnetID]
		desired := planReplicas[subnetID]
		switch {
		case desired > current:
			scaleOutSubnets = append(scaleOutSubnets, instanceGroupScaleOutSubnetInfo{
				Replicas: desired - current,
				SubnetId: subnetID,
			})
		case current > desired:
			scaleInCounts[subnetID] = current - desired
		}
	}
	return scaleInCounts, scaleOutSubnets
}

func topologyReplicaCountsAfterScaleIn(
	currentReplicas map[string]int32,
	scaleInCounts map[string]int32,
) map[string]int32 {
	afterScaleIn := make(map[string]int32, len(currentReplicas))
	for subnetID, replicas := range currentReplicas {
		afterScaleIn[subnetID] = replicas
	}
	for subnetID, scaleInCount := range scaleInCounts {
		remaining := afterScaleIn[subnetID] - scaleInCount
		if remaining <= 0 {
			delete(afterScaleIn, subnetID)
			continue
		}
		afterScaleIn[subnetID] = remaining
	}
	return afterScaleIn
}

func (r *instanceGroupResource) scaleOut(
	ctx context.Context,
	instanceGroupID string,
	standbyPort types.Int32,
	subnets []instanceGroupScaleOutSubnetInfo,
	respDiags *diag.Diagnostics,
) bool {
	subnetInfos := make([]mysqlsdk.MysqlV1ApiScaleOutMysqlInstanceGroupModelSubnetInfoRequestModel, 0, len(subnets))
	for _, subnet := range subnets {
		subnetInfos = append(subnetInfos, *mysqlsdk.NewMysqlV1ApiScaleOutMysqlInstanceGroupModelSubnetInfoRequestModel(
			subnet.Replicas,
			subnet.SubnetId,
		))
	}

	requestGroup := mysqlsdk.NewMysqlV1ApiScaleOutMysqlInstanceGroupModelInstanceGroupRequestModel(subnetInfos)
	if standbyPort.IsNull() || standbyPort.IsUnknown() {
		requestGroup.UnsetStandbyPort()
	} else {
		requestGroup.SetStandbyPort(standbyPort.ValueInt32())
	}
	request := mysqlsdk.NewBodyScaleOutMysqlInstanceGroup(*requestGroup)

	_, httpResp, err := common.ExecuteWithRetryAndAuth[struct{}](ctx, r.kc, respDiags,
		func() (struct{}, *http.Response, error) {
			httpResp, err := r.kc.ApiClient.MySQLInstanceGroupsAPI.
				ScaleOutMysqlInstanceGroup(ctx, instanceGroupID).
				XAuthToken(r.kc.XAuthToken).
				BodyScaleOutMysqlInstanceGroup(*request).
				Execute()
			return struct{}{}, httpResp, err
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "ScaleOutMysqlInstanceGroup", err, respDiags)
		return false
	}
	return true
}

func (r *instanceGroupResource) scaleIn(
	ctx context.Context,
	instanceGroupID string,
	instanceIDs []string,
	respDiags *diag.Diagnostics,
) bool {
	tflog.Debug(ctx, "invoking MySQL instance group scale-in", map[string]any{
		"instance_group_id": instanceGroupID,
		"instance_ids":      instanceIDs,
	})

	request := mysqlsdk.NewBodyScaleInMysqlInstanceGroup(*mysqlsdk.NewMysqlV1ApiScaleInMysqlInstanceGroupModelInstanceGroupRequestModel(instanceIDs))

	_, httpResp, err := common.ExecuteWithRetryAndAuth[struct{}](ctx, r.kc, respDiags,
		func() (struct{}, *http.Response, error) {
			httpResp, err := r.kc.ApiClient.MySQLInstanceGroupsAPI.
				ScaleInMysqlInstanceGroup(ctx, instanceGroupID).
				XAuthToken(r.kc.XAuthToken).
				BodyScaleInMysqlInstanceGroup(*request).
				Execute()
			return struct{}{}, httpResp, err
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "ScaleInMysqlInstanceGroup", err, respDiags)
		return false
	}
	return true
}

func (r *instanceGroupResource) selectScaleInTargetInstanceIDs(
	ctx context.Context,
	instanceGroupID string,
	scaleInCounts map[string]int32,
	respDiags *diag.Diagnostics,
) ([]string, bool) {
	result, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, respDiags,
		func() (*mysqlsdk.GetMySQLInstanceGroupResponseModel, *http.Response, error) {
			return r.kc.ApiClient.MySQLInstanceGroupsAPI.
				GetMysqlInstanceGroup(ctx, instanceGroupID).
				XAuthToken(r.kc.XAuthToken).
				Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "GetMysqlInstanceGroup", err, respDiags)
		return nil, false
	}

	instancesResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, respDiags,
		func() (*mysqlsdk.GetMySQLInstanceGroupInstancesResponseModel, *http.Response, error) {
			return r.kc.ApiClient.MySQLInstanceGroupsAPI.
				ListMysqlInstances(ctx, instanceGroupID).
				XAuthToken(r.kc.XAuthToken).
				Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "ListMysqlInstances", err, respDiags)
		return nil, false
	}

	instanceDetailsByID := map[string]mysqlsdk.InstanceResponseModel{}
	for _, instance := range instancesResp.Instances {
		instanceDetailsByID[instance.Id] = instance
	}

	instancesBySubnetID := map[string][]scaleInTargetInstance{}
	if instances, ok := result.InstanceGroup.GetInstancesOk(); ok && instances != nil {
		for _, standby := range instances.Standby {
			subnetID := standby.GetSubnetId()
			if subnetID == "" {
				continue
			}
			instanceID, ok := utils.GetNullableStringValue(standby.InstanceId)
			if !ok || instanceID == "" {
				continue
			}
			target := scaleInTargetInstance{
				ID:       instanceID,
				SubnetID: subnetID,
			}
			if detail, ok := instanceDetailsByID[instanceID]; ok {
				target.Status = detail.Status
				target.AvailabilityStatus = detail.GetAvailabilityStatus()
				target.CreatedAt = detail.CreatedAt
			}
			instancesBySubnetID[subnetID] = append(instancesBySubnetID[subnetID], target)
		}
	}
	tflog.Debug(ctx, "collected MySQL scale-in target candidates", map[string]any{
		"instance_group_id": instanceGroupID,
		"scale_in_counts":   scaleInCounts,
		"candidates":        scaleInTargetInstanceMapLogFields(instancesBySubnetID),
	})

	subnetIDs := make([]string, 0, len(scaleInCounts))
	for subnetID := range scaleInCounts {
		subnetIDs = append(subnetIDs, subnetID)
	}
	slices.Sort(subnetIDs)

	instanceIDs := []string{}
	for _, subnetID := range subnetIDs {
		availableInstances := instancesBySubnetID[subnetID]
		sortScaleInTargetInstances(availableInstances)
		requiredCount := int(scaleInCounts[subnetID])
		if len(availableInstances) < requiredCount {
			common.AddGeneralError(ctx, r, respDiags, fmt.Sprintf("could not select scale-in target instance IDs in subnet %s: required %d, found %d", subnetID, requiredCount, len(availableInstances)))
			return nil, false
		}
		for _, instance := range availableInstances[:requiredCount] {
			instanceIDs = append(instanceIDs, instance.ID)
		}
	}
	tflog.Debug(ctx, "selected MySQL scale-in target instances", map[string]any{
		"instance_group_id":     instanceGroupID,
		"selected_instance_ids": instanceIDs,
	})
	return instanceIDs, true
}

func scaleInTargetInstanceMapLogFields(instancesBySubnetID map[string][]scaleInTargetInstance) map[string][]map[string]string {
	fieldsBySubnetID := make(map[string][]map[string]string, len(instancesBySubnetID))
	for subnetID, instances := range instancesBySubnetID {
		fieldsBySubnetID[subnetID] = scaleInTargetInstanceLogFields(instances)
	}
	return fieldsBySubnetID
}

func scaleInTargetInstanceLogFields(instances []scaleInTargetInstance) []map[string]string {
	fields := make([]map[string]string, 0, len(instances))
	for _, instance := range instances {
		fields = append(fields, map[string]string{
			"id":                  instance.ID,
			"subnet_id":           instance.SubnetID,
			"status":              string(instance.Status),
			"availability_status": instance.AvailabilityStatus,
			"created_at":          instance.CreatedAt,
		})
	}
	return fields
}

func sortScaleInTargetInstances(instances []scaleInTargetInstance) {
	slices.SortFunc(instances, compareScaleInTargetInstances)
}

func compareScaleInTargetInstances(a, b scaleInTargetInstance) int {
	if diff := cmp.Compare(scaleInStatusPriority(a.Status), scaleInStatusPriority(b.Status)); diff != 0 {
		return diff
	}
	if diff := cmp.Compare(scaleInAvailabilityStatusPriority(a.AvailabilityStatus), scaleInAvailabilityStatusPriority(b.AvailabilityStatus)); diff != 0 {
		return diff
	}
	if diff := compareCreatedAtDesc(a.CreatedAt, b.CreatedAt); diff != 0 {
		return diff
	}
	return strings.Compare(a.ID, b.ID)
}

func scaleInStatusPriority(status mysqlsdk.InstanceStatus) int {
	if status == mysqlsdk.INSTANCESTATUS_ERROR {
		return 0
	}
	return 1
}

func scaleInAvailabilityStatusPriority(status string) int {
	switch status {
	case mysqlAvailabilityStatusWarningSync:
		return 0
	case mysqlAvailabilityStatusDelaying:
		return 1
	default:
		return 2
	}
}

func compareCreatedAtDesc(a, b string) int {
	aTime, aOk := parseScaleInCreatedAt(a)
	bTime, bOk := parseScaleInCreatedAt(b)
	if aOk && bOk && !aTime.Equal(bTime) {
		if aTime.After(bTime) {
			return -1
		}
		return 1
	}
	if a != b {
		return strings.Compare(b, a)
	}
	return 0
}

func parseScaleInCreatedAt(value string) (time.Time, bool) {
	t, err := time.Parse(time.RFC3339Nano, value)
	return t, err == nil
}

func (r *instanceGroupResource) pollScaleInUntilApplied(
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
		status, remaining, ok := r.readScaleInStatus(ctx, instanceGroupID, targets, respDiags)
		if !ok {
			return false
		}

		switch status {
		case string(mysqlsdk.INSTANCEGROUPSTATUS_ERROR), string(mysqlsdk.INSTANCEGROUPSTATUS_TERMINATED):
			common.AddGeneralError(ctx, r, respDiags, fmt.Sprintf("scale in finished with unexpected instance group status %q", status))
			return false
		case string(mysqlsdk.INSTANCEGROUPSTATUS_AVAILABLE), string(mysqlsdk.INSTANCEGROUPSTATUS_PRIMARY_AVAILABLE):
			if !remaining {
				return true
			}
		}

		select {
		case <-ctx.Done():
			common.AddGeneralError(ctx, r, respDiags, "context deadline exceeded")
			return false
		case <-ticker.C:
		}
	}
}

func (r *instanceGroupResource) readScaleInStatus(
	ctx context.Context,
	instanceGroupID string,
	targets map[string]struct{},
	respDiags *diag.Diagnostics,
) (string, bool, bool) {
	result, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, respDiags,
		func() (*mysqlsdk.GetMySQLInstanceGroupResponseModel, *http.Response, error) {
			return r.kc.ApiClient.MySQLInstanceGroupsAPI.
				GetMysqlInstanceGroup(ctx, instanceGroupID).
				XAuthToken(r.kc.XAuthToken).
				Execute()
		},
	)
	if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
		common.AddGeneralError(ctx, r, respDiags, "instance group was not found after scale in")
		return "", false, false
	}
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "GetMysqlInstanceGroup", err, respDiags)
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

func (r *instanceGroupResource) waitInstanceGroupAvailableAfterTopologyUpdate(
	ctx context.Context,
	instanceGroupID string,
	respDiags *diag.Diagnostics,
) bool {
	result, ok := r.pollInstanceGroupUntilStatus(
		ctx,
		instanceGroupID,
		[]string{
			string(mysqlsdk.INSTANCEGROUPSTATUS_AVAILABLE),
			string(mysqlsdk.INSTANCEGROUPSTATUS_PRIMARY_AVAILABLE),
			string(mysqlsdk.INSTANCEGROUPSTATUS_ERROR),
			string(mysqlsdk.INSTANCEGROUPSTATUS_TERMINATED),
		},
		respDiags,
	)
	if !ok {
		return false
	}
	common.CheckResourceAvailableStatus(
		ctx,
		r,
		stringPtr(string(result.Status)),
		[]string{
			string(mysqlsdk.INSTANCEGROUPSTATUS_AVAILABLE),
			string(mysqlsdk.INSTANCEGROUPSTATUS_PRIMARY_AVAILABLE),
		},
		respDiags,
	)
	return !respDiags.HasError()
}

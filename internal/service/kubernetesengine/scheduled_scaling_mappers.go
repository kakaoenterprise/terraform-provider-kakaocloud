// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package kubernetesengine

import (
	"context"
	"strings"
	"time"

	. "terraform-provider-kakaocloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kakaoenterprise/kc-sdk-go/services/kubernetesengine"
)

func mapScheduledScalingBaseModel(
	ctx context.Context,
	base *scheduledScalingBaseModel,
	scheduledScalingResult *kubernetesengine.ScheduledScaleResponseModel,
	respDiags *diag.Diagnostics,
) bool {
	base.CreatedAt = types.StringValue(scheduledScalingResult.CreatedAt.Format(time.RFC3339))
	base.DesiredNodes = types.Int32Value(scheduledScalingResult.DesiredNodes)
	base.Name = types.StringValue(scheduledScalingResult.Name)
	base.ScheduleType = types.StringValue(scheduledScalingResult.ScheduleType)
	base.StartTime = types.StringValue(scheduledScalingResult.StartTime)

	if strings.EqualFold(scheduledScalingResult.ScheduleType, "once") ||
		strings.TrimSpace(scheduledScalingResult.Schedule) == "" {
		base.Schedule = types.StringNull()
	} else {
		base.Schedule = types.StringValue(scheduledScalingResult.Schedule)
	}

	statusObj, statusDiags := ConvertObjectFromModel(
		ctx,
		scheduledScalingResult.Status,
		scheduledScalingStatusAttrTypes,
		func(s kubernetesengine.ScalingStatusResponseModel) any {
			hVals := make([]attr.Value, 0, len(s.Histories))
			for _, h := range s.Histories {
				hObj, hObjDiags := types.ObjectValue(
					scheduledScalingHistoryAttrTypes,
					map[string]attr.Value{
						"description":   types.StringValue(h.Description),
						"occurred_time": types.StringValue(h.OccurredTime),
						"state":         types.StringValue(h.State),
					},
				)
				respDiags.Append(hObjDiags...)
				hVals = append(hVals, hObj)
			}

			hList, lDiags := types.ListValue(
				types.ObjectType{AttrTypes: scheduledScalingHistoryAttrTypes},
				hVals,
			)
			respDiags.Append(lDiags...)

			return scheduledScalingStatusModel{Histories: hList}
		},
	)
	respDiags.Append(statusDiags...)
	base.Status = statusObj

	return !respDiags.HasError()
}

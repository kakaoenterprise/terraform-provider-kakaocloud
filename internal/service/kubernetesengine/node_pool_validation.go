// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package kubernetesengine

import (
	"fmt"
	"math"
	"strings"
	"terraform-provider-kakaocloud/internal/common"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"golang.org/x/net/context"
)

func (r *nodePoolResource) validateAutoscalingModel(ctx context.Context, auto NodePoolAutoscalingModel, resp *resource.ValidateConfigResponse) {
	enabled := auto.IsAutoscalerEnable.ValueBool()

	if !enabled {
		invalid := make([]string, 0)
		if !auto.AutoscalerDesiredNodeCount.IsNull() {
			invalid = append(invalid, "autoscaler_desired_node_count")
		}
		if !auto.AutoscalerMaxNodeCount.IsNull() {
			invalid = append(invalid, "autoscaler_max_node_count")
		}
		if !auto.AutoscalerMinNodeCount.IsNull() {
			invalid = append(invalid, "autoscaler_min_node_count")
		}
		if !auto.AutoscalerScaleDownThreshold.IsNull() {
			invalid = append(invalid, "autoscaler_scale_down_threshold")
		}
		if !auto.AutoscalerScaleDownUnneededTime.IsNull() {
			invalid = append(invalid, "autoscaler_scale_down_unneeded_time")
		}
		if !auto.AutoscalerScaleDownUnreadyTime.IsNull() {
			invalid = append(invalid, "autoscaler_scale_down_unready_time")
		}
		if len(invalid) > 0 {
			common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
				fmt.Sprintf("When 'is_autoscaler_enable' is false, the following fields must be unset: %s", strings.Join(invalid, ", ")),
			)
			return
		}
		return
	}

	missing := make([]string, 0)

	if auto.AutoscalerDesiredNodeCount.IsNull() {
		missing = append(missing, "autoscaler_desired_node_count")
	}
	if auto.AutoscalerMaxNodeCount.IsNull() {
		missing = append(missing, "autoscaler_max_node_count")
	}
	if auto.AutoscalerMinNodeCount.IsNull() {
		missing = append(missing, "autoscaler_min_node_count")
	}
	if auto.AutoscalerScaleDownThreshold.IsNull() {
		missing = append(missing, "autoscaler_scale_down_threshold")
	}
	if auto.AutoscalerScaleDownUnneededTime.IsNull() {
		missing = append(missing, "autoscaler_scale_down_unneeded_time")
	}
	if auto.AutoscalerScaleDownUnreadyTime.IsNull() {
		missing = append(missing, "autoscaler_scale_down_unready_time")
	}

	if len(missing) > 0 {
		common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
			fmt.Sprintf("When 'is_autoscaler_enable' is true, the following fields are required: %s", strings.Join(missing, ", ")),
		)
		return
	}

	desiredCount := auto.AutoscalerDesiredNodeCount.ValueInt32()
	maxCount := auto.AutoscalerMaxNodeCount.ValueInt32()
	minCount := auto.AutoscalerMinNodeCount.ValueInt32()
	th := auto.AutoscalerScaleDownThreshold.ValueFloat32()

	if !hasAtMost2Decimals(th) {
		common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
			"'autoscaler_scale_down_threshold' must have at most 2 decimal places.")
	}

	if maxCount < desiredCount {
		common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
			"'autoscaler_max_node_count' must be greater than or equal to 'autoscaler_desired_node_count'.")
	}
	if maxCount < minCount {
		common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
			"'autoscaler_max_node_count' must be greater than or equal to 'autoscaler_min_node_count'.")
	}
	if minCount > desiredCount {
		common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
			"'autoscaler_min_node_count' must be less than or equal to 'autoscaler_desired_node_count'.")
	}
}

func hasAtMost2Decimals(f float32) bool {
	scaled := float64(f * 100.0)
	return math.Mod(scaled, 1.0) == 0
}

func (r *nodePoolResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var plan, state, config *NodePoolResourceModel
	planDiags := req.Plan.Get(ctx, &plan)
	stateDiags := req.State.Get(ctx, &state)
	configDiags := req.Config.Get(ctx, &config)

	resp.Diagnostics.Append(planDiags...)
	resp.Diagnostics.Append(stateDiags...)
	resp.Diagnostics.Append(configDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if req.Plan.Raw.IsNull() {
		return
	}

	if req.State.Raw.IsNull() {
		return
	}

	if !state.IsBareMetal.IsNull() && state.IsBareMetal.ValueBool() {
		r.validateBM(ctx, *config, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if !plan.MinorVersion.IsNull() && !plan.MinorVersion.IsUnknown() && !plan.MinorVersion.Equal(state.MinorVersion) {
		common.MajorMinorVersionNotDecreasingValidator(plan.MinorVersion.ValueString(), state.MinorVersion.ValueString(), &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}

		if !state.IsUpgradable.ValueBool() {
			common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
				fmt.Sprintf("The version cannot be upgraded. current state version: '%v'", state.MinorVersion.ValueString()))
		}
	}
}

func (r *nodePoolResource) validateBM(ctx context.Context, req NodePoolResourceModel, diags *diag.Diagnostics) {

	if !req.VolumeSize.IsNull() && !req.VolumeSize.IsUnknown() {
		common.AddGeneralError(ctx, r, diags,
			"Bare metal node pools do not support 'volume_size' configuration. Remove 'volume_size' from configuration.")
		return
	}

	if !req.RequestSecurityGroups.IsNull() && !req.RequestSecurityGroups.IsUnknown() {
		common.AddGeneralError(ctx, r, diags,
			"Bare metal node pools do not support 'request_security_groups' configuration. Remove 'request_security_groups' from configuration.")
		return
	}

	if !req.Autoscaling.IsNull() && !req.Autoscaling.IsUnknown() {
		common.AddGeneralError(ctx, r, diags,
			"This node pool flavor is bare metal. Remove the 'autoscaling' block from configuration.")
		return
	}
}

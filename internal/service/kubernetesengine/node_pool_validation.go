package // Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
kubernetesengine

import (
	"fmt"
	"math"
	"strings"
)

func validateAutoscalingModel(auto NodePoolAutoscalingModel) (bool, string) {

	if auto.IsAutoscalerEnable.IsUnknown() || auto.IsAutoscalerEnable.IsNull() {
		return true, ""
	}

	enabled := auto.IsAutoscalerEnable.ValueBool()

	if !enabled {
		invalid := make([]string, 0)
		if !auto.AutoscalerDesiredNodeCount.IsUnknown() && !auto.AutoscalerDesiredNodeCount.IsNull() {
			invalid = append(invalid, "autoscaler_desired_node_count")
		}
		if !auto.AutoscalerMaxNodeCount.IsUnknown() && !auto.AutoscalerMaxNodeCount.IsNull() {
			invalid = append(invalid, "autoscaler_max_node_count")
		}
		if !auto.AutoscalerMinNodeCount.IsUnknown() && !auto.AutoscalerMinNodeCount.IsNull() {
			invalid = append(invalid, "autoscaler_min_node_count")
		}
		if !auto.AutoscalerScaleDownThreshold.IsUnknown() && !auto.AutoscalerScaleDownThreshold.IsNull() {
			invalid = append(invalid, "autoscaler_scale_down_threshold")
		}
		if !auto.AutoscalerScaleDownUnneededTime.IsUnknown() && !auto.AutoscalerScaleDownUnneededTime.IsNull() {
			invalid = append(invalid, "autoscaler_scale_down_unneeded_time")
		}
		if !auto.AutoscalerScaleDownUnreadyTime.IsUnknown() && !auto.AutoscalerScaleDownUnreadyTime.IsNull() {
			invalid = append(invalid, "autoscaler_scale_down_unready_time")
		}
		if len(invalid) > 0 {
			return false, fmt.Sprintf("When is_autoscaler_enable is false, the following fields must be null: %s", strings.Join(invalid, ", "))
		}
		return true, ""
	}

	missing := make([]string, 0)

	if auto.AutoscalerDesiredNodeCount.IsUnknown() || auto.AutoscalerDesiredNodeCount.IsNull() {
		missing = append(missing, "autoscaler_desired_node_count")
	}
	if auto.AutoscalerMaxNodeCount.IsUnknown() || auto.AutoscalerMaxNodeCount.IsNull() {
		missing = append(missing, "autoscaler_max_node_count")
	}
	if auto.AutoscalerMinNodeCount.IsUnknown() || auto.AutoscalerMinNodeCount.IsNull() {
		missing = append(missing, "autoscaler_min_node_count")
	}
	if auto.AutoscalerScaleDownThreshold.IsUnknown() || auto.AutoscalerScaleDownThreshold.IsNull() {
		missing = append(missing, "autoscaler_scale_down_threshold")
	}
	if auto.AutoscalerScaleDownUnneededTime.IsUnknown() || auto.AutoscalerScaleDownUnneededTime.IsNull() {
		missing = append(missing, "autoscaler_scale_down_unneeded_time")
	}
	if auto.AutoscalerScaleDownUnreadyTime.IsUnknown() || auto.AutoscalerScaleDownUnreadyTime.IsNull() {
		missing = append(missing, "autoscaler_scale_down_unready_time")
	}

	if len(missing) > 0 {
		return false, fmt.Sprintf("When is_autoscaler_enable is true, the following fields are required: %s", strings.Join(missing, ", "))
	}

	errors := make([]string, 0)

	desired := auto.AutoscalerDesiredNodeCount.ValueInt32()
	max := auto.AutoscalerMaxNodeCount.ValueInt32()
	min := auto.AutoscalerMinNodeCount.ValueInt32()
	th := float64(auto.AutoscalerScaleDownThreshold.ValueFloat32())
	unneeded := auto.AutoscalerScaleDownUnneededTime.ValueInt32()
	unready := auto.AutoscalerScaleDownUnreadyTime.ValueInt32()

	inRange := func(v, lo, hi int32) bool { return v >= lo && v <= hi }

	if !inRange(desired, 0, 100) {
		errors = append(errors, "autoscaler_desired_node_count must be between 0 and 100.")
	}
	if !inRange(max, 0, 100) {
		errors = append(errors, "autoscaler_max_node_count must be between 0 and 100.")
	}
	if !inRange(min, 0, 100) {
		errors = append(errors, "autoscaler_min_node_count must be between 0 and 100.")
	}

	if !(th > 0.0 && th <= 1.0) {
		errors = append(errors, "autoscaler_scale_down_threshold must be between 0.01 and 1.0.")
	} else {
		if !hasAtMost2Decimals(th) {
			errors = append(errors, "autoscaler_scale_down_threshold must have at most 2 decimal places.")
		}
	}

	if !inRange(unneeded, 1, 86400) {
		errors = append(errors, "autoscaler_scale_down_unneeded_time must be between 1 and 86400.")
	}
	if !inRange(unready, 1, 86400) {
		errors = append(errors, "autoscaler_scale_down_unready_time must be between 1 and 86400.")
	}

	if max < desired {
		errors = append(errors, "autoscaler_max_node_count must be greater than or equal to autoscaler_desired_node_count.")
	}
	if max < min {
		errors = append(errors, "autoscaler_max_node_count must be greater than or equal to autoscaler_min_node_count.")
	}
	if min > desired {
		errors = append(errors, "autoscaler_min_node_count must be less than or equal to autoscaler_desired_node_count.")
	}

	if len(errors) > 0 {
		return false, fmt.Sprintf("Validation failed: %s", strings.Join(errors, " "))
	}

	return true, ""
}

func hasAtMost2Decimals(f float64) bool {

	const eps = 1e-6
	scaled := f * 100.0
	_, frac := math.Modf(scaled)
	return math.Abs(frac) < eps
}

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package loadbalancer

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/kakaoenterprise/kc-sdk-go/services/loadbalancer"
)

// ValidateL7PolicyRuleConfig validates the L7 policy rule configuration
func ValidateL7PolicyRuleConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config loadBalancerL7PolicyRuleResourceModel

	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that key is required for HEADER and COOKIE types
	if !config.Type.IsNull() && !config.Type.IsUnknown() {
		ruleType := config.Type.ValueString()

		// Check if key is required for this rule type
		if ruleType == string(loadbalancer.L7RULETYPE_HEADER) || ruleType == string(loadbalancer.L7RULETYPE_COOKIE) {
			if config.Key.IsNull() || config.Key.IsUnknown() || strings.TrimSpace(config.Key.ValueString()) == "" {
				resp.Diagnostics.AddAttributeError(
					path.Root("key"),
					"Missing required field",
					fmt.Sprintf("The 'key' field is required when 'type' is '%s'", ruleType),
				)
			}
		}
	}

	// Validate that key is not provided for types that don't need it
	if !config.Key.IsNull() && !config.Key.IsUnknown() && strings.TrimSpace(config.Key.ValueString()) != "" {
		if !config.Type.IsNull() && !config.Type.IsUnknown() {
			ruleType := config.Type.ValueString()

			// Check if key is not allowed for this rule type
			if ruleType == string(loadbalancer.L7RULETYPE_PATH) ||
				ruleType == string(loadbalancer.L7RULETYPE_HOST_NAME) ||
				ruleType == string(loadbalancer.L7RULETYPE_FILE_TYPE) {
				resp.Diagnostics.AddAttributeError(
					path.Root("key"),
					"Invalid field",
					fmt.Sprintf("The 'key' field is not allowed when 'type' is '%s'", ruleType),
				)
			}
		}
	}

	// Validate value is not empty
	if config.Value.IsNull() || config.Value.IsUnknown() || strings.TrimSpace(config.Value.ValueString()) == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("value"),
			"Missing required field",
			"The 'value' field is required and cannot be empty",
		)
	}

	// Validate key format when provided
	if !config.Key.IsNull() && !config.Key.IsUnknown() && strings.TrimSpace(config.Key.ValueString()) != "" {
		keyValue := strings.TrimSpace(config.Key.ValueString())

		// Key should not contain spaces or special characters that might cause issues
		if strings.Contains(keyValue, " ") {
			resp.Diagnostics.AddAttributeError(
				path.Root("key"),
				"Invalid key format",
				"The 'key' field cannot contain spaces",
			)
		}

		// Key should not be too long
		if len(keyValue) > 255 {
			resp.Diagnostics.AddAttributeError(
				path.Root("key"),
				"Invalid key length",
				"The 'key' field cannot exceed 255 characters",
			)
		}
	}

	// Validate value format
	if !config.Value.IsNull() && !config.Value.IsUnknown() && strings.TrimSpace(config.Value.ValueString()) != "" {
		valueValue := strings.TrimSpace(config.Value.ValueString())

		// Value should not be too long
		if len(valueValue) > 255 {
			resp.Diagnostics.AddAttributeError(
				path.Root("value"),
				"Invalid value length",
				"The 'value' field cannot exceed 255 characters",
			)
		}
	}
}

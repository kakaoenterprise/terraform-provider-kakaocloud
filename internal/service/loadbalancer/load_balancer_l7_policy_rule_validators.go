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

func ValidateL7PolicyRuleConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config loadBalancerL7PolicyRuleResourceModel

	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !config.Type.IsNull() && !config.Type.IsUnknown() {
		ruleType := config.Type.ValueString()

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

	if !config.Key.IsNull() && !config.Key.IsUnknown() && strings.TrimSpace(config.Key.ValueString()) != "" {
		if !config.Type.IsNull() && !config.Type.IsUnknown() {
			ruleType := config.Type.ValueString()

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

	if config.Value.IsNull() || config.Value.IsUnknown() || strings.TrimSpace(config.Value.ValueString()) == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("value"),
			"Missing required field",
			"The 'value' field is required and cannot be empty",
		)
	}

	if !config.Key.IsNull() && !config.Key.IsUnknown() && strings.TrimSpace(config.Key.ValueString()) != "" {
		keyValue := strings.TrimSpace(config.Key.ValueString())

		if strings.Contains(keyValue, " ") {
			resp.Diagnostics.AddAttributeError(
				path.Root("key"),
				"Invalid key format",
				"The 'key' field cannot contain spaces",
			)
		}

		if len(keyValue) > 255 {
			resp.Diagnostics.AddAttributeError(
				path.Root("key"),
				"Invalid key length",
				"The 'key' field cannot exceed 255 characters",
			)
		}
	}

	if !config.Value.IsNull() && !config.Value.IsUnknown() && strings.TrimSpace(config.Value.ValueString()) != "" {
		valueValue := strings.TrimSpace(config.Value.ValueString())

		if len(valueValue) > 255 {
			resp.Diagnostics.AddAttributeError(
				path.Root("value"),
				"Invalid value length",
				"The 'value' field cannot exceed 255 characters",
			)
		}
	}
}

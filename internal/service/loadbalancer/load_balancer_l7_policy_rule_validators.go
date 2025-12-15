// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package loadbalancer

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"terraform-provider-kakaocloud/internal/common"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/kakaoenterprise/kc-sdk-go/services/loadbalancer"
)

func (r *loadBalancerL7PolicyRuleResource) validateL7PolicyRuleConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config loadBalancerL7PolicyRuleResourceModel

	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.Type.IsUnknown() {
		return
	}

	ruleType := config.Type.ValueString()

	if ruleType == string(loadbalancer.L7RULETYPE_HEADER) || ruleType == string(loadbalancer.L7RULETYPE_COOKIE) {
		if config.Key.IsNull() || !config.Key.IsUnknown() && strings.TrimSpace(config.Key.ValueString()) == "" {
			common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
				fmt.Sprintf("The 'key' field is required when 'type' is '%s'", ruleType),
			)
		}
	} else {
		if !config.Key.IsNull() {
			common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
				fmt.Sprintf("The 'key' field is not allowed when 'type' is '%s'", ruleType),
			)
		}
	}

	if config.Key.IsUnknown() || config.Value.IsUnknown() {
		return
	}

	if strings.TrimSpace(config.Value.ValueString()) == "" {
		common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
			"The 'value' field is required and cannot be empty")
	}

	valueValue := config.Value.ValueString()
	keyValue := ""
	if !config.Key.IsNull() {
		keyValue = config.Key.ValueString()
	}

	switch ruleType {
	case string(loadbalancer.L7RULETYPE_COOKIE):
		pattern := regexp.MustCompile(`^[a-zA-Z0-9-_]{1,32}$`)
		if !pattern.MatchString(keyValue) {
			common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
				"COOKIE key must be 1–32 characters long and contain only a–z, A–Z, 0–9, ‘-’, or ‘_’.")
		}
		pattern = regexp.MustCompile(`^[a-zA-Z0-9()\-=*.?;,+/:&_]{1,255}$`)
		if !pattern.MatchString(valueValue) {
			common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
				"COOKIE value must be 1–255 characters long and may include a–z, A–Z, 0–9, (), -= * . ? ; , + / : & _")
		}

	case string(loadbalancer.L7RULETYPE_HEADER):
		pattern := regexp.MustCompile(`^[a-zA-Z0-9-_]{1,255}$`)
		if !pattern.MatchString(keyValue) {
			common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
				"HEADER key must be 1–255 characters long and contain only a–z, A–Z, 0–9, ‘-’, or ‘_’.")
		}
		pattern = regexp.MustCompile(`^[a-zA-Z0-9\-/+=_.|]{1,255}$`)
		if !pattern.MatchString(valueValue) {
			common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
				"HEADER value must be 1–255 characters long and may include a–z, A–Z, 0–9, ‘-’, ‘/’, ‘+’, ‘=’, ‘_’, ‘.’, or ‘|’.")
		}

	case string(loadbalancer.L7RULETYPE_HOST_NAME):
		if config.CompareType.ValueString() != string(loadbalancer.L7RULECOMPARETYPE_EQUAL_TO) {
			common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
				"HOST_NAME rule type only supports EQUAL_TO compare type.")
		}
		fields := strings.Split(valueValue, ".")
		fieldPattern := regexp.MustCompile(`^[a-zA-Z0-9-]{0,62}[a-zA-Z0-9]$`)
		lastFieldPattern := regexp.MustCompile(`.*[a-zA-Z-].*`)
		for _, f := range fields {
			if !fieldPattern.MatchString(f) || strings.HasPrefix(f, "-") {
				common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
					"EQUAL_TO: Each field must be 1–63 characters long, must not start with a hyphen (-), and must end with an alphanumeric character.")
				break
			}
		}
		if len(fields) == 0 || !lastFieldPattern.MatchString(fields[len(fields)-1]) {
			common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
				"EQUAL_TO: The last field must contain at least one letter or hyphen (-).")
		}

	case string(loadbalancer.L7RULETYPE_FILE_TYPE):
		if config.CompareType.ValueString() != string(loadbalancer.L7RULECOMPARETYPE_EQUAL_TO) {
			common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
				"FILE_TYPE rule type only supports EQUAL_TO compare type.")
		}
		pattern := regexp.MustCompile(`^[a-zA-Z0-9!@#\$%^&{}$begin:math:display$$end:math:display$()_+\-=,.~'` + "`" + `]{1,255}$`)
		if !pattern.MatchString(valueValue) {
			common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
				"Invalid FILE_TYPE value: Must be 1–255 characters long and contain only a–z, A–Z, 0–9, and the following symbols: a-z, A-Z, 0-9, and !@#$%^&{}[]()_+-=,.~'`")
		}

	case string(loadbalancer.L7RULETYPE_PATH):
		pattern := regexp.MustCompile(`^[a-zA-Z0-9./\-_]{1,255}$`)
		if !pattern.MatchString(valueValue) || strings.HasPrefix(valueValue, "-") || strings.HasPrefix(valueValue, "_") {
			common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
				"Invalid PATH value: Must be 1–255 characters long, must not start with ‘-’ or ‘_’, and may include a–z, A–Z, 0–9, ‘.’, ‘-’, ‘/’, or ‘_’.")
		}
	}
}

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package network

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// expandSecurityGroupRules converts types.List into a slice of rule models
func expandSecurityGroupRules(ctx context.Context, ruleList types.Set) ([]securityGroupRuleModel, diag.Diagnostics) {
	var rules []securityGroupRuleModel
	diags := ruleList.ElementsAs(ctx, &rules, false)
	return rules, diags
}

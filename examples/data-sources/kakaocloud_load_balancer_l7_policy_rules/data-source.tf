# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# List all L7 policy rules for a specific policy
data "kakaocloud_load_balancer_l7_policy_rules" "all" {
    id = "your-l7-policy-id-here" # Replace with your L7 policy ID
}

# List L7 policy rules with filters
data "kakaocloud_load_balancer_l7_policy_rules" "filtered" {
    id = "your-l7-policy-id-here" # Replace with your L7 policy ID
}

# Output the L7 policy rules list
output "all_l7_policy_rules" {
  description = "List of L7 policy rules"
  value = data.kakaocloud_load_balancer_l7_policy_rules.all
}

# Output the filtered L7 policy rules list
output "filtered_l7_policy_rules" {
  description = "Filtered list of L7 policy rules"
  value = data.kakaocloud_load_balancer_l7_policy_rules.filtered
}



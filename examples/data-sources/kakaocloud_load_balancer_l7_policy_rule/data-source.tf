# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# Get a specific L7 policy rule by ID
data "kakaocloud_load_balancer_l7_policy_rule" "example" {
  id           = "your-l7-policy-rule-id-here"  # Replace with your L7 policy rule ID
  l7_policy_id = "your-l7-policy-id-here"      # Replace with your L7 policy ID
}

# Output the L7 policy rule
output "l7_policy_rule" {
  value = data.kakaocloud_load_balancer_l7_policy_rule.example
}
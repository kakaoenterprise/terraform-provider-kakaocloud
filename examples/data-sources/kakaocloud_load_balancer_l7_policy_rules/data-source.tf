# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# Example: List all L7 policy rules for a specific policy
data "kakaocloud_load_balancer_l7_policy_rules" "example" {
  id = "2415269a-7142-455a-a7c8-9082dd146c57" # Replace with your L7 policy ID
}

# Output the L7 policy rules list
output "l7_policy_rules_list" {
  description = "List of L7 policy rules"
  value = {
    policy_id   = data.kakaocloud_load_balancer_l7_policy_rules.example.id
    rules_count = data.kakaocloud_load_balancer_l7_policy_rules.example.rules_count
    rules       = data.kakaocloud_load_balancer_l7_policy_rules.example.l7_rules
  }
}

# Example: Filter rules by type
output "path_rules_only" {
  description = "Only PATH type rules"
  value = [
    for rule in data.kakaocloud_load_balancer_l7_policy_rules.example.l7_rules : rule
    if rule.type == "PATH"
  ]
}

# Example: Filter rules by compare type
output "starts_with_rules" {
  description = "Only STARTS_WITH compare type rules"
  value = [
    for rule in data.kakaocloud_load_balancer_l7_policy_rules.example.l7_rules : rule
    if rule.compare_type == "STARTS_WITH"
  ]
}

# Example: Get rules with specific values
output "api_path_rules" {
  description = "Rules that match /api/ paths"
  value = [
    for rule in data.kakaocloud_load_balancer_l7_policy_rules.example.l7_rules : rule
    if rule.type == "PATH" && rule.value == "/api/"
  ]
}

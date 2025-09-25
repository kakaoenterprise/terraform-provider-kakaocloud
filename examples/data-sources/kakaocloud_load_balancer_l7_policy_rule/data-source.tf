# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# Example: Read a specific L7 policy rule
data "kakaocloud_load_balancer_l7_policy_rule" "example" {
  id           = "8a7a1ca5-c687-4a9a-999b-a169ee248ade" # Replace with your L7 policy rule ID
  l7_policy_id = "2415269a-7142-455a-a7c8-9082dd146c57" # Replace with your L7 policy ID
}

# Output the L7 policy rule details
output "l7_policy_rule_details" {
  description = "Details of the L7 policy rule"
  value = {
    id                  = data.kakaocloud_load_balancer_l7_policy_rule.example.id
    l7_policy_id        = data.kakaocloud_load_balancer_l7_policy_rule.example.l7_policy_id
    type                = data.kakaocloud_load_balancer_l7_policy_rule.example.type
    compare_type        = data.kakaocloud_load_balancer_l7_policy_rule.example.compare_type
    key                 = data.kakaocloud_load_balancer_l7_policy_rule.example.key
    value               = data.kakaocloud_load_balancer_l7_policy_rule.example.value
    is_inverted         = data.kakaocloud_load_balancer_l7_policy_rule.example.is_inverted
    provisioning_status = data.kakaocloud_load_balancer_l7_policy_rule.example.provisioning_status
    operating_status    = data.kakaocloud_load_balancer_l7_policy_rule.example.operating_status
    project_id          = data.kakaocloud_load_balancer_l7_policy_rule.example.project_id
  }
}

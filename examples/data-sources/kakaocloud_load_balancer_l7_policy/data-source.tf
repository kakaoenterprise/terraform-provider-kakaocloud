# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# Get a specific load balancer L7 policy by ID
data "kakaocloud_load_balancer_l7_policy" "example" {
  id = "your-l7-policy-id-here"  # Replace with your L7 policy ID
}

# Output the L7 policy
output "l7_policy" {
  value = data.kakaocloud_load_balancer_l7_policy.example
}

# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# Get a specific load balancer target group by ID
data "kakaocloud_load_balancer_target_group" "example" {
  id = "your-target-group-id-here"  # Replace with your target group ID
}

# Output the target group
output "target_group" {
  value = data.kakaocloud_load_balancer_target_group.example
}

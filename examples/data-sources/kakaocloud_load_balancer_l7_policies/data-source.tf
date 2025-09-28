# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# List all load balancer L7 policies for a specific load balancer and listener
data "kakaocloud_load_balancer_l7_policies" "all" {
  load_balancer_id = "your-load-balancer-id-here"  # Replace with your load balancer ID
  listener_id      = "your-listener-id-here"      # Replace with your listener ID
}

# List L7 policies with filters
data "kakaocloud_load_balancer_l7_policies" "filtered" {
  load_balancer_id = "your-load-balancer-id-here"  # Replace with your load balancer ID
  listener_id      = "your-listener-id-here"      # Replace with your listener ID
  filter = [
    {
      name  = "position"
      value = "1"
    },
    {
      name  = "action"
      value = "REDIRECT_TO_URL"
    },
    {
      name  = "provisioning_status"
      value = "ACTIVE"
    },
    {
      name  = "operating_status"
      value = "ONLINE"
    },
    {
      name  = "name"
      value = "your-l7-policy-name-here"  # Replace with your L7 policy name
    },
  ]
}

# Output all L7 policies
output "all_l7_policies" {
  description = "List of all load balancer L7 policies"
  value = data.kakaocloud_load_balancer_l7_policies.all
}

# Output filtered L7 policies
output "filtered_l7_policies" {
  description = "Filtered load balancer L7 policies"
  value = data.kakaocloud_load_balancer_l7_policies.filtered
}

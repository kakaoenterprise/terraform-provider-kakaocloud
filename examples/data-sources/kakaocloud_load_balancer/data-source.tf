# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# Get a specific load balancer by ID
data "kakaocloud_load_balancer" "example" {
  id = "your-load-balancer-id-here"  # Replace with your load balancer ID
}

# Output the fetched data
output "load_balancer_name" {
  value = data.kakaocloud_load_balancer.example
}

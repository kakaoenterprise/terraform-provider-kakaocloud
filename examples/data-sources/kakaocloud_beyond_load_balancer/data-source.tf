# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# Get a specific beyond load balancer by ID
data "kakaocloud_beyond_load_balancer" "example" {
  id = "your-beyond-load-balancer-id-here"  # Replace with your beyond load balancer ID
}

# Output the fetched data
output "beyond_load_balancer_name" {
  value = data.kakaocloud_beyond_load_balancer.example
}

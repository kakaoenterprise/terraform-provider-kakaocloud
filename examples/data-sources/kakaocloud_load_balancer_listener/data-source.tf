# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# Get a specific load balancer listener by ID
data "kakaocloud_load_balancer_listener" "example" {
  id = "your-listener-id-here"  # Replace with your listener ID
}

#
output "listener" {
  value = data.kakaocloud_load_balancer_listener.example
}

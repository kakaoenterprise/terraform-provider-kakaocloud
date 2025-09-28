# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# Get a specific load balancer health monitor by ID
data "kakaocloud_load_balancer_health_monitor" "example" {
  id = "your-health-monitor-id-here"  # Replace with your health monitor ID
}

# Output the health monitor
output "health_monitor" {
  value = data.kakaocloud_load_balancer_health_monitor.example
}
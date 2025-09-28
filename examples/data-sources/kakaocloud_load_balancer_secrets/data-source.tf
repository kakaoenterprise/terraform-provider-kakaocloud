# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# List all available load balancer secrets
data "kakaocloud_load_balancer_secrets" "all" {
  # No filters - get all available secrets
}

# List load balancer secrets with filters
data "kakaocloud_load_balancer_secrets" "filtered" {
  filter = [
    {
      name  = "name"
      value = "your-secret-name"  # Replace with your secret name
    },
    {
      name = "created_at"
      value = "2024-01-01T00:00:00Z"  # Replace with your start date
    },
    {
      name = "updated_at"
      value = "2024-12-31T23:59:59Z"  # Replace with your end date
    },
    {
      name  = "expiration"
      value = "2024-12-31T23:59:59Z"  # Replace with your expiration date
    }
  ]
}

# Output all load balancer secrets
output "all_load_balancer_secrets" {
  description = "List of all available load balancer secrets"
  value = data.kakaocloud_load_balancer_secrets.all
}

# Output filtered load balancer secrets
output "filtered_load_balancer_secrets" {
  description = "List of filtered load balancer secrets"
  value = data.kakaocloud_load_balancer_secrets.filtered
}

# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# List all public IPs
data "kakaocloud_public_ips" "all" {
  # No filters - get all public IPs
}

# List public IPs with comprehensive filters
data "kakaocloud_public_ips" "filtered" {
  filter = [
    {
      name  = "id"
      value = "your-public-ip-id"  # Replace with your public IP ID
    },
    {
      name  = "status"
      value = "ACTIVE"  # ACTIVE, DOWN, ERROR
    },
    {
      name  = "public_ip"
      value = "1.2.3.4"  # Replace with your public IP address
    },
    {
      name  = "related_resource_id"
      value = "your-related-resource-id"  # Replace with your related resource ID
    },
    {
      name  = "related_resource_name"
      value = "your-related-resource-name"  # Replace with your related resource name
    },
    {
      name  = "created_at"
      value = "2024-01-01T00:00:00Z"  # Replace with creation time (RFC3339 format)
    },
    {
      name  = "updated_at"
      value = "2024-12-31T23:59:59Z"  # Replace with update time (RFC3339 format)
    }
  ]
}

# Output all public IPs
output "all_public_ips" {
  description = "List of all public IPs"
  value = {
    count = length(data.kakaocloud_public_ips.all.public_ips)
    ids   = data.kakaocloud_public_ips.all.public_ips[*].id
    names = data.kakaocloud_public_ips.all.public_ips[*].name
  }
}

# Output filtered public IPs
output "filtered_public_ips" {
  description = "List of filtered public IPs"
  value = {
    count = length(data.kakaocloud_public_ips.filtered.public_ips)
    ids   = data.kakaocloud_public_ips.filtered.public_ips[*].id
    names = data.kakaocloud_public_ips.filtered.public_ips[*].name
  }
}

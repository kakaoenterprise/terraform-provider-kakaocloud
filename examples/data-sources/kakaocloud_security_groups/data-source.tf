# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# List all security groups
data "kakaocloud_security_groups" "all" {
  # No filters - get all security groups
}

# List security groups with comprehensive filters
data "kakaocloud_security_groups" "filtered" {
  filter = [
    {
      name  = "name"
      value = "your-security-group-name"  # Replace with your security group name
    },
    {
      name  = "id"
      value = "your-security-group-id"  # Replace with your security group ID
    },
    {
      name  = "created_at"
      value = "2024-01-01T00:00:00Z"  # Replace with creation time (RFC3339 format)
    }
  ]
}

# Output all security groups
output "all_security_groups" {
  description = "List of all security groups"
  value = {
    count = length(data.kakaocloud_security_groups.all.security_groups)
    ids   = data.kakaocloud_security_groups.all.security_groups[*].id
    names = data.kakaocloud_security_groups.all.security_groups[*].name
  }
}

# Output filtered security groups
output "filtered_security_groups" {
  description = "List of filtered security groups"
  value = {
    count = length(data.kakaocloud_security_groups.filtered.security_groups)
    ids   = data.kakaocloud_security_groups.filtered.security_groups[*].id
    names = data.kakaocloud_security_groups.filtered.security_groups[*].name
  }
}

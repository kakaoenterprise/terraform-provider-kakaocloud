# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# List all VPCs
data "kakaocloud_vpcs" "all" {
  # No filters - get all VPCs
}

# List VPCs with comprehensive filters
data "kakaocloud_vpcs" "filtered" {
  filter = [
    {
      name  = "id"
      value = "your-vpc-id"  # Replace with your VPC ID
    },
    {
      name  = "name"
      value = "your-vpc-name"  # Replace with your VPC name
    },
    {
      name  = "cidr_block"
      value = "10.0.0.0/16"  # Replace with your CIDR block
    },
    {
      name  = "provisioning_status"
      value = "ACTIVE"  # ACTIVE, BUILDING, DELETED, ERROR, PENDING_CREATE, PENDING_DELETE, PENDING_UPDATE
    },
    {
      name  = "is_default"
      value = "false"  # true, false
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

# Output all VPCs
output "all_vpcs" {
  description = "List of all VPCs"
  value = {
    count = length(data.kakaocloud_vpcs.all.vpcs)
    ids   = data.kakaocloud_vpcs.all.vpcs[*].id
    names = data.kakaocloud_vpcs.all.vpcs[*].name
  }
}

# Output filtered VPCs
output "filtered_vpcs" {
  description = "List of filtered VPCs"
  value = {
    count = length(data.kakaocloud_vpcs.filtered.vpcs)
    ids   = data.kakaocloud_vpcs.filtered.vpcs[*].id
    names = data.kakaocloud_vpcs.filtered.vpcs[*].name
  }
}

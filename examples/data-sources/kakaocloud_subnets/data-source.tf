# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# List all subnets
data "kakaocloud_subnets" "all" {
  # No filters - get all subnets
}

# List subnets with comprehensive filters
data "kakaocloud_subnets" "filtered" {
  filter = [
    {
      name  = "id"
      value = "your-subnet-id"  # Replace with your subnet ID
    },
    {
      name  = "name"
      value = "your-subnet-name"  # Replace with your subnet name
    },
    {
      name  = "availability_zone"
      value = "kr-central-2-a"  # Replace with your availability zone
    },
    {
      name  = "provisioning_status"
      value = "ACTIVE"  # ACTIVE, BUILDING, DELETED, ERROR, PENDING_CREATE, PENDING_DELETE, PENDING_UPDATE
    },
    {
      name  = "operating_status"
      value = "ONLINE"  # ONLINE, OFFLINE, DEGRADED, ERROR
    },
    {
      name  = "cidr_block"
      value = "10.0.1.0/24"  # Replace with your subnet CIDR
    },
    {
      name  = "vpc_id"
      value = "your-vpc-id"  # Replace with your VPC ID
    },
    {
      name  = "vpc_name"
      value = "your-vpc-name"  # Replace with your VPC name
    },
    {
      name  = "route_table_id"
      value = "your-route-table-id"  # Replace with your route table ID
    },
    {
      name  = "route_table_name"
      value = "your-route-table-name"  # Replace with your route table name
    },
    {
      name  = "is_shared"
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

# Output all subnets
output "all_subnets" {
  description = "List of all subnets"
  value = {
    count = length(data.kakaocloud_subnets.all.subnets)
    ids   = data.kakaocloud_subnets.all.subnets[*].id
    names = data.kakaocloud_subnets.all.subnets[*].name
  }
}

# Output filtered subnets
output "filtered_subnets" {
  description = "List of filtered subnets"
  value = {
    count = length(data.kakaocloud_subnets.filtered.subnets)
    ids   = data.kakaocloud_subnets.filtered.subnets[*].id
    names = data.kakaocloud_subnets.filtered.subnets[*].name
  }
}

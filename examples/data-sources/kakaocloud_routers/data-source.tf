# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# List all routers
data "kakaocloud_routers" "all" {
  # No filters - get all routers
}

# List routers with comprehensive filters
data "kakaocloud_routers" "filtered" {
  filter = [
    {
      name  = "id"
      value = "your-route-table-id"  # Replace with your route table ID
    },
    {
      name  = "name"
      value = "your-route-table-name"  # Replace with your route table name
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
      name  = "provisioning_status"
      value = "ACTIVE"  # ACTIVE, BUILDING, DELETED, ERROR, PENDING_CREATE, PENDING_DELETE, PENDING_UPDATE
    },
    {
      name  = "vpc_provisioning_status"
      value = "ACTIVE"  # ACTIVE, BUILDING, DELETED, ERROR, PENDING_CREATE, PENDING_DELETE, PENDING_UPDATE
    },
    {
      name  = "subnet_id"
      value = "your-subnet-id"  # Replace with your subnet ID
    },
    {
      name  = "subnet_name"
      value = "your-subnet-name"  # Replace with your subnet name
    },
    {
      name  = "association_count"
      value = "1"  # Replace with association count
    },
    {
      name  = "destination"
      value = "0.0.0.0/0"  # Replace with destination CIDR
    },
    {
      name  = "created_at"
      value = "2024-01-01T00:00:00Z"  # Replace with creation time (RFC3339 format)
    }
  ]
}

# Output all routers
output "all_routers" {
  description = "List of all routers"
  value = {
    count = length(data.kakaocloud_routers.all.routers)
    ids   = data.kakaocloud_routers.all.routers[*].id
    names = data.kakaocloud_routers.all.routers[*].name
  }
}

# Output filtered routers
output "filtered_routers" {
  description = "List of filtered routers"
  value = {
    count = length(data.kakaocloud_routers.filtered.routers)
    ids   = data.kakaocloud_routers.filtered.routers[*].id
    names = data.kakaocloud_routers.filtered.routers[*].name
  }
}

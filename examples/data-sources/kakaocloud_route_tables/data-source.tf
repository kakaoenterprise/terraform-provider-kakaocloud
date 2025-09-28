# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# List all route tables
data "kakaocloud_route_tables" "all" {
  # No filters - get all route tables
}

# List route tables with comprehensive filters
data "kakaocloud_route_tables" "filtered" {
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
      value = "1"  # Number of associated subnets
    },
    {
      name  = "destination"
      value = "0.0.0.0/0"  # Route destination (e.g., default route)
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

# Output all route tables
output "all_route_tables" {
  description = "List of all route tables"
  value = {
    count = length(data.kakaocloud_route_tables.all.route_tables)
    ids   = data.kakaocloud_route_tables.all.route_tables[*].id
    names = data.kakaocloud_route_tables.all.route_tables[*].name
  }
}

# Output filtered route tables
output "filtered_route_tables" {
  description = "List of filtered route tables"
  value = {
    count = length(data.kakaocloud_route_tables.filtered.route_tables)
    ids   = data.kakaocloud_route_tables.filtered.route_tables[*].id
    names = data.kakaocloud_route_tables.filtered.route_tables[*].name
  }
}

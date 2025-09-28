# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# Get a specific route table by ID
data "kakaocloud_route_table" "example" {
  id = "your-route-table-id-here"  # Replace with your route table ID
}

# Output the route table information
output "route_table_example" {
  description = "Information about the example route table"
  value = kakaocloud_route_table.example
}

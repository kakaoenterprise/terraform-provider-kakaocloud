# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# Get a specific instance flavor by ID
data "kakaocloud_instance_flavor" "example" {
  id = "your-flavor-id-here"  # Replace with your flavor ID
}

# Output the flavor information
output "flavor_example" {
  description = "Information about the example flavor"
  value = kakaocloud_instance_flavor.example
}

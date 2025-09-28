# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# Get a specific VPC by ID
data "kakaocloud_vpc" "example" {
  id = "your-vpc-id-here"  # Replace with your VPC ID
}

# Output the VPC information
output "vpc_example" {
  description = "Information about the example VPC"
  value = kakaocloud_vpc.example
}

# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# Get a specific subnet by ID
data "kakaocloud_subnet" "example" {
  id = "your-subnet-id-here"  # Replace with your subnet ID
}

# Output the subnet information
output "subnet_example" {
  description = "Information about the example subnet"
  value = kakaocloud_subnet.example
}

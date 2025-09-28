# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# Get subnet share information by subnet ID
data "kakaocloud_subnet_share" "example" {
  id = "your-subnet-id-here"  # Replace with your subnet ID
}

# Output the subnet share information
output "subnet_share_example" {
  description = "Information about the example subnet share"
  value = {
    id = data.kakaocloud_subnet_share.example.id
    projects = data.kakaocloud_subnet_share.example.projects
  }
}

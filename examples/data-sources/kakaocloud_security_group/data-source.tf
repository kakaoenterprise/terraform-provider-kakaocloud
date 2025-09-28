# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# Get a specific security group by ID
data "kakaocloud_security_group" "example" {
  id = "your-security-group-id-here"  # Replace with your security group ID
}

# Output the security group information
output "security_group_example" {
  description = "Information about the example security group"
  value = kakaocloud_security_group.example
}

# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# Get a specific public IP by ID
data "kakaocloud_public_ip" "example" {
  id = "your-public-ip-id-here"  # Replace with your public IP ID
}

# Output the public IP information
output "public_ip_example" {
  description = "Information about the example public IP"
  value = kakaocloud_public_ip.example
}

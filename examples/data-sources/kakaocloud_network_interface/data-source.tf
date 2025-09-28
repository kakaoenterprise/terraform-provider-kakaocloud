# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# Get a specific network interface by ID
data "kakaocloud_network_interface" "example" {
  id = "your-network-interface-id-here"  # Replace with your network interface ID
}

# Output the network interface information
output "network_interface_example" {
  description = "Information about the example network interface"
  value = kakaocloud_network_interface.example
}

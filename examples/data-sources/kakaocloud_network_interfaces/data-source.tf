# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# List all network interfaces
data "kakaocloud_network_interfaces" "all" {
  # No filters - get all network interfaces
}

# List network interfaces with comprehensive filters
data "kakaocloud_network_interfaces" "filtered" {
  filter = [
    {
      name  = "id"
      value = "your-network-interface-id"  # Replace with your network interface ID
    },
    {
      name  = "name"
      value = "your-network-interface-name"  # Replace with your network interface name
    },
    {
      name  = "status"
      value = "AVAILABLE"  # AVAILABLE, IN_USE
    },
    {
      name  = "private_ip"
      value = "10.0.1.100"  # Replace with your private IP
    },
    {
      name  = "public_ip"
      value = "203.0.113.1"  # Replace with your public IP
    },
    {
      name  = "device_id"
      value = "your-device-id"  # Replace with your device ID
    },
    {
      name  = "device_owner"
      value = "compute:nova"  # Replace with your device owner
    },
    {
      name  = "subnet_id"
      value = "your-subnet-id"  # Replace with your subnet ID
    },
    {
      name  = "mac_address"
      value = "aa:bb:cc:dd:ee:ff"  # Replace with your MAC address
    },
    {
      name  = "security_group_id"
      value = "your-security-group-id"  # Replace with your security group ID
    },
    {
      name  = "security_group_name"
      value = "your-security-group-name"  # Replace with your security group name
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

# Output all network interfaces
output "all_network_interfaces" {
  description = "List of all network interfaces"
  value = {
    count = length(data.kakaocloud_network_interfaces.all.network_interfaces)
    ids   = data.kakaocloud_network_interfaces.all.network_interfaces[*].id
    names = data.kakaocloud_network_interfaces.all.network_interfaces[*].name
  }
}

# Output filtered network interfaces
output "filtered_network_interfaces" {
  description = "List of filtered network interfaces"
  value = {
    count = length(data.kakaocloud_network_interfaces.filtered.network_interfaces)
    ids   = data.kakaocloud_network_interfaces.filtered.network_interfaces[*].id
    names = data.kakaocloud_network_interfaces.filtered.network_interfaces[*].name
  }
}

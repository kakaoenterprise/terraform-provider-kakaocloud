# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# List all networks
data "kakaocloud_networks" "all" {
  # No filters - get all networks
}

# List networks with comprehensive filters
data "kakaocloud_networks" "filtered" {
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
      value = "ACTIVE"  # ACTIVE, BUILDING, DOWN, ERROR
    },
    {
      name  = "private_ip"
      value = "192.168.1.100"  # Replace with your private IP
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

# Output all networks
output "all_networks" {
  description = "List of all networks"
  value = {
    count = length(data.kakaocloud_networks.all.networks)
    ids   = data.kakaocloud_networks.all.networks[*].id
    names = data.kakaocloud_networks.all.networks[*].name
  }
}

# Output filtered networks
output "filtered_networks" {
  description = "List of filtered networks"
  value = {
    count = length(data.kakaocloud_networks.filtered.networks)
    ids   = data.kakaocloud_networks.filtered.networks[*].id
    names = data.kakaocloud_networks.filtered.networks[*].name
  }
}

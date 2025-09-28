# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# List all instance flavors
data "kakaocloud_instance_flavors" "all" {
  # No filters - get all instance flavors
}

# List instance flavors with comprehensive filters
data "kakaocloud_instance_flavors" "filtered" {
  filter = [
    {
      name  = "id"
      value = "your-flavor-id"  # Replace with your flavor ID
    },
    {
      name  = "name"
      value = "your-flavor-name"  # Replace with your flavor name
    },
    {
      name  = "is_burstable"
      value = "false"  # true, false
    },
    {
      name  = "vcpus"
      value = "1"  # Replace with number of vCPUs
    },
    {
      name  = "architecture"
      value = "x86_64"  # Replace with architecture
    },
    {
      name  = "memory_mb"
      value = "1024"  # Replace with memory size in MB
    },
    {
      name  = "instance_type"
      value = "COMPUTE"  # COMPUTE, MEMORY, STORAGE, GPU
    },
    {
      name  = "instance_family"
      value = "your-instance-family"  # Replace with instance family
    },
    {
      name  = "instance_size"
      value = "your-instance-size"  # Replace with instance size
    },
    {
      name  = "manufacturer"
      value = "your-manufacturer"  # Replace with manufacturer
    },
    {
      name  = "maximum_network_interfaces"
      value = "2"  # Replace with maximum network interfaces (integer)
    },
    {
      name  = "processor"
      value = "your-processor"  # Replace with processor
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

# Output all instance flavors
output "all_flavors" {
  description = "List of all instance flavors"
  value = {
    count = length(data.kakaocloud_instance_flavors.all.flavors)
    ids   = data.kakaocloud_instance_flavors.all.flavors[*].id
    names = data.kakaocloud_instance_flavors.all.flavors[*].name
  }
}

# Output filtered instance flavors
output "filtered_flavors" {
  description = "List of filtered instance flavors"
  value = {
    count = length(data.kakaocloud_instance_flavors.filtered.flavors)
    ids   = data.kakaocloud_instance_flavors.filtered.flavors[*].id
    names = data.kakaocloud_instance_flavors.filtered.flavors[*].name
  }
}
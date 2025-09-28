# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# List all instances
data "kakaocloud_instances" "all" {
  # No filters - get all instances
}

# List instances with comprehensive filters
data "kakaocloud_instances" "filtered" {
  filter = [
    {
      name  = "name"
      value = "your-instance-name"  # Replace with your instance name
    },
    {
      name  = "id"
      value = "your-instance-id"  # Replace with your instance ID
    },
    {
      name  = "status"
      value = "ACTIVE"  # ACTIVE, BUILDING, DELETED, ERROR, HARD_REBOOT, MIGRATING, PAUSED, REBOOT, RESCUE, RESIZE, REVERT_RESIZE, SHELVED, SHELVED_OFFLOADED, SHUTOFF, SOFT_DELETED, SUSPENDED, UNKNOWN, VERIFY_RESIZE
    },
    {
      name  = "vm_state"
      value = "active"  # active, building, deleted, error, hard_reboot, migrating, paused, reboot, rescue, resize, revert_resize, shelved, shelved_offloaded, shutoff, soft_deleted, suspended, unknown, verify_resize
    },
    {
      name  = "flavor_name"
      value = "your-flavor-name"  # Replace with your flavor name
    },
    {
      name  = "image_name"
      value = "your-image-name"  # Replace with your image name
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
      name  = "availability_zone"
      value = "kr-central-2-a"  # Replace with your availability zone
    },
    {
      name  = "instance_type"
      value = "COMPUTE"  # COMPUTE, MEMORY, STORAGE, GPU
    },
    {
      name  = "user_id"
      value = "your-user-id"  # Replace with your user ID
    },
    {
      name  = "hostname"
      value = "your-hostname"  # Replace with your hostname
    },
    {
      name  = "os_type"
      value = "linux"  # linux, windows
    },
    {
      name  = "is_hadoop"
      value = "false"  # true, false
    },
    {
      name  = "is_k8se"
      value = "false"  # true, false
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

# Output all instances
output "all_instances" {
  description = "List of all instances"
  value = {
    count = length(data.kakaocloud_instances.all.instances)
    ids   = data.kakaocloud_instances.all.instances[*].id
    names = data.kakaocloud_instances.all.instances[*].name
  }
}

# Output filtered instances
output "filtered_instances" {
  description = "List of filtered instances"
  value = {
    count = length(data.kakaocloud_instances.filtered.instances)
    ids   = data.kakaocloud_instances.filtered.instances[*].id
    names = data.kakaocloud_instances.filtered.instances[*].name
  }
}
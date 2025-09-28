# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# List all volumes
data "kakaocloud_volumes" "all" {
  # No filters - get all volumes
}

# List volumes with comprehensive filters
data "kakaocloud_volumes" "filtered" {
  filter = [
    {
      name  = "name"
      value = "your-volume-name"  # Replace with your volume name
    },
    {
      name  = "id"
      value = "your-volume-id"  # Replace with your volume ID
    },
    {
      name  = "status"
      value = "available"  # available, creating, deleting, error, error_deleting, error_restoring, in-use, restoring, uploading
    },
    {
      name  = "instance_id"
      value = "your-instance-id"  # Replace with your instance ID
    },
    {
      name  = "mount_point"
      value = "/dev/vdb"  # Replace with your mount point
    },
    {
      name  = "type"
      value = "your-volume-type"  # Replace with your volume type
    },
    {
      name  = "size"
      value = "100"  # Replace with volume size in GB
    },
    {
      name  = "availability_zone"
      value = "kr-central-2-a"  # Replace with your availability zone
    },
    {
      name  = "instance_name"
      value = "your-instance-name"  # Replace with your instance name
    },
    {
      name  = "volume_type"
      value = "your-volume-type"  # Replace with your volume type
    },
    {
      name  = "attach_status"
      value = "attached"  # attached, detached
    },
    {
      name  = "is_bootable"
      value = "true"  # true, false
    },
    {
      name  = "is_encrypted"
      value = "false"  # true, false
    },
    {
      name  = "is_root"
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

# Output all volumes
output "all_volumes" {
  description = "List of all volumes"
  value = {
    count = length(data.kakaocloud_volumes.all.volumes)
    ids   = data.kakaocloud_volumes.all.volumes[*].id
    names = data.kakaocloud_volumes.all.volumes[*].name
  }
}

# Output filtered volumes
output "filtered_volumes" {
  description = "List of filtered volumes"
  value = {
    count = length(data.kakaocloud_volumes.filtered.volumes)
    ids   = data.kakaocloud_volumes.filtered.volumes[*].id
    names = data.kakaocloud_volumes.filtered.volumes[*].name
  }
}
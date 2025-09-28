# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# List all volume snapshots
data "kakaocloud_volume_snapshots" "all" {
  # No filters - get all volume snapshots
}

# List volume snapshots with comprehensive filters
data "kakaocloud_volume_snapshots" "filtered" {
  filter = [
    {
      name  = "name"
      value = "your-snapshot-name"  # Replace with your snapshot name
    },
    {
      name  = "id"
      value = "your-snapshot-id"  # Replace with your snapshot ID
    },
    {
      name  = "volume_id"
      value = "your-volume-id"  # Replace with your volume ID
    },
    {
      name  = "status"
      value = "available"  # available, creating, deleting, error, error_deleting, restoring
    },
    {
      name  = "is_incremental"
      value = "false"  # true or false
    },
    {
      name  = "is_dependent_snapshot"
      value = "false"  # true or false
    }
    {
      name  = "schedule_id"
      value = "your-schedule-id"  # Replace with your schedule ID
    },
    {
      name  = "parent_id"
      value = "your-parent-id"  # Replace with your parent ID
    }
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

# Output all volume snapshots
output "all_volume_snapshots" {
  description = "List of all volume snapshots"
  value = {
    count = length(data.kakaocloud_volume_snapshots.all.volume_snapshots)
    ids   = data.kakaocloud_volume_snapshots.all.volume_snapshots[*].id
    names = data.kakaocloud_volume_snapshots.all.volume_snapshots[*].name
  }
}

# Output filtered volume snapshots
output "filtered_volume_snapshots" {
  description = "List of filtered volume snapshots"
  value = {
    count = length(data.kakaocloud_volume_snapshots.filtered.volume_snapshots)
    ids   = data.kakaocloud_volume_snapshots.filtered.volume_snapshots[*].id
    names = data.kakaocloud_volume_snapshots.filtered.volume_snapshots[*].name
  }
}
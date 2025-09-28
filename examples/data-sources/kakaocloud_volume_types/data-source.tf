# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# List all volume types
data "kakaocloud_volume_types" "all" {
  # No filters supported - returns all available volume types
}

# Output all volume types
output "all_volume_types" {
  description = "List of all volume types"
  value = {
    count = length(data.kakaocloud_volume_types.all.volume_types)
    ids   = data.kakaocloud_volume_types.all.volume_types[*].id
    names = data.kakaocloud_volume_types.all.volume_types[*].name
    descriptions = data.kakaocloud_volume_types.all.volume_types[*].description
  }
}
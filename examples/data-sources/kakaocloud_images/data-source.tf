# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# List all images
data "kakaocloud_images" "all" {
  # No filters - get all images
}

# List images with comprehensive filters
data "kakaocloud_images" "filtered" {
  filter = [
    {
      name  = "name"
      value = "your-image-name"  # Replace with your image name
    },
    {
      name  = "id"
      value = "your-image-id"  # Replace with your image ID
    },
    {
      name  = "image_type"
      value = "snapshot"  # snapshot, backup, custom, etc.
    },
    {
      name  = "instance_type"
      value = "COMPUTE"  # COMPUTE, MEMORY, STORAGE, GPU
    },
    {
      name  = "size"
      value = "1073741824"  # Replace with image size in bytes
    },
    {
      name  = "min_disk"
      value = "20"  # Replace with minimum disk size in GB
    },
    {
      name  = "disk_format"
      value = "qcow2"  # Replace with disk format (qcow2, raw, etc.)
    },
    {
      name  = "status"
      value = "active"  # Replace with image status
    },
    {
      name  = "visibility"
      value = "public"  # public, private, shared
    },
    {
      name  = "os_type"
      value = "LINUX"  # LINUX, WINDOWS
    },
    {
      name  = "image_member_status"
      value = "accepted"  # accepted, pending, rejected
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

# Output all images
output "all_images" {
  description = "List of all images"
  value = {
    count = length(data.kakaocloud_images.all.images)
    ids   = data.kakaocloud_images.all.images[*].id
    names = data.kakaocloud_images.all.images[*].name
  }
}

# Output filtered images
output "filtered_images" {
  description = "List of filtered images"
  value = {
    count = length(data.kakaocloud_images.filtered.images)
    ids   = data.kakaocloud_images.filtered.images[*].id
    names = data.kakaocloud_images.filtered.images[*].name
  }
}

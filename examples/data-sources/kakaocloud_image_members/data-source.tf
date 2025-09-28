# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# List image members for a specific image
data "kakaocloud_image_members" "example" {
  image_id = "your-image-id-here"  # Replace with your image ID
}

# Output image members for the first image
output "image_members_example" {
  description = "List of image members for the example image"
  value = {
    image_id = data.kakaocloud_image_members.example.image_id
    count = length(data.kakaocloud_image_members.example.members)
    members = data.kakaocloud_image_members.example.members
  }
}

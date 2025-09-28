# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# Get a specific image by ID
data "kakaocloud_image" "example" {
  id = "your-image-id-here"  # Replace with your image ID
}

# Output the image information
output "image_example" {
  description = "Information about the example image"
  value = kakaocloud_image.example
}

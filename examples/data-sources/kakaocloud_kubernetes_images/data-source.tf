# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# List all Kubernetes images
data "kakaocloud_kubernetes_images" "all" {
  # No filters - get all Kubernetes images
}

# List Kubernetes images with comprehensive filters
data "kakaocloud_kubernetes_images" "filtered" {
  filter = [
    {
      name  = "os_distro"
      value = "ubuntu"  # Replace with OS distribution
    },
    {
      name  = "instance_type"
      value = "COMPUTE"  # COMPUTE, MEMORY, STORAGE, GPU
    },
    {
      name  = "is_gpu_type"
      value = "false"  # true, false
    },
    {
      name  = "k8s_version"
      value = "1.28.0"  # Replace with your Kubernetes version
    }
  ]
}

# Output all Kubernetes images
output "all_kubernetes_images" {
  description = "List of all Kubernetes images"
  value = {
    count = length(data.kakaocloud_kubernetes_images.all.images)
    ids   = data.kakaocloud_kubernetes_images.all.images[*].id
    names = data.kakaocloud_kubernetes_images.all.images[*].name
  }
}

# Output filtered Kubernetes images
output "filtered_kubernetes_images" {
  description = "List of filtered Kubernetes images"
  value = {
    count = length(data.kakaocloud_kubernetes_images.filtered.images)
    ids   = data.kakaocloud_kubernetes_images.filtered.images[*].id
    names = data.kakaocloud_kubernetes_images.filtered.images[*].name
  }
}

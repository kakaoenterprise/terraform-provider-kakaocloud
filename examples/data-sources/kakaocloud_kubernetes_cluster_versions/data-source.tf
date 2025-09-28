# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# List all available Kubernetes cluster versions
data "kakaocloud_kubernetes_cluster_versions" "all" {
  # No parameters - returns all available Kubernetes versions
}

# Output all available Kubernetes cluster versions
output "all_kubernetes_cluster_versions" {
  description = "List of all available Kubernetes cluster versions"
  value = {
    count = length(data.kakaocloud_kubernetes_cluster_versions.all.versions)
    versions = data.kakaocloud_kubernetes_cluster_versions.all.versions
  }
}
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# List all Kubernetes clusters
data "kakaocloud_kubernetes_clusters" "all" {
  # No filters supported - returns all available Kubernetes clusters
}

# Output all Kubernetes clusters
output "all_kubernetes_clusters" {
  description = "List of all Kubernetes clusters"
  value = {
    count = length(data.kakaocloud_kubernetes_clusters.all.clusters)
    ids   = data.kakaocloud_kubernetes_clusters.all.clusters[*].id
    names = data.kakaocloud_kubernetes_clusters.all.clusters[*].name
    versions = data.kakaocloud_kubernetes_clusters.all.clusters[*].version
    statuses = data.kakaocloud_kubernetes_clusters.all.clusters[*].status
  }
}

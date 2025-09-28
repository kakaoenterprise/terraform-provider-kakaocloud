# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# List all node pools for a specific cluster
data "kakaocloud_kubernetes_node_pools" "example" {
  cluster_name = "your-cluster-name-here"  # Replace with your cluster name
}

# Output node pools for the cluster
output "kubernetes_node_pools_example" {
  description = "List of node pools for the example cluster"
  value = {
    cluster_name = data.kakaocloud_kubernetes_node_pools.example.cluster_name
    count = length(data.kakaocloud_kubernetes_node_pools.example.node_pools)
    node_pools = data.kakaocloud_kubernetes_node_pools.example.node_pools
  }
}

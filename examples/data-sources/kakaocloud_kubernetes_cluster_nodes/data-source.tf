# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# List all nodes in a Kubernetes cluster
data "kakaocloud_kubernetes_cluster_nodes" "all_cluster_nodes" {
  cluster_name = "your-cluster-name-here"  # Replace with your cluster name
}

# List nodes in a specific node pool
data "kakaocloud_kubernetes_cluster_nodes" "node_pool_nodes" {
  cluster_name   = "your-cluster-name-here"  # Replace with your cluster name
  node_pool_name = "your-node-pool-name-here"  # Replace with your node pool name
}

# Output all cluster nodes
output "all_cluster_nodes" {
  description = "List of all nodes in the cluster"
  value = {
    cluster_name = data.kakaocloud_kubernetes_cluster_nodes.all_cluster_nodes.cluster_name
    count = length(data.kakaocloud_kubernetes_cluster_nodes.all_cluster_nodes.nodes)
    nodes = data.kakaocloud_kubernetes_cluster_nodes.all_cluster_nodes.nodes
  }
}

# Output node pool nodes
output "node_pool_nodes" {
  description = "List of nodes in the specific node pool"
  value = {
    cluster_name = data.kakaocloud_kubernetes_cluster_nodes.node_pool_nodes.cluster_name
    node_pool_name = data.kakaocloud_kubernetes_cluster_nodes.node_pool_nodes.node_pool_name
    count = length(data.kakaocloud_kubernetes_cluster_nodes.node_pool_nodes.nodes)
    nodes = data.kakaocloud_kubernetes_cluster_nodes.node_pool_nodes.nodes
  }
}


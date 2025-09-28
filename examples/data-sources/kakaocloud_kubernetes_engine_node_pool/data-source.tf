# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# Get a specific node pool by cluster name and node pool name
data "kakaocloud_kubernetes_engine_node_pool" "example" {
  cluster_name = "your-cluster-name-here"  # Replace with your cluster name
  name         = "your-node-pool-name-here"  # Replace with your node pool name
}

# Output the node pool information
output "node_pool_example" {
  description = "Information about the example node pool"
  value = kakaocloud_kubernetes_engine_node_pool.example
}

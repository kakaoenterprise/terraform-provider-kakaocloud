# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# Get a specific Kubernetes cluster by name
data "kakaocloud_kubernetes_engine_cluster" "example" {
  name = "your-cluster-name-here"  # Replace with your cluster name
}

# Output the cluster information
output "cluster_example" {
  description = "Information about the example Kubernetes cluster"
  value = kakaocloud_kubernetes_engine_cluster.example
}

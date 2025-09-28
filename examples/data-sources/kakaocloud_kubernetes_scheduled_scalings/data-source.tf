# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# List scheduled scalings for a specific node pool
data "kakaocloud_kubernetes_scheduled_scalings" "example" {
  cluster_name   = "your-cluster-name-here"  # Replace with your cluster name
  node_pool_name = "your-node-pool-name-here"  # Replace with your node pool name
}
# Output scheduled scalings for the first node pool
output "scheduled_scalings_example" {
  description = "List of scheduled scalings for the example node pool"
  value = {
    cluster_name = data.kakaocloud_kubernetes_scheduled_scalings.example.cluster_name
    node_pool_name = data.kakaocloud_kubernetes_scheduled_scalings.example.node_pool_name
    count = length(data.kakaocloud_kubernetes_scheduled_scalings.example.scheduled_scaling)
    scheduled_scalings = data.kakaocloud_kubernetes_scheduled_scalings.example.scheduled_scaling
  }
}

# Output specific scheduled scaling details
output "scheduled_scaling_details" {
  description = "Details of scheduled scalings"
  value = {
    names = data.kakaocloud_kubernetes_scheduled_scalings.example.scheduled_scaling[*].name
    schedules = data.kakaocloud_kubernetes_scheduled_scalings.example.scheduled_scaling[*].schedule
    schedule_types = data.kakaocloud_kubernetes_scheduled_scalings.example.scheduled_scaling[*].schedule_type
    desired_nodes = data.kakaocloud_kubernetes_scheduled_scalings.example.scheduled_scaling[*].desired_nodes
    start_times = data.kakaocloud_kubernetes_scheduled_scalings.example.scheduled_scaling[*].start_time
  }
}

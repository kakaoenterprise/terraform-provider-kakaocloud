# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# Get kubeconfig for a specific cluster
data "kakaocloud_kubernetes_kubeconfig" "example" {
  cluster_name = "your-cluster-name-here"  # Replace with your cluster name
}

# Output kubeconfig for the first cluster
output "kubeconfig_example" {
  description = "Kubeconfig for the example cluster"
  value = {
    cluster_name = data.kakaocloud_kubernetes_kubeconfig.example.cluster_name
    kubeconfig_yaml = data.kakaocloud_kubernetes_kubeconfig.example.kubeconfig_yaml
    api_version = data.kakaocloud_kubernetes_kubeconfig.example.api_version
    kind = data.kakaocloud_kubernetes_kubeconfig.example.kind
    current_context = data.kakaocloud_kubernetes_kubeconfig.example.current_context
  }
}
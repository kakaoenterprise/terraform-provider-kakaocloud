# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# List all available load balancer flavors
data "kakaocloud_load_balancer_flavors" "all" {
  # No filters supported - returns all available flavors
}

# Output all load balancer flavors
output "all_load_balancer_flavors" {
  description = "List of all available load balancer flavors"
  value = {
    count = length(data.kakaocloud_load_balancer_flavors.all.flavors)
    ids   = data.kakaocloud_load_balancer_flavors.all.flavors[*].id
    names = data.kakaocloud_load_balancer_flavors.all.flavors[*].name
  }
}

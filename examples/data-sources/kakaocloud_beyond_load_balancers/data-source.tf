# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# List all beyond load balancers
data "kakaocloud_beyond_load_balancers" "all" {
  # No filters - get all beyond load balancers
}

# List beyond load balancers with filters
data "kakaocloud_beyond_load_balancers" "filtered" {
  filter = [
    {
      name  = "id"
      value = "your-beyond-load-balancer-id"  # Replace with your beyond load balancer ID
    },
    {
      name  = "name"
      value = "your-beyond-load-balancer-name"  # Replace with your beyond load balancer name
    },
    {
      name  = "dns_name"
      value = "your-beyond-load-balancer-dns-name"  # Replace with your beyond load balancer DNS name
    },
    {
      name  = "scheme"
      value = "internal"
    },
    {
      name  = "provisioning_status"
      value = "ACTIVE"
    },
    {
      name  = "operating_status"
      value = "ONLINE"
    },
    {
      name  = "type"
      value = "ALB"
    },
    {
      name  = "vpc_id"
      value = "your-vpc-id-here"  # Replace with your VPC ID
    },
    {
      name  = "vpc_name"
      value = "your-vpc-name"  # Replace with your VPC name
    },
    {
      name  = "created_at"
      value = "2021-01-01T00:00:00Z"  # Replace with your created at
    },
    {
      name  = "updated_at"
      value = "2021-01-01T00:00:00Z"  # Replace with your updated at
    }
  ]
}

# Output beyond load balancers
output "beyond_load_balancers" {
  description = "List of beyond load balancers"
  value = data.kakaocloud_beyond_load_balancers.all
}

# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# List all load balancers
data "kakaocloud_load_balancers" "all" {
  # No filters - get all load balancers
}

# List load balancers with filters
data "kakaocloud_load_balancers" "filtered" {
  filter = [
    {
      name  = "id"
      value = "your-load-balancer-id-here"  # Replace with your load balancer ID
    },
    {
      name  = "name"
      value = "your-load-balancer-name"  # Replace with your load balancer name
    },
    {
      name  = "private_vip"
      value = "your-private-vip-here"  # Replace with your private VIP
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
      name  = "subnet_id"
      value = "your-subnet-id-here"  # Replace with your subnet ID
    },
    {
      name  = "subnet_cidr_block"
      value = "your-subnet-cidr-block-here"  # Replace with your subnet CIDR block
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
      name  = "availability_zone"
      value = "your-availability-zone-here"  # Replace with your availability zone
    },
    {
      name  = "beyond_load_balancer_name"
      value = "your-beyond-load-balancer-name"  # Replace with your beyond load balancer name
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

# Output all load balancers
output "all_load_balancers" {
  description = "List of all load balancers"
  value = data.kakaocloud_load_balancers.all
}

# Output filtered load balancers
output "filtered_load_balancers" {
  description = "List of filtered load balancers"
  value = data.kakaocloud_load_balancers.filtered
}
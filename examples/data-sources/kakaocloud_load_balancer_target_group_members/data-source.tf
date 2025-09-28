# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# List all target group members for a specific target group
data "kakaocloud_load_balancer_target_group_members" "example" {
  target_group_id = "your-target-group-id-here"  # Replace with your target group ID
}

# List target group members with filters
data "kakaocloud_load_balancer_target_group_members" "filtered" {
  target_group_id = "your-target-group-id-here"  # Replace with your target group ID
  
  filter = [
    {
      name  = "id"
      value = "your-target-group-member-id-here"  # Replace with your target group member ID
    },
    {
      name  = "name"
      value = "your-target-group-member-name-here"  # Replace with your target group member name
    },
    {
      name  = "protocol"
      value = "HTTP"
    },
    {
      name = "availability_zone"
      value = "your-availability-zone-here"  # Replace with your availability zone
    },
    {
      name = "load_balancer_algorithm"
      value = "ROUND_ROBIN"
    },
    {
      name = "load_balancer_name"
      value = "your-load-balancer-name-here"  # Replace with your load balancer name
    },
    {
      name = "load_balancer_id"
      value = "your-load-balancer-id-here"  # Replace with your load balancer ID
    },
    {
      name = "listener_protocol"
      value = "HTTP"
    },
    {
      name = "vpc_name"
      value = "your-vpc-name-here"  # Replace with your VPC name
    },
    {
      name = "vpc_id"
      value = "your-vpc-id-here"  # Replace with your VPC ID
    },
    {
      name = "subnet_name"
      value = "your-subnet-name-here"  # Replace with your subnet name
    },
    {
      name = "subnet_id"
      value = "your-subnet-id-here"  # Replace with your subnet ID
    },
    {
      name = "health_monitor_id"
      value = "your-health-monitor-id-here"  # Replace with your health monitor ID
    },
    {
      name = "created_at"
      value = "2021-01-01T00:00:00Z"  # Replace with your created at
    },
    {
      name = "updated_at"
      value = "2021-01-01T00:00:00Z"  # Replace with your updated at
    }
  ]
}

# List target group members by instance
data "kakaocloud_load_balancer_target_group_members" "by_instance" {
  target_group_id = "your-target-group-id-here"  # Replace with your target group ID
  
  filter = [
    {
      name  = "instance_id"
      value = "your-instance-id-here"  # Replace with your instance ID
    },
    {
      name  = "vpc_id"
      value = "your-vpc-id-here"  # Replace with your VPC ID
    },
    {
      name  = "operating_status"
      value = "ONLINE"
    }
  ]
}

# Output all target group members
output "target_group_members" {
  description = "List of target group members"
  value = {
    count = length(data.kakaocloud_load_balancer_target_group_members.example.members)
    ids   = data.kakaocloud_load_balancer_target_group_members.example.members[*].id
    names = data.kakaocloud_load_balancer_target_group_members.example.members[*].name
    addresses = data.kakaocloud_load_balancer_target_group_members.example.members[*].address
  }
}

# Output filtered target group members
output "filtered_target_group_members" {
  description = "List of filtered target group members"
  value = {
    count = length(data.kakaocloud_load_balancer_target_group_members.filtered.members)
    ids   = data.kakaocloud_load_balancer_target_group_members.filtered.members[*].id
    names = data.kakaocloud_load_balancer_target_group_members.filtered.members[*].name
    addresses = data.kakaocloud_load_balancer_target_group_members.filtered.members[*].address
  }
}

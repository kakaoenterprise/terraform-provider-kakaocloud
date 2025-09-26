# Copyright (c) HashiCorp, Inc.

# List all instances or filtered instances.
data "kakaocloud_instances" "example_list" {
  filter = [
    {
      name  = "name"
      value = "example"
    }
  ]
}
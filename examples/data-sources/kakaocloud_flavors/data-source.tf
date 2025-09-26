# Copyright (c) HashiCorp, Inc.

# List all flavors or filtered flavors.
data "kakaocloud_flavors" "example_list" {
  filter = [
    {
      name  = "name"
      value = "example"
    }
  ]
}
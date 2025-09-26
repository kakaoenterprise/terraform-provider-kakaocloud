# Copyright (c) HashiCorp, Inc.

# List all volume snapshots or filtered volume snapshots.
data "kakaocloud_volume_snapshots" "example_list" {
  filter = [
    {
      name  = "name"
      value = "example"
    }
  ]
}
# Copyright (c) HashiCorp, Inc.

# Manage example volume snapshot.
resource "kakaocloud_volume_snapshot" "example" {
  volume_id      = "xxxxxx-yyyy-zzzz-81e8-836ebe461ba6"
  name           = "snapshot_example"
  is_incremental = false
}
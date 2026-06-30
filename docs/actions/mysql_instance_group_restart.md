---
page_title: "kakaocloud_mysql_instance_group_restart Action - kakaocloud"
subcategory: "MySQL"
description: |-
  The kakaocloud_mysql_instance_group_restart action restarts selected instances in a KakaoCloud MySQL instance group.
---

# kakaocloud_mysql_instance_group_restart (Action)

The `kakaocloud_mysql_instance_group_restart` action restarts selected instances in a KakaoCloud MySQL instance group.

Use this action to restart one or more MySQL instances by instance ID.

-> **Note:**  - After a restart, the primary and standby roles may change. <br/> - Check the current `primary_subnet_info` and `standby_subnet_info` values on the related `kakaocloud_mysql_instance_group` resource. <br/> - Manually update the Terraform `.tf` configuration to match the current KakaoCloud placement.

## Example Usage

```hcl
action "kakaocloud_mysql_instance_group_restart" "example" {
  config {
    instance_group_id = "<your-mysql-instance-group-id>"
    instance_ids      = ["<your-mysql-instance-id>"]
  }
}
```

## Argument Reference

- `instance_group_id` (Required, String) MySQL instance group ID.
- `instance_ids` (Required, Set of String) IDs of the MySQL instances to restart.

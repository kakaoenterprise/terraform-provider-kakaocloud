---
page_title: "kakaocloud_mysql_instance_group_switchover Action - kakaocloud"
subcategory: "MySQL"
description: |-
  The kakaocloud_mysql_instance_group_switchover action performs a switchover for a KakaoCloud MySQL instance group.
---

# kakaocloud_mysql_instance_group_switchover (Action)

The `kakaocloud_mysql_instance_group_switchover` action performs a switchover for a KakaoCloud MySQL instance group.

Use this action to promote a standby instance and change the primary role during planned operations.

-> **Note:**  - After a switchover, the primary and standby roles are exchanged. <br/> - Check the current `primary_subnet_info` and `standby_subnet_info` values on the related `kakaocloud_mysql_instance_group` resource. <br/> - Manually update the Terraform `.tf` configuration to match the current KakaoCloud placement.

## Example Usage

```hcl
action "kakaocloud_mysql_instance_group_switchover" "example" {
  config {
    instance_group_id = "<your-mysql-instance-group-id>"
  }
}
```

## Argument Reference

- `instance_group_id` (Required, String) MySQL instance group ID.

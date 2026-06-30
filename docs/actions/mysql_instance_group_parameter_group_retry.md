---
page_title: "kakaocloud_mysql_instance_group_parameter_group_retry Action - kakaocloud"
subcategory: "MySQL"
description: |-
  The kakaocloud_mysql_instance_group_parameter_group_retry action retries parameter group application for a KakaoCloud MySQL instance group.
---

# kakaocloud_mysql_instance_group_parameter_group_retry (Action)

The `kakaocloud_mysql_instance_group_parameter_group_retry` action retries parameter group application for a KakaoCloud
MySQL instance group.

Use this action when a previous parameter group apply operation needs to be retried.

## Example Usage

```hcl
action "kakaocloud_mysql_instance_group_parameter_group_retry" "example" {
  config {
    instance_group_id = "<your-mysql-instance-group-id>"
  }
}
```

## Argument Reference

- `instance_group_id` (Required, String) MySQL instance group ID.

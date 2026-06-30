---
page_title: "kakaocloud_mysql_instance_group_scale_in Action - kakaocloud"
subcategory: "MySQL"
description: |-
  The kakaocloud_mysql_instance_group_scale_in action removes selected standby instances from a KakaoCloud MySQL instance group.
---

# kakaocloud_mysql_instance_group_scale_in (Action)

The `kakaocloud_mysql_instance_group_scale_in` action removes selected standby instances from a KakaoCloud MySQL
instance group.

Use this action when you need to scale in a MySQL instance group by specifying the instance IDs to remove.

-> **Note:**  - After scale-in, manually reduce `replicas` in the Terraform `.tf` configuration for the related `kakaocloud_mysql_instance_group` resource. <br/> - Reduce `replicas` by the number of instances removed from the scaled-in subnet.

## Example Usage

```hcl
action "kakaocloud_mysql_instance_group_scale_in" "example" {
  config {
    instance_group_id = "<your-mysql-instance-group-id>"
    instance_ids      = ["<your-standby-mysql-instance-id>"]
  }
}
```

## Argument Reference

- `instance_group_id` (Required, String) MySQL instance group ID.
- `instance_ids` (Required, Set of String) IDs of the standby MySQL instances to remove.

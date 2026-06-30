---
page_title: "kakaocloud_mysql_instance_group_restorable_time Data Source - kakaocloud"
subcategory: "MySQL"
description: |-
  The kakaocloud_mysql_instance_group_restorable_time data source retrieves the restorable time range for a KakaoCloud MySQL instance group.
---

# kakaocloud_mysql_instance_group_restorable_time (Data Source)

The `kakaocloud_mysql_instance_group_restorable_time` data source retrieves the restorable time range for a KakaoCloud
MySQL instance group.

Use this data source before point-in-time restore operations to confirm the earliest and latest restore points available
for an instance group.

## Example Usage

```hcl
# Get the restorable time range for a MySQL instance group
data "kakaocloud_mysql_instance_group_restorable_time" "example" {
  instance_group_id = "<your-mysql-instance-group-id>"
}

output "mysql_restorable_time" {
  value = {
    from_time = data.kakaocloud_mysql_instance_group_restorable_time.example.restorable_time.from_time
    to_time   = data.kakaocloud_mysql_instance_group_restorable_time.example.restorable_time.to_time
  }
}
```

## Argument Reference

- `instance_group_id` (Required, String) MySQL instance group ID.

- `timeouts` (Optional, Attributes) Custom timeout settings. (See [below for nested schema](#nestedatt--timeouts).)

## Attribute Reference

The following attributes are exported:

- `restorable_time` (Attributes) Available point-in-time restore range. (See [below for nested schema](#nestedatt--restorable_time).)

<a id="nestedatt--timeouts"></a>
### Nested Schema for `timeouts`

- `read` (Optional, String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).


<a id="nestedatt--restorable_time"></a>
### Nested Schema for `restorable_time`

- `from_time` (String) Earliest restorable time.
- `to_time` (String) Latest restorable time.

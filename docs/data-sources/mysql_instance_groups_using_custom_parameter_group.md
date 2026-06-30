---
page_title: "kakaocloud_mysql_instance_groups_using_custom_parameter_group Data Source - kakaocloud"
subcategory: "MySQL"
description: |-
  The kakaocloud_mysql_instance_groups_using_custom_parameter_group data source retrieves KakaoCloud MySQL instance groups using a custom parameter group.
---

# kakaocloud_mysql_instance_groups_using_custom_parameter_group (Data Source)

The `kakaocloud_mysql_instance_groups_using_custom_parameter_group` data source retrieves KakaoCloud MySQL instance
groups that use a specific custom parameter group.

Use this data source to understand which instance groups are affected before updating or deleting a custom parameter
group.

## Example Usage

```hcl
# List instance groups using a custom parameter group
data "kakaocloud_mysql_instance_groups_using_custom_parameter_group" "example" {
  custom_parameter_group_id = "<your-mysql-custom-parameter-group-id>"
}

output "mysql_instance_groups_using_custom_parameter_group" {
  value = [
    for group in data.kakaocloud_mysql_instance_groups_using_custom_parameter_group.example.instance_groups : group.id
  ]
}
```

## Argument Reference

- `custom_parameter_group_id` (Required, String) Custom parameter group ID.

- `timeouts` (Optional, Attributes) Custom timeout settings. (See [below for nested schema](#nestedatt--timeouts).)

## Attribute Reference

The following attributes are exported:

- `instance_groups` (Attributes List) MySQL instance groups using the custom parameter group. (See [below for nested schema](#nestedatt--instance_groups).)

<a id="nestedatt--timeouts"></a>
### Nested Schema for `timeouts`

- `read` (Optional, String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).


<a id="nestedatt--instance_groups"></a>
### Nested Schema for `instance_groups`

- `engine_version` (String) MySQL engine version.
- `flavor_id` (String) Flavor ID used by the instance group.
- `id` (String) MySQL instance group ID.
- `instance_group_type` (String) MySQL instance group type.
- `is_multi_az` (Boolean) Whether the instance group uses multiple availability zones.
- `name` (String) MySQL instance group name.
- `parameter_group_status` (String) Parameter group apply status for the instance group.
- `status` (String) MySQL instance group status.

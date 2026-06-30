---
page_title: "kakaocloud_mysql_custom_parameter_groups Data Source - kakaocloud"
subcategory: "MySQL"
description: |-
  The kakaocloud_mysql_custom_parameter_groups data source retrieves KakaoCloud MySQL custom parameter groups.
---

# kakaocloud_mysql_custom_parameter_groups (Data Source)

The `kakaocloud_mysql_custom_parameter_groups` data source retrieves KakaoCloud MySQL custom parameter groups in the
current project.

Use this data source to discover custom parameter groups and select one for a MySQL instance group configuration.

## Example Usage

```hcl
# List custom parameter groups
data "kakaocloud_mysql_custom_parameter_groups" "example" {
}

output "mysql_custom_parameter_group_ids" {
  value = [
    for group in data.kakaocloud_mysql_custom_parameter_groups.example.custom_parameter_groups : group.id
  ]
}
```

## Argument Reference

- `timeouts` (Optional, Attributes) Custom timeout settings. (See [below for nested schema](#nestedatt--timeouts).)

## Attribute Reference

The following attributes are exported:

- `custom_parameter_groups` (Attributes List) List of custom parameter groups. (See [below for nested schema](#nestedatt--custom_parameter_groups).)

<a id="nestedatt--timeouts"></a>
### Nested Schema for `timeouts`

- `read` (Optional, String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).


<a id="nestedatt--custom_parameter_groups"></a>
### Nested Schema for `custom_parameter_groups`

- `default_parameter_group_id` (String) ID of the default parameter group associated with this custom parameter group.
- `description` (String) Custom parameter group description.
- `engine_version` (String) MySQL engine version of the parameter group.
- `exist_error_sync` (Boolean) Whether an error exists while synchronizing parameter values.
- `id` (String) Custom parameter group ID.
- `instance_group_count` (Number) Number of MySQL instance groups using this parameter group.
- `is_rollback_possible` (Boolean) Whether the latest parameter change can be rolled back.
- `name` (String) Custom parameter group name.

---
page_title: "kakaocloud_mysql_default_parameter_groups Data Source - kakaocloud"
subcategory: "MySQL"
description: |-
  The kakaocloud_mysql_default_parameter_groups data source retrieves KakaoCloud MySQL default parameter groups.
---

# kakaocloud_mysql_default_parameter_groups (Data Source)

The `kakaocloud_mysql_default_parameter_groups` data source retrieves KakaoCloud MySQL default parameter groups.

Use this data source to discover available default parameter groups by engine version and use them as sources for
custom parameter groups or MySQL instance groups.

## Example Usage

```hcl
# List default parameter groups
data "kakaocloud_mysql_default_parameter_groups" "example" {
}

output "mysql_default_parameter_group_ids" {
  value = [
    for group in data.kakaocloud_mysql_default_parameter_groups.example.default_parameter_groups : group.id
  ]
}
```

## Argument Reference

- `timeouts` (Optional, Attributes) Custom timeout settings. (See [below for nested schema](#nestedatt--timeouts).)

## Attribute Reference

The following attributes are exported:

- `default_parameter_groups` (Attributes List) List of default parameter groups. (See [below for nested schema](#nestedatt--default_parameter_groups).)

<a id="nestedatt--timeouts"></a>
### Nested Schema for `timeouts`

- `read` (Optional, String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).


<a id="nestedatt--default_parameter_groups"></a>
### Nested Schema for `default_parameter_groups`

- `description` (String) Default parameter group description.
- `engine_version` (String) MySQL engine version of the parameter group.
- `exist_engine_version_mismatch` (Boolean) Whether any associated instance group has an engine version mismatch.
- `exist_error_sync` (Boolean) Whether an error exists while synchronizing parameter values.
- `id` (String) Default parameter group ID.
- `instance_group_count` (Number) Number of MySQL instance groups using this parameter group.
- `name` (String) Default parameter group name.

---
page_title: "kakaocloud_mysql_custom_parameter_group Resource - kakaocloud"
subcategory: "MySQL"
description: |-
  The kakaocloud_mysql_custom_parameter_group resource allows you to create and manage custom parameter groups for KakaoCloud MySQL.
---

# kakaocloud_mysql_custom_parameter_group (Resource)

The `kakaocloud_mysql_custom_parameter_group` resource allows you to create and manage custom parameter groups for
KakaoCloud MySQL.

Use this resource to derive a custom parameter group from an existing default or custom parameter group, and to manage
parameter overrides that are applied to MySQL instance groups.

-> **Note:** <br/> - `apply_mode` is sent to the API only when parameter changes are requested. Setting or changing only `apply_mode` does not call the remote update API, but Terraform may still show a plan diff to align state and configuration. To reduce unnecessary plan diffs, configure `apply_mode` only when parameter changes are needed. <br/> - `parameter_overrides = []` resets all parameter overrides. `parameter_overrides = null` means Terraform does not manage overrides and does not reset them.

## Example Usage

```hcl
# kakaocloud_mysql_custom_parameter_group Terraform Resource Example

data "kakaocloud_mysql_engine_versions" "example" {}
data "kakaocloud_mysql_default_parameter_groups" "example" {}

# Basic Usage (kakaocloud_mysql_custom_parameter_group)
resource "kakaocloud_mysql_custom_parameter_group" "example" {
  name        = "example"
  description = "Example MySQL Custom Parameter Group"
  source_parameter_group_id = one([for group in data.kakaocloud_mysql_default_parameter_groups.example.default_parameter_groups : group.id
    if group.engine_version == one([for version in data.kakaocloud_mysql_engine_versions.example.engine_versions : version.engine_version if version.engine_version == "8.4.8"])
  ])
  source_parameter_group_type = "DEFAULT"
}
```

## Argument Reference

- `name` (Required, String) Custom parameter group name.
- `source_parameter_group_id` (Required, String) ID of the source parameter group to copy from.
- `source_parameter_group_type` (Required, String) Type of the source parameter group. (Possible values: `DEFAULT`, `CUSTOM`)

- `apply_mode` (Optional, String) How parameter changes are applied to associated MySQL instance groups. This is sent to the API only when parameter changes are requested. (Possible values: `SEQUENTIAL`, `PARALLEL`)
- `description` (Optional, String) Custom parameter group description.
- `parameter_overrides` (Optional, Attributes Set) Parameter values to override from the source parameter group. Set `[]` to reset all overrides, or `null` to leave override management disabled. (See [below for nested schema](#nestedatt--parameter_overrides).)
- `timeouts` (Optional, Attributes) Custom timeout settings. (See [below for nested schema](#nestedatt--timeouts).)

## Attribute Reference

- `default_parameter_group_id` (String) ID of the default parameter group associated with this custom parameter group.
- `engine_version` (String) MySQL engine version of the parameter group.
- `exist_error_sync` (Boolean) Whether an error exists while synchronizing parameter values.
- `id` (String) Custom parameter group ID.
- `instance_group_count` (Number) Number of MySQL instance groups using this parameter group.
- `is_rollback_possible` (Boolean) Whether the latest parameter change can be rolled back.

<a id="nestedatt--parameter_overrides"></a>
### Nested Schema for `parameter_overrides`

- `key` (Required, String) Parameter key.

- `value` (Optional, String) Parameter value to apply.


<a id="nestedatt--timeouts"></a>
### Nested Schema for `timeouts`

- `create` (Optional, String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).
- `delete` (Optional, String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Setting a timeout for a Delete operation is only applicable if changes are saved into state before the destroy operation occurs.
- `read` (Optional, String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Read operations occur during any refresh or planning operation when refresh is enabled.
- `update` (Optional, String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).


## Import

Import is supported using the following syntax:

The [`terraform import` command](https://developer.hashicorp.com/terraform/cli/commands/import) can be used, for
example:

```shell
$ terraform import kakaocloud_mysql_custom_parameter_group.example <mysql_custom_parameter_group_id>
```

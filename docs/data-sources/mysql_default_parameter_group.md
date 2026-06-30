---
page_title: "kakaocloud_mysql_default_parameter_group Data Source - kakaocloud"
subcategory: "MySQL"
description: |-
  The kakaocloud_mysql_default_parameter_group data source retrieves detailed information about a specific KakaoCloud MySQL default parameter group.
---

# kakaocloud_mysql_default_parameter_group (Data Source)

The `kakaocloud_mysql_default_parameter_group` data source retrieves detailed information about a specific KakaoCloud
MySQL default parameter group by its ID.

Use this data source to inspect default parameters before creating a custom parameter group or selecting a parameter
group for an instance group.

## Example Usage

```hcl
# Get a specific default parameter group by ID
data "kakaocloud_mysql_default_parameter_group" "example" {
  id = "<your-mysql-default-parameter-group-id>"
}

output "mysql_default_parameter_group_info" {
  value = {
    id              = data.kakaocloud_mysql_default_parameter_group.example.id
    name            = data.kakaocloud_mysql_default_parameter_group.example.name
    engine_version  = data.kakaocloud_mysql_default_parameter_group.example.engine_version
    parameter_count = length(data.kakaocloud_mysql_default_parameter_group.example.parameters)
  }
}
```

## Argument Reference

- `id` (Required, String) Default parameter group ID.

- `timeouts` (Optional, Attributes) Custom timeout settings. (See [below for nested schema](#nestedatt--timeouts).)

## Attribute Reference

The following attributes are exported:

- `description` (String) Default parameter group description.
- `engine_version` (String) MySQL engine version of the parameter group.
- `exist_engine_version_mismatch` (Boolean) Whether any associated instance group has an engine version mismatch.
- `exist_error_sync` (Boolean) Whether an error exists while synchronizing parameter values.
- `instance_group_count` (Number) Number of MySQL instance groups using this parameter group.
- `name` (String) Default parameter group name.
- `parameters` (Attributes List) Parameters included in the group. (See [below for nested schema](#nestedatt--parameters).)

<a id="nestedatt--timeouts"></a>
### Nested Schema for `timeouts`

- `read` (Optional, String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).


<a id="nestedatt--parameters"></a>
### Nested Schema for `parameters`

- `data_type` (String) Parameter data type.
- `default_parameter_value` (String) Default value for the parameter.
- `is_editable` (Boolean) Whether the parameter can be edited.
- `is_required` (Boolean) Whether the parameter is required.
- `key` (String) Parameter key.
- `parameter_type` (String) Parameter type.
- `validation_value_format` (String) Expected value format for validation.
- `value` (String) Current parameter value.

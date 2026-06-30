---
page_title: "kakaocloud_mysql_custom_parameter_group_events Data Source - kakaocloud"
subcategory: "MySQL"
description: |-
  The kakaocloud_mysql_custom_parameter_group_events data source retrieves events for a KakaoCloud MySQL custom parameter group.
---

# kakaocloud_mysql_custom_parameter_group_events (Data Source)

The `kakaocloud_mysql_custom_parameter_group_events` data source retrieves events for a KakaoCloud MySQL custom
parameter group.

Use this data source to review recent parameter group events and operational history.

## Example Usage

```hcl
# List events for a custom parameter group
data "kakaocloud_mysql_custom_parameter_group_events" "example" {
  custom_parameter_group_id = "<your-mysql-custom-parameter-group-id>"
}

output "mysql_custom_parameter_group_events" {
  value = data.kakaocloud_mysql_custom_parameter_group_events.example.events
}
```

## Argument Reference

- `custom_parameter_group_id` (Required, String) Custom parameter group ID.

- `timeouts` (Optional, Attributes) Custom timeout settings. (See [below for nested schema](#nestedatt--timeouts).)

## Attribute Reference

The following attributes are exported:

- `events` (Attributes List) Events for the custom parameter group. (See [below for nested schema](#nestedatt--events).)

<a id="nestedatt--timeouts"></a>
### Nested Schema for `timeouts`

- `read` (Optional, String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).


<a id="nestedatt--events"></a>
### Nested Schema for `events`

- `created_at` (String) Time when the event was created.
- `description` (String) Event description.
- `name` (String) Event name.

---
page_title: "kakaocloud_mysql_default_parameter_group_events Data Source - kakaocloud"
subcategory: "MySQL"
description: |-
  The kakaocloud_mysql_default_parameter_group_events data source retrieves events for a KakaoCloud MySQL default parameter group.
---

# kakaocloud_mysql_default_parameter_group_events (Data Source)

The `kakaocloud_mysql_default_parameter_group_events` data source retrieves events for a KakaoCloud MySQL default
parameter group.

Use this data source to review parameter group events and identify instance groups affected by default parameter group
changes.

## Example Usage

```hcl
# List events for a default parameter group
data "kakaocloud_mysql_default_parameter_group_events" "example" {
  default_parameter_group_id = "<your-mysql-default-parameter-group-id>"
}

output "mysql_default_parameter_group_events" {
  value = data.kakaocloud_mysql_default_parameter_group_events.example.events
}
```

## Argument Reference

- `default_parameter_group_id` (Required, String) Default parameter group ID.

- `timeouts` (Optional, Attributes) Custom timeout settings. (See [below for nested schema](#nestedatt--timeouts).)

## Attribute Reference

The following attributes are exported:

- `events` (Attributes List) Events for the default parameter group. (See [below for nested schema](#nestedatt--events).)

<a id="nestedatt--timeouts"></a>
### Nested Schema for `timeouts`

- `read` (Optional, String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).


<a id="nestedatt--events"></a>
### Nested Schema for `events`

- `created_at` (String) Time when the event was created.
- `description` (String) Event description.
- `name` (String) Event name.

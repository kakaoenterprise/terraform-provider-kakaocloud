---
page_title: "kakaocloud_mysql_flavors Data Source - kakaocloud"
subcategory: "MySQL"
description: |-
  The kakaocloud_mysql_flavors data source retrieves available KakaoCloud MySQL flavors.
---

# kakaocloud_mysql_flavors (Data Source)

The `kakaocloud_mysql_flavors` data source retrieves available KakaoCloud MySQL flavors.

Use this data source to select a compute shape for MySQL instance group creation. Set `show_all` to `true` when you need
to include deprecated flavors in the result.

## Example Usage

```hcl
# List available MySQL flavors
data "kakaocloud_mysql_flavors" "example" {
  show_all = true
}

output "mysql_flavors" {
  value = [
    for flavor in data.kakaocloud_mysql_flavors.example.flavors : {
      id         = flavor.id
      name       = flavor.name
      vcpus      = flavor.vcpus
      memory_mb  = flavor.memory_mb
      deprecated = flavor.deprecated
    }
  ]
}
```

## Argument Reference

- `show_all` (Optional, Boolean) Whether to include all flavors, including deprecated flavors.
- `timeouts` (Optional, Attributes) Custom timeout settings. (See [below for nested schema](#nestedatt--timeouts).)

## Attribute Reference

The following attributes are exported:

- `flavors` (Attributes List) List of MySQL flavors. (See [below for nested schema](#nestedatt--flavors).)

<a id="nestedatt--timeouts"></a>
### Nested Schema for `timeouts`

- `read` (Optional, String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).


<a id="nestedatt--flavors"></a>
### Nested Schema for `flavors`

- `availability_zones` (List of String) Availability zones where the flavor is available.
- `deprecated` (Boolean) Whether the flavor is deprecated.
- `family` (String) Flavor family.
- `group` (String) Flavor group.
- `id` (String) Flavor ID.
- `memory` (Number) Memory size.
- `memory_mb` (Number) Memory size in MB.
- `name` (String) Flavor name.
- `type` (String) Flavor type.
- `vcpus` (Number) Number of vCPUs.

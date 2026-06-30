---
page_title: "kakaocloud_mysql_engine_versions Data Source - kakaocloud"
subcategory: "MySQL"
description: |-
  The kakaocloud_mysql_engine_versions data source retrieves available KakaoCloud MySQL engine versions.
---

# kakaocloud_mysql_engine_versions (Data Source)

The `kakaocloud_mysql_engine_versions` data source retrieves available KakaoCloud MySQL engine versions.

Use this data source to select an engine version for MySQL instance group creation or to inspect supported licenses.

## Example Usage

```hcl
# List available MySQL engine versions
data "kakaocloud_mysql_engine_versions" "example" {
}

output "mysql_engine_versions" {
  value = [
    for version in data.kakaocloud_mysql_engine_versions.example.engine_versions : version.engine_version
  ]
}
```

## Argument Reference

- `timeouts` (Optional, Attributes) Custom timeout settings. (See [below for nested schema](#nestedatt--timeouts).)

## Attribute Reference

The following attributes are exported:

- `engine_versions` (Attributes List) List of available MySQL engine versions. (See [below for nested schema](#nestedatt--engine_versions).)

<a id="nestedatt--timeouts"></a>
### Nested Schema for `timeouts`

- `read` (Optional, String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).


<a id="nestedatt--engine_versions"></a>
### Nested Schema for `engine_versions`

- `engine_version` (String) MySQL engine version.
- `license` (String) License associated with the engine version.

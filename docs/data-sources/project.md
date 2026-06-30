---
page_title: "kakaocloud_project Data Source - kakaocloud"
subcategory: "Project"
description: |-
  Retrieves information about the current project in KakaoCloud.
---

# kakaocloud_project (Data Source)

The `kakaocloud_project` data source retrieves metadata for the current project in KakaoCloud.

A project represents an isolated administrative and billing unit within a domain.  
Use this data source to reference the project ID, project name, and associated domain information in other Terraform resources without managing the project lifecycle.

## Example Usage

```hcl
data "kakaocloud_project" "example" {
  # Configuration options
}
```

## Argument Reference

- `timeouts` (Optional, Attributes) Read operation timeout configuration. (see [below for nested schema](#nestedatt--timeouts))

## Attribute Reference

The following attributes are exported:

- `domain` (Attributes) Domain information to which the project belongs. (see [below for nested schema](#nestedatt--domain))
- `id` (String) ID of the project.
- `name` (String) Name of the project.

<a id="nestedatt--timeouts"></a>
### Nested Schema for `timeouts`

- `read` (Optional, String) A string that can be parsed as a duration consisting of numbers and unit suffixes, such as `"30s"` or `"2h45m"`.   - Valid time units are `"s"` (seconds), `"m"` (minutes), and `"h"` (hours).


<a id="nestedatt--domain"></a>
### Nested Schema for `domain`

- `id` (String) ID of the domain.
- `name` (String) Name of the domain.

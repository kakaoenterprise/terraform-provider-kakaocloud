---
page_title: "kakaocloud_transit_gateway_shares Data Source - kakaocloud"
subcategory: "Transit Gateway"
description: |-
  Retrieves a list of projects that a Transit Gateway (TGW) is shared with in KakaoCloud.
---

# kakaocloud_transit_gateway_shares (Data Source)

The `kakaocloud_transit_gateway_shares` data source retrieves a list of projects that a specific Transit Gateway (TGW) is shared with in KakaoCloud.

A Transit Gateway can be shared across multiple projects to enable centralized and cross-project networking.  
Use this data source to list shared projects for a given TGW, verify sharing status, and reference project-level metadata without managing the sharing lifecycle.

## Example Usage

```terraform
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# List projects that a Transit Gateway is shared with
data "kakaocloud_transit_gateway_shares" "example" {
  tgw_id = "<your-transit-gateway-id>"
}

# Output all shared projects
output "tgw_shared_projects" {
  description = "Projects that the Transit Gateway is shared with"
  value = [
    for p in data.kakaocloud_transit_gateway_shares.example.shared_projects : {
      id         = p.id
      name       = p.name
      is_enabled = p.is_enabled
      domain_id  = p.domain_id
    }
  ]
}
```

## Argument Reference

- `tgw_id` (Required, String) ID of the Transit Gateway for which shared projects are retrieved.

- `timeouts` (Optional, Attributes) Read operation timeout configuration. (see [below for nested schema](#nestedatt--timeouts))

## Attribute Reference

The following attributes are exported:

- `shared_projects` (Attributes List) List of projects that the Transit Gateway is shared with. (see [below for nested schema](#nestedatt--shared_projects))

<a id="nestedatt--timeouts"></a>
### Nested Schema for `timeouts`

- `read` (Optional, String) Duration to wait for the read operation, such as `"30s"` or `"2h45m"`.


<a id="nestedatt--shared_projects"></a>
### Nested Schema for `shared_projects`

- `created_at` (String) Time when the resource was created<br/> - ISO_8601 format<br/> - UTC standard
- `description` (String) Description of the project.
- `disabled_at` (String) Timestamp when sharing was disabled, if applicable.
- `domain_id` (String) ID of the domain that owns the project.
- `id` (String) Project ID.
- `is_enabled` (Boolean) Whether the Transit Gateway sharing is currently enabled.
- `name` (String) Project name.
- `nickname` (String) Project nickname.

---
page_title: "kakaocloud_transit_gateway_route_table Data Source - kakaocloud"
subcategory: "Transit Gateway"
description: |-
  Retrieves detailed information about a specific Transit Gateway (TGW) route table in KakaoCloud.
---

# kakaocloud_transit_gateway_route_table (Data Source)

The `kakaocloud_transit_gateway_route_table` data source retrieves metadata and configuration details for a specific Transit Gateway (TGW) route table in KakaoCloud.

A TGW route table defines routing rules and attachment associations that control how traffic is forwarded through a Transit Gateway.  
Use this data source to reference an existing TGW route table by its ID and inspect its routes, associations, and default behaviors without managing its lifecycle.

## Example Usage

```terraform
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# Retrieve a specific Transit Gateway route table by ID
data "kakaocloud_transit_gateway_route_table" "example" {
  id = "<your-tgw-route-table-id>"
}

# Output basic route table information
output "tgw_route_table_basic" {
  description = "Basic information of the Transit Gateway route table"
  value = {
    id                                = data.kakaocloud_transit_gateway_route_table.example.id
    name                              = data.kakaocloud_transit_gateway_route_table.example.name
    provisioning_status               = data.kakaocloud_transit_gateway_route_table.example.provisioning_status
    is_default_association_route_table = data.kakaocloud_transit_gateway_route_table.example.is_default_association_route_table
    is_default_propagation_route_table = data.kakaocloud_transit_gateway_route_table.example.is_default_propagation_route_table
  }
}

# Output route table associations
output "tgw_route_table_associations" {
  description = "Attachment associations of the Transit Gateway route table"
  value       = data.kakaocloud_transit_gateway_route_table.example.associations
}

# Output routes configured in the route table
output "tgw_route_table_routes" {
  description = "Routes configured in the Transit Gateway route table"
  value       = data.kakaocloud_transit_gateway_route_table.example.routes
}
```

## Argument Reference

- `id` (Required, String) ID of the Transit Gateway route table to retrieve.

- `timeouts` (Optional, Attributes) Read operation timeout configuration. (see [below for nested schema](#nestedatt--timeouts))

## Attribute Reference

The following attributes are exported:

- `associations` (Attributes List) List of attachment associations applied to the route table. (see [below for nested schema](#nestedatt--associations))
- `created_at` (String) Time when the resource was created<br/> - ISO_8601 format<br/> - UTC standard
- `is_default_association_route_table` (Boolean) Whether this route table is the default association route table.
- `is_default_propagation_route_table` (Boolean) Whether this route table is the default propagation route table.
- `name` (String) Route table name.
- `project_id` (String) Project ID that owns the route table.
- `project_name` (String) Project name that owns the route table.
- `provisioning_status` (String) Current provisioning status of the route table.
- `region` (String) Region where the route table is located.
- `routes` (Attributes List) List of routes configured in the route table. (see [below for nested schema](#nestedatt--routes))
- `tgw_id` (String) Transit Gateway ID associated with the route table.
- `tgw_name` (String) Transit Gateway name.
- `updated_at` (String) Time when the resource was last updated<br/> - ISO_8601 format<br/> - UTC standard

<a id="nestedatt--timeouts"></a>
### Nested Schema for `timeouts`

- `read` (Optional, String) Duration to wait for the read operation, such as `"30s"` or `"2h45m"`.


<a id="nestedatt--associations"></a>
### Nested Schema for `associations`

- `id` (String) Association ID.
- `provisioning_status` (String) Provisioning status of the association.
- `resource` (Attributes) Resource associated with the route table. (see [below for nested schema](#nestedatt--associations--resource))
- `resource_attachment_id` (String) Attachment ID of the associated resource.
- `resource_id` (String) ID of the associated resource.
- `resource_type` (String) Type of the associated resource.
- `tgw_attachment_id` (String) Transit Gateway attachment ID.
- `tgw_route_table_id` (String) Transit Gateway route table ID.

<a id="nestedatt--associations--resource"></a>
### Nested Schema for `associations.resource`

- `cidr_block` (String) CIDR block of the associated resource.
- `id` (String) Resource ID.
- `name` (String) Resource name.
- `project_id` (String) Project ID that owns the resource.
- `project_name` (String) Project name that owns the resource.
- `provisioning_status` (String) Provisioning status of the resource.



<a id="nestedatt--routes"></a>
### Nested Schema for `routes`

- `destination_cidr_block` (String) Destination CIDR block for the route.
- `id` (String) Route ID.
- `provisioning_status` (String) Provisioning status of the route.
- `resource` (Attributes) Resource targeted by the route. (see [below for nested schema](#nestedatt--routes--resource))
- `resource_attachment_id` (String) Attachment ID of the target resource.
- `resource_id` (String) ID of the target resource.
- `resource_type` (String) Type of the target resource.
- `route_type` (String) Route type.
- `tgw_attachment_id` (String) Transit Gateway attachment ID used by the route.
- `tgw_route_table_id` (String) Transit Gateway route table ID.

<a id="nestedatt--routes--resource"></a>
### Nested Schema for `routes.resource`

- `cidr_block` (String) CIDR block of the target resource.
- `id` (String) Resource ID.
- `name` (String) Resource name.
- `project_id` (String) Project ID that owns the resource.
- `project_name` (String) Project name that owns the resource.
- `provisioning_status` (String) Provisioning status of the resource.




---
page_title: "kakaocloud_transit_gateway Data Source - kakaocloud"
subcategory: "Transit Gateway"
description: |-
  Retrieves detailed information about a specific Transit Gateway (TGW) in KakaoCloud.
---

# kakaocloud_transit_gateway (Data Source)

The `kakaocloud_transit_gateway` data source retrieves metadata and configuration details for a specific Transit Gateway (TGW) in KakaoCloud.

A Transit Gateway provides centralized routing between multiple VPCs and supports shared and hybrid networking scenarios.  
Use this data source to reference an existing Transit Gateway by its ID and inspect its attachments, route tables, sharing status, and configuration options without managing its lifecycle.

## Example Usage

```terraform
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# Retrieve a specific Transit Gateway by ID
data "kakaocloud_transit_gateway" "example" {
  id = "<your-transit-gateway-id>"
}

# Output basic Transit Gateway information
output "tgw_basic" {
  description = "Basic information of the Transit Gateway"
  value = {
    id                  = data.kakaocloud_transit_gateway.example.id
    name                = data.kakaocloud_transit_gateway.example.name
    region              = data.kakaocloud_transit_gateway.example.region
    provisioning_status = data.kakaocloud_transit_gateway.example.provisioning_status
    is_shared            = data.kakaocloud_transit_gateway.example.is_shared
  }
}

# Output attached resources
output "tgw_attachments" {
  description = "Attachments connected to the Transit Gateway"
  value       = data.kakaocloud_transit_gateway.example.attachments
}

# Output associated route tables
output "tgw_route_tables" {
  description = "Route tables associated with the Transit Gateway"
  value       = data.kakaocloud_transit_gateway.example.route_tables
}
```

## Argument Reference

- `id` (Required, String) ID of the Transit Gateway to retrieve.

- `timeouts` (Optional, Attributes) Read operation timeout configuration. (see [below for nested schema](#nestedatt--timeouts))

## Attribute Reference

The following attributes are exported:

- `attachments` (Attributes List) List of attachments connected to the Transit Gateway. (see [below for nested schema](#nestedatt--attachments))
- `created_at` (String) Time when the resource was created<br/> - ISO_8601 format<br/> - UTC standard
- `is_shared` (Boolean) Whether the Transit Gateway is shared with other projects.
- `name` (String) Transit Gateway name.
- `options` (Attributes) Transit Gateway configuration options. (see [below for nested schema](#nestedatt--options))
- `owner_project_id` (String) Owner project ID of the Transit Gateway.
- `owner_project_name` (String) Owner project name of the Transit Gateway.
- `project_id` (String) Project ID that owns the Transit Gateway.
- `project_name` (String) Project name that owns the Transit Gateway.
- `provisioning_status` (String) Current provisioning status of the Transit Gateway.
- `region` (String) Region where the Transit Gateway is located.
- `route_tables` (Attributes List) List of route tables associated with the Transit Gateway. (see [below for nested schema](#nestedatt--route_tables))
- `updated_at` (String) Timestamp when the Transit Gateway was last updated.

<a id="nestedatt--timeouts"></a>
### Nested Schema for `timeouts`

- `read` (Optional, String) Duration to wait for the read operation, such as `"30s"` or `"2h45m"`.


<a id="nestedatt--attachments"></a>
### Nested Schema for `attachments`

- `created_at` (String) Time when the resource was created<br/> - ISO_8601 format<br/> - UTC standard
- `id` (String) Transit Gateway attachment ID.
- `provisioning_status` (String) Provisioning status of the attachment.
- `resource_id` (String) ID of the attached resource.
- `resource_name` (String) Name of the attached resource.
- `resource_type` (String) Type of the attached resource.
- `tgw_id` (String) Transit Gateway ID.
- `updated_at` (String) Time when the resource was last updated<br/> - ISO_8601 format<br/> - UTC standard


<a id="nestedatt--options"></a>
### Nested Schema for `options`

- `association_default_route_table_id` (String) Default route table ID used for attachment association.
- `is_auto_accept_shared_attachments` (Boolean) Whether attachments from shared projects are automatically accepted.
- `is_default_route_table_association` (Boolean) Whether attachments are automatically associated with the default route table.


<a id="nestedatt--route_tables"></a>
### Nested Schema for `route_tables`

- `created_at` (String) Time when the resource was created<br/> - ISO_8601 format<br/> - UTC standard
- `id` (String) Route table ID.
- `is_default_association_route_table` (Boolean) Whether this is the default association route table.
- `is_default_propagation_route_table` (Boolean) Whether this is the default propagation route table.
- `name` (String) Route table name.
- `project_id` (String) Project ID that owns the route table.
- `provisioning_status` (String) Route table provisioning status.
- `region` (String) Region where the route table is located.
- `tgw_id` (String) Associated Transit Gateway ID.
- `updated_at` (String) Time when the resource was last updated<br/> - ISO_8601 format<br/> - UTC standard



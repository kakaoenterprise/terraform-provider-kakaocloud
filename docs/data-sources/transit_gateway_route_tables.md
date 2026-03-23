---
page_title: "kakaocloud_transit_gateway_route_tables Data Source - kakaocloud"
subcategory: "Transit Gateway"
description: |-
  Retrieves a list of Transit Gateway (TGW) route tables in KakaoCloud.
---

# kakaocloud_transit_gateway_route_tables (Data Source)

The `kakaocloud_transit_gateway_route_tables` data source retrieves a list of Transit Gateway (TGW) route tables in KakaoCloud.

A TGW route table defines routing rules and attachment associations that control how traffic is forwarded through a Transit Gateway.  
Use this data source to list and filter existing TGW route tables and to reference their routing configurations, associations, and default behaviors without managing their lifecycle.

## Available Filters

| Filter                | Type                | Description |
|-----------------------|---------------------|-------------|
| `id`                  | string              | Route table ID |
| `name`                | string              | Route table name |
| `tgw_id`              | string              | Transit Gateway ID |
| `tgw_name`            | string              | Transit Gateway name |
| `provisioning_status` | ProvisioningStatus  | Route table provisioning status <br/>Possible values: `ACTIVE`, `DELETED`, `ERROR`, `PENDING_CREATE`, `PENDING_UPDATE`, `PENDING_DELETE` |
| `created_at`          | string              | Resource creation time <br/>- ISO_8601 format <br/>- UTC |
| `updated_at`          | string              | Last updated time of the resource <br/>- ISO_8601 format <br/>- UTC |

## Example Usage

```terraform
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# List all Transit Gateway route tables
data "kakaocloud_transit_gateway_route_tables" "all" {}

# List Transit Gateway route tables filtered by project ID
data "kakaocloud_transit_gateway_route_tables" "by_project" {
  filter = [
    {
      name  = "project_id"
      value = "<your-project-id>"
    }
  ]
}

# Output all route tables
output "all_tgw_route_tables" {
  description = "All Transit Gateway route tables"
  value       = data.kakaocloud_transit_gateway_route_tables.all.transit_gateway_route_tables
}

# Output filtered route tables with key fields
output "filtered_tgw_route_tables" {
  description = "Transit Gateway route tables filtered by project ID"
  value = [
    for rt in data.kakaocloud_transit_gateway_route_tables.by_project.transit_gateway_route_tables : {
      id                                = rt.id
      name                              = rt.name
      provisioning_status               = rt.provisioning_status
      is_default_association_route_table = rt.is_default_association_route_table
      is_default_propagation_route_table = rt.is_default_propagation_route_table
      tgw_name                          = rt.tgw_name
    }
  ]
}
```

## Argument Reference

- `filter` (Optional, Attributes List) One or more filters used to limit the returned TGW route tables. (see [below for nested schema](#nestedatt--filter))
- `timeouts` (Optional, Attributes) Read operation timeout configuration. (see [below for nested schema](#nestedatt--timeouts))

## Attribute Reference

The following attributes are exported:

- `transit_gateway_route_tables` (Attributes List) List of Transit Gateway route tables returned by the API. (see [below for nested schema](#nestedatt--transit_gateway_route_tables))

<a id="nestedatt--filter"></a>
### Nested Schema for `filter`

- `name` (Required, String) Name of the field to filter by (for example, `project_id`).

- `value` (Optional, String) Value to match for the specified filter name.


<a id="nestedatt--timeouts"></a>
### Nested Schema for `timeouts`

- `read` (Optional, String) Duration to wait for the read operation.


<a id="nestedatt--transit_gateway_route_tables"></a>
### Nested Schema for `transit_gateway_route_tables`

- `associations` (Attributes List) List of attachment associations for the route table. (see [below for nested schema](#nestedatt--transit_gateway_route_tables--associations))
- `created_at` (String) Time when the resource was created<br/> - ISO_8601 format<br/> - UTC standard
- `id` (String) Route table ID.
- `is_default_association_route_table` (Boolean) Whether this is the default association route table.
- `is_default_propagation_route_table` (Boolean) Whether this is the default propagation route table.
- `name` (String) Route table name.
- `project_id` (String) Project ID that owns the route table.
- `project_name` (String) Project name that owns the route table.
- `provisioning_status` (String) Current provisioning status of the route table.
- `region` (String) Region where the route table is located.
- `routes` (Attributes List) List of routes configured in the route table. (see [below for nested schema](#nestedatt--transit_gateway_route_tables--routes))
- `tgw_id` (String) Transit Gateway ID associated with the route table.
- `tgw_name` (String) Transit Gateway name.
- `updated_at` (String) Time when the resource was last updated<br/> - ISO_8601 format<br/> - UTC standard

<a id="nestedatt--transit_gateway_route_tables--associations"></a>
### Nested Schema for `transit_gateway_route_tables.associations`

- `id` (String) Association ID.
- `provisioning_status` (String) Provisioning status of the association.
- `resource` (Attributes) Resource associated with the route table. (see [below for nested schema](#nestedatt--transit_gateway_route_tables--associations--resource))
- `resource_attachment_id` (String) Attachment ID of the associated resource.
- `resource_id` (String) ID of the associated resource.
- `resource_type` (String) Type of the associated resource.
- `tgw_route_table_id` (String) Transit Gateway route table ID.

<a id="nestedatt--transit_gateway_route_tables--associations--resource"></a>
### Nested Schema for `transit_gateway_route_tables.associations.resource`

- `cidr_block` (String) CIDR block of the associated resource.
- `id` (String) Resource ID.
- `name` (String) Resource name.
- `project_id` (String) Project ID that owns the resource.
- `project_name` (String) Project name that owns the resource.
- `provisioning_status` (String) Provisioning status of the resource.



<a id="nestedatt--transit_gateway_route_tables--routes"></a>
### Nested Schema for `transit_gateway_route_tables.routes`

- `destination_cidr_block` (String) Destination CIDR block for the route.
- `id` (String) Route ID.
- `provisioning_status` (String) Provisioning status of the route.
- `resource` (Attributes) Resource targeted by the route. (see [below for nested schema](#nestedatt--transit_gateway_route_tables--routes--resource))
- `resource_attachment_id` (String) Attachment ID of the target resource.
- `resource_id` (String) ID of the target resource.
- `resource_type` (String) Type of the target resource.
- `route_type` (String) Route type.
- `tgw_route_table_id` (String) Transit Gateway route table ID.

<a id="nestedatt--transit_gateway_route_tables--routes--resource"></a>
### Nested Schema for `transit_gateway_route_tables.routes.resource`

- `cidr_block` (String) CIDR block of the target resource.
- `id` (String) Resource ID.
- `name` (String) Resource name.
- `project_id` (String) Project ID that owns the resource.
- `project_name` (String) Project name that owns the resource.
- `provisioning_status` (String) Provisioning status of the resource.





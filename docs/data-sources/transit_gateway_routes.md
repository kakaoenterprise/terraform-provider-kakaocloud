---
page_title: "kakaocloud_transit_gateway_routes Data Source - kakaocloud"
subcategory: "Transit Gateway"
description: |-
  The kakaocloud_transit_gateway_routes data source retrieves route entries from a specified Transit Gateway route table in KakaoCloud.
---

# kakaocloud_transit_gateway_routes (Data Source)

The `kakaocloud_transit_gateway_routes` data source retrieves a list of routes in a specified Transit Gateway route table in KakaoCloud.

## Available Filters

| Filter                         | Type                | Description |
|--------------------------------|---------------------|-------------|
| `destination_cidr_block`       | string              | Destination CIDR block |
| `route_type`                   | string              | Route type (e.g. `static`, `propagated`) |
| `provisioning_status`          | ProvisioningStatus  | Provisioning status of the route <br/>Possible values: `ACTIVE`, `DELETED`, `ERROR`, `PENDING_CREATE`, `PENDING_UPDATE`, `PENDING_DELETE` |
| `resource_type`                | string              | Type of the associated resource |
| `resource_id`                  | string              | ID of the associated resource |
| `resource_name`                | string              | Name of the associated resource |
| `resource_provisioning_status` | string              | Provisioning status of the associated resource |
| `resource_attachment_id`       | string              | ID of the associated attachment |

## Example Usage

```terraform
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

data "kakaocloud_transit_gateway_routes" "example" {
  route_table_id = "<your-route-table-id>"

  filter = [
    {
      name  = "destination_cidr_block"
      value = "10.0.0.0/16"
    },
    {
      name  = "route_type"
      value = "static"
    },
    {
      name  = "provisioning_status"
      value = "ACTIVE"
    }
  ]
}

output "tgw_routes" {
  description = "Filtered Transit Gateway routes"
  value       = data.kakaocloud_transit_gateway_routes.example.transit_gateway_routes
}
```

## Argument Reference

- `route_table_id` (Required, String) ID of the Transit Gateway route table to retrieve routes from.

- `filter` (Optional, Attributes List) Filters to narrow down the returned routes. (see [below for nested schema](#nestedatt--filter))
- `timeouts` (Optional, Attributes) Custom timeout settings. (see [below for nested schema](#nestedatt--timeouts))

## Attribute Reference

The following attributes are exported:

- `transit_gateway_routes` (Attributes List) List of routes associated with the specified Transit Gateway route table. (see [below for nested schema](#nestedatt--transit_gateway_routes))

<a id="nestedatt--filter"></a>
### Nested Schema for `filter`

- `name` (Required, String) Name of the attribute to filter by.

- `value` (Optional, String) Value to match for the specified filter attribute.


<a id="nestedatt--timeouts"></a>
### Nested Schema for `timeouts`

- `read` (Optional, String) A string that can be parsed as a duration, such as `30s` or `2h45m`.


<a id="nestedatt--transit_gateway_routes"></a>
### Nested Schema for `transit_gateway_routes`

- `destination_cidr_block` (String) Destination CIDR block of the route.
- `id` (String) Unique ID of the Transit Gateway route.
- `provisioning_status` (String) Provisioning status of the route.
- `resource` (Attributes) Resource associated with the route. (see [below for nested schema](#nestedatt--transit_gateway_routes--resource))
- `resource_attachment_id` (String) ID of the resource attachment.
- `resource_id` (String) ID of the associated resource.
- `resource_type` (String) Type of the associated resource.
- `route_table_id` (String) ID of the Transit Gateway route table.
- `route_type` (String) Type of the route.
- `tgw_route_table_id` (String) Transit Gateway route table ID.

<a id="nestedatt--transit_gateway_routes--resource"></a>
### Nested Schema for `transit_gateway_routes.resource`

- `cidr_block` (String) CIDR block of the associated resource.
- `id` (String) Unique ID of the resource.
- `name` (String) Name of the resource.
- `project_id` (String) Project ID that owns the resource.
- `project_name` (String) Project name that owns the resource.
- `provisioning_status` (String) Provisioning status of the resource.




---
page_title: "kakaocloud_transit_gateways Data Source - kakaocloud"
subcategory: "Transit Gateway"
description: |-
  Retrieves a list of Transit Gateways (TGWs) in KakaoCloud.
---

# kakaocloud_transit_gateways (Data Source)

The `kakaocloud_transit_gateways` data source retrieves a list of Transit Gateways (TGWs) in KakaoCloud.

A Transit Gateway provides centralized routing between multiple VPCs and supports shared and hybrid networking scenarios.  
Use this data source to list and filter existing Transit Gateways and reference their attachments, route tables, and sharing configuration without managing their lifecycle.

## Available Filters

| Filter                  | Type                | Description |
|-------------------------|---------------------|-------------|
| `id`                    | string              | Transit Gateway ID |
| `name`                  | string              | Transit Gateway name |
| `region`                | Region              | Region where the Transit Gateway is located <br/>Possible values: `kr-central-2` |
| `is_shared`             | boolean             | Whether the Transit Gateway is shared <br/>Example: `true` |
| `provisioning_status`   | ProvisioningStatus  | Transit Gateway provisioning status <br/>Possible values: `ACTIVE`, `DELETED`, `ERROR`, `PENDING_CREATE`, `PENDING_UPDATE`, `PENDING_DELETE` |
| `created_at`            | string              | Resource creation time <br/>- ISO_8601 format <br/>- UTC |
| `updated_at`            | string              | Last updated time of the resource <br/>- ISO_8601 format <br/>- UTC |

## Example Usage

```terraform
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# List all Transit Gateways
data "kakaocloud_transit_gateways" "example" {}

# Filter Transit Gateways by project ID
data "kakaocloud_transit_gateways" "by_project" {
  filter = [
    {
      name  = "project_id"
      value = "<your-project-id>"
    }
  ]
}

# Output Transit Gateway IDs and names
output "transit_gateways_basic" {
  description = "Transit Gateway IDs and names"
  value = [
    for tgw in data.kakaocloud_transit_gateways.example.transit_gateways : {
      id   = tgw.id
      name = tgw.name
      region = tgw.region
      status = tgw.provisioning_status
    }
  ]
}

# Output route tables for each Transit Gateway
output "transit_gateways_route_tables" {
  description = "Route tables associated with each Transit Gateway"
  value = {
    for tgw in data.kakaocloud_transit_gateways.example.transit_gateways :
    tgw.id => tgw.route_tables
  }
}
```

## Argument Reference

- `filter` (Optional, Attributes List) One or more filters used to limit the returned Transit Gateways. (see [below for nested schema](#nestedatt--filter))
- `timeouts` (Optional, Attributes) Read operation timeout configuration. (see [below for nested schema](#nestedatt--timeouts))

## Attribute Reference

The following attributes are exported:

- `transit_gateways` (Attributes List) List of Transit Gateways returned by the API. (see [below for nested schema](#nestedatt--transit_gateways))

<a id="nestedatt--filter"></a>
### Nested Schema for `filter`

- `name` (Required, String) Name of the field to filter by (for example, `project_id`).

- `value` (Optional, String) Value to match for the specified filter name.


<a id="nestedatt--timeouts"></a>
### Nested Schema for `timeouts`

- `read` (Optional, String) Duration to wait for the read operation, such as `"30s"` or `"2h45m"`.


<a id="nestedatt--transit_gateways"></a>
### Nested Schema for `transit_gateways`

- `attachments` (Attributes List) List of attachments connected to the Transit Gateway. (see [below for nested schema](#nestedatt--transit_gateways--attachments))
- `created_at` (String) Time when the resource was created<br/> - ISO_8601 format<br/> - UTC standard
- `id` (String) Transit Gateway ID.
- `is_shared` (Boolean) Whether the Transit Gateway is shared with other projects.
- `name` (String) Transit Gateway name.
- `options` (Attributes) Transit Gateway configuration options. (see [below for nested schema](#nestedatt--transit_gateways--options))
- `owner_project_id` (String) Owner project ID of the Transit Gateway.
- `owner_project_name` (String) Owner project name of the Transit Gateway.
- `project_id` (String) Project ID that owns the Transit Gateway.
- `project_name` (String) Project name that owns the Transit Gateway.
- `provisioning_status` (String) Current provisioning status of the Transit Gateway.
- `region` (String) Region where the Transit Gateway is located.
- `route_tables` (Attributes List) List of route tables associated with the Transit Gateway. (see [below for nested schema](#nestedatt--transit_gateways--route_tables))
- `updated_at` (String) Time when the resource was last updated<br/> - ISO_8601 format<br/> - UTC standard

<a id="nestedatt--transit_gateways--attachments"></a>
### Nested Schema for `transit_gateways.attachments`

- `created_at` (String) Time when the resource was created<br/> - ISO_8601 format<br/> - UTC standard
- `id` (String) Transit Gateway attachment ID.
- `provisioning_status` (String) Provisioning status of the attachment.
- `resource_id` (String) ID of the attached resource.
- `resource_name` (String) Name of the attached resource.
- `resource_type` (String) Type of the attached resource.
- `tgw_id` (String) Transit Gateway ID.
- `updated_at` (String) Time when the resource was last updated<br/> - ISO_8601 format<br/> - UTC standard


<a id="nestedatt--transit_gateways--options"></a>
### Nested Schema for `transit_gateways.options`

- `association_default_route_table_id` (String) Default route table ID used for attachment association.
- `is_auto_accept_shared_attachments` (Boolean) Whether attachments from shared projects are automatically accepted.
- `is_default_route_table_association` (Boolean) Whether attachments are automatically associated with the default route table.


<a id="nestedatt--transit_gateways--route_tables"></a>
### Nested Schema for `transit_gateways.route_tables`

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




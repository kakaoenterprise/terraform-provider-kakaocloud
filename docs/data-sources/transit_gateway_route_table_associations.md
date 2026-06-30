---
page_title: "kakaocloud_transit_gateway_route_table_associations Data Source - kakaocloud"
subcategory: "Transit Gateway"
description: |-
  Retrieves Transit Gateway (TGW) route table associations in KakaoCloud.
---

# kakaocloud_transit_gateway_route_table_associations (Data Source)

The `kakaocloud_transit_gateway_route_table_associations` data source retrieves a list of associations associated with a specified Transit Gateway (TGW) route table in KakaoCloud.

You can use this data source to inspect which attachments or resources are currently associated with a TGW route table and to filter the results using supported query parameters.

## Available Filters

Each filter can be specified using the `filter` block, where the `name` value corresponds to one of the query parameters listed below.

| Filter                   | Type               | Description |
|--------------------------|--------------------|-------------|
| `resource_name`          | string             | Name of the associated resource |
| `resource_id`            | string             | ID of the associated resource |
| `resource_provisioning_status` | string     | Provisioning status of the resource |
| `resource_type`          | ResourceType       | Type of the associated resource <br/>Possible values: `VPC` |
| `provisioning_status`    | ProvisioningStatus | Provisioning status of the association <br/>Possible values: `ACTIVE`, `DELETED`, `ERROR`, `PENDING_CREATE`, `PENDING_UPDATE`, `PENDING_DELETE` |
| `resource_attachment_id` | string             | ID of the associated attachment |

## Example Usage

```hcl
data "kakaocloud_transit_gateway_route_table_associations" "example" {
  route_table_id = kakaocloud_transit_gateway_route_table.example.id
}
```

### Example Usage with filter

```hcl
data "kakaocloud_transit_gateway_route_table_associations" "example_with_filter" {
  route_table_id = kakaocloud_transit_gateway_route_table.example.id

  filter = [
    {
      name  = "resource_type"
      value = "VPC"
    }
  ]
}
```

## Argument Reference

- `route_table_id` (Required, String) Transit Gateway route table ID to retrieve associations for.

- `filter` (Optional, Attributes List) One or more filters to apply to the results. (see [below for nested schema](#nestedatt--filter))
- `timeouts` (Optional, Attributes) Timeout configuration for read operations. (see [below for nested schema](#nestedatt--timeouts))

## Attribute Reference

The following attributes are exported:

- `associations` (Attributes List) List of route table associations. (see [below for nested schema](#nestedatt--associations))

<a id="nestedatt--filter"></a>
### Nested Schema for `filter`

- `name` (Required, String) Filter field name.

- `value` (Optional, String) Filter value to match.


<a id="nestedatt--timeouts"></a>
### Nested Schema for `timeouts`

- `read` (Optional, String) A string that can be parsed as a duration consisting of numbers and unit suffixes, such as "30s" or "2h45m".   Valid time units are "s" (seconds), "m" (minutes), "h" (hours).


<a id="nestedatt--associations"></a>
### Nested Schema for `associations`

- `id` (String) Association ID.
- `provisioning_status` (String) Provisioning status of the association.
- `resource` (Attributes) Associated resource information. (see [below for nested schema](#nestedatt--associations--resource))
- `resource_attachment_id` (String) Resource attachment ID.
- `resource_id` (String) Associated resource ID.
- `resource_type` (String) Associated resource type (for example, VPC).
- `route_table_id` (String) Route table ID.
- `tgw_route_table_id` (String) Transit Gateway route table ID.

<a id="nestedatt--associations--resource"></a>
### Nested Schema for `associations.resource`

- `cidr_block` (String) CIDR block of the associated resource.
- `id` (String) Resource ID.
- `name` (String) Resource name.
- `project_id` (String) Project ID that owns the resource.
- `project_name` (String) Project name that owns the resource.
- `provisioning_status` (String) Provisioning status of the resource.

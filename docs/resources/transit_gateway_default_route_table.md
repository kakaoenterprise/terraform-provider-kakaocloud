---
page_title: "kakaocloud_transit_gateway_default_route_table Resource - kakaocloud"
subcategory: "Transit Gateway"
description: |-
  Manages kakaocloud_transit_gateway_default_route_table. This resource designates a specific route table as the default route table for a Transit Gateway.
---

# kakaocloud_transit_gateway_default_route_table (Resource)

Manages kakaocloud_transit_gateway_default_route_table. This resource sets a specific route table as the default route table of a Transit Gateway (TGW). When a route table is configured as the default, newly created attachments are automatically associated with this route table unless explicitly specified otherwise. Only one default route table can exist per Transit Gateway. Updating this resource changes the default association behavior for future attachments but does not automatically modify existing attachments.

## Example Usage

```terraform
resource "kakaocloud_transit_gateway_default_route_table" "example" {
  route_table_id = kakaocloud_transit_gateway_route_table.example.id
  tgw_id         = kakaocloud_transit_gateway.example.id
}
```

## Argument Reference

- `route_table_id` (Required, String) The ID of the route table to designate as the default route table for the Transit Gateway.
- `tgw_id` (Required, String) The ID of the Transit Gateway for which the default route table is configured.

- `timeouts` (Optional, Attributes) Configuration block for specifying timeouts for create, read, update, and delete operations. (see [below for nested schema](#nestedatt--timeouts))

<a id="nestedatt--timeouts"></a>
### Nested Schema for `timeouts`

- `create` (Optional, String) A string that can be parsed as a duration consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s", "m", and "h".
- `delete` (Optional, String) A string that can be parsed as a duration consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s", "m", and "h".
- `read` (Optional, String) A string that can be parsed as a duration consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s", "m", and "h".
- `update` (Optional, String) A string that can be parsed as a duration consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s", "m", and "h".


## Import

Import is supported using the following syntax:

The [`terraform import` command](https://developer.hashicorp.com/terraform/cli/commands/import) can be used, for example:

```shell
# kakaocloud_transit_gateway_default_route_table import script
terraform import kakaocloud_transit_gateway_default_route_table.example <tgw_id>
```


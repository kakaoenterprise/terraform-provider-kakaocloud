---
page_title: "kakaocloud_transit_gateway_route_table_association Resource - kakaocloud"
subcategory: "Transit Gateway"
description: |-
  Manages kakaocloud_transit_gateway_route_table_association
---

# kakaocloud_transit_gateway_route_table_association (Resource)

Manages kakaocloud_transit_gateway_route_table_association. This resource associates a Transit Gateway attachment with a specific Transit Gateway route table.

## Example Usage

```terraform
resource "kakaocloud_transit_gateway_route_table_association" "example" {
  route_table_id    = kakaocloud_transit_gateway_route_table.example.id
  tgw_attachment_id = kakaocloud_transit_gateway_attachment.example.id
}
```

## Argument Reference

- `route_table_id` (Required, String) The ID of the Transit Gateway route table to associate with the attachment.
- `tgw_attachment_id` (Required, String) The ID of the Transit Gateway attachment to associate with the route table.

- `timeouts` (Optional, Attributes) Timeout configuration for create, read, update, and delete operations. (see [below for nested schema](#nestedatt--timeouts))

## Attribute Reference

- `id` (String) The unique identifier of the Transit Gateway route table association.
- `provisioning_status` (String) The current provisioning status of the association.

<a id="nestedatt--timeouts"></a>
### Nested Schema for `timeouts`

- `create` (Optional, String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).
- `delete` (Optional, String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Setting a timeout for a Delete operation is only applicable if changes are saved into state before the destroy operation occurs.
- `read` (Optional, String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Read operations occur during any refresh or planning operation when refresh is enabled.
- `update` (Optional, String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).


## Import

Import is supported using the following syntax:

The [`terraform import` command](https://developer.hashicorp.com/terraform/cli/commands/import) can be used, for example:

```shell
# kakaocloud_transit_gateway_route_table_association import script
terraform import kakaocloud_transit_gateway_route_table_association.example <route_table_id>/<tgw_attachment_id>
```

---
page_title: "kakaocloud_transit_gateway_route_table Resource - kakaocloud"
subcategory: "Transit Gateway"
description: |-
  Creates and manages a Transit Gateway (TGW) route table in KakaoCloud.
---

# kakaocloud_transit_gateway_route_table (Resource)

The `kakaocloud_transit_gateway_route_table` resource creates and manages a Transit Gateway (TGW) route table in KakaoCloud.

A TGW route table defines routing rules and attachment associations that control how traffic is forwarded between resources connected to a Transit Gateway.  
You can use this resource to create custom route tables, associate TGW attachments, and define routes for inter-VPC or cross-project traffic.

## Example Usage

```terraform
resource "kakaocloud_transit_gateway_route_table" "example" {
  name   = "tgwrt-example"
  tgw_id = kakaocloud_transit_gateway.example.id
}
```

## Argument Reference

- `name` (Required, String) Name of the Transit Gateway route table.
- `tgw_id` (Required, String) ID of the Transit Gateway that owns the route table.

- `routes` (Optional, Attributes Set) List of routes configured in the route table. (see [below for nested schema](#nestedatt--routes))
- `timeouts` (Optional, Attributes) Timeout configuration for create, read, update, and delete operations. (see [below for nested schema](#nestedatt--timeouts))

## Attribute Reference

- `associations` (Attributes List) List of attachment associations applied to the route table. (see [below for nested schema](#nestedatt--associations))
- `created_at` (String) Time when the resource was created<br/> - ISO_8601 format<br/> - UTC standard
- `id` (String) Transit Gateway route table ID.
- `is_default_association_route_table` (Boolean) Whether this route table is the default association route table.
- `is_default_propagation_route_table` (Boolean) Whether this route table is the default propagation route table.
- `project_id` (String) Project ID that owns the route table.
- `project_name` (String) Project name that owns the route table.
- `provisioning_status` (String) Current provisioning status of the route table.
- `region` (String) Region where the route table is located.
- `tgw_name` (String) Name of the associated Transit Gateway.
- `updated_at` (String) Time when the resource was last updated<br/> - ISO_8601 format<br/> - UTC standard

<a id="nestedatt--routes"></a>
### Nested Schema for `routes`

- `destination_cidr_block` (Required, String) Destination CIDR block for the route.
- `tgw_attachment_id` (Required, String) Transit Gateway attachment ID.
- `id` (String) Route ID.


<a id="nestedatt--timeouts"></a>
### Nested Schema for `timeouts`

- `create` (Optional, String) Maximum duration to wait for the create operation.
- `delete` (Optional, String) Maximum duration to wait for the delete operation, applied only if state is saved before destroy.
- `read` (Optional, String) Maximum duration to wait for read operations during refresh or planning.
- `update` (Optional, String) Maximum duration to wait for the update operation.


<a id="nestedatt--associations"></a>
### Nested Schema for `associations`

- `id` (String) Association ID.
- `tgw_attachment_id` (String) Transit Gateway attachment ID.


## Import

Import is supported using the following syntax:

The [`terraform import` command](https://developer.hashicorp.com/terraform/cli/commands/import) can be used, for example:

```shell
$ terraform import kakaocloud_transit_gateway_route_table.example <transit_gateway_route_table_id>
```

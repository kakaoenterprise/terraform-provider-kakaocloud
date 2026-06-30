---
page_title: "kakaocloud_transit_gateway Resource - kakaocloud"
subcategory: "Transit Gateway"
description: |-
  Creates and manages a Transit Gateway (TGW) in KakaoCloud.
---

# kakaocloud_transit_gateway (Resource)

The `kakaocloud_transit_gateway` resource creates and manages a Transit Gateway (TGW) in KakaoCloud.

A Transit Gateway provides centralized routing between multiple VPCs and supports shared and hybrid networking scenarios.  
You can use this resource to create a TGW and configure whether to automatically accept shared attachments.
(To configure the default route table, use the `kakaocloud_transit_gateway_default_route_table resource`.)

## Example Usage

```terraform
resource "kakaocloud_transit_gateway" "example" {
  name = "tgw-example"

  options = {
    is_auto_accept_shared_attachments = false
  }
}
```

## Argument Reference

- `name` (Required, String) Name of the Transit Gateway.
- `options` (Required, Attributes) Configuration options that control Transit Gateway behavior. (see [below for nested schema](#nestedatt--options))

- `timeouts` (Optional, Attributes) Timeout configuration for create, read, update, and delete operations. (see [below for nested schema](#nestedatt--timeouts))

## Attribute Reference

- `created_at` (String) Time when the resource was created<br/> - ISO_8601 format<br/> - UTC standard
- `id` (String) ID of the Transit Gateway.
- `is_shared` (Boolean) Indicates whether the Transit Gateway is shared with other projects.
- `owner_project_id` (String) ID of the project that owns the Transit Gateway.
- `owner_project_name` (String) Name of the project that owns the Transit Gateway.
- `project_id` (String) ID of the project where the Transit Gateway is created.
- `project_name` (String) Name of the project where the Transit Gateway is created.
- `provisioning_status` (String) Current provisioning status of the Transit Gateway.
- `region` (String) Region where the Transit Gateway is located.
- `updated_at` (String) Time when the resource was last updated<br/> - ISO_8601 format<br/> - UTC standard

<a id="nestedatt--options"></a>
### Nested Schema for `options`

- `is_auto_accept_shared_attachments` (Required, Boolean) Whether attachments from shared projects are automatically accepted.
- `association_default_route_table_id` (String) ID of the default route table to associate with attachments.
- `is_default_route_table_association` (Boolean) Whether attachments are automatically associated with the default route table.


<a id="nestedatt--timeouts"></a>
### Nested Schema for `timeouts`

- `create` (Optional, String) Maximum duration to wait for the create operation.
- `delete` (Optional, String) Maximum duration to wait for the delete operation. This applies only if the resource state is saved before the destroy operation occurs.
- `read` (Optional, String) Maximum duration to wait for read operations during refresh or planning.
- `update` (Optional, String) Maximum duration to wait for the update operation.


## Import

Import is supported using the following syntax:

The [`terraform import` command](https://developer.hashicorp.com/terraform/cli/commands/import) can be used, for example:

```shell
$ terraform import kakaocloud_transit_gateway.example <transit_gateway_id>
```

---
page_title: "kakaocloud_transit_gateway Resource - kakaocloud"
subcategory: "Transit Gateway"
description: |-
  Creates and manages a Transit Gateway (TGW) in KakaoCloud.
---

# kakaocloud_transit_gateway (Resource)

The `kakaocloud_transit_gateway` resource creates and manages a Transit Gateway (TGW) in KakaoCloud.

A Transit Gateway provides centralized routing between multiple VPCs and supports shared and hybrid networking scenarios.  
You can use this resource to create a TGW, configure its default routing behavior, and control how attachments and shared projects are handled.

## Example Usage

```terraform
resource "kakaocloud_transit_gateway" "example" {
  name = "tgw-example"

  options = {
    is_auto_accept_shared_attachments = false
    is_default_route_table_association = false
  }
}
```

## Argument Reference

- `name` (Required, String) Name of the Transit Gateway.
- `options` (Required, Attributes) Configuration options that control Transit Gateway behavior. (see [below for nested schema](#nestedatt--options))

- `timeouts` (Optional, Attributes) Timeout configuration for create, read, update, and delete operations. (see [below for nested schema](#nestedatt--timeouts))

## Attribute Reference

- `attachments` (Attributes List) List of attachments connected to the Transit Gateway. (see [below for nested schema](#nestedatt--attachments))
- `created_at` (String) Time when the resource was created<br/> - ISO_8601 format<br/> - UTC standard
- `id` (String) ID of the Transit Gateway.
- `is_shared` (Boolean) Indicates whether the Transit Gateway is shared with other projects.
- `owner_project_id` (String) ID of the project that owns the Transit Gateway.
- `owner_project_name` (String) Name of the project that owns the Transit Gateway.
- `project_id` (String) ID of the project where the Transit Gateway is created.
- `project_name` (String) Name of the project where the Transit Gateway is created.
- `provisioning_status` (String) Current provisioning status of the Transit Gateway.
- `region` (String) Region where the Transit Gateway is located.
- `route_tables` (Attributes List) List of route tables associated with the Transit Gateway. (see [below for nested schema](#nestedatt--route_tables))
- `updated_at` (String) Time when the resource was last updated<br/> - ISO_8601 format<br/> - UTC standard

<a id="nestedatt--options"></a>
### Nested Schema for `options`

- `is_auto_accept_shared_attachments` (Required, Boolean) Whether attachments from shared projects are automatically accepted.
- `is_default_route_table_association` (Required, Boolean) Whether attachments are automatically associated with the default route table.

- `association_default_route_table_id` (Optional, String) ID of the default route table to associate with attachments.


<a id="nestedatt--timeouts"></a>
### Nested Schema for `timeouts`

- `create` (Optional, String) Maximum duration to wait for the create operation.
- `delete` (Optional, String) Maximum duration to wait for the delete operation. This applies only if the resource state is saved before the destroy operation occurs.
- `read` (Optional, String) Maximum duration to wait for read operations during refresh or planning.
- `update` (Optional, String) Maximum duration to wait for the update operation.


<a id="nestedatt--attachments"></a>
### Nested Schema for `attachments`

- `created_at` (String) Time when the attachment was created <br/> - ISO_8601 format <br/> - UTC standard
- `id` (String) Unique ID of the Transit Gateway attachment
- `provisioning_status` (String) Provisioning status of the attachment
- `resource_id` (String) ID of the resource attached to the Transit Gateway
- `resource_name` (String) Name of the resource attached to the Transit Gateway
- `resource_type` (String) Type of the attached resource (for example, VPC)
- `tgw_id` (String) ID of the Transit Gateway associated with the attachment
- `updated_at` (String) Time when the attachment was last updated <br/> - ISO_8601 format <br/> - UTC standard


<a id="nestedatt--route_tables"></a>
### Nested Schema for `route_tables`

- `created_at` (String) Time when the route table was created <br/> - ISO_8601 format <br/> - UTC standard
- `id` (String) Unique ID of the Transit Gateway route table
- `is_default_association_route_table` (Boolean) Indicates whether this route table is the default association route table
- `is_default_propagation_route_table` (Boolean) Indicates whether this route table is the default propagation route table
- `name` (String) Name of the Transit Gateway route table
- `project_id` (String) ID of the project that owns the route table
- `provisioning_status` (String) Provisioning status of the route table
- `region` (String) Region where the route table is located
- `tgw_id` (String) ID of the Transit Gateway associated with the route table
- `updated_at` (String) Time when the route table was last updated <br/> - ISO_8601 format <br/> - UTC standard


## Import

Import is supported using the following syntax:

The [`terraform import` command](https://developer.hashicorp.com/terraform/cli/commands/import) can be used, for example:

```shell
$ terraform import kakaocloud_transit_gateway.example <transit_gateway_id>
```


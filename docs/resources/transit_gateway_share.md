---
page_title: "kakaocloud_transit_gateway_share Resource - kakaocloud"
subcategory: "Transit Gateway"
description: |-
  Shares a Transit Gateway (TGW) with another project in KakaoCloud.
---

# kakaocloud_transit_gateway_share (Resource)

The `kakaocloud_transit_gateway_share` resource shares a Transit Gateway (TGW) with another project in KakaoCloud.

Transit Gateway sharing enables cross-project networking by allowing a TGW owned by one project to be used by other projects.  
Use this resource to grant another project access to a Transit Gateway. The shared project can then create attachments to the shared TGW, subject to approval and policy settings.

## Example Usage

```terraform
data "kakaocloud_project" "example" {}

resource "kakaocloud_transit_gateway_share" "example" {
  target_project_id = data.kakaocloud_project.example.id
  tgw_id            = kakaocloud_transit_gateway.example.id
}
```

## Argument Reference

- `target_project_id` (Required, String) ID of the target project to share the Transit Gateway with.
- `tgw_id` (Required, String) ID of the Transit Gateway to share.

- `timeouts` (Optional, Attributes) Timeout configuration for create, read, update, and delete operations. (see [below for nested schema](#nestedatt--timeouts))

## Attribute Reference

- `id` (String) ID of the Transit Gateway share resource.

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
$ terraform import kakaocloud_transit_gateway_share.example <tgw_id>/<target_project_id>
```

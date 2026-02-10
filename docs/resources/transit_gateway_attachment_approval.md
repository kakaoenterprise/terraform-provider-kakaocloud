---
page_title: "kakaocloud_transit_gateway_attachment_approval Resource - kakaocloud"
subcategory: "Transit Gateway"
description: |-
  Approves a pending Transit Gateway (TGW) attachment in KakaoCloud.
---

# kakaocloud_transit_gateway_attachment_approval (Resource)

The `kakaocloud_transit_gateway_attachment_approval` resource approves a pending Transit Gateway (TGW) attachment in KakaoCloud.

A TGW attachment approval is required when Transit Gateway sharing or attachment approval is configured to manual mode.  
By approving the attachment, the attached VPC becomes active and traffic can be routed through the Transit Gateway.

## Example Usage


```terraform
# Approve an existing Transit Gateway attachment

resource "kakaocloud_transit_gateway_attachment_approval" "example" {
  attachment_id = kakaocloud_transit_gateway_attachment.example.id
}
```

## Argument Reference

- `attachment_id` (Required, String) ID of the Transit Gateway attachment to approve.

- `timeouts` (Optional, Attributes) Timeout configuration for create, read, update, and delete operations. (see [below for nested schema](#nestedatt--timeouts))

## Attribute Reference

- `cidr_block` (String) CIDR block of the attached VPC or network resource.
- `created_at` (String) Time when the resource was created<br/> - ISO_8601 format<br/> - UTC standard
- `id` (String) ID of this approval resource.
- `project_id` (String) Project ID that owns the attachment.
- `provisioning_status` (String) Provisioning status of the attachment after approval.
- `tgw_id` (String) ID of the Transit Gateway.
- `tgw_project_id` (String) Project ID that owns the Transit Gateway.
- `updated_at` (String) Time when the resource was last updated<br/> - ISO_8601 format<br/> - UTC standard
- `vpc_id` (String) ID of the attached VPC.
- `vpc_name` (String) Name of the attached VPC.

<a id="nestedatt--timeouts"></a>
### Nested Schema for `timeouts`

- `create` (Optional, String) A string that can be parsed as a duration such as `"30s"` or `"2h45m"`.
- `delete` (Optional, String) A string that can be parsed as a duration such as `"30s"` or `"2h45m"`.   Setting a delete timeout is only applicable if state is saved before destroy.
- `read` (Optional, String) A string that can be parsed as a duration such as `"30s"` or `"2h45m"`.
- `update` (Optional, String) A string that can be parsed as a duration such as `"30s"` or `"2h45m"`.


## Import

Import is supported using the following syntax:

The [`terraform import` command](https://developer.hashicorp.com/terraform/cli/commands/import) can be used, for example:

```shell
$ terraform import kakaocloud_transit_gateway_attachment_approval.example <attachment_approval_id>
```


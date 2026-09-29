---
page_title: "kakaocloud_internet_gateway Resource - kakaocloud"
subcategory: "VPC"
description: |-
  Creates and manages an Internet Gateway in KakaoCloud.
---

# kakaocloud_internet_gateway (Resource)

The `kakaocloud_internet_gateway` resource creates and manages an Internet Gateway in KakaoCloud.

An Internet Gateway enables communication between a VPC and the internet. Use this resource to configure the Internet Gateway's name and description and to retrieve its NAT IP address, connection status, and associated VPC information.

## Example Usage

```terraform
# kakaocloud_internet_gateway Terraform Resource Example

# Basic Usage (kakaocloud_internet_gateway)
resource "kakaocloud_internet_gateway" "example" {}
```

## Argument Reference

- `description` (Optional, String) Internet Gateway description.
- `name` (Optional, String) Internet Gateway name.
- `timeouts` (Optional, Attributes) Custom timeout settings. (see [below for nested schema](#nestedatt--timeouts))

## Attribute Reference

- `attachment` (Attributes) Attachment information for the VPC connection. (see [below for nested schema](#nestedatt--attachment))
- `attachment_status` (String) Connection status of the Internet Gateway attachment. Possible values: `ATTACHED`, `ATTACHING`, `DETACHING`, `DETACHED`.
- `created_at` (String) Time when the resource was created.<br/> - ISO_8601 format<br/> - UTC standard
- `domain_id` (String) Domain ID that owns the Internet Gateway.
- `id` (String) Internet Gateway ID.
- `nat_ip` (String) Public NAT IP address assigned to the Internet Gateway.
- `project_id` (String) Project ID that owns the Internet Gateway.
- `project_name` (String) Project name that owns the Internet Gateway.
- `provisioning_status` (String) Provisioning status of the Internet Gateway.
- `updated_at` (String) Time when the resource was last updated.<br/> - ISO_8601 format<br/> - UTC standard
- `vpc` (Attributes) Connected VPC information. (see [below for nested schema](#nestedatt--vpc))

<a id="nestedatt--timeouts"></a>
### Nested Schema for `timeouts`

- `create` (Optional, String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).
- `delete` (Optional, String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Setting a timeout for a Delete operation is only applicable if changes are saved into state before the destroy operation occurs.
- `read` (Optional, String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Read operations occur during any refresh or planning operation when refresh is enabled.
- `update` (Optional, String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).


<a id="nestedatt--attachment"></a>
### Nested Schema for `attachment`

- `id` (String) Internet Gateway attachment ID.
- `provisioning_status` (String) Provisioning status of the attachment.


<a id="nestedatt--vpc"></a>
### Nested Schema for `vpc`

- `cidr_block` (String) IPv4 CIDR block of the connected VPC.
- `id` (String) Connected VPC ID.
- `name` (String) Connected VPC name.


## Import

Import is supported using the following syntax:

The [`terraform import` command](https://developer.hashicorp.com/terraform/cli/commands/import) can be used, for example:

```shell
$ terraform import kakaocloud_internet_gateway.example <internet_gateway_id>
```

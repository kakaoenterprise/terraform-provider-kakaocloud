---
page_title: "kakaocloud_internet_gateway_attachment Resource - kakaocloud"
subcategory: "VPC"
description: |-
  Creates and manages an attachment between an Internet Gateway and a VPC in KakaoCloud.
---

# kakaocloud_internet_gateway_attachment (Resource)

The `kakaocloud_internet_gateway_attachment` resource creates and manages an attachment between an Internet Gateway and a VPC in KakaoCloud.

Use this resource to connect an existing Internet Gateway to a VPC. The attachment controls the VPC's internet connectivity and exposes its provisioning status.

## Example Usage

```terraform
# kakaocloud_internet_gateway_attachment Terraform Resource Example

resource "kakaocloud_vpc" "example" {
  name       = "example-vpc"
  cidr_block = "10.0.0.0/16"
}

resource "kakaocloud_internet_gateway" "example" {}

# Basic Usage (kakaocloud_internet_gateway_attachment)
resource "kakaocloud_internet_gateway_attachment" "example" {
  internet_gateway_id = kakaocloud_internet_gateway.example.id
  vpc_id              = kakaocloud_vpc.example.id
}
```

## Argument Reference

- `internet_gateway_id` (Required, String) ID of the Internet Gateway to attach.
- `vpc_id` (Required, String) ID of the VPC to attach to the Internet Gateway.

- `timeouts` (Optional, Attributes) Custom timeout settings. (see [below for nested schema](#nestedatt--timeouts))

## Attribute Reference

- `id` (String) Internet Gateway attachment ID.
- `provisioning_status` (String) Provisioning status of the attachment.

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
$ terraform import kakaocloud_internet_gateway_attachment.example <internet_gateway_id>/<vpc_id>
```

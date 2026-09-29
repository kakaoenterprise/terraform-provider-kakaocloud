---
page_title: "kakaocloud_route_table_ingress_routing Resource - kakaocloud"
subcategory: "VPC"
description: |-
  Creates and manages ingress routing between a route table and an Internet Gateway in a KakaoCloud VPC.
---

# kakaocloud_route_table_ingress_routing (Resource)

The `kakaocloud_route_table_ingress_routing` resource creates and manages ingress routing between a route table and an Internet Gateway in a KakaoCloud VPC.

Use this resource to route traffic entering through an Internet Gateway according to the routes in a specified route table. The Internet Gateway must be attached to the VPC before ingress routing can be enabled.

-> **Note:** The Internet Gateway attachment must be created before ingress routing. Referencing the Internet Gateway and VPC IDs does not establish a dependency on `kakaocloud_internet_gateway_attachment`, so Terraform does not wait for the attachment to be created. Add `depends_on = [kakaocloud_internet_gateway_attachment.example]`, as shown in the example, to enforce the required order: attachment then ingress routing during creation, and ingress routing then attachment during deletion.

-> **Note:** If the attachment and ingress routing are managed in separate Terraform states, `depends_on` cannot enforce their order. Create the attachment before the ingress routing, and delete the ingress routing before the attachment.

-> **Note:** A route table can have only one ingress routing connection. If ingress routing is already configured for the route table, import the existing connection instead of creating another resource.

## Example Usage

```terraform
# kakaocloud_route_table_ingress_routing Terraform Resource Example

resource "kakaocloud_vpc" "example" {
  name       = "example-vpc"
  cidr_block = "10.0.0.0/16"
}

resource "kakaocloud_internet_gateway" "example" {}

resource "kakaocloud_internet_gateway_attachment" "example" {
  internet_gateway_id = kakaocloud_internet_gateway.example.id
  vpc_id              = kakaocloud_vpc.example.id
}

resource "kakaocloud_route_table" "example" {
  name   = "example-route-table"
  vpc_id = kakaocloud_vpc.example.id
}

# Basic Usage (kakaocloud_route_table_ingress_routing)
resource "kakaocloud_route_table_ingress_routing" "example" {
  internet_gateway_id = kakaocloud_internet_gateway.example.id
  route_table_id      = kakaocloud_route_table.example.id
  vpc_id              = kakaocloud_vpc.example.id

  depends_on = [kakaocloud_internet_gateway_attachment.example]
}
```

## Argument Reference

- `internet_gateway_id` (Required, String) ID of the Internet Gateway connected to the VPC. Changing this value replaces the resource.
- `route_table_id` (Required, String) ID of the route table to use for ingress routing. Changing this value replaces the resource.
- `vpc_id` (Required, String) ID of the VPC that contains the route table and is attached to the Internet Gateway. Changing this value replaces the resource.

- `timeouts` (Optional, Attributes) Custom timeout settings. (see [below for nested schema](#nestedatt--timeouts))

## Attribute Reference

- `id` (String) Ingress routing connection ID.
- `provisioning_status` (String) Provisioning status of the ingress routing connection.

<a id="nestedatt--timeouts"></a>
### Nested Schema for `timeouts`

- `create` (Optional, String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).
- `delete` (Optional, String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Setting a timeout for a Delete operation is only applicable if changes are saved into state before the destroy operation occurs.
- `read` (Optional, String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Read operations occur during any refresh or planning operation when refresh is enabled.


## Import

Import is supported using the following syntax:

The [`terraform import` command](https://developer.hashicorp.com/terraform/cli/commands/import) can be used, for example:

```shell
$ terraform import kakaocloud_route_table_ingress_routing.example <route_table_id>
```

-> **Note:** The `id` attribute is the ingress routing connection ID, but the import ID is the route table ID.

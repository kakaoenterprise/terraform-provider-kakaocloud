---
page_title: "kakaocloud_internet_gateway Data Source - kakaocloud"
subcategory: "VPC"
description: |-
  Retrieves detailed information about a specific Internet Gateway in KakaoCloud.
---

# kakaocloud_internet_gateway (Data Source)

The `kakaocloud_internet_gateway` data source retrieves detailed information about a specific Internet Gateway in KakaoCloud.

An Internet Gateway provides connectivity between a VPC and the internet. Use this data source to look up an existing Internet Gateway by its ID and reference its VPC, attachment, and NAT IP address without managing its lifecycle.

## Example Usage

```terraform
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# Retrieve a specific Internet Gateway by ID
data "kakaocloud_internet_gateway" "example" {
  id = "<your-internet-gateway-id>"
}

# Output basic Internet Gateway information
output "internet_gateway_basic" {
  description = "Basic information about the Internet Gateway"
  value = {
    id                  = data.kakaocloud_internet_gateway.example.id
    name                = data.kakaocloud_internet_gateway.example.name
    nat_ip              = data.kakaocloud_internet_gateway.example.nat_ip
    provisioning_status = data.kakaocloud_internet_gateway.example.provisioning_status
  }
}

# Output the connected VPC
output "internet_gateway_vpc" {
  description = "VPC connected to the Internet Gateway"
  value       = data.kakaocloud_internet_gateway.example.vpc
}

# Output attachment information
output "internet_gateway_attachment" {
  description = "Attachment information for the Internet Gateway"
  value       = data.kakaocloud_internet_gateway.example.attachment
}
```

## Argument Reference

- `id` (Required, String) Internet Gateway ID.

- `timeouts` (Optional, Attributes) Custom timeout settings. (see [below for nested schema](#nestedatt--timeouts))

## Attribute Reference

The following attributes are exported:

- `attachment` (Attributes) Attachment information for the VPC connection. (see [below for nested schema](#nestedatt--attachment))
- `created_at` (String) Time when the resource was created.<br/> - ISO_8601 format<br/> - UTC standard
- `description` (String) Internet Gateway description.
- `domain_id` (String) Domain ID that owns the Internet Gateway.
- `name` (String) Internet Gateway name.
- `nat_ip` (String) Public NAT IP address assigned to the Internet Gateway.
- `project_id` (String) Project ID that owns the Internet Gateway.
- `project_name` (String) Project name that owns the Internet Gateway.
- `provisioning_status` (String) Provisioning status of the Internet Gateway.
- `updated_at` (String) Time when the resource was last updated.<br/> - ISO_8601 format<br/> - UTC standard
- `vpc` (Attributes) Connected VPC information. (see [below for nested schema](#nestedatt--vpc))

<a id="nestedatt--timeouts"></a>
### Nested Schema for `timeouts`

- `read` (Optional, String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).


<a id="nestedatt--attachment"></a>
### Nested Schema for `attachment`

- `id` (String) Internet Gateway attachment ID.
- `provisioning_status` (String) Provisioning status of the attachment.


<a id="nestedatt--vpc"></a>
### Nested Schema for `vpc`

- `cidr_block` (String) IPv4 CIDR block of the connected VPC.
- `id` (String) Connected VPC ID.
- `name` (String) Connected VPC name.

---
page_title: "kakaocloud_internet_gateways Data Source - kakaocloud"
subcategory: "VPC"
description: |-
  Retrieves a list of Internet Gateways in KakaoCloud.
---

# kakaocloud_internet_gateways (Data Source)

The `kakaocloud_internet_gateways` data source retrieves a list of Internet Gateways in KakaoCloud.

Use this data source to list or filter existing Internet Gateways and reference their VPC, attachment, and NAT IP address without managing their lifecycle.

## Available Filters

| Filter | Type | Description |
|--------|------|-------------|
| `id` | string | Internet Gateway ID |
| `name` | string | Internet Gateway name |
| `vpc_id` | string | ID of the connected VPC |
| `vpc_name` | string | Name of the connected VPC |
| `vpc_cidr_block` | string | IPv4 CIDR block of the connected VPC |
| `attachment_id` | string | Internet Gateway attachment ID |
| `attachment_provisioning_status` | string | Attachment provisioning status |
| `nat_ip` | string | Public NAT IP address assigned to the Internet Gateway |
| `provisioning_status` | string | Internet Gateway provisioning status |
| `created_at` | string | Resource creation time<br/>- ISO_8601 format<br/>- UTC |
| `updated_at` | string | Resource last updated time<br/>- ISO_8601 format<br/>- UTC |

## Example Usage

```terraform
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# List all Internet Gateways
data "kakaocloud_internet_gateways" "all" {}

# List Internet Gateways connected to a specific VPC
data "kakaocloud_internet_gateways" "by_vpc" {
  filter = [
    {
      name  = "vpc_id"
      value = "<your-vpc-id>"
    }
  ]
}

# Output all Internet Gateways
output "all_internet_gateways" {
  description = "All Internet Gateways"
  value       = data.kakaocloud_internet_gateways.all.igws
}

# Output selected information for Internet Gateways connected to the VPC
output "internet_gateways_by_vpc" {
  description = "Internet Gateways connected to the specified VPC"
  value = [
    for igw in data.kakaocloud_internet_gateways.by_vpc.igws : {
      id                  = igw.id
      name                = igw.name
      nat_ip              = igw.nat_ip
      provisioning_status = igw.provisioning_status
      vpc                 = igw.vpc
      attachment          = igw.attachment
    }
  ]
}
```

## Argument Reference

- `filter` (Optional, Attributes List) Filters to narrow down the returned results. (see [below for nested schema](#nestedatt--filter))
- `timeouts` (Optional, Attributes) Custom timeout settings. (see [below for nested schema](#nestedatt--timeouts))

## Attribute Reference

The following attributes are exported:

- `igws` (Attributes List) List of Internet Gateways. (see [below for nested schema](#nestedatt--igws))

<a id="nestedatt--filter"></a>
### Nested Schema for `filter`

- `name` (Required, String) Name of the attribute to filter by.

- `value` (Optional, String) Value to match for the specified filter attribute.


<a id="nestedatt--timeouts"></a>
### Nested Schema for `timeouts`

- `read` (Optional, String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).


<a id="nestedatt--igws"></a>
### Nested Schema for `igws`

- `attachment` (Attributes) Attachment information for the VPC connection. (see [below for nested schema](#nestedatt--igws--attachment))
- `attachment_status` (String) Connection status of the Internet Gateway attachment. Possible values: `ATTACHED`, `ATTACHING`, `DETACHING`, `DETACHED`.
- `created_at` (String) Time when the resource was created.<br/> - ISO_8601 format<br/> - UTC standard
- `description` (String) Internet Gateway description.
- `domain_id` (String) Domain ID that owns the Internet Gateway.
- `id` (String) Internet Gateway ID.
- `name` (String) Internet Gateway name.
- `nat_ip` (String) Public NAT IP address assigned to the Internet Gateway.
- `project_id` (String) Project ID that owns the Internet Gateway.
- `project_name` (String) Project name that owns the Internet Gateway.
- `provisioning_status` (String) Provisioning status of the Internet Gateway.
- `updated_at` (String) Time when the resource was last updated.<br/> - ISO_8601 format<br/> - UTC standard
- `vpc` (Attributes) Connected VPC information. (see [below for nested schema](#nestedatt--igws--vpc))

<a id="nestedatt--igws--attachment"></a>
### Nested Schema for `igws.attachment`

- `id` (String) Internet Gateway attachment ID.
- `provisioning_status` (String) Provisioning status of the attachment.


<a id="nestedatt--igws--vpc"></a>
### Nested Schema for `igws.vpc`

- `cidr_block` (String) IPv4 CIDR block of the connected VPC.
- `id` (String) Connected VPC ID.
- `name` (String) Connected VPC name.

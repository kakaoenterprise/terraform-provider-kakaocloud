---
page_title: "kakaocloud_transit_gateway_attachment Data Source - kakaocloud"
subcategory: "Transit Gateway"
description: |-
  Retrieves detailed information about a specific Transit Gateway (TGW) attachment in KakaoCloud.
---

# kakaocloud_transit_gateway_attachment (Data Source)

The `kakaocloud_transit_gateway_attachment` data source retrieves metadata and configuration details for a specific Transit Gateway (TGW) attachment in KakaoCloud.

A TGW attachment represents a connection between a Transit Gateway and a networking resource, such as a VPC and its subnets.  
Use this data source to reference an existing attachment by its ID and consume its attributes in other Terraform configurations.

## Example Usage

```terraform
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# Retrieve a specific Transit Gateway attachment by ID
data "kakaocloud_transit_gateway_attachment" "example" {
  id = "<your-tgw-attachment-id>"
}

# Output basic attachment information
output "tgw_attachment_basic" {
  description = "Basic information of the Transit Gateway attachment"
  value = {
    id                  = data.kakaocloud_transit_gateway_attachment.example.id
    provisioning_status = data.kakaocloud_transit_gateway_attachment.example.provisioning_status
    resource_type       = data.kakaocloud_transit_gateway_attachment.example.resource_type
    resource_name       = data.kakaocloud_transit_gateway_attachment.example.resource_name
  }
}

# Output attached subnets
output "tgw_attachment_resources" {
  description = "Subnets attached to the Transit Gateway attachment"
  value       = data.kakaocloud_transit_gateway_attachment.example.resources
}

# Output associated route table
output "tgw_attachment_route_table" {
  description = "Route table associated with the attachment"
  value       = data.kakaocloud_transit_gateway_attachment.example.route_table
}
```

## Argument Reference

- `id` (Required, String) ID of the Transit Gateway attachment to retrieve.

- `timeouts` (Optional, Attributes) Read operation timeout configuration. (see [below for nested schema](#nestedatt--timeouts))

## Attribute Reference

The following attributes are exported:

- `created_at` (String) Time when the resource was created<br/> - ISO_8601 format<br/> - UTC standard
- `project_id` (String) Project ID that owns the attachment.
- `project_name` (String) Project name that owns the attachment.
- `provisioning_status` (String) Current provisioning status of the attachment.
- `resource_cidr_block` (String) CIDR block of the attached resource.
- `resource_id` (String) ID of the attached resource.
- `resource_name` (String) Name of the attached resource.
- `resource_type` (String) Type of the attached resource.
- `resources` (Attributes List) List of attached resources (subnets). (see [below for nested schema](#nestedatt--resources))
- `route_table` (Attributes) Route table associated with the attachment. (see [below for nested schema](#nestedatt--route_table))
- `tgw` (Attributes) Transit Gateway associated with the attachment. (see [below for nested schema](#nestedatt--tgw))
- `updated_at` (String) Time when the resource was last updated<br/> - ISO_8601 format<br/> - UTC standard

<a id="nestedatt--timeouts"></a>
### Nested Schema for `timeouts`

- `read` (Optional, String) Duration to wait for read operations, such as `"30s"` or `"2h45m"`.


<a id="nestedatt--resources"></a>
### Nested Schema for `resources`

- `availability_zone` (String) Availability zone of the attached subnet.
- `cidr_block` (String) CIDR block assigned to the subnet.
- `created_at` (String) Time when the resource was created<br/> - ISO_8601 format<br/> - UTC standard
- `description` (String) Description of the subnet.
- `id` (String) Subnet ID.
- `name` (String) Subnet name.
- `operating_status` (String) Current operating status of the subnet.
- `provisioning_status` (String) Provisioning status of the subnet.
- `updated_at` (String) Time when the resource was last updated<br/> - ISO_8601 format<br/> - UTC standard
- `vpc_id` (String) VPC ID to which the subnet belongs.


<a id="nestedatt--route_table"></a>
### Nested Schema for `route_table`

- `created_at` (String) Time when the resource was created<br/> - ISO_8601 format<br/> - UTC standard
- `id` (String) Route table ID.
- `is_default_association_route_table` (Boolean) Whether this is the default association route table.
- `is_default_propagation_route_table` (Boolean) Whether this is the default propagation route table.
- `name` (String) Route table name.
- `project_id` (String) Project ID that owns the route table.
- `provisioning_status` (String) Provisioning status of the route table.
- `region` (String) Region where the route table is located.
- `tgw_id` (String) Transit Gateway ID associated with the route table.
- `updated_at` (String) Time when the resource was last updated<br/> - ISO_8601 format<br/> - UTC standard


<a id="nestedatt--tgw"></a>
### Nested Schema for `tgw`

- `id` (String) Transit Gateway ID.
- `name` (String) Transit Gateway name.
- `project_id` (String) Project ID that owns the Transit Gateway.
- `project_name` (String) Project name that owns the Transit Gateway.

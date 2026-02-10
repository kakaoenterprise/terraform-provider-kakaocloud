---
page_title: "kakaocloud_transit_gateway_attachments Data Source - kakaocloud"
subcategory: "Transit Gateway"
description: |-
  Retrieves a list of Transit Gateway (TGW) attachments in KakaoCloud.
---

# kakaocloud_transit_gateway_attachments (Data Source)

The `kakaocloud_transit_gateway_attachments` data source retrieves a list of Transit Gateway (TGW) attachments in KakaoCloud.

A TGW attachment represents a connection between a Transit Gateway and a network resource, such as a VPC or its subnets.  
Use this data source to list and filter existing TGW attachments and reference their details in other Terraform configurations without managing their lifecycle.

## Available Filters

| Filter                | Type                | Description |
|-----------------------|---------------------|-------------|
| `id`                  | string              | TGW attachment ID |
| `tgw_id`              | string              | Transit Gateway ID |
| `tgw_name`            | string              | Transit Gateway name |
| `provisioning_status` | ProvisioningStatus  | Attachment provisioning status <br/>Possible values: `ACTIVE`, `DELETED`, `ERROR`, `PENDING_CREATE`, `PENDING_UPDATE`, `PENDING_DELETE` |
| `resource_id`         | string              | ID of the attached resource |
| `resource_name`       | string              | Name of the attached resource |
| `route_table_id`      | string              | ID of the associated route table |
| `route_table_name`    | string              | Name of the associated route table |
| `created_at`          | string              | Resource creation time <br/>- ISO_8601 format <br/>- UTC |
| `updated_at`          | string              | Resource last updated time <br/>- ISO_8601 format <br/>- UTC |

## Example Usage

```terraform
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# List all Transit Gateway attachments
data "kakaocloud_transit_gateway_attachments" "all" {}

# List Transit Gateway attachments filtered by project ID
data "kakaocloud_transit_gateway_attachments" "by_project" {
  filter = [
    {
      name  = "project_id"
      value = "<your-project-id>"
    }
  ]
}

# Output all attachments
output "all_tgw_attachments" {
  description = "All Transit Gateway attachments"
  value       = data.kakaocloud_transit_gateway_attachments.all.transit_gateway_attachments
}

# Output filtered attachments
output "filtered_tgw_attachments" {
  description = "Transit Gateway attachments filtered by project ID"
  value = [
    for att in data.kakaocloud_transit_gateway_attachments.by_project.transit_gateway_attachments : {
      id                  = att.id
      provisioning_status = att.provisioning_status
      resource_type       = att.resource_type
      resource_name       = att.resource_name
      project_name        = att.project_name
    }
  ]
}
```

## Argument Reference

- `filter` (Optional, Attributes List) One or more filters used to limit the returned TGW attachments. (see [below for nested schema](#nestedatt--filter))
- `timeouts` (Optional, Attributes) Read operation timeout configuration. (see [below for nested schema](#nestedatt--timeouts))

## Attribute Reference

The following attributes are exported:

- `transit_gateway_attachments` (Attributes List) List of Transit Gateway attachments returned by the API. (see [below for nested schema](#nestedatt--transit_gateway_attachments))

<a id="nestedatt--filter"></a>
### Nested Schema for `filter`

- `name` (Required, String) Filter field name (for example, `project_id`).

- `value` (Optional, String) Filter value to match.


<a id="nestedatt--timeouts"></a>
### Nested Schema for `timeouts`

- `read` (Optional, String) Maximum duration to wait for the read operation.


<a id="nestedatt--transit_gateway_attachments"></a>
### Nested Schema for `transit_gateway_attachments`

- `created_at` (String) Time when the resource was created<br/> - ISO_8601 format<br/> - UTC standard
- `id` (String) TGW attachment ID.
- `project_id` (String) Project ID that owns the attachment.
- `project_name` (String) Project name that owns the attachment.
- `provisioning_status` (String) Current provisioning status of the attachment.
- `resource_cidr_block` (String) CIDR block of the attached resource.
- `resource_id` (String) ID of the attached resource.
- `resource_name` (String) Name of the attached resource.
- `resource_type` (String) Type of the attached resource.
- `resources` (Attributes List) List of attached subnet resources. (see [below for nested schema](#nestedatt--transit_gateway_attachments--resources))
- `route_table` (Attributes) Route table associated with the attachment. (see [below for nested schema](#nestedatt--transit_gateway_attachments--route_table))
- `tgw` (Attributes) Transit Gateway associated with the attachment. (see [below for nested schema](#nestedatt--transit_gateway_attachments--tgw))
- `updated_at` (String) Time when the resource was last updated<br/> - ISO_8601 format<br/> - UTC standard

<a id="nestedatt--transit_gateway_attachments--resources"></a>
### Nested Schema for `transit_gateway_attachments.resources`

- `availability_zone` (String) Availability zone of the subnet.
- `cidr_block` (String) CIDR block of the subnet.
- `created_at` (String) Time when the resource was created<br/> - ISO_8601 format<br/> - UTC standard
- `description` (String) Subnet description.
- `id` (String) Subnet ID.
- `name` (String) Subnet name.
- `operating_status` (String) Current operating status of the subnet.
- `provisioning_status` (String) Provisioning status of the subnet.
- `updated_at` (String) Time when the resource was last updated<br/> - ISO_8601 format<br/> - UTC standard
- `vpc_id` (String) VPC ID to which the subnet belongs.


<a id="nestedatt--transit_gateway_attachments--route_table"></a>
### Nested Schema for `transit_gateway_attachments.route_table`

- `created_at` (String) Time when the resource was created<br/> - ISO_8601 format<br/> - UTC standard
- `id` (String) Route table ID.
- `is_default_association_route_table` (Boolean) Whether this is the default association route table.
- `is_default_propagation_route_table` (Boolean) Whether this is the default propagation route table.
- `name` (String) Route table name.
- `project_id` (String) Project ID that owns the route table.
- `provisioning_status` (String) Route table provisioning status.
- `region` (String) Region where the route table is located.
- `tgw_id` (String) Associated Transit Gateway ID.
- `updated_at` (String) Time when the resource was last updated<br/> - ISO_8601 format<br/> - UTC standard


<a id="nestedatt--transit_gateway_attachments--tgw"></a>
### Nested Schema for `transit_gateway_attachments.tgw`

- `id` (String) Transit Gateway ID.
- `name` (String) Transit Gateway name.
- `project_id` (String) Project ID that owns the Transit Gateway.
- `project_name` (String) Project name that owns the Transit Gateway.




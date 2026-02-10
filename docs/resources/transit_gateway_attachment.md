---
page_title: "kakaocloud_transit_gateway_attachment Resource - kakaocloud"
subcategory: "Transit Gateway"
description: |-
  Creates and manages a Transit Gateway (TGW) attachment in KakaoCloud.
---

# kakaocloud_transit_gateway_attachment (Resource)

The `kakaocloud_transit_gateway_attachment` resource creates and manages a Transit Gateway (TGW) attachment in KakaoCloud.

A TGW attachment establishes a connection between a Transit Gateway and a network resource, such as a VPC, by specifying one or more subnets.  
After the attachment is created, it may remain in a pending state until it is approved, depending on the Transit Gateway’s sharing and attachment approval configuration.

## Example Usage

```terraform
resource "kakaocloud_transit_gateway_attachment" "example" {
  resource_id = kakaocloud_vpc.example.id
  subnet_ids  = [
    kakaocloud_subnet.example.id
  ]
  tgw_id      = kakaocloud_transit_gateway.example.id
}
```

## Argument Reference

- `resource_id` (Required, String) ID of the resource to attach, such as a VPC.
- `subnet_ids` (Required, Set of String) One or more subnet IDs to include in the attachment.
- `tgw_id` (Required, String) ID of the Transit Gateway to attach the resource to.

- `timeouts` (Optional, Attributes) Timeout configuration for create, read, update, and delete operations. (see [below for nested schema](#nestedatt--timeouts))

## Attribute Reference

- `created_at` (String) Time when the resource was created<br/> - ISO_8601 format<br/> - UTC standard
- `id` (String) Transit Gateway attachment ID.
- `project_id` (String) Project ID that owns the attachment.
- `project_name` (String) Project name that owns the attachment.
- `provisioning_status` (String) Current provisioning status of the attachment.
- `resource_cidr_block` (String) CIDR block of the attached resource.
- `resource_name` (String) Name of the attached resource.
- `resource_type` (String) Type of the attached resource.
- `resources` (Attributes List) List of attached subnet resources. (see [below for nested schema](#nestedatt--resources))
- `route_table` (Attributes) Route table associated with the attachment. (see [below for nested schema](#nestedatt--route_table))
- `tgw` (Attributes) Transit Gateway associated with the attachment. (see [below for nested schema](#nestedatt--tgw))
- `updated_at` (String) Time when the resource was last updated<br/> - ISO_8601 format<br/> - UTC standard

<a id="nestedatt--timeouts"></a>
### Nested Schema for `timeouts`

- `create` (Optional, String) Duration to wait for the create operation.
- `delete` (Optional, String) Duration to wait for the delete operation. This is only applied if the resource state is saved before destroy.
- `read` (Optional, String) Duration to wait for read operations during refresh or planning.
- `update` (Optional, String) Duration to wait for the update operation.


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


## Import

Import is supported using the following syntax:

The [`terraform import` command](https://developer.hashicorp.com/terraform/cli/commands/import) can be used, for example:

```shell
$ terraform import kakaocloud_transit_gateway_attachment.example <transit_gateway_attachment_id>
```


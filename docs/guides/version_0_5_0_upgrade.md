---
page_title: "KakaoCloud Provider Version 0.5.0 Upgrade Guide"
subcategory: "Guides"
description: |-
  Describes the required configuration changes when upgrading the KakaoCloud Terraform Provider to version 0.5.0.
---

# KakaoCloud Provider Version 0.5.0 Upgrade Guide

Version 0.5.0 of the KakaoCloud Terraform Provider changes the schemas of the VPC and Route Table resources and removes
the Transit Gateway Attachment Approval resource.

Existing Terraform configurations may fail during `terraform validate` or `terraform plan` if they use the removed
arguments. Apply the changes in this guide before upgrading the provider.

## Summary of Changes

| Resource | Change | Required action |
|----------|--------|-----------------|
| `kakaocloud_vpc` | The `subnet` argument was removed. | Remove `subnet` from the VPC resource and manage the subnet with a separate `kakaocloud_subnet` resource if necessary. |
| `kakaocloud_route_table` | The `instance` value was removed from `request_routes.target_type`. | Set `target_type` to `network-interface` and set `target_id` to a Network Interface ID. |
| `kakaocloud_transit_gateway_attachment_approval` | The resource was removed and replaced by the `kakaocloud_transit_gateway_attachment_approve` action. | Remove the resource from the configuration and Terraform state before upgrading the provider. |

## Before You Upgrade

Back up the current Terraform state:

```shell
$ terraform state pull > terraform-state-backup.json
```

Review the backup and keep it in a secure location because Terraform state can contain sensitive information.

## Remove the Transit Gateway Attachment Approval Resource

The `kakaocloud_transit_gateway_attachment_approval` resource has been removed and replaced by the
`kakaocloud_transit_gateway_attachment_approve` action.

Before upgrading the provider, remove every `kakaocloud_transit_gateway_attachment_approval` block from the Terraform
configuration. Also remove each corresponding resource instance from Terraform state:

```shell
$ terraform state rm kakaocloud_transit_gateway_attachment_approval.example
```

The `terraform state rm` command removes the resource from Terraform state without deleting the Transit Gateway
attachment in KakaoCloud. If the resource uses `count` or `for_each`, specify and remove every instance address.

Complete the following steps for every affected approval resource:

1. Confirm that the Transit Gateway attachment has already been approved.
2. Back up the current Terraform state.
3. Remove the `kakaocloud_transit_gateway_attachment_approval` block from the configuration.
4. Remove the corresponding resource instance from Terraform state.
5. Confirm that the removed resource address no longer appears in `terraform state list`.

## Remove `subnet` from the VPC Resource

The `subnet` argument, which previously allowed a subnet to be defined inside a `kakaocloud_vpc` resource, has been
removed. In version 0.5.0, VPCs and subnets must be managed as separate resources.

If `subnet` remains in the VPC configuration, Terraform returns an `Unsupported argument` error.

### Configuration Before Version 0.5.0

```terraform
resource "kakaocloud_vpc" "example" {
  name       = "example-vpc"
  cidr_block = "10.0.0.0/16"

  subnet = {
    availability_zone = "kr-central-2-a"
    cidr_block        = "10.0.1.0/24"
  }
}
```

### Configuration for Version 0.5.0

Remove `subnet` from the VPC resource:

```terraform
resource "kakaocloud_vpc" "example" {
  name       = "example-vpc"
  cidr_block = "10.0.0.0/16"
}
```

To continue managing the subnet with Terraform, declare a separate `kakaocloud_subnet` resource:

```terraform
resource "kakaocloud_subnet" "example" {
  vpc_id            = kakaocloud_vpc.example.id
  name              = "example-subnet"
  cidr_block        = "10.0.1.0/24"
  availability_zone = "kr-central-2-a"
}
```

### Import an Existing Subnet

If the subnet already exists, import it into the separate resource by using its Subnet ID:

```shell
$ terraform import kakaocloud_subnet.example <subnet_id>
```

~> **Warning:** If you add a `kakaocloud_subnet` resource without importing the existing subnet, Terraform may attempt
to create another subnet with the same or an overlapping CIDR block. Check the result of `terraform plan` before
applying any changes.

Complete the following steps for every affected VPC:

1. Remove `subnet` from the `kakaocloud_vpc` resource.
2. Decide whether Terraform should continue managing the existing subnet.
3. If Terraform should manage it, declare a separate `kakaocloud_subnet` resource.
4. Import the existing subnet into that resource.
5. Confirm that `terraform plan` does not propose creating or deleting the subnet.

## Change Instance Route Targets to Network Interfaces

The `instance` value has been removed from `kakaocloud_route_table.request_routes.target_type`. A route that previously
targeted an Instance must now target a Network Interface attached to that Instance.

| Argument | Before version 0.5.0 | Version 0.5.0 |
|----------|----------------------|---------------|
| `target_type` | `instance` | `network-interface` |
| `target_id` | Instance ID | Network Interface ID attached to the Instance |

Provider validation or the API request fails if the configuration continues to use `target_type = "instance"`. Changing
only `target_type` is not sufficient: `target_id` must also be changed from the Instance ID to the corresponding Network
Interface ID.

### Configuration Before Version 0.5.0

```terraform
resource "kakaocloud_route_table" "example" {
  name   = "example-route-table"
  vpc_id = kakaocloud_vpc.example.id

  request_routes = [
    {
      destination = "10.10.0.0/16"
      target_type = "instance"
      target_id   = kakaocloud_instance.example.id
    }
  ]
}
```

### Configuration for Version 0.5.0

Use the Network Interface ID exposed by the Instance resource:

```terraform
resource "kakaocloud_route_table" "example" {
  name   = "example-route-table"
  vpc_id = kakaocloud_vpc.example.id

  request_routes = [
    {
      destination = "10.10.0.0/16"
      target_type = "network-interface"
      target_id   = kakaocloud_instance.example.subnets[0].network_interface_id
    }
  ]
}
```

If the Network Interface is managed as a separate Terraform resource, reference its ID directly:

```terraform
resource "kakaocloud_route_table" "example" {
  name   = "example-route-table"
  vpc_id = kakaocloud_vpc.example.id

  request_routes = [
    {
      destination = "10.10.0.0/16"
      target_type = "network-interface"
      target_id   = kakaocloud_network_interface.example.id
    }
  ]
}
```

Complete the following steps for every affected route:

1. Find each route that uses `target_type = "instance"`.
2. Identify the Network Interface attached to the Instance.
3. Change `target_type` to `network-interface`.
4. Change `target_id` from the Instance ID to the Network Interface ID.
5. Confirm that `terraform plan` does not propose unintended Route Table or route replacement.

~> **Warning:** `target_id` must contain a Network Interface ID, not an Instance ID.

## Upgrade Procedure

1. Back up the current Terraform state.
2. Remove all `kakaocloud_transit_gateway_attachment_approval` resources from the configuration and Terraform state.
3. Apply all other required configuration changes described in this guide.
4. Update the provider version constraint:

   ```terraform
   terraform {
     required_providers {
       kakaocloud = {
         source  = "kakaoenterprise/kakaocloud"
         version = "~> 0.5.0"
       }
     }
   }
   ```

5. Upgrade the provider:

   ```shell
   $ terraform init -upgrade
   ```

6. Import any existing subnets that are now represented by separate `kakaocloud_subnet` resources.
7. Validate the configuration:

   ```shell
   $ terraform validate
   ```

8. Review the execution plan:

   ```shell
   $ terraform plan
   ```

   Verify that the plan does not include any unintended changes to the following resources:

   - VPC deletion or replacement
   - Subnet deletion or duplicate creation
   - Route Table deletion or replacement
   - Unintended deletion of existing routes
   - Instance or Network Interface changes

9. Apply the changes only when the execution plan matches the intended migration:

   ```shell
   $ terraform apply
   ```

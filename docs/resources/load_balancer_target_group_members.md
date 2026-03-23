---
page_title: "kakaocloud_load_balancer_target_group_members Resource - kakaocloud"
subcategory: "Load Balancer"
description: |-
  Manages kakaocloud_load_balancer_target_group_members
---

# kakaocloud_load_balancer_target_group_members (Resource)

Manages `kakaocloud_load_balancer_target_group_members`.  
This resource manages the members (targets) associated with a specific Load Balancer target group and keeps the target group membership synchronized with the configuration defined in Terraform.

> ⚠️ Note:
> - Do not use this resource together with individual `kakaocloud_load_balancer_target_group_member` resources for the same target group. Doing so will cause conflicts.

## Example Usage

```terraform
# kakaocloud_load_balancer_target_group_members terraform resource example

resource "kakaocloud_load_balancer_target_group_members" "example" {
  target_group_id = kakaocloud_load_balancer_target_group.example.id

  members = [
    {
      address       = "10.0.1.10"
      protocol_port = 80
      subnet_id     = kakaocloud_subnet.example.id
    },
    {
      address       = "10.0.1.11"
      protocol_port = 80
      subnet_id     = kakaocloud_subnet.example.id
    }
  ]
}
```

## Argument Reference

- `members` (Required, Attributes List) A list of target members to register in the target group. (see [below for nested schema](#nestedatt--members))
- `target_group_id` (Required, String) The ID of the target group in which the members will be managed.

- `timeouts` (Optional, Attributes) Timeout configuration for create, read, update, and delete operations. (see [below for nested schema](#nestedatt--timeouts))

<a id="nestedatt--members"></a>
### Nested Schema for `members`

- `address` (Required, String) The IP address of the target instance.
- `protocol_port` (Required, Number) The port number on which the target instance receives traffic.
- `subnet_id` (Required, String) The ID of the subnet where the target instance is located.

- `monitor_port` (Optional, Number) The port used for health checks.
- `name` (Optional, String) A name assigned to the target member.
- `weight` (Optional, Number) The traffic distribution weight of the target member.


<a id="nestedatt--timeouts"></a>
### Nested Schema for `timeouts`

- `create` (Optional, String) A string that can be parsed as a duration such as "30s" or "2h45m". Valid time units are "s", "m", and "h".
- `delete` (Optional, String) A string that can be parsed as a duration such as "30s" or "2h45m". This timeout applies only if changes are saved into state before the destroy operation occurs.
- `read` (Optional, String) A string that can be parsed as a duration such as "30s" or "2h45m". Read operations occur during refresh or planning when refresh is enabled.
- `update` (Optional, String) A string that can be parsed as a duration such as "30s" or "2h45m". Valid time units are "s", "m", and "h".


## Import

This resource supports import using the following format:

```bash
terraform import kakaocloud_load_balancer_target_group_members.example <target_group_id>
```

> ⚠️ Note: When importing, ensure the `members` list in your configuration is **sorted by IP address (`address`)**.  
> Any difference in order compared to the actual target group state will result in a diff and **force resource replacement** during the next apply.


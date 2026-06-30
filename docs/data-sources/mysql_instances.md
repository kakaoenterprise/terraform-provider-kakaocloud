---
page_title: "kakaocloud_mysql_instances Data Source - kakaocloud"
subcategory: "MySQL"
description: |-
  The kakaocloud_mysql_instances data source retrieves instances that belong to a KakaoCloud MySQL instance group.
---

# kakaocloud_mysql_instances (Data Source)

The `kakaocloud_mysql_instances` data source retrieves instances that belong to a KakaoCloud MySQL instance group.

Use this data source to inspect primary and standby instances, roles, statuses, and instance IDs for operational tasks
such as restart, scale-in, or log export actions.

## Example Usage

```hcl
# List instances in a MySQL instance group
data "kakaocloud_mysql_instances" "example" {
  instance_group_id = "<your-mysql-instance-group-id>"
}

output "mysql_instances" {
  value = [
    for instance in data.kakaocloud_mysql_instances.example.instances : {
      id     = instance.id
      role   = instance.role
      status = instance.status
    }
  ]
}
```

## Argument Reference

- `instance_group_id` (Required, String) MySQL instance group ID.

- `timeouts` (Optional, Attributes) Custom timeout settings. (See [below for nested schema](#nestedatt--timeouts).)

## Attribute Reference

The following attributes are exported:

- `instances` (Attributes List) List of MySQL instances in the instance group. (See [below for nested schema](#nestedatt--instances).)

<a id="nestedatt--timeouts"></a>
### Nested Schema for `timeouts`

- `read` (Optional, String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).


<a id="nestedatt--instances"></a>
### Nested Schema for `instances`

- `availability_status` (String) Availability status of the MySQL instance.
- `created_at` (String) Time when the instance was created.
- `data_disk_usage` (Number) Current data disk usage.
- `id` (String) MySQL instance ID.
- `instance_group_id` (String) MySQL instance group ID.
- `instance_group_name` (String) MySQL instance group name.
- `log_disk_usage` (Number) Current log disk usage.
- `name` (String) MySQL instance name.
- `project_id` (String) Project ID where the instance exists.
- `role` (String) Instance role, such as primary or standby.
- `spec_content` (Attributes) MySQL instance specification. (See [below for nested schema](#nestedatt--instances--spec_content).)
- `start_time` (String) Time when the instance was started.
- `status` (String) MySQL instance status.
- `status_content` (Attributes) Additional status details for the instance. (See [below for nested schema](#nestedatt--instances--status_content).)
- `updated_at` (String) Time when the instance was last updated.

<a id="nestedatt--instances--spec_content"></a>
### Nested Schema for `instances.spec_content`

- `availability_zone` (String) Availability zone of the instance.
- `data_disk_size` (Number) Data disk size, in GB.
- `engine_version` (String) MySQL engine version.
- `flavor_id` (String) MySQL flavor ID.
- `log_disk_size` (Number) Log disk size, in GB.
- `network_ports` (Attributes List) Network ports attached to the instance. (See [below for nested schema](#nestedatt--instances--spec_content--network_ports).)

<a id="nestedatt--instances--spec_content--network_ports"></a>
### Nested Schema for `instances.spec_content.network_ports`

- `security_group_ids` (Set of String) Security group IDs associated with the network port.
- `subnet_id` (String) Subnet ID associated with the network port.



<a id="nestedatt--instances--status_content"></a>
### Nested Schema for `instances.status_content`

- `needs_restart` (Boolean) Whether the instance must be restarted for pending changes to take effect.
- `needs_restart_reason` (List of String) Reasons why the instance needs a restart.

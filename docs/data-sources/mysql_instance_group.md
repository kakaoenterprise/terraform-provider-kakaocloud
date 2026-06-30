---
page_title: "kakaocloud_mysql_instance_group Data Source - kakaocloud"
subcategory: "MySQL"
description: |-
  The kakaocloud_mysql_instance_group data source retrieves detailed information about a specific KakaoCloud MySQL instance group.
---

# kakaocloud_mysql_instance_group (Data Source)

The `kakaocloud_mysql_instance_group` data source retrieves detailed information about a specific KakaoCloud MySQL
instance group by its ID.

It returns configuration and runtime information such as engine version, network settings, endpoints, backup schedule,
instances, parameter group, and status.

## Example Usage

```hcl
# Get a specific MySQL instance group by ID
data "kakaocloud_mysql_instance_group" "example" {
  id = "<your-mysql-instance-group-id>"
}

output "mysql_instance_group_info" {
  value = {
    id                     = data.kakaocloud_mysql_instance_group.example.id
    name                   = data.kakaocloud_mysql_instance_group.example.name
    status                 = data.kakaocloud_mysql_instance_group.example.status
    engine_version         = data.kakaocloud_mysql_instance_group.example.spec_content.engine_version
    primary_port           = data.kakaocloud_mysql_instance_group.example.spec_content.primary_port
    primary_subnet_id      = data.kakaocloud_mysql_instance_group.example.network_info.primary_subnet_info.subnet_id
    primary_instance_id    = try(data.kakaocloud_mysql_instance_group.example.instances.primary.instance_id, null)
    standby_instance_count = try(length(data.kakaocloud_mysql_instance_group.example.instances.standby), 0)
  }
}
```

## Argument Reference

- `id` (Required, String) MySQL instance group ID.

- `timeouts` (Optional, Attributes) Custom timeout settings. (See [below for nested schema](#nestedatt--timeouts).)

## Attribute Reference

The following attributes are exported:

- `backup_schedule` (Attributes) Backup schedule configuration. (See [below for nested schema](#nestedatt--backup_schedule).)
- `created_at` (String) Time when the instance group was created.
- `creator` (String) User who created the instance group.
- `description` (String) MySQL instance group description.
- `endpoint` (List of String) Connection endpoints for the MySQL instance group.
- `extra_info` (Attributes) Additional MySQL settings. (See [below for nested schema](#nestedatt--extra_info).)
- `instances` (Attributes) Primary and standby instance information. (See [below for nested schema](#nestedatt--instances).)
- `is_multi_az` (Boolean) Whether the instance group uses multiple availability zones.
- `license` (String) MySQL license information.
- `name` (String) MySQL instance group name.
- `network_info` (Attributes) Current network configuration. (See [below for nested schema](#nestedatt--network_info).)
- `parameter_group` (Attributes) Parameter group applied to the instance group. (See [below for nested schema](#nestedatt--parameter_group).)
- `project_id` (String) Project ID where the instance group exists.
- `source_backup_id` (String) Source backup ID when the instance group was restored from a backup.
- `spec_content` (Attributes) MySQL instance group specification. (See [below for nested schema](#nestedatt--spec_content).)
- `status` (String) MySQL instance group status.
- `updated_at` (String) Time when the instance group was last updated.

<a id="nestedatt--timeouts"></a>
### Nested Schema for `timeouts`

- `read` (Optional, String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).


<a id="nestedatt--backup_schedule"></a>
### Nested Schema for `backup_schedule`

- `enabled` (Boolean) Whether automated backups are enabled.
- `expiry_duration` (Number) Backup retention period, in days.
- `id` (String) Backup schedule ID.
- `start_time` (String) Backup schedule start time. (Format: `HH:mm` (UTC))
- `type` (String) Backup schedule type. (Possible values: `DAY`)


<a id="nestedatt--extra_info"></a>
### Nested Schema for `extra_info`

- `use_case_sensitive_table_names` (Boolean) Whether table names are treated as case-sensitive.


<a id="nestedatt--instances"></a>
### Nested Schema for `instances`

- `primary` (Attributes) Primary MySQL instance. (See [below for nested schema](#nestedatt--instances--primary).)
- `standby` (Attributes List) Standby MySQL instances. (See [below for nested schema](#nestedatt--instances--standby).)

<a id="nestedatt--instances--primary"></a>
### Nested Schema for `instances.primary`

- `availability_zone` (String) Availability zone of the instance.
- `instance_id` (String) MySQL instance ID.
- `subnet_id` (String) Subnet ID where the instance is placed.


<a id="nestedatt--instances--standby"></a>
### Nested Schema for `instances.standby`

- `availability_zone` (String) Availability zone of the standby instance.
- `instance_id` (String) Standby MySQL instance ID.
- `subnet_id` (String) Subnet ID where the standby instance is placed.



<a id="nestedatt--network_info"></a>
### Nested Schema for `network_info`

- `primary_subnet_info` (Attributes) Current primary subnet placement. (See [below for nested schema](#nestedatt--network_info--primary_subnet_info).)
- `security_group_ids` (Set of String) Security group IDs associated with the MySQL instance group.
- `standby_subnet_info` (Attributes List) Current standby subnet placement. (See [below for nested schema](#nestedatt--network_info--standby_subnet_info).)

<a id="nestedatt--network_info--primary_subnet_info"></a>
### Nested Schema for `network_info.primary_subnet_info`

- `availability_zone` (String) Availability zone of the subnet.
- `replicas` (Number) Number of MySQL instances in the subnet.
- `subnet_id` (String) Subnet ID.


<a id="nestedatt--network_info--standby_subnet_info"></a>
### Nested Schema for `network_info.standby_subnet_info`

- `availability_zone` (String) Availability zone of the standby subnet.
- `replicas` (Number) Number of standby MySQL instances in the subnet.
- `subnet_id` (String) Standby subnet ID.



<a id="nestedatt--parameter_group"></a>
### Nested Schema for `parameter_group`

- `apply_status` (String) Parameter group apply status.
- `engine_version` (String) MySQL engine version of the parameter group.
- `id` (String) Parameter group ID.
- `is_engine_version_mismatch` (Boolean) Whether the parameter group engine version differs from the instance group engine version.
- `type` (String) Parameter group type.


<a id="nestedatt--spec_content"></a>
### Nested Schema for `spec_content`

- `data_disk_size` (Number) Data disk size, in GB.
- `database_user_name` (String) Initial database user name.
- `engine_version` (String) MySQL engine version.
- `flavor_id` (String) MySQL flavor ID.
- `instance_group_type` (String) MySQL instance group type.
- `log_disk_size` (Number) Log disk size, in GB.
- `memory` (Number) Memory size.
- `node_size` (Number) Number of nodes in the instance group.
- `primary_port` (Number) Port for the primary MySQL instance.
- `standby_port` (Number) Port for standby MySQL instances.
- `vcpu` (Number) Number of vCPUs.

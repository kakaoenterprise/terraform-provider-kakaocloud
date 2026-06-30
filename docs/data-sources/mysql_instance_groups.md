---
page_title: "kakaocloud_mysql_instance_groups Data Source - kakaocloud"
subcategory: "MySQL"
description: |-
  The kakaocloud_mysql_instance_groups data source retrieves KakaoCloud MySQL instance groups.
---

# kakaocloud_mysql_instance_groups (Data Source)

The `kakaocloud_mysql_instance_groups` data source retrieves KakaoCloud MySQL instance groups in the current project.

Use this data source to discover existing instance groups and select IDs or status information for downstream
Terraform configuration.

## Example Usage

```hcl
# List MySQL instance groups
data "kakaocloud_mysql_instance_groups" "example" {
}

output "mysql_instance_groups" {
  value = [
    for group in data.kakaocloud_mysql_instance_groups.example.instance_groups : {
      id          = group.id
      name        = group.name
      status      = group.status
      is_multi_az = group.is_multi_az
    }
  ]
}
```

## Argument Reference

- `timeouts` (Optional, Attributes) Custom timeout settings. (See [below for nested schema](#nestedatt--timeouts).)

## Attribute Reference

The following attributes are exported:

- `instance_groups` (Attributes List) List of MySQL instance groups. (See [below for nested schema](#nestedatt--instance_groups).)

<a id="nestedatt--timeouts"></a>
### Nested Schema for `timeouts`

- `read` (Optional, String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).


<a id="nestedatt--instance_groups"></a>
### Nested Schema for `instance_groups`

- `backup_schedule` (Attributes) Backup schedule configuration. (See [below for nested schema](#nestedatt--instance_groups--backup_schedule).)
- `created_at` (String) Time when the instance group was created.
- `creator` (String) User who created the instance group.
- `description` (String) MySQL instance group description.
- `endpoint` (List of String) Connection endpoints for the MySQL instance group.
- `extra_info` (Attributes) Additional MySQL settings. (See [below for nested schema](#nestedatt--instance_groups--extra_info).)
- `id` (String) MySQL instance group ID.
- `instances` (Attributes) Primary and standby instance IDs. (See [below for nested schema](#nestedatt--instance_groups--instances).)
- `is_multi_az` (Boolean) Whether the instance group uses multiple availability zones.
- `license` (String) MySQL license information.
- `name` (String) MySQL instance group name.
- `network_info` (Attributes) Current network configuration. (See [below for nested schema](#nestedatt--instance_groups--network_info).)
- `parameter_group` (Attributes) Parameter group applied to the instance group. (See [below for nested schema](#nestedatt--instance_groups--parameter_group).)
- `project_id` (String) Project ID where the instance group exists.
- `source_backup_id` (String) Source backup ID when the instance group was restored from a backup.
- `spec_content` (Attributes) MySQL instance group specification. (See [below for nested schema](#nestedatt--instance_groups--spec_content).)
- `status` (String) MySQL instance group status.
- `updated_at` (String) Time when the instance group was last updated.

<a id="nestedatt--instance_groups--backup_schedule"></a>
### Nested Schema for `instance_groups.backup_schedule`

- `enabled` (Boolean) Whether automated backups are enabled.
- `expiry_duration` (Number) Backup retention period, in days.
- `id` (String) Backup schedule ID.
- `start_time` (String) Backup schedule start time. (Format: `HH:mm` (UTC))
- `type` (String) Backup schedule type. (Possible values: `DAY`)


<a id="nestedatt--instance_groups--extra_info"></a>
### Nested Schema for `instance_groups.extra_info`

- `use_case_sensitive_table_names` (Boolean) Whether table names are treated as case-sensitive.


<a id="nestedatt--instance_groups--instances"></a>
### Nested Schema for `instance_groups.instances`

- `primary` (Attributes) Primary MySQL instance. (See [below for nested schema](#nestedatt--instance_groups--instances--primary).)
- `standby` (Attributes List) Standby MySQL instances. (See [below for nested schema](#nestedatt--instance_groups--instances--standby).)

<a id="nestedatt--instance_groups--instances--primary"></a>
### Nested Schema for `instance_groups.instances.primary`

- `instance_id` (String) Primary MySQL instance ID.


<a id="nestedatt--instance_groups--instances--standby"></a>
### Nested Schema for `instance_groups.instances.standby`

- `instance_id` (String) Standby MySQL instance ID.



<a id="nestedatt--instance_groups--network_info"></a>
### Nested Schema for `instance_groups.network_info`

- `primary_subnet_info` (Attributes) Current primary subnet placement. (See [below for nested schema](#nestedatt--instance_groups--network_info--primary_subnet_info).)
- `security_group_ids` (Set of String) Security group IDs associated with the MySQL instance group.
- `standby_subnet_info` (Attributes List) Current standby subnet placement. (See [below for nested schema](#nestedatt--instance_groups--network_info--standby_subnet_info).)

<a id="nestedatt--instance_groups--network_info--primary_subnet_info"></a>
### Nested Schema for `instance_groups.network_info.primary_subnet_info`

- `availability_zone` (String) Availability zone of the subnet.
- `replicas` (Number) Number of MySQL instances in the subnet.
- `subnet_id` (String) Subnet ID.


<a id="nestedatt--instance_groups--network_info--standby_subnet_info"></a>
### Nested Schema for `instance_groups.network_info.standby_subnet_info`

- `availability_zone` (String) Availability zone of the standby subnet.
- `replicas` (Number) Number of standby MySQL instances in the subnet.
- `subnet_id` (String) Standby subnet ID.



<a id="nestedatt--instance_groups--parameter_group"></a>
### Nested Schema for `instance_groups.parameter_group`

- `apply_status` (String) Parameter group apply status.
- `engine_version` (String) MySQL engine version of the parameter group.
- `id` (String) Parameter group ID.
- `is_engine_version_mismatch` (Boolean) Whether the parameter group engine version differs from the instance group engine version.
- `type` (String) Parameter group type.


<a id="nestedatt--instance_groups--spec_content"></a>
### Nested Schema for `instance_groups.spec_content`

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

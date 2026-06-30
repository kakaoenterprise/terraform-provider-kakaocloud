---
page_title: "kakaocloud_mysql_instance_group Resource - kakaocloud"
subcategory: "MySQL"
description: |-
  The kakaocloud_mysql_instance_group resource allows you to create and manage KakaoCloud MySQL instance groups.
---

# kakaocloud_mysql_instance_group (Resource)

The `kakaocloud_mysql_instance_group` resource allows you to create and manage KakaoCloud MySQL instance groups.

You can configure engine version, flavor, storage sizes, network placement, backup schedule, parameter group, and
high-availability topology. The resource also exposes instance, endpoint, network, status, and parameter group metadata
returned by KakaoCloud MySQL.

-> **Note:** `spec_content.database_user_password` is used only during initial creation. It is required for new instance groups, write-only, sensitive, and later changes are not applied to the remote instance group. <br/> `source` is only used when creating a new instance group from a backup or another instance group. Do not configure `source` when importing an already-restored instance group. <br/> Restoring from an `INSTANCE_GROUP` source uses point-in-time recovery (PITR), not a current-state clone. `source.time` is required and must be within the source instance group's restorable time range. <br/> The restorable time range can be queried only when automated backups and binary logs (binlog) are ready. <br/> `INSTANCE_GROUP` restores support only single-instance configurations: set `desired_network_info.primary_subnet_info.replicas` to `1`, and leave `desired_network_info.standby_subnet_info` and `spec_content.standby_port` as `null`. 

## Example Usage

```hcl
# kakaocloud_mysql_instance_group Terraform Resource Example

data "kakaocloud_mysql_flavors" "example" {}
data "kakaocloud_mysql_engine_versions" "example" {}
data "kakaocloud_mysql_default_parameter_groups" "example" {}

# Basic Usage (kakaocloud_mysql_instance_group)
resource "kakaocloud_mysql_instance_group" "example" {
  name        = "example"
  description = "Example MySQL Instance Group"

  parameter_group = {
    type = "DEFAULT"
    id = one([for group in data.kakaocloud_mysql_default_parameter_groups.example.default_parameter_groups : group.id
      if group.engine_version == one([for version in data.kakaocloud_mysql_engine_versions.example.engine_versions : version.engine_version if version.engine_version == "8.4.8"])
    ])
  }

  spec_content = {
    engine_version         = one([for version in data.kakaocloud_mysql_engine_versions.example.engine_versions : version.engine_version if version.engine_version == "8.4.8"])
    flavor_id              = one([for flavor in data.kakaocloud_mysql_flavors.example.flavors : flavor.id if flavor.name == "m2a.large"])
    data_disk_size         = 100
    log_disk_size          = 100
    database_user_name     = local.username
    database_user_password = local.password
    primary_port           = 3306
    standby_port           = 3307
  }

  desired_network_info = {
    primary_subnet_info = {
      replicas  = 2
      subnet_id = kakaocloud_subnet.example_a.id
    }
    standby_subnet_info = [
      {
        replicas  = 1
        subnet_id = kakaocloud_subnet.example_b.id
      }
    ]
    security_group_ids = [kakaocloud_security_group.example.id]
  }

  extra_info = {
    use_case_sensitive_table_names = true
  }

  backup_schedule = {
    # type = "DAY"
    # start_time = "01:00"
    # expiry_duration = 3
    enabled = false
  }
}
```

## Argument Reference

- `backup_schedule` (Required, Attributes) Backup schedule configuration for the MySQL instance group. (See [below for nested schema](#nestedatt--backup_schedule).)
- `desired_network_info` (Required, Attributes) Desired network placement and security groups. (See [below for nested schema](#nestedatt--desired_network_info).)
- `name` (Required, String) MySQL instance group name.
- `parameter_group` (Required, Attributes) Parameter group to apply to the MySQL instance group. (See [below for nested schema](#nestedatt--parameter_group).)
- `spec_content` (Required, Attributes) MySQL instance group specification. (See [below for nested schema](#nestedatt--spec_content).)

- `description` (Optional, String) MySQL instance group description.
- `extra_info` (Optional, Attributes) Additional MySQL settings. (See [below for nested schema](#nestedatt--extra_info).)
- `source` (Optional, Attributes) Restore source used only when creating an instance group from a backup or another instance group. Do not configure this when importing an already-restored instance group. (See [below for nested schema](#nestedatt--source).)
- `timeouts` (Optional, Attributes) Custom timeout settings. (See [below for nested schema](#nestedatt--timeouts).)

## Attribute Reference

- `created_at` (String) Time when the instance group was created.
- `creator` (String) User who created the instance group.
- `endpoint` (List of String) Connection endpoints for the MySQL instance group.
- `id` (String) MySQL instance group ID.
- `instances` (Attributes) Primary and standby instance information. (See [below for nested schema](#nestedatt--instances).)
- `is_multi_az` (Boolean) Whether the instance group uses multiple availability zones.
- `license` (String) MySQL license information.
- `network_info` (Attributes) Current network configuration. (See [below for nested schema](#nestedatt--network_info).)
- `project_id` (String) Project ID where the instance group exists.
- `source_backup_id` (String) Source backup ID when the instance group was restored from a backup.
- `status` (String) MySQL instance group status.
- `updated_at` (String) Time when the instance group was last updated.

<a id="nestedatt--backup_schedule"></a>
### Nested Schema for `backup_schedule`

- `enabled` (Required, Boolean) Whether automated backups are enabled.

- `expiry_duration` (Optional, Number) Backup retention period, in days.
- `start_time` (Optional, String) Backup schedule start time. (Format: `HH:mm` (UTC))
- `type` (Optional, String) Backup schedule type. (Possible values: `DAY`)
- `id` (String) Backup schedule ID.


<a id="nestedatt--desired_network_info"></a>
### Nested Schema for `desired_network_info`

- `primary_subnet_info` (Required, Attributes) Primary subnet placement. (See [below for nested schema](#nestedatt--desired_network_info--primary_subnet_info).)
- `security_group_ids` (Required, Set of String) Security group IDs to associate with the MySQL instance group.

- `standby_subnet_info` (Optional, Attributes Set) Standby subnet placement for multi-AZ configurations. (See [below for nested schema](#nestedatt--desired_network_info--standby_subnet_info).)

<a id="nestedatt--desired_network_info--primary_subnet_info"></a>
### Nested Schema for `desired_network_info.primary_subnet_info`

- `replicas` (Required, Number) Number of MySQL instances to place in the subnet.
- `subnet_id` (Required, String) Subnet ID.


<a id="nestedatt--desired_network_info--standby_subnet_info"></a>
### Nested Schema for `desired_network_info.standby_subnet_info`

- `replicas` (Required, Number) Number of standby MySQL instances to place in the subnet.
- `subnet_id` (Required, String) Standby subnet ID.



<a id="nestedatt--parameter_group"></a>
### Nested Schema for `parameter_group`

- `id` (Required, String) Parameter group ID.
- `type` (Required, String) Parameter group type. (Possible values: `DEFAULT`, `CUSTOM`)
- `apply_status` (String) Parameter group apply status.
- `engine_version` (String) MySQL engine version of the parameter group.
- `is_engine_version_mismatch` (Boolean) Whether the parameter group engine version differs from the instance group engine version.


<a id="nestedatt--spec_content"></a>
### Nested Schema for `spec_content`

- `data_disk_size` (Required, Number) Data disk size, in GB.
- `database_user_name` (Required, String) Initial database user name.
- `engine_version` (Required, String) MySQL engine version.
- `flavor_id` (Required, String) MySQL flavor ID.
- `log_disk_size` (Required, Number) Log disk size, in GB.
- `primary_port` (Required, Number) Port for the primary MySQL instance.

- `database_user_password` (Optional, String) Initial database user password. This value is required for new instance group creation, write-only, sensitive, and used only during initial creation. Later changes are not applied to the remote instance group.
- `standby_port` (Optional, Number) Port for standby MySQL instances.
- `instance_group_type` (String) MySQL instance group type.
- `memory` (Number) Memory size.
- `node_size` (Number) Number of nodes in the instance group.
- `vcpu` (Number) Number of vCPUs.


<a id="nestedatt--extra_info"></a>
### Nested Schema for `extra_info`

- `use_case_sensitive_table_names` (Optional, Boolean) Whether table names are treated as case-sensitive.


<a id="nestedatt--source"></a>
### Nested Schema for `source`

- `id` (Required, String) Source backup ID or source instance group ID.
- `type` (Required, String) Restore source type. (Possible values: `BACKUP`, `INSTANCE_GROUP`)

- `time` (Optional, String) Point-in-time restore timestamp. Required when `type` is `INSTANCE_GROUP`; the value must be within the source instance group's restorable time range.


<a id="nestedatt--timeouts"></a>
### Nested Schema for `timeouts`

- `create` (Optional, String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).
- `delete` (Optional, String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Setting a timeout for a Delete operation is only applicable if changes are saved into state before the destroy operation occurs.
- `read` (Optional, String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Read operations occur during any refresh or planning operation when refresh is enabled.
- `update` (Optional, String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).


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
- `standby_subnet_info` (Attributes Set) Current standby subnet placement. (See [below for nested schema](#nestedatt--network_info--standby_subnet_info).)

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



## Import

Import is supported using the following syntax:

The [`terraform import` command](https://developer.hashicorp.com/terraform/cli/commands/import) can be used, for
example:

```shell
$ terraform import kakaocloud_mysql_instance_group.example <mysql_instance_group_id>
```

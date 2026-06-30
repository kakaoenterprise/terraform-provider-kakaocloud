---
page_title: "kakaocloud_mysql_backups Data Source - kakaocloud"
subcategory: "MySQL"
description: |-
  The kakaocloud_mysql_backups data source retrieves KakaoCloud MySQL backups.
---

# kakaocloud_mysql_backups (Data Source)

The `kakaocloud_mysql_backups` data source retrieves KakaoCloud MySQL backups.

You can optionally filter backups by MySQL instance group ID. Use this data source to list backup IDs, inspect backup
status, or pass backup information to restore workflows.

## Example Usage

```hcl
# List backups for a MySQL instance group
data "kakaocloud_mysql_backups" "example" {
  instance_group_id = "<your-mysql-instance-group-id>"
}

output "mysql_backup_ids" {
  value = [for backup in data.kakaocloud_mysql_backups.example.backups : backup.id]
}
```

## Argument Reference

- `instance_group_id` (Optional, String) MySQL instance group ID used to filter backups.
- `timeouts` (Optional, Attributes) Custom timeout settings. (See [below for nested schema](#nestedatt--timeouts).)

## Attribute Reference

The following attributes are exported:

- `backups` (Attributes List) List of MySQL backups. (See [below for nested schema](#nestedatt--backups).)

<a id="nestedatt--timeouts"></a>
### Nested Schema for `timeouts`

- `read` (Optional, String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).


<a id="nestedatt--backups"></a>
### Nested Schema for `backups`

- `created_at` (String) Time when the backup was created.
- `creator_name` (String) Name of the user who created the backup.
- `description` (String) Backup description.
- `disk_size` (Number) Disk size of the source MySQL instance group, in GB.
- `engine_version` (String) MySQL engine version of the source instance group.
- `expire_at` (String) Time when the backup expires.
- `expiry_duration` (Number) Backup retention period, in days.
- `extra_info` (Attributes) Additional MySQL backup settings. (See [below for nested schema](#nestedatt--backups--extra_info).)
- `id` (String) Backup ID.
- `instance_group_id` (String) ID of the source MySQL instance group.
- `instance_group_name` (String) Name of the source MySQL instance group.
- `name` (String) Backup name.
- `project_id` (String) Project ID where the backup exists.
- `size` (Number) Backup size, in bytes.
- `started_at` (String) Time when backup creation started.
- `status` (String) Backup status.
- `type` (String) Backup type.
- `updated_at` (String) Time when the backup was last updated.

<a id="nestedatt--backups--extra_info"></a>
### Nested Schema for `backups.extra_info`

- `use_case_sensitive_table_names` (Boolean) Whether table names are treated as case-sensitive.

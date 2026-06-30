---
page_title: "kakaocloud_mysql_backup Data Source - kakaocloud"
subcategory: "MySQL"
description: |-
  The kakaocloud_mysql_backup data source retrieves detailed information about a specific KakaoCloud MySQL backup.
---

# kakaocloud_mysql_backup (Data Source)

The `kakaocloud_mysql_backup` data source retrieves detailed information about a specific KakaoCloud MySQL backup by
its ID.

Use this data source when you need to reference an existing backup, inspect its status, or use backup metadata in other
Terraform configurations.

## Example Usage

```hcl
# Get a specific MySQL backup by ID
data "kakaocloud_mysql_backup" "example" {
  id = "<your-mysql-backup-id>"
}

output "mysql_backup_info" {
  value = {
    id                = data.kakaocloud_mysql_backup.example.id
    name              = data.kakaocloud_mysql_backup.example.name
    instance_group_id = data.kakaocloud_mysql_backup.example.instance_group_id
    status            = data.kakaocloud_mysql_backup.example.status
    type              = data.kakaocloud_mysql_backup.example.type
  }
}
```

## Argument Reference

- `id` (Required, String) Backup ID.

- `timeouts` (Optional, Attributes) Custom timeout settings. (See [below for nested schema](#nestedatt--timeouts).)

## Attribute Reference

The following attributes are exported:

- `created_at` (String) Time when the backup was created.
- `creator_name` (String) Name of the user who created the backup.
- `description` (String) Backup description.
- `disk_size` (Number) Disk size of the source MySQL instance group, in GB.
- `engine_version` (String) MySQL engine version of the source instance group.
- `expire_at` (String) Time when the backup expires.
- `expiry_duration` (Number) Backup retention period, in days.
- `extra_info` (Attributes) Additional MySQL backup settings. (See [below for nested schema](#nestedatt--extra_info).)
- `instance_group_id` (String) ID of the source MySQL instance group.
- `instance_group_name` (String) Name of the source MySQL instance group.
- `name` (String) Backup name.
- `project_id` (String) Project ID where the backup exists.
- `size` (Number) Backup size, in bytes.
- `started_at` (String) Time when backup creation started.
- `status` (String) Backup status.
- `type` (String) Backup type.
- `updated_at` (String) Time when the backup was last updated.

<a id="nestedatt--timeouts"></a>
### Nested Schema for `timeouts`

- `read` (Optional, String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).


<a id="nestedatt--extra_info"></a>
### Nested Schema for `extra_info`

- `use_case_sensitive_table_names` (Boolean) Whether table names are treated as case-sensitive.

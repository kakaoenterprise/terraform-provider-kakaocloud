---
page_title: "kakaocloud_mysql_backup Resource - kakaocloud"
subcategory: "MySQL"
description: |-
  The kakaocloud_mysql_backup resource allows you to create and manage manual backups for KakaoCloud MySQL instance groups.
---

# kakaocloud_mysql_backup (Resource)

The `kakaocloud_mysql_backup` resource allows you to create and manage manual backups for KakaoCloud MySQL instance
groups.

Use this resource when you need to create a point-in-time backup that can be tracked in Terraform state.

## Example Usage

```hcl
# kakaocloud_mysql_backup Terraform Resource Example

# Basic Usage (kakaocloud_mysql_backup)
resource "kakaocloud_mysql_backup" "example" {
  name              = "example"
  instance_group_id = kakaocloud_mysql_instance_group.example.id
}
```

## Argument Reference

- `instance_group_id` (Required, String) ID of the MySQL instance group to back up.
- `name` (Required, String) Backup name.

- `timeouts` (Optional, Attributes) Custom timeout settings. (See [below for nested schema](#nestedatt--timeouts).)

## Attribute Reference

- `created_at` (String) Time when the backup was created.
- `creator_name` (String) Name of the user who created the backup.
- `description` (String) Backup description.
- `disk_size` (Number) Disk size of the source MySQL instance group, in GB.
- `engine_version` (String) MySQL engine version of the source instance group.
- `expire_at` (String) Time when the backup expires.
- `expiry_duration` (Number) Backup retention period, in days.
- `extra_info` (Attributes) Additional MySQL backup settings. (See [below for nested schema](#nestedatt--extra_info).)
- `id` (String) Backup ID.
- `instance_group_name` (String) Name of the source MySQL instance group.
- `project_id` (String) Project ID where the backup exists.
- `size` (Number) Backup size, in bytes.
- `started_at` (String) Time when backup creation started.
- `status` (String) Backup status.
- `type` (String) Backup type.
- `updated_at` (String) Time when the backup was last updated.

<a id="nestedatt--timeouts"></a>
### Nested Schema for `timeouts`

- `create` (Optional, String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).
- `delete` (Optional, String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Setting a timeout for a Delete operation is only applicable if changes are saved into state before the destroy operation occurs.
- `read` (Optional, String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Read operations occur during any refresh or planning operation when refresh is enabled.
- `update` (Optional, String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).


<a id="nestedatt--extra_info"></a>
### Nested Schema for `extra_info`

- `use_case_sensitive_table_names` (Boolean) Whether table names are treated as case-sensitive.


## Import

Import is supported using the following syntax:

The [`terraform import` command](https://developer.hashicorp.com/terraform/cli/commands/import) can be used, for
example:

```shell
$ terraform import kakaocloud_mysql_backup.example <mysql_backup_id>
```

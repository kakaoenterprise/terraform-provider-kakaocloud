---
page_title: "kakaocloud_mysql_instance_export_logs Action - kakaocloud"
subcategory: "MySQL"
description: |-
  The kakaocloud_mysql_instance_export_logs action exports logs from a KakaoCloud MySQL instance to an object storage bucket.
---

# kakaocloud_mysql_instance_export_logs (Action)

The `kakaocloud_mysql_instance_export_logs` action exports logs from a KakaoCloud MySQL instance to an object storage
bucket.

Use this action to request log export for a specific MySQL instance and log type over an optional date range.

## Example Usage

```hcl
action "kakaocloud_mysql_instance_export_logs" "example" {
  config {
    instance_group_id      = "<your-mysql-instance-group-id>"
    instance_id            = "<your-mysql-instance-id>"
    bucket                 = "<your-object-storage-bucket>"
    path                   = "/mysql/logs"
    user_credential_id     = "<your-user-credential-id>"
    user_credential_secret = "<your-user-credential-secret>"

    log_infos = [
      {
        log_type = "SLOW_LOG"
        # start_date = "yyyy-mm-dd"
        # end_date   = "yyyy-mm-dd"
      }
    ]
  }
}
```

## Argument Reference

- `bucket` (Required, String) Object storage bucket name where exported logs are stored.
- `instance_group_id` (Required, String) MySQL instance group ID.
- `instance_id` (Required, String) MySQL instance ID to export logs from.
- `log_infos` (Required, Attributes List) Log export targets and optional date ranges. (See [below for nested schema](#nestedatt--log_infos).)
- `path` (Required, String) Object storage path where exported logs are written.
- `user_credential_id` (Required, String) User credential ID used to access the object storage bucket.
- `user_credential_secret` (Required, String) User credential secret used to access the object storage bucket.

<a id="nestedatt--log_infos"></a>
### Nested Schema for `log_infos`

- `log_type` (Required, String) Type of MySQL log to export. (Possible values: `GENERAL_LOG`, `SLOW_LOG`, `ERROR_LOG`, `BIN_LOG`)

- `end_date` (Optional, String) End date of the log export range. Use `yyyy-mm-dd` format. For non-`BIN_LOG` log types, the date must be within the last 7 days. Do not set this when `log_type` is `BIN_LOG`.
- `start_date` (Optional, String) Start date of the log export range. Use `yyyy-mm-dd` format. For non-`BIN_LOG` log types, the date must be within the last 7 days. Do not set this when `log_type` is `BIN_LOG`.

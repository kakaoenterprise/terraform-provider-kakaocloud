---
page_title: "kakaocloud_transit_gateway_attachment_reject Action - kakaocloud"
subcategory: "Transit Gateway"
description: |-
  Rejects a pending Transit Gateway attachment in KakaoCloud.
---

# kakaocloud_transit_gateway_attachment_reject (Action)

The `kakaocloud_transit_gateway_attachment_reject` action rejects a pending Transit Gateway attachment in KakaoCloud.

Use this action when you do not want to accept an attachment that requires manual approval, such as an attachment requested from a project that shares the Transit Gateway. The action waits until the rejected attachment is deleted.

## Example Usage

```hcl
action "kakaocloud_transit_gateway_attachment_reject" "example" {
  config {
    attachment_id = "<your-transit-gateway-attachment-id>"
  }
}
```

## Argument Reference

- `attachment_id` (Required, String) ID of the Transit Gateway attachment to reject.

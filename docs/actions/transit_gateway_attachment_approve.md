---
page_title: "kakaocloud_transit_gateway_attachment_approve Action - kakaocloud"
subcategory: "Transit Gateway"
description: |-
  Approves a pending Transit Gateway attachment in KakaoCloud.
---

# kakaocloud_transit_gateway_attachment_approve (Action)

The `kakaocloud_transit_gateway_attachment_approve` action approves a pending Transit Gateway attachment in KakaoCloud.

Use this action when an attachment requires manual approval, such as an attachment requested from a project that shares the Transit Gateway. The action waits for the approval process to complete and verifies that the attachment becomes active.

## Example Usage

```hcl
action "kakaocloud_transit_gateway_attachment_approve" "example" {
  config {
    attachment_id = "<your-transit-gateway-attachment-id>"
  }
}
```

## Argument Reference

- `attachment_id` (Required, String) ID of the Transit Gateway attachment to approve.

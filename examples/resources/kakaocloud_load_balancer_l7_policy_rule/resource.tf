# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# Example: Create an L7 policy rule for path-based routing
resource "kakaocloud_load_balancer_l7_policy_rule" "path_rule" {
  l7_policy_id = "2415269a-7142-455a-a7c8-9082dd146c57" # Replace with your L7 policy ID
  type         = "PATH"
  compare_type = "STARTS_WITH"
  value        = "/api/"
  is_inverted  = false

  # Optional: Configure timeouts
  timeouts {
    create = "5m"
    update = "5m"
    delete = "5m"
  }
}

# Example: Create an L7 policy rule for header-based routing
resource "kakaocloud_load_balancer_l7_policy_rule" "header_rule" {
  l7_policy_id = "2415269a-7142-455a-a7c8-9082dd146c57" # Replace with your L7 policy ID
  type         = "HEADER"
  compare_type = "EQUAL_TO"
  key          = "User-Agent"
  value        = "MobileApp/1.0"
  is_inverted  = false

  # Optional: Configure timeouts
  timeouts {
    create = "5m"
    update = "5m"
    delete = "5m"
  }
}

# Example: Create an L7 policy rule for host-based routing
resource "kakaocloud_load_balancer_l7_policy_rule" "host_rule" {
  l7_policy_id = "2415269a-7142-455a-a7c8-9082dd146c57" # Replace with your L7 policy ID
  type         = "HOST_NAME"
  compare_type = "EQUAL_TO"
  value        = "api.example.com"
  is_inverted  = false

  # Optional: Configure timeouts
  timeouts {
    create = "5m"
    update = "5m"
    delete = "5m"
  }
}

# Example: Create an L7 policy rule for cookie-based routing
resource "kakaocloud_load_balancer_l7_policy_rule" "cookie_rule" {
  l7_policy_id = "2415269a-7142-455a-a7c8-9082dd146c57" # Replace with your L7 policy ID
  type         = "COOKIE"
  compare_type = "CONTAINS"
  key          = "session_type"
  value        = "premium"
  is_inverted  = false

  # Optional: Configure timeouts
  timeouts {
    create = "5m"
    update = "5m"
    delete = "5m"
  }
}

# Example: Create an L7 policy rule for file type-based routing
resource "kakaocloud_load_balancer_l7_policy_rule" "file_type_rule" {
  l7_policy_id = "2415269a-7142-455a-a7c8-9082dd146c57" # Replace with your L7 policy ID
  type         = "FILE_TYPE"
  compare_type = "EQUAL_TO"
  value        = "jpg"
  is_inverted  = false

  # Optional: Configure timeouts
  timeouts {
    create = "5m"
    update = "5m"
    delete = "5m"
  }
}

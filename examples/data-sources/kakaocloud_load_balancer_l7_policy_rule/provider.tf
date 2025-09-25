# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

terraform {
  required_providers {
    kakaocloud = {
      source = "kakaocloud.com/dev/kakaocloud"
    }
  }
}

provider "kakaocloud" {
  # Configure provider credentials, e.g., by setting the X_AUTH_TOKEN environment variable
  # x_auth_token = "your-auth-token-here"
}

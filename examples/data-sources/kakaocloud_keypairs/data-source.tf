# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# List all keypairs
data "kakaocloud_keypairs" "all" {
  # No filters - get all keypairs
}

# List keypairs with comprehensive filters
data "kakaocloud_keypairs" "filtered" {
  filter = [
    {
      name  = "id"
      value = "your-keypair-id"  # Replace with your keypair ID
    },
    {
      name  = "name"
      value = "your-keypair-name"  # Replace with your keypair name
    },
    {
      name  = "type"
      value = "ssh"  # ssh, x509
    },
    {
      name  = "fingerprint"
      value = "your-keypair-fingerprint"  # Replace with your keypair fingerprint
    },
    {
      name  = "created_at"
      value = "2024-01-01T00:00:00Z"  # Replace with creation time (RFC3339 format)
    }
  ]
}

# Output all keypairs
output "all_keypairs" {
  description = "List of all keypairs"
  value = {
    count = length(data.kakaocloud_keypairs.all.keypairs)
    ids   = data.kakaocloud_keypairs.all.keypairs[*].id
    names = data.kakaocloud_keypairs.all.keypairs[*].name
  }
}

# Output filtered keypairs
output "filtered_keypairs" {
  description = "List of filtered keypairs"
  value = {
    count = length(data.kakaocloud_keypairs.filtered.keypairs)
    ids   = data.kakaocloud_keypairs.filtered.keypairs[*].id
    names = data.kakaocloud_keypairs.filtered.keypairs[*].name
  }
}

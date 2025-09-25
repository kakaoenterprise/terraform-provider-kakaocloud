#!/bin/bash

# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# Example import script for kakaocloud_load_balancer_l7_policy_rule
# Usage: ./import.sh <l7_policy_id>,<rule_id>

if [ $# -ne 1 ]; then
    echo "Usage: $0 <l7_policy_id>,<rule_id>"
    echo "Example: $0 2415269a-7142-455a-a7c8-9082dd146c57,8a7a1ca5-c687-4a9a-999b-a169ee248ade"
    exit 1
fi

IMPORT_ID=$1

# Import the L7 policy rule
terraform import kakaocloud_load_balancer_l7_policy_rule.path_rule "$IMPORT_ID"

echo "Import completed for L7 policy rule with ID: $IMPORT_ID"
echo "Note: You may need to update the resource configuration to match the imported resource."

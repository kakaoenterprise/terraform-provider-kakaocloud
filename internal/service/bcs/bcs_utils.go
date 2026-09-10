package // Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
bcs

import "terraform-provider-kakaocloud/internal/common"

var instanceFixedStatuses = []string{
	common.InstanceStatusActive,
	common.InstanceStatusError,
}

var instanceAvailableStatuses = []string{
	common.InstanceStatusActive,
	common.InstanceStatusStopped,
	common.InstanceStatusShelved,
}

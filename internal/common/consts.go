// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package common

import "time"

const (
	DefaultCreateTimeout  = 60 * time.Minute
	DefaultReadTimeout    = 30 * time.Minute
	DefaultUpdateTimeout  = 30 * time.Minute
	DefaultDeleteTimeout  = 30 * time.Minute
	DefaultPollingTimeout = 10 * time.Second
)

const (
	ActionC = "create"
	ActionR = "read"
	ActionU = "update"
	ActionD = "delete"
)

const (
	InstanceStatusActive  = "active"
	InstanceStatusStopped = "stopped"
	InstanceStatusShelved = "shelved_offloaded"
	InstanceStatusError   = "error"
)

const (
	InstanceTypeVM = "vm"
	InstanceTypeBM = "bm"
)

type AllowedInstanceStatus string

var AllInstanceStatuses = []AllowedInstanceStatus{
	InstanceStatusActive,
	InstanceStatusStopped,
	InstanceStatusShelved,
}

func IsInstanceValidStatus(s string) bool {
	for _, v := range AllInstanceStatuses {
		if string(v) == s {
			return true
		}
	}
	return false
}

const (
	VolumeStatusAvailable    = "available"
	VolumeStatusInUse        = "in_use"
	VolumeStatusError        = "error"
	VolumeStatusErrorRestore = "error_restoring"
)

const (
	VolumeSnapshotStatusAvailable = "available"
	VolumeSnapshotStatusError     = "error"
)

const (
	VpcProvisioningStatusActive = "ACTIVE"
	VpcProvisioningStatusError  = "ERROR"
)

const (
	NetworkInterfaceStatusAvailable = "available"
	NetworkInterfaceStatusInUse     = "in_use"
)

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package common

import "time"

const (
	DefaultCreateTimeout  = 60 * time.Minute
	DefaultReadTimeout    = 30 * time.Minute
	DefaultUpdateTimeout  = 30 * time.Minute
	DefaultDeleteTimeout  = 30 * time.Minute
	DefaultPollingTimeout = 1 * time.Minute

	LongCreateTimeout = 24 * time.Hour
	LongUpdateTimeout = 10 * time.Hour
	LongDeleteTimeout = 10 * time.Hour

	defaultActionProgressInterval = 10 * time.Second
)

const (
	ActionC      = "create"
	ActionR      = "read"
	ActionU      = "update"
	ActionD      = "delete"
	ActionInvoke = "invoke"
)

const (
	ServiceBCS  = "bcs"
	ServiceVPC  = "vpc"
	ServiceLB   = "lb"
	ServiceK8SE = "k8se"
)

const (
	ServiceMySQL = "mysql"
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
	VolumeStatusReserved     = "reserved"
	VolumeStatusDeleting     = "deleting"
)

const (
	VolumeSnapshotStatusAvailable = "available"
	VolumeSnapshotStatusError     = "error"
	VolumeSnapshotStatusDeleting  = "deleting"
)

const (
	VpcProvisioningStatusActive   = "ACTIVE"
	VpcProvisioningStatusError    = "ERROR"
	VpcProvisioningStatusDeleting = "PENDING_DELETE"
)

const (
	NetworkInterfaceStatusAvailable = "available"
	NetworkInterfaceStatusInUse     = "in_use"
)

const (
	PublicIpAvailable = "available"
	PublicIpInUse     = "in_use"
	PublicIpAttaching = "attaching"
)

const (
	SecurityGroupAvailable = "ACTIVE"
	SecurityGroupCreating  = "CREATING"
	SecurityGroupDeleting  = "DELETING"
	SecurityGroupError     = "ERROR"
)

const (
	ImageStatusActive = "active"
)

const (
	LoadBalancerProvisioningStatusActive   = "ACTIVE"
	LoadBalancerProvisioningStatusError    = "ERROR"
	LoadBalancerProvisioningStatusDeleted  = "DELETED"
	LoadBalancerProvisioningStatusDeleting = "PENDING_DELETE"
)

const (
	ClusterStatusProvisioned  = "Provisioned"
	ClusterStatusProvisioning = "Provisioning"
	ClusterStatusFailed       = "Failed"
	ClusterStatusDeleting     = "Deleting"
)

const (
	TgwStatusActive         = "ACTIVE"
	TgwStatusError          = "ERROR"
	TgwStatusInUse          = "IN_USE"
	TgwStatusInactive       = "INACTIVE"
	TgwStatusAvaliable      = "AVAILABLE"
	TgwStatusPendingApprove = "PENDING_APPROVE"
	TgwStatusDeleting       = "PENDING_DELETE"
)

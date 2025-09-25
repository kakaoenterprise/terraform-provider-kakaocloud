// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package utils

import (
	"net"

	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func Contains(list []string, val string) bool {
	for _, v := range list {
		if v == val {
			return true
		}
	}
	return false
}

func MergeDataSourceSchemaAttributes(
	base map[string]dschema.Attribute,
	override map[string]dschema.Attribute,
) map[string]dschema.Attribute {
	merged := make(map[string]dschema.Attribute, len(base)+len(override))

	for k, v := range base {
		merged[k] = v
	}
	for k, v := range override {
		merged[k] = v
	}

	return merged
}

func MergeResourceSchemaAttributes(
	base map[string]rschema.Attribute,
	override map[string]rschema.Attribute,
) map[string]rschema.Attribute {
	merged := make(map[string]rschema.Attribute, len(base)+len(override))

	for k, v := range base {
		merged[k] = v
	}
	for k, v := range override {
		merged[k] = v
	}

	return merged
}

func MergeAttributes[T any](base, override map[string]T) map[string]T {
	out := make(map[string]T, len(base)+len(override))
	for k, v := range base {
		out[k] = v
	}
	for k, v := range override {
		out[k] = v
	}
	return out
}

// CompareCIDRs compares two CIDR strings.
func CompareCIDRs(cidrA, cidrB string) int {
	ipA, netA, errA := net.ParseCIDR(cidrA)
	ipB, netB, errB := net.ParseCIDR(cidrB)

	// Fallback to string comparison if parsing fails
	if errA != nil || errB != nil {
		if cidrA < cidrB {
			return -1
		} else if cidrA > cidrB {
			return 1
		}
		return 0
	}

	cmp := CompareIPs(ipA, ipB)
	if cmp != 0 {
		return cmp
	}

	_, bitsA := netA.Mask.Size()
	_, bitsB := netB.Mask.Size()
	if bitsA < bitsB {
		return -1
	} else if bitsA > bitsB {
		return 1
	}
	return 0
}

func CompareIPs(a, b net.IP) int {
	a = a.To16()
	b = b.To16()
	for i := 0; i < len(a); i++ {
		if a[i] < b[i] {
			return -1
		}
		if a[i] > b[i] {
			return 1
		}
	}
	return 0
}

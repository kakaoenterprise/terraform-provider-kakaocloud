// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package common

const (
	AvailabilityZoneKr1a = "kr-central-1-a"
	AvailabilityZoneKr1b = "kr-central-1-b"
	AvailabilityZoneKr2a = "kr-central-2-a"
	AvailabilityZoneKr2b = "kr-central-2-b"
	AvailabilityZoneKr2c = "kr-central-2-c"
	AvailabilityZoneKr2d = "kr-central-2-d"
)

var zoneMatrix = map[string]map[string][]string{
	ServiceRealmStage: {
		RegionKR2: {AvailabilityZoneKr2a, AvailabilityZoneKr2b},
	},
	ServiceRealmPublic: {
		RegionKR2: {AvailabilityZoneKr2a, AvailabilityZoneKr2b, AvailabilityZoneKr2c, AvailabilityZoneKr2d},
	},
	ServiceRealmGov: {
		RegionKR1: {AvailabilityZoneKr1a, AvailabilityZoneKr1b},
	},
}

func AvailabilityZonesFor(serviceRealm, region string) ([]string, bool) {
	regions, ok := zoneMatrix[serviceRealm]
	if !ok {
		return nil, false
	}
	zones, ok := regions[region]
	if !ok {
		return nil, false
	}

	out := make([]string, len(zones))
	copy(out, zones)
	return out, true
}

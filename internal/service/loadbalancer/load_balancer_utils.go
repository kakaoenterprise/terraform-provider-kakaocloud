// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package loadbalancer

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/kakaoenterprise/kc-sdk-go/services/loadbalancer"
)

func ToScheme(v string) (*loadbalancer.Scheme, error) {
	scheme := loadbalancer.Scheme(strings.ToLower(v))
	for _, allowed := range loadbalancer.AllowedSchemeEnumValues {
		if scheme == allowed {
			return &scheme, nil
		}
	}
	return nil, fmt.Errorf("invalid scheme: %s (allowed: %v)", v, loadbalancer.AllowedSchemeEnumValues)
}
func ToProvisioningStatus(v string) (*loadbalancer.ProvisioningStatus, error) {
	ps := loadbalancer.ProvisioningStatus(strings.ToUpper(v))
	for _, allowed := range loadbalancer.AllowedProvisioningStatusEnumValues {
		if ps == allowed {
			return &ps, nil
		}
	}
	return nil, fmt.Errorf("invalid ProvisioningStatus: %s (allowed: %v)", v, loadbalancer.AllowedProvisioningStatusEnumValues)
}
func ToLoadBalancerOperatingStatus(v string) (*loadbalancer.LoadBalancerOperatingStatus, error) {
	os := loadbalancer.LoadBalancerOperatingStatus(strings.ToUpper(v))
	for _, allowed := range loadbalancer.AllowedLoadBalancerOperatingStatusEnumValues {
		if os == allowed {
			return &os, nil
		}
	}
	return nil, fmt.Errorf("invalid LoadBalancerOperatingStatus: %s (allowed: %v)", v, loadbalancer.AllowedLoadBalancerOperatingStatusEnumValues)
}
func ToLoadBalancerType(v string) (*loadbalancer.LoadBalancerType, error) {
	lbType := loadbalancer.LoadBalancerType(strings.ToUpper(v))
	for _, allowed := range loadbalancer.AllowedLoadBalancerTypeEnumValues {
		if lbType == allowed {
			return &lbType, nil
		}
	}
	return nil, fmt.Errorf("invalid LoadBalancerType: %s (allowed: %v)", v, loadbalancer.AllowedLoadBalancerTypeEnumValues)
}

func ToLoadBalancerProtocol(v string) (*loadbalancer.Protocol, error) {
	protocol := loadbalancer.Protocol(strings.ToUpper(v))
	for _, allowed := range loadbalancer.AllowedProtocolEnumValues {
		if protocol == allowed {
			return &protocol, nil
		}
	}
	return nil, fmt.Errorf("invalid LoadBalancerType: %s (allowed: %v)", v, loadbalancer.AllowedLoadBalancerTypeEnumValues)
}

func ToL7PolicyAction(v string) (*loadbalancer.L7PolicyAction, error) {
	action := loadbalancer.L7PolicyAction(strings.ToUpper(v))
	for _, allowed := range loadbalancer.AllowedL7PolicyActionEnumValues {
		if action == allowed {
			return &action, nil
		}
	}
	return nil, fmt.Errorf("invalid L7PolicyAction: %s (allowed: %v)", v, loadbalancer.AllowedL7PolicyActionEnumValues)
}

func ParseInt32(v string) (*int32, error) {
	i, err := strconv.ParseInt(v, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid integer: %s", v)
	}
	result := int32(i)
	return &result, nil
}

func ToTargetGroupProtocol(v string) (*loadbalancer.TargetGroupProtocol, error) {
	protocol := loadbalancer.TargetGroupProtocol(strings.ToUpper(v))
	for _, allowed := range loadbalancer.AllowedTargetGroupProtocolEnumValues {
		if protocol == allowed {
			return &protocol, nil
		}
	}
	return nil, fmt.Errorf("invalid TargetGroupProtocol: %s (allowed: %v)", v, loadbalancer.AllowedTargetGroupProtocolEnumValues)
}

func ToTargetGroupAlgorithm(v string) (*loadbalancer.TargetGroupAlgorithm, error) {
	algorithm := loadbalancer.TargetGroupAlgorithm(strings.ToUpper(v))
	for _, allowed := range loadbalancer.AllowedTargetGroupAlgorithmEnumValues {
		if algorithm == allowed {
			return &algorithm, nil
		}
	}
	return nil, fmt.Errorf("invalid TargetGroupAlgorithm: %s (allowed: %v)", v, loadbalancer.AllowedTargetGroupAlgorithmEnumValues)
}

func ToAvailabilityZone(v string) (*loadbalancer.AvailabilityZone, error) {
	az := loadbalancer.AvailabilityZone(v)
	for _, allowed := range loadbalancer.AllowedAvailabilityZoneEnumValues {
		if az == allowed {
			return &az, nil
		}
	}
	return nil, fmt.Errorf("invalid AvailabilityZone: %s (allowed: %v)", v, loadbalancer.AllowedAvailabilityZoneEnumValues)
}

func ToListenerProtocol(v string) (*loadbalancer.Protocol, error) {
	protocol := loadbalancer.Protocol(v)
	for _, allowed := range loadbalancer.AllowedProtocolEnumValues {
		if protocol == allowed {
			return &protocol, nil
		}
	}
	return nil, fmt.Errorf("invalid Protocol: %s (allowed: %v)", v, loadbalancer.AllowedProtocolEnumValues)
}

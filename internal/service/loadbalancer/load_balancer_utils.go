// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package loadbalancer

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"terraform-provider-kakaocloud/internal/common"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/kakaoenterprise/kc-sdk-go/services/loadbalancer"
	"golang.org/x/net/context"
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
	protocol := loadbalancer.Protocol(strings.ToUpper(v))
	for _, allowed := range loadbalancer.AllowedProtocolEnumValues {
		if protocol == allowed {
			return &protocol, nil
		}
	}
	return nil, fmt.Errorf("invalid Protocol: %s (allowed: %v)", v, loadbalancer.AllowedProtocolEnumValues)
}

func CheckLoadBalancerStatus(
	ctx context.Context,
	loadBalancerId string,
	allowError bool,
	r resource.Resource,
	kc *common.KakaoCloudClient,
	diags *diag.Diagnostics,
) bool {
	interval := 3 * time.Second
	lbResult, ok := common.PollUntilResult(
		ctx,
		r,
		interval,
		"load balancer",
		loadBalancerId,
		[]string{common.LoadBalancerProvisioningStatusActive, common.LoadBalancerProvisioningStatusError, common.LoadBalancerProvisioningStatusDeleting},
		diags,
		func(ctx context.Context) (*loadbalancer.BnsLoadBalancerV1ApiGetLoadBalancerModelLoadBalancerModel, *http.Response, error) {
			respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, kc, diags,
				func() (*loadbalancer.BnsLoadBalancerV1ApiGetLoadBalancerModelResponseLoadBalancerModel, *http.Response, error) {
					return kc.ApiClient.LoadBalancerAPI.
						GetLoadBalancer(ctx, loadBalancerId).
						XAuthToken(kc.XAuthToken).
						Execute()
				},
			)
			if err != nil {
				return nil, httpResp, err
			}
			return &respModel.LoadBalancer, httpResp, nil
		},
		func(lb *loadbalancer.BnsLoadBalancerV1ApiGetLoadBalancerModelLoadBalancerModel) string {
			return string(*lb.ProvisioningStatus.Get())
		},
	)
	if !ok {
		for _, d := range diags.Errors() {
			if strings.Contains(d.Detail(), "context deadline exceeded") {
				common.AddGeneralError(ctx, r, diags,
					fmt.Sprintf("Load balancer %s did not reach one of the following states: '%v'.", loadBalancerId, []string{common.LoadBalancerProvisioningStatusActive, common.LoadBalancerProvisioningStatusError}))
				return false
			}
		}
	}
	if lbResult != nil && string(*lbResult.ProvisioningStatus.Get()) == common.LoadBalancerProvisioningStatusDeleting ||
		!allowError && lbResult != nil && string(*lbResult.ProvisioningStatus.Get()) == common.LoadBalancerProvisioningStatusError {
		common.AddGeneralError(ctx, r, diags,
			fmt.Sprintf("Load balancer %s is in %v state", loadBalancerId, string(*lbResult.ProvisioningStatus.Get())))
		return false
	}
	if diags.HasError() {
		return false
	}
	return true
}

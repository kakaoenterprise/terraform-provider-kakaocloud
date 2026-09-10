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
	"github.com/kakaoenterprise/kc-sdk-go/v2/services/loadbalancer"
	"golang.org/x/net/context"
)

func ToScheme(v string) (*loadbalancer.BeyondLoadBalancerScheme, error) {
	scheme := loadbalancer.BeyondLoadBalancerScheme(strings.ToLower(v))
	for _, allowed := range loadbalancer.AllowedBeyondLoadBalancerSchemeEnumValues {
		if scheme == allowed {
			return &scheme, nil
		}
	}
	return nil, fmt.Errorf("invalid scheme: %s (allowed: %v)", v, loadbalancer.AllowedBeyondLoadBalancerSchemeEnumValues)
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
	return new(int32(i)), nil
}

func ToLoadBalancerPoolProtocol(v string) (*loadbalancer.LoadBalancerPoolProtocol, error) {
	protocol := loadbalancer.LoadBalancerPoolProtocol(strings.ToUpper(v))
	for _, allowed := range loadbalancer.AllowedLoadBalancerPoolProtocolEnumValues {
		if protocol == allowed {
			return &protocol, nil
		}
	}
	return nil, fmt.Errorf("invalid LoadBalancerPoolProtocol: %s (allowed: %v)", v, loadbalancer.AllowedLoadBalancerPoolProtocolEnumValues)
}

func ToLoadBalancerPoolAlgorithm(v string) (*loadbalancer.LoadBalancerPoolAlgorithm, error) {
	algorithm := loadbalancer.LoadBalancerPoolAlgorithm(strings.ToUpper(v))
	for _, allowed := range loadbalancer.AllowedLoadBalancerPoolAlgorithmEnumValues {
		if algorithm == allowed {
			return &algorithm, nil
		}
	}
	return nil, fmt.Errorf("invalid LoadBalancerPoolAlgorithm: %s (allowed: %v)", v, loadbalancer.AllowedLoadBalancerPoolAlgorithmEnumValues)
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

var loadBalancerAvailableAndErrorStatuses = []string{
	common.LoadBalancerProvisioningStatusActive,
	common.LoadBalancerProvisioningStatusError,
}

var loadBalancerPollingStatuses = append(
	append([]string(nil), loadBalancerAvailableAndErrorStatuses...),
	common.LoadBalancerProvisioningStatusDeleting,
)

type loadBalancerPollExpectedValues struct {
	name              *string
	description       *string
	accessLogsEnabled *bool
	accessLogBucket   *string
}

func pollLoadBalancerStatus(
	ctx context.Context,
	loadBalancerId string,
	allowError bool,
	r resource.Resource,
	kc *common.KakaoCloudClient,
	diags *diag.Diagnostics,
	expectedValues ...loadBalancerPollExpectedValues,
) (*loadbalancer.LoadBalancer, bool) {
	var expected *loadBalancerPollExpectedValues
	if len(expectedValues) > 0 {
		expected = &expectedValues[0]
	}

	interval := 3 * time.Second
	lbResult, ok := common.PollUntilResult(
		ctx,
		r,
		interval,
		"load balancer",
		loadBalancerId,
		loadBalancerPollingStatuses,
		diags,
		func(ctx context.Context) (*loadbalancer.LoadBalancer, *http.Response, error) {
			respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, kc, diags,
				func() (*loadbalancer.GetLoadBalancerResponse, *http.Response, error) {
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
		func(lb *loadbalancer.LoadBalancer) string {
			if lb == nil || lb.ProvisioningStatus.Get() == nil {
				return ""
			}

			status := string(*lb.ProvisioningStatus.Get())
			if status != common.LoadBalancerProvisioningStatusActive || expected == nil {
				return status
			}

			if expected.name != nil && (lb.Name.Get() == nil || *lb.Name.Get() != *expected.name) {
				return "UPDATING"
			}
			if expected.description != nil {
				actualDescription := ""
				if lb.Description.Get() != nil {
					actualDescription = *lb.Description.Get()
				}
				if actualDescription != *expected.description {
					return "UPDATING"
				}
			}
			if expected.accessLogsEnabled != nil {
				accessLogs := lb.AccessLogs.Get()
				if *expected.accessLogsEnabled {
					if accessLogs == nil || expected.accessLogBucket == nil || accessLogs.Bucket != *expected.accessLogBucket {
						return "UPDATING"
					}
				} else if accessLogs != nil {
					return "UPDATING"
				}
			}

			return status
		},
	)
	if !ok {
		for _, d := range diags.Errors() {
			if strings.Contains(d.Detail(), "context deadline exceeded") {
				common.AddGeneralError(ctx, r, diags,
					fmt.Sprintf("Load balancer %s did not reach one of the following states: '%v'.", loadBalancerId, loadBalancerAvailableAndErrorStatuses))
				return nil, false
			}
		}
	}
	if lbResult != nil && string(*lbResult.ProvisioningStatus.Get()) == common.LoadBalancerProvisioningStatusDeleting ||
		!allowError && lbResult != nil && string(*lbResult.ProvisioningStatus.Get()) == common.LoadBalancerProvisioningStatusError {
		common.AddGeneralError(ctx, r, diags,
			fmt.Sprintf("Load balancer %s is in %v state", loadBalancerId, string(*lbResult.ProvisioningStatus.Get())))
		return nil, false
	}
	if diags.HasError() {
		return nil, false
	}
	return lbResult, true
}

func findLbAndListenerByL7PolicyId(ctx context.Context, r resource.Resource, kc *common.KakaoCloudClient, l7PolicyId string, diags *diag.Diagnostics) (*string, *string, bool) {
	limit := int32(1000)
	offset := int32(0)

	for {
		listenersResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, kc, diags,
			func() (*loadbalancer.ListListenersResponse, *http.Response, error) {
				return kc.ApiClient.LoadBalancerListenerAPI.ListListeners(ctx).Limit(limit).Offset(offset).XAuthToken(kc.XAuthToken).Execute()
			},
		)

		if err != nil {
			common.AddApiActionError(ctx, r, httpResp, "ListListeners", err, diags)
			return nil, nil, false
		}

		for _, listener := range listenersResp.Listeners {
			protocol := listener.Protocol.Get()
			if protocol != nil &&
				(*protocol == loadbalancer.PROTOCOL_HTTP || *protocol == loadbalancer.PROTOCOL_TERMINATED_HTTPS) &&
				listener.L7Policies != nil {
				for _, policy := range listener.L7Policies {
					if policy.Id.Get() != nil && *policy.Id.Get() == l7PolicyId {
						return listener.LoadBalancerId.Get(), &listener.Id, true
					}
				}
			}
		}

		pagination := listenersResp.GetPagination()
		total := pagination.GetTotal()
		if total <= int64(limit+offset) {
			break
		}
		offset += limit
	}

	common.AddGeneralError(ctx, r, diags, fmt.Sprintf("L7 policy %s not found in any listener", l7PolicyId))
	return nil, nil, false
}

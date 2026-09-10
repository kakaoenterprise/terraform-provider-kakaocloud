// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package bcs

import (
	"net/http"
	"terraform-provider-kakaocloud/internal/common"
	"terraform-provider-kakaocloud/internal/utils"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kakaoenterprise/kc-sdk-go/v2/services/bcs"
	"golang.org/x/net/context"
)

const instanceSynchronizedStatus = "SYNCHRONIZED"

type instanceExpectedVolume struct {
	size                *int32
	deleteOnTermination *bool
}

type instanceExpectedState struct {
	name                *types.String
	description         *types.String
	flavorID            *string
	isHyperThreading    *bool
	volumes             map[string]instanceExpectedVolume
	networkInterfaceIDs map[string]struct{}
}

func (r *instanceResource) updateStatus(
	ctx context.Context,
	instanceId string,
	plan string,
	state string,
	resp *diag.Diagnostics,
) bool {
	if plan == state {
		return true
	}

	if !common.IsInstanceValidStatus(state) {
		result, ok := r.pollInstanceUntilStatus(
			ctx,
			instanceId,
			instanceAvailableStatuses,
			resp,
		)
		if !ok || resp.HasError() {
			return false
		}
		state = result.GetStatus()
	}

	type transitionFunc func(ctx context.Context, instanceId string, resp *diag.Diagnostics) bool

	transitions := map[string]map[string][]transitionFunc{
		common.InstanceStatusActive: {
			common.InstanceStatusStopped: {
				r.startInstance,
			},
			common.InstanceStatusShelved: {
				r.unshelveInstance,
			},
		},
		common.InstanceStatusStopped: {
			common.InstanceStatusActive: {
				r.stopInstance,
			},
			common.InstanceStatusShelved: {
				r.unshelveInstance,
				r.stopInstance,
			},
		},
		common.InstanceStatusShelved: {
			common.InstanceStatusActive: {
				r.shelveInstance,
			},
			common.InstanceStatusStopped: {
				r.startInstance,
				r.shelveInstance,
			},
		},
	}

	if next, ok := transitions[plan][state]; ok {
		for _, fn := range next {
			if !fn(ctx, instanceId, resp) {
				return false
			}
		}
	}
	return true
}

func (r *instanceResource) startInstance(
	ctx context.Context,
	instanceId string,
	resp *diag.Diagnostics,
) bool {
	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, resp,
		func() (struct{}, *http.Response, error) {
			_, httpResp, err := r.kc.ApiClient.InstanceAPI.StartInstance(ctx, instanceId).
				XAuthToken(r.kc.XAuthToken).
				Execute()
			return struct{}{}, httpResp, err
		},
	)

	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "StartInstance", err, resp)
		return false
	}

	result, ok := r.pollInstanceUntilStatus(
		ctx,
		instanceId,
		instanceFixedStatuses,
		resp,
	)
	if !ok || resp.HasError() {
		return false
	}

	common.CheckResourceAvailableStatus(ctx, r, result.Status.Get(), []string{common.InstanceStatusActive}, resp)
	if resp.HasError() {
		return false
	}

	return true
}

func (r *instanceResource) stopInstance(
	ctx context.Context,
	instanceId string,
	resp *diag.Diagnostics,
) bool {
	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, resp,
		func() (interface{}, *http.Response, error) {
			return r.kc.ApiClient.InstanceAPI.StopInstance(ctx, instanceId).
				XAuthToken(r.kc.XAuthToken).
				Execute()
		},
	)

	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "StopInstance", err, resp)
		return false
	}
	result, ok := r.pollInstanceUntilStatus(
		ctx,
		instanceId,
		[]string{common.InstanceStatusStopped, common.InstanceStatusError},
		resp,
	)
	if !ok || resp.HasError() {
		return false
	}

	common.CheckResourceAvailableStatus(ctx, r, result.Status.Get(), []string{common.InstanceStatusStopped}, resp)
	if resp.HasError() {
		return false
	}

	return true
}

func (r *instanceResource) shelveInstance(
	ctx context.Context,
	instanceId string,
	resp *diag.Diagnostics,
) bool {
	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, resp,
		func() (interface{}, *http.Response, error) {
			return r.kc.ApiClient.InstanceAPI.ShelveInstance(ctx, instanceId).
				XAuthToken(r.kc.XAuthToken).
				Execute()
		},
	)

	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "ShelveInstance", err, resp)
		return false
	}
	result, ok := r.pollInstanceUntilStatus(
		ctx,
		instanceId,
		[]string{common.InstanceStatusShelved, common.InstanceStatusError},
		resp,
	)
	if !ok || resp.HasError() {
		return false
	}

	common.CheckResourceAvailableStatus(ctx, r, result.Status.Get(), []string{common.InstanceStatusShelved}, resp)
	if resp.HasError() {
		return false
	}

	return true
}

func (r *instanceResource) unshelveInstance(
	ctx context.Context,
	instanceId string,
	resp *diag.Diagnostics,
) bool {
	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, resp,
		func() (interface{}, *http.Response, error) {
			return r.kc.ApiClient.InstanceAPI.UnshelveInstance(ctx, instanceId).
				XAuthToken(r.kc.XAuthToken).
				Execute()
		},
	)

	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "UnshelveInstance", err, resp)
		return false
	}
	result, ok := r.pollInstanceUntilStatus(
		ctx,
		instanceId,
		instanceFixedStatuses,
		resp,
	)
	if !ok || resp.HasError() {
		return false
	}

	common.CheckResourceAvailableStatus(ctx, r, result.Status.Get(), []string{common.InstanceStatusActive}, resp)
	if resp.HasError() {
		return false
	}

	return true
}

func (r *instanceResource) pollInstanceUntilStatus(
	ctx context.Context,
	instanceId string,
	targetStatuses []string,
	diag *diag.Diagnostics,
	expectedValues ...instanceExpectedState,
) (*bcs.Instance, bool) {
	var expected *instanceExpectedState
	pollingTargetStatuses := targetStatuses
	if len(expectedValues) > 0 {
		expected = &expectedValues[0]
		pollingTargetStatuses = []string{
			instanceSynchronizedStatus,
			common.InstanceStatusError,
		}
	}

	return common.PollUntilResult(
		ctx,
		r,
		3*time.Second,
		"instance",
		instanceId,
		pollingTargetStatuses,
		diag,
		func(ctx context.Context) (*bcs.Instance, *http.Response, error) {
			respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, diag,
				func() (*bcs.GetInstanceResponse, *http.Response, error) {
					return r.kc.ApiClient.InstanceAPI.
						GetInstance(ctx, instanceId).
						XAuthToken(r.kc.XAuthToken).
						Execute()
				},
			)
			if err != nil {
				return nil, httpResp, err
			}
			return &respModel.Instance, httpResp, nil
		},
		func(v *bcs.Instance) string {
			instanceType := v.InstanceType.Get()
			if instanceType == nil {
				return ""
			}
			if *instanceType != common.InstanceTypeBM && len(v.AttachedVolumes) == 0 {
				return ""
			}
			if len(v.Addresses) == 0 {
				return ""
			}
			if v.Status.Get() == nil {
				return ""
			}

			status := *v.Status.Get()
			if expected == nil || status == common.InstanceStatusError {
				return status
			}
			if instanceStatusIsTarget(status, targetStatuses) && instanceMatchesExpectedState(v, expected) {
				return instanceSynchronizedStatus
			}
			return status
		},
	)
}

func instanceStatusIsTarget(status string, targetStatuses []string) bool {
	for _, targetStatus := range targetStatuses {
		if status == targetStatus {
			return true
		}
	}
	return false
}

func instanceMatchesExpectedState(instance *bcs.Instance, expected *instanceExpectedState) bool {
	if expected.name != nil && !utils.ConvertNullableString(instance.Name).Equal(*expected.name) {
		return false
	}
	if expected.description != nil && !utils.ConvertNullableString(instance.Description).Equal(*expected.description) {
		return false
	}
	if expected.flavorID != nil {
		flavor := instance.Flavor.Get()
		if flavor == nil || flavor.Id.Get() == nil || *flavor.Id.Get() != *expected.flavorID {
			return false
		}
	}
	if expected.isHyperThreading != nil {
		if instance.IsHyperThreading.Get() == nil || *instance.IsHyperThreading.Get() != *expected.isHyperThreading {
			return false
		}
	}
	if expected.volumes != nil && !instanceVolumesMatchExpected(instance.AttachedVolumes, expected.volumes) {
		return false
	}
	if expected.networkInterfaceIDs != nil && !instanceNetworkInterfacesMatchExpected(instance.Addresses, expected.networkInterfaceIDs) {
		return false
	}
	return true
}

func instanceVolumesMatchExpected(actual []bcs.AttachedVolume, expected map[string]instanceExpectedVolume) bool {
	if len(actual) != len(expected) {
		return false
	}
	for _, volume := range actual {
		if volume.Id.Get() == nil {
			return false
		}
		expectedVolume, ok := expected[*volume.Id.Get()]
		if !ok {
			return false
		}
		if expectedVolume.size != nil && volume.Size != *expectedVolume.size {
			return false
		}
		if expectedVolume.deleteOnTermination != nil {
			if volume.IsDeleteOnTermination.Get() == nil || *volume.IsDeleteOnTermination.Get() != *expectedVolume.deleteOnTermination {
				return false
			}
		}
	}
	return true
}

func instanceNetworkInterfacesMatchExpected(actual []bcs.InstanceAddress, expected map[string]struct{}) bool {
	if len(actual) != len(expected) {
		return false
	}
	for _, address := range actual {
		if address.NetworkInterfaceId.Get() == nil {
			return false
		}
		if _, ok := expected[*address.NetworkInterfaceId.Get()]; !ok {
			return false
		}
	}
	return true
}

func buildInstanceExpectedVolumes(volumes []instanceVolumeModel) map[string]instanceExpectedVolume {
	expected := make(map[string]instanceExpectedVolume, len(volumes))
	for _, volume := range volumes {
		if volume.Id.IsNull() || volume.Id.IsUnknown() {
			continue
		}

		expectedVolume := instanceExpectedVolume{}
		if !volume.Size.IsNull() && !volume.Size.IsUnknown() {
			size := volume.Size.ValueInt32()
			expectedVolume.size = &size
		}
		if !volume.IsDeleteOnTermination.IsNull() && !volume.IsDeleteOnTermination.IsUnknown() {
			deleteOnTermination := volume.IsDeleteOnTermination.ValueBool()
			expectedVolume.deleteOnTermination = &deleteOnTermination
		}
		expected[volume.Id.ValueString()] = expectedVolume
	}
	return expected
}

func buildInstanceExpectedNetworkInterfaceIDs(subnets []instanceSubnetModel) map[string]struct{} {
	expected := make(map[string]struct{}, len(subnets))
	for _, subnet := range subnets {
		if !subnet.NetworkInterfaceId.IsNull() && !subnet.NetworkInterfaceId.IsUnknown() {
			expected[subnet.NetworkInterfaceId.ValueString()] = struct{}{}
		}
	}
	return expected
}

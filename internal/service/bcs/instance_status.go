// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bcs

import (
	"net/http"
	"terraform-provider-kakaocloud/internal/common"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/kakaoenterprise/kc-sdk-go/services/bcs"
	"golang.org/x/net/context"
)

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

	// wait to fixed status
	if !common.IsInstanceValidStatus(state) {
		result, ok := r.pollInstanceUntilStatus(
			ctx,
			instanceId,
			[]string{
				common.InstanceStatusActive,
				common.InstanceStatusStopped,
				common.InstanceStatusShelved,
			},
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
			_, httpResp, err := r.kc.ApiClient.InstanceRunAnActionAPI.StartInstance(ctx, instanceId).
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
		[]string{common.InstanceStatusActive, common.InstanceStatusError},
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
			return r.kc.ApiClient.InstanceRunAnActionAPI.StopInstance(ctx, instanceId).
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
			return r.kc.ApiClient.InstanceRunAnActionAPI.ShelveInstance(ctx, instanceId).
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
			return r.kc.ApiClient.InstanceRunAnActionAPI.UnshelveInstance(ctx, instanceId).
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
		[]string{common.InstanceStatusActive, common.InstanceStatusError},
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
) (*bcs.BcsInstanceV1ApiGetInstanceModelInstanceModel, bool) {
	return common.PollUntilResult(
		ctx,
		r,
		2*time.Second,
		targetStatuses,
		diag,
		func(ctx context.Context) (*bcs.BcsInstanceV1ApiGetInstanceModelInstanceModel, *http.Response, error) {
			respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, diag,
				func() (*bcs.ResponseInstanceModel, *http.Response, error) {
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
		func(v *bcs.BcsInstanceV1ApiGetInstanceModelInstanceModel) string {
			return *v.Status.Get()
		},
	)
}

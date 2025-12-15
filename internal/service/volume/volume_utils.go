package // Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
volume

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"terraform-provider-kakaocloud/internal/common"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/kakaoenterprise/kc-sdk-go/services/volume"
)

var (
	StatusesReadyGetOrUpdateForSize = []string{
		common.VolumeStatusAvailable,
		common.VolumeStatusInUse,
	}

	StatusesReadyForNameDesc = []string{
		common.VolumeStatusAvailable,
		common.VolumeStatusInUse,
		common.VolumeStatusReserved,
		common.VolumeStatusError,
		common.VolumeStatusErrorRestore,
	}

	StatusesReadyToDelete = []string{
		common.VolumeStatusAvailable,
		common.VolumeStatusError,
		common.VolumeStatusErrorRestore,
	}
)

func CheckVolumeStatus(
	ctx context.Context,
	kc *common.KakaoCloudClient,
	r interface{},
	volumeId string,
	targetStatuses []string,
	diags *diag.Diagnostics,
) (*volume.BcsVolumeV1ApiGetVolumeModelVolumeModel, bool) {
	result, ok := common.PollUntilResult(
		ctx,
		r,
		3*time.Second,
		"volume",
		volumeId,
		[]string{
			common.VolumeStatusAvailable,
			common.VolumeStatusInUse,
			common.VolumeStatusError,
			common.VolumeStatusErrorRestore,
			common.VolumeStatusDeleting,
			common.VolumeStatusReserved,
		},
		diags,
		func(ctx context.Context) (*volume.BcsVolumeV1ApiGetVolumeModelVolumeModel, *http.Response, error) {
			respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, kc, diags,
				func() (*volume.BcsVolumeV1ApiGetVolumeModelResponseVolumeModel, *http.Response, error) {
					return kc.ApiClient.VolumeAPI.
						GetVolume(ctx, volumeId).
						XAuthToken(kc.XAuthToken).
						Execute()
				},
			)
			if err != nil {
				return nil, httpResp, err
			}
			return &respModel.Volume, httpResp, nil
		},
		func(v *volume.BcsVolumeV1ApiGetVolumeModelVolumeModel) string {
			return *v.Status.Get()
		},
	)
	if !ok {
		for _, d := range diags.Errors() {
			if strings.Contains(d.Detail(), "context deadline exceeded") {
				common.AddGeneralError(ctx, r, diags,
					fmt.Sprintf("Volume '%s' did not reach one of the following states: '%v'.", volumeId, []string{common.VolumeStatusAvailable, common.VolumeStatusInUse}))
				return nil, false
			}
		}
		return nil, false
	}
	common.CheckResourceAvailableStatus(ctx, r, result.Status.Get(), targetStatuses, diags)
	if diags.HasError() {
		return nil, false
	}

	return result, ok
}

func CheckVolumeSnapshotStatus(
	ctx context.Context,
	kc *common.KakaoCloudClient,
	r interface{},
	volumeSnapshotId string,
	allowError bool,
	diags *diag.Diagnostics,
) (*volume.BcsVolumeV1ApiGetSnapshotModelVolumeSnapshotModel, bool) {
	snapshotResult, ok := common.PollUntilResult(
		ctx,
		r,
		3*time.Second,
		"volume snapshot",
		volumeSnapshotId,
		[]string{common.VolumeSnapshotStatusAvailable, common.VolumeSnapshotStatusError, common.VolumeSnapshotStatusDeleting},
		diags,
		func(ctx context.Context) (*volume.BcsVolumeV1ApiGetSnapshotModelVolumeSnapshotModel, *http.Response, error) {
			respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, kc, diags,
				func() (*volume.BcsVolumeV1ApiGetSnapshotModelResponseVolumeSnapshotModel, *http.Response, error) {
					return kc.ApiClient.VolumeSnapshotAPI.GetSnapshot(ctx, volumeSnapshotId).
						XAuthToken(kc.XAuthToken).Execute()
				},
			)
			if err != nil {
				return nil, httpResp, err
			}
			return &respModel.Snapshot, httpResp, nil
		},
		func(v *volume.BcsVolumeV1ApiGetSnapshotModelVolumeSnapshotModel) string {
			return *v.Status.Get()
		},
	)
	if !ok {
		for _, d := range diags.Errors() {
			if strings.Contains(d.Detail(), "context deadline exceeded") {
				common.AddGeneralError(ctx, r, diags,
					fmt.Sprintf("Volume snapshot '%s' did not reach the status '%s'.", volumeSnapshotId, common.VolumeSnapshotStatusAvailable))
				return nil, false
			}
		}
		return nil, false
	}

	if allowError {
		common.CheckResourceAvailableStatus(ctx, r, snapshotResult.Status.Get(), []string{common.VolumeSnapshotStatusAvailable, common.VolumeSnapshotStatusError}, diags)
	} else {
		common.CheckResourceAvailableStatus(ctx, r, snapshotResult.Status.Get(), []string{common.VolumeSnapshotStatusAvailable}, diags)
	}
	if diags.HasError() {
		return nil, false
	}

	return snapshotResult, ok
}

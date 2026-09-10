package // Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
vpc

import (
	"fmt"
	"net/http"
	"terraform-provider-kakaocloud/internal/common"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/kakaoenterprise/kc-sdk-go/v2/services/vpc"
	"golang.org/x/net/context"
)

var vpcAvailableAndErrorStatuses = []string{
	common.VpcProvisioningStatusActive,
	common.VpcProvisioningStatusError,
	common.VpcProvisioningStatusDeleting,
}

var subnetAvailableAndErrorStatuses = []string{
	common.VpcProvisioningStatusActive,
	common.VpcProvisioningStatusError,
}

func pollVpcStatus(
	ctx context.Context,
	r resource.Resource,
	kc *common.KakaoCloudClient,
	vpcId string,
	respDiags *diag.Diagnostics,
) bool {
	interval := 1 * time.Second
	result, ok := common.PollUntilResult(
		ctx,
		r,
		interval,
		"vpc",
		vpcId,
		vpcAvailableAndErrorStatuses,
		respDiags,
		func(ctx context.Context) (*vpc.Vpc, *http.Response, error) {
			respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, kc, respDiags,
				func() (*vpc.GetVpcResponse, *http.Response, error) {
					return kc.ApiClient.VPCAPI.
						GetVpc(ctx, vpcId).
						XAuthToken(kc.XAuthToken).
						Execute()
				},
			)
			if err != nil {
				return nil, httpResp, err
			}
			return &respModel.Vpc, httpResp, nil
		},
		func(v *vpc.Vpc) string {
			return string(*v.ProvisioningStatus.Get())
		},
	)
	if !ok {
		common.AddGeneralError(ctx, r, respDiags,
			fmt.Sprintf("VPC did not reach the status '%v'.", common.VpcProvisioningStatusActive),
		)
		return false
	}
	status := *result.ProvisioningStatus.Get()
	if status != common.VpcProvisioningStatusActive {
		common.AddGeneralError(ctx, r, respDiags,
			fmt.Sprintf("VPC status is '%v'.", status),
		)
	}
	if respDiags.HasError() {
		return false
	}
	return true
}

func pollSubnetStatus(
	ctx context.Context,
	r resource.Resource,
	kc *common.KakaoCloudClient,
	subnetId string,
	respDiags *diag.Diagnostics,
) bool {
	interval := 1 * time.Second
	result, ok := common.PollUntilResult(
		ctx,
		r,
		interval,
		"subnet",
		subnetId,
		subnetAvailableAndErrorStatuses,
		respDiags,
		func(ctx context.Context) (*vpc.Subnet, *http.Response, error) {
			respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, kc, respDiags,
				func() (*vpc.GetSubnetResponse, *http.Response, error) {
					return kc.ApiClient.VPCSubnetAPI.
						GetSubnet(ctx, subnetId).
						XAuthToken(kc.XAuthToken).
						Execute()
				},
			)
			if err != nil {
				return nil, httpResp, err
			}
			return &respModel.Subnet, httpResp, nil
		},
		func(v *vpc.Subnet) string {
			return string(*v.ProvisioningStatus.Get())
		},
	)
	if !ok {
		common.AddGeneralError(ctx, r, respDiags,
			fmt.Sprintf("Subnet did not reach the status '%v'.", common.VpcProvisioningStatusActive),
		)
		return false
	}
	status := *result.ProvisioningStatus.Get()
	if status == common.VpcProvisioningStatusError {
		common.AddGeneralError(ctx, r, respDiags,
			fmt.Sprintf("Subnet status is '%v'.", status),
		)
	}
	if respDiags.HasError() {
		return false
	}
	return true
}

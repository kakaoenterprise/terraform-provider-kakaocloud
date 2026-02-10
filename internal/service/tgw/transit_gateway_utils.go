// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package tgw

import (
	"fmt"
	"net/http"
	"strings"
	"terraform-provider-kakaocloud/internal/common"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/kakaoenterprise/kc-sdk-go/services/tgw"
	"golang.org/x/net/context"
)

func ToTgwProvisioningStatus(v string) (*tgw.ProvisioningStatus, error) {
	ps := tgw.ProvisioningStatus(strings.ToUpper(v))

	for _, allowed := range tgw.AllowedProvisioningStatusEnumValues {
		if ps == allowed {
			return &ps, nil
		}
	}
	return nil, fmt.Errorf("invalid provisioning status: %s (allowed: %v)", v, tgw.AllowedProvisioningStatusEnumValues)
}

func ToRegion(v string) (*tgw.Region, error) {
	region := tgw.Region(strings.ToLower(v))

	for _, allowed := range tgw.AllowedRegionEnumValues {
		if region == allowed {
			return &region, nil
		}
	}
	return nil, fmt.Errorf("invalid region: %s (allowed: %v)", v, tgw.AllowedRegionEnumValues)
}

func ToResourceType(v string) (*tgw.ResourceType, error) {
	rt := tgw.ResourceType(strings.ToLower(v))

	for _, allowed := range tgw.AllowedResourceTypeEnumValues {
		if rt == allowed {
			return &rt, nil
		}
	}
	return nil, fmt.Errorf("invalid resource type: %s (allowed: %v)", v, tgw.AllowedResourceTypeEnumValues)
}

func pollTgw(
	ctx context.Context,
	kc *common.KakaoCloudClient,
	resource interface{},
	tgwId string,
	targetStatuses []string,
	diags *diag.Diagnostics,
) (*tgw.BnsTgwV1ApiGetTransitGatewayModelTgwResponseModel, bool) {
	result, ok := common.PollUntilResult(
		ctx, resource, 3*time.Second, "transit gateway", tgwId, targetStatuses, diags,
		func(ctx context.Context) (*tgw.BnsTgwV1ApiGetTransitGatewayModelTgwResponseModel, *http.Response, error) {
			respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, kc, diags,
				func() (*tgw.GetTgwResponseModel, *http.Response, error) {
					return kc.ApiClient.TgwsAPI.
						GetTransitGateway(ctx, tgwId).
						XAuthToken(kc.XAuthToken).
						Execute()
				},
			)
			if err != nil {
				return nil, httpResp, err
			}
			return &respModel.Tgw, httpResp, nil
		},
		func(v *tgw.BnsTgwV1ApiGetTransitGatewayModelTgwResponseModel) string {
			return string(*v.ProvisioningStatus.Get())
		},
	)
	if !ok {
		for _, d := range diags.Errors() {
			if strings.Contains(d.Detail(), "context deadline exceeded") {
				common.AddGeneralError(ctx, resource, diags,
					fmt.Sprintf("Transit gateway %s did not reach one of the following states: '%v'.", tgwId, targetStatuses))
				return result, false
			}
		}
	}
	return result, ok
}

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
	"github.com/kakaoenterprise/kc-sdk-go/v2/services/tgw"
	"golang.org/x/net/context"
)

func ToTGWProvisioningStatus(v string) (*tgw.TGWProvisioningStatus, error) {
	return tgw.NewTGWProvisioningStatusFromValue(strings.ToUpper(v))
}

func ToTGWAttachmentProvisioningStatus(v string) (*tgw.TGWAttachmentProvisioningStatus, error) {
	return tgw.NewTGWAttachmentProvisioningStatusFromValue(strings.ToUpper(v))
}

func ToTGWRouteTableProvisioningStatus(v string) (*tgw.TGWRouteTableProvisioningStatus, error) {
	return tgw.NewTGWRouteTableProvisioningStatusFromValue(strings.ToUpper(v))
}

func ToTGWRouteProvisioningStatus(v string) (*tgw.TGWRouteProvisioningStatus, error) {
	return tgw.NewTGWRouteProvisioningStatusFromValue(strings.ToUpper(v))
}

func ToTGWAssociationProvisioningStatus(v string) (*tgw.TGWAssociationProvisioningStatus, error) {
	return tgw.NewTGWAssociationProvisioningStatusFromValue(strings.ToUpper(v))
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

var tgwAvailableStatuses = []string{
	common.TgwStatusActive,
	common.TgwStatusInUse,
	common.TgwStatusInactive,
	common.TgwStatusAvailable,
}

var tgwAvailableAndErrorStatuses = append(
	append([]string(nil), tgwAvailableStatuses...),
	common.TgwStatusError,
)

var tgwDeleteStatuses = append(
	append([]string(nil), tgwAvailableAndErrorStatuses...),
	common.TgwStatusDeleting,
)

var tgwAttachmentAvailableAndErrorStatuses = append(
	append([]string(nil), tgwAvailableAndErrorStatuses...),
	common.TgwStatusPendingApprove,
)

func pollTgw(
	ctx context.Context,
	kc *common.KakaoCloudClient,
	resource interface{},
	tgwId string,
	targetStatuses []string,
	diags *diag.Diagnostics,
) (*tgw.Tgw, bool) {
	result, ok := common.PollUntilResult(
		ctx, resource, 3*time.Second, "transit gateway", tgwId, targetStatuses, diags,
		func(ctx context.Context) (*tgw.Tgw, *http.Response, error) {
			respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, kc, diags,
				func() (*tgw.GetTransitGatewayResponse, *http.Response, error) {
					return kc.ApiClient.TransitGatewayAPI.
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
		func(v *tgw.Tgw) string {
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

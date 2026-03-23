// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package tgw

import (
	"context"
	"net/http"
	"terraform-provider-kakaocloud/internal/common"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kakaoenterprise/kc-sdk-go/services/tgw"
)

func (r *transitGatewayRouteTableResource) updateRouteTableRoutes(
	ctx context.Context,
	tgwId string,
	tgwRouteTableId string,
	plans *[]tgwRouteTableRequestRouteModel,
	states *[]tgwRouteTableRequestRouteModel,
	resp *diag.Diagnostics,
) bool {
	stateMap := make(map[string]tgwRouteTableRequestRouteModel)
	if states != nil {
		for _, s := range *states {
			if !s.Id.IsNull() && !s.Id.IsUnknown() {
				stateMap[s.Id.ValueString()] = s
			}
		}
	}
	planMap := make(map[string]tgwRouteTableRequestRouteModel)
	if plans != nil {
		for _, s := range *plans {
			if !s.Id.IsNull() && !s.Id.IsUnknown() {
				planMap[s.Id.ValueString()] = s
			}
		}
	}

	if states != nil {
		for _, stateRoute := range *states {
			if _, exists := planMap[stateRoute.Id.ValueString()]; !exists || !stateRoute.DestinationCidrBlock.Equal(planMap[stateRoute.Id.ValueString()].DestinationCidrBlock) {
				_, ok := pollTgw(ctx, r.kc, r, tgwId, []string{common.TgwStatusActive, common.TgwStatusError, common.TgwStatusInUse, common.TgwStatusInactive, common.TgwStatusAvaliable}, resp)
				if !ok || resp.HasError() {
					return false
				}
				ok = r.deleteRoute(ctx, tgwRouteTableId, stateRoute.Id.ValueString(), resp)
				if !ok {
					return false
				}
			}
		}
	}

	if plans != nil {
		for i := range *plans {
			planRoute := &(*plans)[i]

			_, ok := pollTgw(ctx, r.kc, r, tgwId, []string{common.TgwStatusActive, common.TgwStatusError, common.TgwStatusInUse, common.TgwStatusInactive, common.TgwStatusAvaliable}, resp)
			if !ok || resp.HasError() {
				return false
			}

			stateRoute, exists := stateMap[planRoute.Id.ValueString()]
			if !exists || !planRoute.DestinationCidrBlock.Equal(stateRoute.DestinationCidrBlock) {
				ok := r.addRoute(ctx, tgwRouteTableId, planRoute, resp)
				if !ok {
					return false
				}
			} else if !planRoute.TgwAttachmentId.Equal(stateRoute.TgwAttachmentId) {
				ok := r.updateRoute(ctx, tgwRouteTableId, planRoute, resp)
				if !ok {
					return false
				}
			}
		}
	}

	return true
}

func (r *transitGatewayRouteTableResource) addRoute(ctx context.Context, tgwRouteTableId string, route *tgwRouteTableRequestRouteModel, diag *diag.Diagnostics) bool {
	createReq := tgw.NewBnsTgwV1ApiCreateTgwRouteModelTgwRouteRequestModel(
		route.DestinationCidrBlock.ValueString(),
		route.TgwAttachmentId.ValueString(),
	)

	body := tgw.CreateTgwRouteTableRouteRequestModel{
		Route: *createReq,
	}

	respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, diag,
		func() (*tgw.BnsTgwV1ApiCreateTgwRouteModelCreateTgwRouteTableRouteResponseModel, *http.Response, error) {
			return r.kc.ApiClient.RouteTablesAPI.CreateTgwRoute(ctx, tgwRouteTableId).
				XAuthToken(r.kc.XAuthToken).
				CreateTgwRouteTableRouteRequestModel(body).
				Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "CreateTgwRoute", err, diag)
		return false
	}
	route.Id = types.StringValue(respModel.Route.Id)

	ok := r.pollRouteActive(ctx, tgwRouteTableId, route.Id.ValueString(), diag)
	if !ok {
		return false
	}
	return true
}

func (r *transitGatewayRouteTableResource) updateRoute(ctx context.Context, tgwRouteTableId string, route *tgwRouteTableRequestRouteModel, diag *diag.Diagnostics) bool {
	updateReq := tgw.NewBnsTgwV1ApiUpdateTgwRouteModelTgwRouteRequestModel(
		route.TgwAttachmentId.ValueString(),
	)

	body := tgw.UpdateTgwRouteTableRouteRequestModel{
		Route: *updateReq,
	}

	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, diag,
		func() (*tgw.BnsTgwV1ApiUpdateTgwRouteModelCreateTgwRouteTableRouteResponseModel, *http.Response, error) {
			return r.kc.ApiClient.RouteTablesAPI.UpdateTgwRouteTableRoute(
				ctx, tgwRouteTableId, route.Id.ValueString(),
			).XAuthToken(r.kc.XAuthToken).
				UpdateTgwRouteTableRouteRequestModel(body).
				Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "UpdateTgwRouteTableRoute", err, diag)
		return false
	}

	ok := r.pollRouteActive(ctx, tgwRouteTableId, route.Id.ValueString(), diag)
	if !ok {
		return false
	}
	return true
}

func (r *transitGatewayRouteTableResource) deleteRoute(ctx context.Context, tgwRouteTableId, routeId string, diag *diag.Diagnostics) bool {
	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, diag,
		func() (interface{}, *http.Response, error) {
			httpResp, err := r.kc.ApiClient.RouteTablesAPI.DeleteTgwRoute(ctx, tgwRouteTableId, routeId).
				XAuthToken(r.kc.XAuthToken).
				Execute()
			return nil, httpResp, err
		},
	)
	if !(httpResp != nil && httpResp.StatusCode == 404) && err != nil {
		common.AddApiActionError(ctx, r, httpResp, "DeleteTgwRoute", err, diag)
		return false
	}
	ok := r.pollRouteDelete(ctx, tgwRouteTableId, routeId, diag)
	if !ok {
		return false
	}
	return true
}

func (r *transitGatewayRouteTableResource) pollRouteActive(
	ctx context.Context,
	tgwRouteTableId string,
	routeId string,
	resp *diag.Diagnostics,
) bool {
	for {
		time.Sleep(5 * time.Second)
		respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, resp,
			func() (*tgw.GetTgwRouteTableRoutesResponseModel, *http.Response, error) {
				return r.kc.ApiClient.RouteTablesAPI.
					ListTgwRoutes(ctx, tgwRouteTableId).
					XAuthToken(r.kc.XAuthToken).
					Execute()
			},
		)
		if err != nil {
			common.AddApiActionError(ctx, r, httpResp, "ListTgwRoutes", err, resp)
			return false
		}

		for _, respRoute := range respModel.Routes {
			if respRoute.Id.IsSet() && *respRoute.Id.Get() == routeId {
				status := *respRoute.ProvisioningStatus.Get()
				if status == common.TgwStatusActive {
					return true
				}
			}
		}
	}
}

func (r *transitGatewayRouteTableResource) pollRouteDelete(
	ctx context.Context,
	tgwRouteTableId string,
	routeId string,
	resp *diag.Diagnostics,
) bool {
	for {
		time.Sleep(5 * time.Second)
		respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, resp,
			func() (*tgw.GetTgwRouteTableRoutesResponseModel, *http.Response, error) {
				return r.kc.ApiClient.RouteTablesAPI.
					ListTgwRoutes(ctx, tgwRouteTableId).
					XAuthToken(r.kc.XAuthToken).
					Execute()
			},
		)
		if err != nil {
			if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
				return true
			}
			common.AddApiActionError(ctx, r, httpResp, "ListTgwRoutes", err, resp)
			return false
		}

		deleted := true
		for _, respRoute := range respModel.Routes {
			if respRoute.Id.IsSet() && *respRoute.Id.Get() == routeId {
				deleted = false
				break
			}
		}

		if deleted {
			return true
		}
	}
}

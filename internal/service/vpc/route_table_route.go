// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package vpc

import (
	"net/http"
	"terraform-provider-kakaocloud/internal/common"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kakaoenterprise/kc-sdk-go/v2/services/vpc"
	"golang.org/x/net/context"
)

func (r *routeTableResource) updateRouteTableRoutes(
	ctx context.Context,
	routeTableId string,
	vpcId string,
	plans *[]routeTableRequestRouteModel,
	states *[]routeTableRequestRouteModel,
	resp *diag.Diagnostics,
) bool {
	for _, stateRoute := range *states {
		needToDelete := true
		routeId := stateRoute.Id.ValueString()
		for _, planRoute := range *plans {
			if planRoute.Destination.Equal(stateRoute.Destination) && planRoute.TargetId.Equal(stateRoute.TargetId) && planRoute.TargetType.Equal(stateRoute.TargetType) {
				needToDelete = false
				break
			}
		}
		if needToDelete {
			ok := r.deleteRoute(ctx, routeTableId, vpcId, routeId, resp)
			if !ok {
				return false
			}
		}
	}

	for i := range *plans {
		needToAdd := true
		planRoute := &(*plans)[i]
		for _, stateRoute := range *states {
			if planRoute.Destination.Equal(stateRoute.Destination) && planRoute.TargetId.Equal(stateRoute.TargetId) && planRoute.TargetType.Equal(stateRoute.TargetType) {
				needToAdd = false
				break
			}
		}
		if needToAdd {
			ok := r.addRoute(ctx, routeTableId, vpcId, planRoute, resp)
			if !ok {
				return false
			}
		}
	}

	return true
}

func (r *routeTableResource) addRoute(ctx context.Context, routeTableId, vpcId string, route *routeTableRequestRouteModel, diag *diag.Diagnostics) bool {
	req := vpc.AddRoute{
		RouteType:   vpc.RouteType(route.TargetType.ValueString()),
		TargetId:    route.TargetId.ValueString(),
		Destination: route.Destination.ValueString(),
	}

	body := *vpc.NewAddRouteRequest(req)

	ok := pollVpcStatus(ctx, r, r.kc, vpcId, diag)
	if !ok || diag.HasError() {
		return false
	}

	resp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, diag,
		func() (*vpc.AddRouteResponse, *http.Response, error) {
			return r.kc.ApiClient.VPCRouteTableAPI.AddRoute(ctx, routeTableId).
				XAuthToken(r.kc.XAuthToken).
				AddRouteRequest(body).
				Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "AddRoute", err, diag)
		return false
	}
	route.Id = types.StringValue(resp.VpcRoute.Id)
	return true
}

func (r *routeTableResource) updateRoute(ctx context.Context, routeTableId, vpcId string, route *routeTableRequestRouteModel, diag *diag.Diagnostics) bool {
	req := vpc.UpdateRoute{
		RouteType:   vpc.RouteType(route.TargetType.ValueString()),
		TargetId:    route.TargetId.ValueString(),
		Destination: route.Destination.ValueString(),
	}

	ok := pollVpcStatus(ctx, r, r.kc, vpcId, diag)
	if !ok || diag.HasError() {
		return false
	}

	body := *vpc.NewUpdateRouteRequest(req)
	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, diag,
		func() (*vpc.UpdateRouteResponse, *http.Response, error) {
			return r.kc.ApiClient.VPCRouteTableAPI.UpdateRoute(ctx, routeTableId, route.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				UpdateRouteRequest(body).
				Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "UpdateRoute", err, diag)
		return false
	}
	return true
}

func (r *routeTableResource) deleteRoute(ctx context.Context, routeTableId, vpcId, routeId string, diag *diag.Diagnostics) bool {
	ok := pollVpcStatus(ctx, r, r.kc, vpcId, diag)
	if !ok || diag.HasError() {
		return false
	}

	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, diag,
		func() (interface{}, *http.Response, error) {
			httpResp, err := r.kc.ApiClient.VPCRouteTableAPI.DeleteRoute(ctx, routeTableId, routeId).
				XAuthToken(r.kc.XAuthToken).
				Execute()
			return nil, httpResp, err
		},
	)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			return true
		}
		common.AddApiActionError(ctx, r, httpResp, "DeleteRoute", err, diag)
		return false
	}
	ok = r.pollRouteDelete(ctx, routeTableId, routeId, diag)
	if !ok {
		return false
	}

	return true
}

func (r *routeTableResource) pollRouteDelete(
	ctx context.Context,
	routeTableId string,
	routeId string,
	resp *diag.Diagnostics,
) bool {
	for {
		respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, resp,
			func() (*vpc.GetRouteTableResponse, *http.Response, error) {
				return r.kc.ApiClient.VPCRouteTableAPI.
					GetRouteTable(ctx, routeTableId).
					XAuthToken(r.kc.XAuthToken).
					Execute()
			},
		)
		if err != nil {
			if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
				time.Sleep(2 * time.Second)
				continue
			}
			common.AddApiActionError(ctx, r, httpResp, "GetRouteTable", err, resp)
			return false
		}

		deleted := true
		for _, respRoute := range respModel.VpcRouteTable.Routes {
			if respRoute.Id.Get() != nil && *respRoute.Id.Get() == routeId {
				deleted = false
				break
			}
		}

		if deleted {
			return true
		}
		time.Sleep(2 * time.Second)
	}
}

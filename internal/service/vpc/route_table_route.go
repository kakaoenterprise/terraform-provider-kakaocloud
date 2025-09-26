// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package vpc

import (
	"net/http"
	"strings"
	"terraform-provider-kakaocloud/internal/common"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kakaoenterprise/kc-sdk-go/services/vpc"
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
	req := vpc.CreateRouteModel{
		RouteType:   vpc.RouteTableRouteType(route.TargetType.ValueString()),
		TargetId:    route.TargetId.ValueString(),
		Destination: route.Destination.ValueString(),
	}

	body := *vpc.NewBodyAddRoute(req)

	ok := checkVpcStatus(ctx, r, r.kc, vpcId, diag)
	if !ok || diag.HasError() {
		return false
	}

	resp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, diag,
		func() (*vpc.BnsVpcV1ApiAddRouteModelResponseRouteModel, *http.Response, error) {
			return r.kc.ApiClient.VPCRouteTableRouteAPI.AddRoute(ctx, routeTableId).
				XAuthToken(r.kc.XAuthToken).
				BodyAddRoute(body).
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
	req := vpc.EditRouteModel{
		RouteType:   vpc.RouteTableRouteType(strings.ToLower(route.TargetType.ValueString())),
		TargetId:    route.TargetId.ValueString(),
		Destination: route.Destination.ValueString(),
	}

	ok := checkVpcStatus(ctx, r, r.kc, vpcId, diag)
	if !ok || diag.HasError() {
		return false
	}

	body := *vpc.NewBodyUpdateRoute(req)
	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, diag,
		func() (*vpc.BnsVpcV1ApiUpdateRouteModelResponseRouteModel, *http.Response, error) {
			return r.kc.ApiClient.VPCRouteTableRouteAPI.UpdateRoute(ctx, routeTableId, route.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				BodyUpdateRoute(body).
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
	ok := checkVpcStatus(ctx, r, r.kc, vpcId, diag)
	if !ok || diag.HasError() {
		return false
	}

	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, diag,
		func() (interface{}, *http.Response, error) {
			httpResp, err := r.kc.ApiClient.VPCRouteTableRouteAPI.DeleteRoute(ctx, routeTableId, routeId).
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
	return true
}

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

func (r *transitGatewayRouteTableResource) updateRouteTableAssociations(
	ctx context.Context,
	tgwRouteTableId string,
	plans *[]tgwRouteTableRequestAssociationModel,
	states *[]tgwRouteTableRequestAssociationModel,
	resp *diag.Diagnostics,
) bool {
	stateMap := make(map[string]tgwRouteTableRequestAssociationModel)
	for _, s := range *states {
		if !s.Id.IsNull() && !s.Id.IsUnknown() {
			stateMap[s.TgwAttachmentId.ValueString()] = s
		}
	}
	planMap := make(map[string]tgwRouteTableRequestAssociationModel)
	if plans != nil {
		for _, s := range *plans {
			if !s.Id.IsNull() && !s.Id.IsUnknown() {
				planMap[s.TgwAttachmentId.ValueString()] = s
			}
		}
	}

	for _, stateAssociation := range *states {
		if _, exists := planMap[stateAssociation.TgwAttachmentId.ValueString()]; !exists {
			ok := r.deleteAssociation(ctx, tgwRouteTableId, stateAssociation.Id.ValueString(), resp)
			if !ok {
				return false
			}
		}
	}

	if plans != nil {
		for _, planAssociation := range *plans {
			if _, exists := stateMap[planAssociation.TgwAttachmentId.ValueString()]; !exists {
				ok := r.addAssociation(ctx, tgwRouteTableId, &planAssociation, resp)
				if !ok {
					return false
				}
			}
		}
	}

	return true
}

func (r *transitGatewayRouteTableResource) addAssociation(ctx context.Context, tgwRouteTableId string, association *tgwRouteTableRequestAssociationModel, diag *diag.Diagnostics) bool {
	assocReq := tgw.TgwRouteTableAssociationRequestModel{
		TgwAttachmentId: association.TgwAttachmentId.ValueString(),
	}

	body := tgw.CreateTgwRouteTableAssociationRequestModel{
		Association: assocReq,
	}

	respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, diag,
		func() (*tgw.CreateTgwRouteTableAssociationResponseModel, *http.Response, error) {
			return r.kc.ApiClient.RouteTablesAPI.CreateTgwRouteTableAssociation(ctx, tgwRouteTableId).
				XAuthToken(r.kc.XAuthToken).
				CreateTgwRouteTableAssociationRequestModel(body).
				Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "CreateTgwRouteTableAssociation", err, diag)
		return false
	}

	association.Id = types.StringValue(respModel.Association.Id)

	ok := r.pollAssociationActive(ctx, tgwRouteTableId, association.Id.ValueString(), diag)
	if !ok {
		return false
	}
	return true
}

func (r *transitGatewayRouteTableResource) deleteAssociation(ctx context.Context, tgwRouteTableId, associationId string, diag *diag.Diagnostics) bool {
	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, diag,
		func() (interface{}, *http.Response, error) {
			httpResp, err := r.kc.ApiClient.RouteTablesAPI.DeleteTgwRouteTableAssociation(
				ctx,
				tgwRouteTableId,
				associationId,
			).XAuthToken(r.kc.XAuthToken).Execute()
			return nil, httpResp, err
		},
	)

	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "DeleteTgwRouteTableAssociation", err, diag)
		return false
	}
	ok := r.pollAssociationDelete(ctx, tgwRouteTableId, associationId, diag)
	if !ok {
		return false
	}

	return true
}

func (r *transitGatewayRouteTableResource) pollAssociationActive(
	ctx context.Context,
	tgwRouteTableId string,
	associationId string,
	resp *diag.Diagnostics,
) bool {
	for {
		time.Sleep(5 * time.Second)
		respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, resp,
			func() (*tgw.GetTgwRouteTableAssociationsResponseModel, *http.Response, error) {
				return r.kc.ApiClient.RouteTablesAPI.
					ListTgwRouteTableAssociations(ctx, tgwRouteTableId).
					XAuthToken(r.kc.XAuthToken).
					Execute()
			},
		)
		if err != nil {
			common.AddApiActionError(ctx, r, httpResp, "ListTgwRouteTableAssociations", err, resp)
			return false
		}

		for _, respRoute := range respModel.Associations {
			if respRoute.Id.IsSet() && *respRoute.Id.Get() == associationId {
				status := *respRoute.ProvisioningStatus.Get()
				if status == common.TgwStatusActive {
					return true
				}
			}
		}
	}
}

func (r *transitGatewayRouteTableResource) pollAssociationDelete(
	ctx context.Context,
	tgwRouteTableId string,
	associationId string,
	resp *diag.Diagnostics,
) bool {
	for {
		time.Sleep(5 * time.Second)
		respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, resp,
			func() (*tgw.GetTgwRouteTableAssociationsResponseModel, *http.Response, error) {
				return r.kc.ApiClient.RouteTablesAPI.
					ListTgwRouteTableAssociations(ctx, tgwRouteTableId).
					XAuthToken(r.kc.XAuthToken).
					Execute()
			},
		)
		if err != nil {
			if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
				return true
			}
			common.AddApiActionError(ctx, r, httpResp, "ListTgwRouteTableAssociations", err, resp)
			return false
		}

		deleted := true
		for _, respRoute := range respModel.Associations {
			if respRoute.Id.IsSet() && *respRoute.Id.Get() == associationId {
				deleted = false
				break
			}
		}

		if deleted {
			return true
		}
	}
}

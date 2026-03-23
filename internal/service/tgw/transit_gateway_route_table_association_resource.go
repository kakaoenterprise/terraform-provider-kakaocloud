// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package tgw

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"terraform-provider-kakaocloud/internal/common"
	"terraform-provider-kakaocloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kakaoenterprise/kc-sdk-go/services/tgw"
)

var (
	_ resource.ResourceWithConfigure   = &transitGatewayRouteTableAssociationResource{}
	_ resource.ResourceWithImportState = &transitGatewayRouteTableAssociationResource{}
)

func NewTransitGatewayRouteTableAssociationResource() resource.Resource {
	return &transitGatewayRouteTableAssociationResource{}
}

type transitGatewayRouteTableAssociationResource struct {
	kc *common.KakaoCloudClient
}

func (r *transitGatewayRouteTableAssociationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_transit_gateway_route_table_association"
}

func (r *transitGatewayRouteTableAssociationResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: utils.MergeResourceSchemaAttributes(
			transitGatewayRouteTableAssociationsResourceSchemaAttributes,
			map[string]schema.Attribute{
				"timeouts": timeouts.AttributesAll(ctx),
			},
		),
	}
}

func (r *transitGatewayRouteTableAssociationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan transitGatewayRouteTableAssociationResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, ok := pollRouteTableUntilAvailable(ctx, r.kc, r, plan.RouteTableId.ValueString(), &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	tgwId := result.TgwId.Get()

	mutex := common.LockForID(*tgwId)
	mutex.Lock()
	defer mutex.Unlock()

	timeout, diags := plan.Timeouts.Create(ctx, common.DefaultCreateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	_, ok = pollTgw(ctx, r.kc, r, *tgwId, []string{"ACTIVE", "ERROR"}, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	routeTableId := plan.RouteTableId.ValueString()
	attachmentId := plan.TgwAttachmentId.ValueString()

	currentRouteTableId, err := r.fetchAttachmentRouteTable(ctx, attachmentId, &resp.Diagnostics)
	if err != nil || resp.Diagnostics.HasError() {
		return
	}
	if currentRouteTableId != "" {
		if currentRouteTableId == routeTableId {
			asso, err := r.fetchRouteTableAssociation(ctx, routeTableId, attachmentId, &resp.Diagnostics)
			if err != nil || resp.Diagnostics.HasError() {
				return
			}

			if asso == nil {
				common.AddGeneralError(ctx, r, &resp.Diagnostics,
					fmt.Sprintf("route table association not found (404) for attachment %s in route table %s", attachmentId, routeTableId))
				return
			}

			plan.Id = asso.AssociationId
			status, ok := r.pollAssociationActive(ctx, routeTableId, plan.Id.ValueString(), &resp.Diagnostics)
			if !ok {
				return
			}
			plan.ProvisioningStatus = types.StringValue(status)

			diags = resp.State.Set(ctx, plan)
			resp.Diagnostics.Append(diags...)
			return
		} else {
			asso, err := r.fetchRouteTableAssociation(ctx, currentRouteTableId, attachmentId, &resp.Diagnostics)
			if err != nil || resp.Diagnostics.HasError() {
				return
			}

			ok := deleteRouteTableAssociation(ctx, r.kc, r, currentRouteTableId, asso.AssociationId.ValueString(), &resp.Diagnostics)
			if !ok || resp.Diagnostics.HasError() {
				return
			}

			_, ok = pollTgw(ctx, r.kc, r, *tgwId, []string{"ACTIVE", "ERROR"}, &resp.Diagnostics)
			if !ok || resp.Diagnostics.HasError() {
				return
			}
		}
	}

	createReq := tgw.TgwRouteTableAssociationRequestModel{
		TgwAttachmentId: attachmentId,
	}

	body := tgw.CreateTgwRouteTableAssociationRequestModel{
		Association: createReq,
	}

	respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*tgw.CreateTgwRouteTableAssociationResponseModel, *http.Response, error) {
			return r.kc.ApiClient.RouteTablesAPI.CreateTgwRouteTableAssociation(ctx, routeTableId).
				XAuthToken(r.kc.XAuthToken).
				CreateTgwRouteTableAssociationRequestModel(body).
				Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "CreateTgwRouteTableAssociation", err, &resp.Diagnostics)
		return
	}
	plan.Id = types.StringValue(respModel.Association.Id)

	status, ok := r.pollAssociationActive(ctx, routeTableId, plan.Id.ValueString(), &resp.Diagnostics)
	if !ok {
		return
	}
	plan.ProvisioningStatus = types.StringValue(status)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *transitGatewayRouteTableAssociationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state transitGatewayRouteTableAssociationResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	timeout, diags := state.Timeouts.Read(ctx, common.DefaultReadTimeout)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	asso, err := r.fetchRouteTableAssociation(ctx, state.RouteTableId.ValueString(), state.TgwAttachmentId.ValueString(), &resp.Diagnostics)
	if err != nil || resp.Diagnostics.HasError() {
		return
	}

	if asso == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Id = asso.AssociationId
	state.ProvisioningStatus = asso.ProvisioningStatus

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *transitGatewayRouteTableAssociationResource) Update(ctx context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	common.AddGeneralError(
		ctx, r, &resp.Diagnostics,
		"Updates are not supported for route_table_association.",
	)
}

func (r *transitGatewayRouteTableAssociationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state transitGatewayRouteTableAssociationResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, ok := pollRouteTableUntilAvailable(ctx, r.kc, r, state.RouteTableId.ValueString(), &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	tgwId := result.TgwId.Get()

	mutex := common.LockForID(*tgwId)
	mutex.Lock()
	defer mutex.Unlock()

	timeout, diags := state.Timeouts.Delete(ctx, common.DefaultDeleteTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	_, ok = pollTgw(ctx, r.kc, r, *tgwId, []string{"ACTIVE", "ERROR"}, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	ok = deleteRouteTableAssociation(ctx, r.kc, r, state.RouteTableId.ValueString(), state.Id.ValueString(), &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}
}

func (r *transitGatewayRouteTableAssociationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*common.KakaoCloudClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *common.KakaoCloudClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.kc = client
}

func (r *transitGatewayRouteTableAssociationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {

	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		common.AddImportFormatError(ctx, r, &resp.Diagnostics,
			"Expected import ID in the format: route_table_id/tgw_attachment_id")
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("route_table_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("tgw_attachment_id"), parts[1])...)
}

func (r *transitGatewayRouteTableAssociationResource) fetchAttachmentRouteTable(ctx context.Context, tgwAttachmentId string, resp *diag.Diagnostics) (string, error) {

	respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, resp,
		func() (*tgw.GetTgwAttachmentResponseModel, *http.Response, error) {
			return r.kc.ApiClient.AttachmentsAPI.GetTgwAttachment(ctx, tgwAttachmentId).
				XAuthToken(r.kc.XAuthToken).
				Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "GetTgwAttachment", err, resp)
		return "", err
	}

	var routeTableId = ""

	if respModel.Attachment.RouteTable.IsSet() {
		routeTableId = respModel.Attachment.RouteTable.Get().GetId()
	}

	return routeTableId, nil
}

func (r *transitGatewayRouteTableAssociationResource) fetchRouteTableAssociation(ctx context.Context, routeTableId, tgwAttachmentId string, resp *diag.Diagnostics) (*routeTableAssociationFetchResultModel, error) {

	var fetchResult *routeTableAssociationFetchResultModel

	listResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, resp,
		func() (*tgw.GetTgwRouteTableAssociationsResponseModel, *http.Response, error) {
			return r.kc.ApiClient.RouteTablesAPI.
				ListTgwRouteTableAssociations(ctx, routeTableId).
				Limit(1000).
				XAuthToken(r.kc.XAuthToken).
				Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "ListTgwRouteTableAssociations", err, resp)
		return nil, err
	}

	for _, asso := range listResp.Associations {
		assoAttachmentId := asso.ResourceAttachmentId.Get()
		if *assoAttachmentId == tgwAttachmentId {
			fetchResult = &routeTableAssociationFetchResultModel{
				AssociationId:      utils.ConvertNullableString(asso.Id),
				TgwAttachmentId:    types.StringValue(tgwAttachmentId),
				ProvisioningStatus: utils.ConvertNullableString(asso.ProvisioningStatus),
			}
			break
		}
	}

	return fetchResult, nil
}

func (r *transitGatewayRouteTableAssociationResource) pollAssociationActive(
	ctx context.Context,
	tgwRouteTableId string,
	associationId string,
	resp *diag.Diagnostics,
) (string, bool) {
	for {
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
			return "", false
		}

		for _, respRoute := range respModel.Associations {
			if respRoute.Id.IsSet() && *respRoute.Id.Get() == associationId {
				status := *respRoute.ProvisioningStatus.Get()
				if status == common.TgwStatusActive {
					return string(status), true
				}
			}
		}
	}
}

func deleteRouteTableAssociation(
	ctx context.Context,
	kc *common.KakaoCloudClient,
	resource interface{},
	routeTableId, associationId string,
	resp *diag.Diagnostics,
) bool {
	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, kc, resp,
		func() (interface{}, *http.Response, error) {
			httpResp, err := kc.ApiClient.RouteTablesAPI.DeleteTgwRouteTableAssociation(ctx, routeTableId, associationId).
				XAuthToken(kc.XAuthToken).
				Execute()
			return nil, httpResp, err
		},
	)
	if !(httpResp != nil && httpResp.StatusCode == 404) && err != nil {
		common.AddApiActionError(ctx, resource, httpResp, "DeleteTgwRouteTableAssociation", err, resp)
		return false
	}

	ok := pollAssociationDelete(ctx, kc, resource, routeTableId, associationId, resp)
	if !ok {
		return false
	}

	return true
}

func pollAssociationDelete(
	ctx context.Context,
	kc *common.KakaoCloudClient,
	resource interface{},
	tgwRouteTableId string,
	associationId string,
	resp *diag.Diagnostics,
) bool {
	for {
		respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, kc, resp,
			func() (*tgw.GetTgwRouteTableAssociationsResponseModel, *http.Response, error) {
				return kc.ApiClient.RouteTablesAPI.
					ListTgwRouteTableAssociations(ctx, tgwRouteTableId).
					XAuthToken(kc.XAuthToken).
					Execute()
			},
		)
		if err != nil {
			if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
				return true
			}
			common.AddApiActionError(ctx, resource, httpResp, "ListTgwRouteTableAssociations", err, resp)
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

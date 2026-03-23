// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package tgw

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-kakaocloud/internal/common"
	"terraform-provider-kakaocloud/internal/utils"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	_ "github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/kakaoenterprise/kc-sdk-go/services/tgw"
)

var (
	_ resource.ResourceWithConfigure   = &transitGatewayDefaultRouteTableResource{}
	_ resource.ResourceWithImportState = &transitGatewayDefaultRouteTableResource{}
)

func NewTransitGatewayDefaultRouteTableResource() resource.Resource {
	return &transitGatewayDefaultRouteTableResource{}
}

type transitGatewayDefaultRouteTableResource struct {
	kc *common.KakaoCloudClient
}

func (r *transitGatewayDefaultRouteTableResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_transit_gateway_default_route_table"
}

type transitGatewayDefaultRouteTableModel struct {
	TgwId        types.String   `tfsdk:"tgw_id"`
	RouteTableId types.String   `tfsdk:"route_table_id"`
	Timeouts     timeouts.Value `tfsdk:"timeouts"`
}

func (r *transitGatewayDefaultRouteTableResource) Schema(
	ctx context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"tgw_id": schema.StringAttribute{
				Required:   true,
				Validators: common.UuidValidator(),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"route_table_id": schema.StringAttribute{
				Required:   true,
				Validators: common.UuidValidator(),
			},
			"timeouts": timeouts.AttributesAll(ctx),
		},
	}
}

func (r *transitGatewayDefaultRouteTableResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan transitGatewayDefaultRouteTableModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	timeout, diags := plan.Timeouts.Create(ctx, common.DefaultCreateTimeout)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*tgw.GetTgwResponseModel, *http.Response, error) {
			return r.kc.ApiClient.TgwsAPI.GetTransitGateway(ctx, plan.TgwId.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				Execute()
		},
	)

	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "GetTransitGateway", err, &resp.Diagnostics)
		return
	}

	respOption := respModel.Tgw.Options.Get()

	if ok := r.updateDefaultRouteTable(ctx, plan.TgwId.ValueString(), plan.RouteTableId.ValueStringPointer(), *respOption, &resp.Diagnostics); !ok {
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *transitGatewayDefaultRouteTableResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state transitGatewayDefaultRouteTableModel
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

	respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*tgw.GetTgwResponseModel, *http.Response, error) {
			return r.kc.ApiClient.TgwsAPI.GetTransitGateway(ctx, state.TgwId.ValueString()).
				XAuthToken(r.kc.XAuthToken).Execute()
		},
	)

	if httpResp != nil && httpResp.StatusCode == 404 {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "GetTransitGateway", err, &resp.Diagnostics)
		return
	}

	state.RouteTableId = utils.ConvertNullableString(respModel.Tgw.Options.Get().AssociationDefaultRouteTableId)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *transitGatewayDefaultRouteTableResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan transitGatewayDefaultRouteTableModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	timeout, diags := plan.Timeouts.Update(ctx, common.DefaultUpdateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*tgw.GetTgwResponseModel, *http.Response, error) {
			return r.kc.ApiClient.TgwsAPI.GetTransitGateway(ctx, plan.TgwId.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				Execute()
		},
	)

	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "GetTransitGateway", err, &resp.Diagnostics)
		return
	}

	respOption := respModel.Tgw.Options.Get()

	if ok := r.updateDefaultRouteTable(ctx, plan.TgwId.ValueString(), plan.RouteTableId.ValueStringPointer(), *respOption, &resp.Diagnostics); !ok {
		return
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *transitGatewayDefaultRouteTableResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state transitGatewayDefaultRouteTableModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	timeout, diags := state.Timeouts.Delete(ctx, common.DefaultDeleteTimeout)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*tgw.GetTgwResponseModel, *http.Response, error) {
			return r.kc.ApiClient.TgwsAPI.GetTransitGateway(ctx, state.TgwId.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				Execute()
		},
	)

	if httpResp != nil && httpResp.StatusCode == 404 {
		return
	}
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "GetTransitGateway", err, &resp.Diagnostics)
		return
	}

	respOption := respModel.Tgw.Options.Get()

	if ok := r.updateDefaultRouteTable(ctx, state.TgwId.ValueString(), nil, *respOption, &resp.Diagnostics); !ok {
		return
	}

}

func (r *transitGatewayDefaultRouteTableResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*common.KakaoCloudClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *common.KakaoCloudClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.kc = client
}

func (r *transitGatewayDefaultRouteTableResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("tgw_id"), req, resp)
}

func (r *transitGatewayDefaultRouteTableResource) updateDefaultRouteTable(
	ctx context.Context,
	tgwID string,
	routeTableID *string,
	respOption tgw.OptionResponseModel,
	diags *diag.Diagnostics,
) bool {
	mutex := common.LockForID(tgwID)
	mutex.Lock()
	defer mutex.Unlock()

	_, ok := pollTgw(ctx, r.kc, r, tgwID, []string{common.TgwStatusActive, common.TgwStatusError, common.TgwStatusInUse, common.TgwStatusInactive, common.TgwStatusAvaliable}, diags)
	if !ok || diags.HasError() {
		return false
	}

	isAutoAcceptSharedAttachments := respOption.IsAutoAcceptSharedAttachments.Get()

	updateReq := tgw.NewBnsTgwV1ApiUpdateTransitGatewayModelTgwRequestModel()
	optionsReq := tgw.BnsTgwV1ApiUpdateTransitGatewayModelTgwOptionRequestModel{
		IsAutoAcceptSharedAttachments: *isAutoAcceptSharedAttachments,
	}

	if routeTableID == nil {
		optionsReq.SetIsDefaultRouteTableAssociation(false)
		optionsReq.SetAssociationDefaultRouteTableIdNil()
	} else {
		optionsReq.SetIsDefaultRouteTableAssociation(true)
		optionsReq.SetAssociationDefaultRouteTableId(*routeTableID)
	}
	updateReq.SetOptions(optionsReq)

	body := tgw.UpdateTgwRequestModel{
		Tgw: *updateReq,
	}

	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, diags,
		func() (*tgw.BnsTgwV1ApiUpdateTransitGatewayModelCreateTgwResponseModel, *http.Response, error) {
			return r.kc.ApiClient.TgwsAPI.UpdateTransitGateway(ctx, tgwID).
				XAuthToken(r.kc.XAuthToken).
				UpdateTgwRequestModel(body).
				Execute()
		},
	)

	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "UpdateTransitGateway", err, diags)
		return false
	}

	time.Sleep(3 * time.Second)

	result, ok := pollTgw(ctx, r.kc, r, tgwID, []string{common.TgwStatusActive, common.TgwStatusError}, diags)
	if !ok || diags.HasError() {
		return ok
	}
	common.CheckResourceAvailableStatus(ctx, r, (*string)(result.ProvisioningStatus.Get()), []string{common.TgwStatusActive}, diags)
	if diags.HasError() {
		return false
	}

	return true
}

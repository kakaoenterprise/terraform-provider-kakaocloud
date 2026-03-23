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
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/kakaoenterprise/kc-sdk-go/services/tgw"
)

var (
	_ resource.ResourceWithConfigure      = &transitGatewayResource{}
	_ resource.ResourceWithImportState    = &transitGatewayResource{}
	_ resource.ResourceWithValidateConfig = &transitGatewayResource{}
)

func NewTransitGatewayResource() resource.Resource {
	return &transitGatewayResource{}
}

type transitGatewayResource struct {
	kc *common.KakaoCloudClient
}

func (r *transitGatewayResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_transit_gateway"
}

func (r *transitGatewayResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: utils.MergeResourceSchemaAttributes(
			transitGatewayResourceSchemaAttributes,
			map[string]schema.Attribute{
				"timeouts": timeouts.AttributesAll(ctx),
			},
		),
	}
}

func (r *transitGatewayResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan transitGatewayResourceModel
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

	var optionsModel tgwOptionsModel
	diags = plan.Options.As(ctx, &optionsModel, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	optionsReq := tgw.BnsTgwV1ApiCreateTransitGatewayModelTgwOptionRequestModel{
		IsAutoAcceptSharedAttachments:  optionsModel.IsAutoAcceptSharedAttachments.ValueBool(),
		IsDefaultRouteTableAssociation: false,
	}

	createReq := tgw.BnsTgwV1ApiCreateTransitGatewayModelTgwRequestModel{
		Name:    plan.Name.ValueString(),
		Options: optionsReq,
	}

	body := tgw.CreateTgwRequestModel{
		Tgw: createReq,
	}

	respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*tgw.BnsTgwV1ApiCreateTransitGatewayModelCreateTgwResponseModel, *http.Response, error) {
			return r.kc.ApiClient.TgwsAPI.CreateTransitGateway(ctx).
				XAuthToken(r.kc.XAuthToken).
				CreateTgwRequestModel(body).
				Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "CreateTransitGateway", err, &resp.Diagnostics)
		return
	}

	plan.Id = types.StringValue(respModel.Tgw.Id)

	result, ok := pollTgw(ctx, r.kc, r, plan.Id.ValueString(), []string{common.TgwStatusActive, common.TgwStatusError}, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}
	common.CheckResourceAvailableStatus(ctx, r, (*string)(result.ProvisioningStatus.Get()), []string{common.TgwStatusActive}, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	ok = mapTransitGatewayResourceBaseModel(ctx, &plan.transitGatewayResourceBaseModel, result, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *transitGatewayResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state transitGatewayResourceModel
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
			return r.kc.ApiClient.TgwsAPI.GetTransitGateway(ctx, state.Id.ValueString()).
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

	tgwResult := respModel.Tgw
	ok := mapTransitGatewayResourceBaseModel(ctx, &state.transitGatewayResourceBaseModel, &tgwResult, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *transitGatewayResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state transitGatewayResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	mutex := common.LockForID(plan.Id.ValueString())
	mutex.Lock()
	defer mutex.Unlock()

	timeout, diags := plan.Timeouts.Update(ctx, common.DefaultUpdateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	_, ok := pollTgw(ctx, r.kc, r, plan.Id.ValueString(), []string{common.TgwStatusActive, common.TgwStatusError, common.TgwStatusInUse, common.TgwStatusInactive, common.TgwStatusAvaliable}, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	var planOption, stateOption tgwOptionsModel
	diags = plan.Options.As(ctx, &planOption, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	diags = state.Options.As(ctx, &stateOption, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	needUpdateOption := false
	if !planOption.IsAutoAcceptSharedAttachments.Equal(stateOption.IsAutoAcceptSharedAttachments) {
		needUpdateOption = true
	}

	if !plan.Name.Equal(state.Name) || needUpdateOption {
		updateReq := tgw.NewBnsTgwV1ApiUpdateTransitGatewayModelTgwRequestModel()

		if !plan.Name.Equal(state.Name) {
			updateReq.SetName(plan.Name.ValueString())
		}
		if needUpdateOption {
			optionsReq := tgw.BnsTgwV1ApiUpdateTransitGatewayModelTgwOptionRequestModel{
				IsAutoAcceptSharedAttachments:  planOption.IsAutoAcceptSharedAttachments.ValueBool(),
				IsDefaultRouteTableAssociation: planOption.IsDefaultRouteTableAssociation.ValueBool(),
			}
			if planOption.IsDefaultRouteTableAssociation.ValueBool() {
				optionsReq.SetAssociationDefaultRouteTableId(planOption.AssociationDefaultRouteTableId.ValueString())
			}
			updateReq.SetOptions(optionsReq)
		}
		body := tgw.UpdateTgwRequestModel{
			Tgw: *updateReq,
		}

		_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
			func() (*tgw.BnsTgwV1ApiUpdateTransitGatewayModelCreateTgwResponseModel, *http.Response, error) {
				return r.kc.ApiClient.TgwsAPI.UpdateTransitGateway(ctx, plan.Id.ValueString()).
					XAuthToken(r.kc.XAuthToken).
					UpdateTgwRequestModel(body).
					Execute()
			},
		)
		if err != nil {
			common.AddApiActionError(ctx, r, httpResp, "UpdateTransitGateway", err, &resp.Diagnostics)
			return
		}

		time.Sleep(3 * time.Second)

		result, ok := pollTgw(ctx, r.kc, r, plan.Id.ValueString(), []string{common.TgwStatusActive, common.TgwStatusError, common.TgwStatusInUse, common.TgwStatusInactive, common.TgwStatusAvaliable}, &resp.Diagnostics)
		if !ok || resp.Diagnostics.HasError() {
			return
		}
		common.CheckResourceAvailableStatus(ctx, r, (*string)(result.ProvisioningStatus.Get()), []string{common.TgwStatusActive, common.TgwStatusInUse, common.TgwStatusInactive, common.TgwStatusAvaliable}, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}

		ok = mapTransitGatewayResourceBaseModel(ctx, &state.transitGatewayResourceBaseModel, result, &resp.Diagnostics)
		if !ok || resp.Diagnostics.HasError() {
			return
		}
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *transitGatewayResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state transitGatewayResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	mutex := common.LockForID(state.Id.ValueString())
	mutex.Lock()
	defer mutex.Unlock()

	timeout, diags := state.Timeouts.Delete(ctx, common.DefaultDeleteTimeout)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	_, ok := pollTgw(ctx, r.kc, r, state.Id.ValueString(), []string{common.TgwStatusActive, common.TgwStatusError, common.TgwStatusInUse, common.TgwStatusInactive, common.TgwStatusAvaliable, common.TgwStatusDeleting}, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (interface{}, *http.Response, error) {
			httpResp, err := r.kc.ApiClient.TgwsAPI.DeleteTransitGateway(ctx, state.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				Execute()
			return nil, httpResp, err
		},
	)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			return
		}
		common.AddApiActionError(ctx, r, httpResp, "DeleteTransitGateway", err, &resp.Diagnostics)
		return
	}

	common.PollUntilDeletion(ctx, r, 10*time.Second, &resp.Diagnostics, func(ctx context.Context) (bool, *http.Response, error) {
		_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
			func() (interface{}, *http.Response, error) {
				_, httpResp, err := r.kc.ApiClient.TgwsAPI.
					GetTransitGateway(ctx, state.Id.ValueString()).
					XAuthToken(r.kc.XAuthToken).
					Execute()
				return nil, httpResp, err
			},
		)
		return false, httpResp, err
	})
}

func (r *transitGatewayResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *transitGatewayResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *transitGatewayResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config transitGatewayResourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.Options.IsNull() || config.Options.IsUnknown() {
		return
	}

	var option tgwOptionsModel
	diags = config.Options.As(ctx, &option, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !option.IsDefaultRouteTableAssociation.IsUnknown() && !option.IsDefaultRouteTableAssociation.ValueBool() && !option.AssociationDefaultRouteTableId.IsNull() {
		common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
			"Invalid Configuration: When 'is_default_route_table_association' is false, 'association_default_route_table_id' cannot be specified .")
	}
	if !option.IsDefaultRouteTableAssociation.IsUnknown() && option.IsDefaultRouteTableAssociation.ValueBool() && option.AssociationDefaultRouteTableId.IsNull() {
		common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
			"Invalid Configuration: When 'is_default_route_table_association' is true, 'association_default_route_table_id' must be provided.")
	}
}

func (r *transitGatewayResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var config, state *transitGatewayResourceModel

	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if req.Plan.Raw.IsNull() {
		return
	}

	var configOption, stateOption tgwOptionsModel
	diags = config.Options.As(ctx, &configOption, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if req.State.Raw.IsNull() && !req.Plan.Raw.IsNull() {
		if !configOption.AssociationDefaultRouteTableId.IsNull() && !configOption.AssociationDefaultRouteTableId.IsUnknown() {
			common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
				"Invalid Configuration: When creating a transit gateway, 'association_default_route_table_id' cannot be specified.")
			return
		}
	}

	if !req.State.Raw.IsNull() && !req.Plan.Raw.IsNull() {
		diags = state.Options.As(ctx, &stateOption, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		if configOption.IsDefaultRouteTableAssociation.ValueBool() &&
			configOption.AssociationDefaultRouteTableId.IsNull() &&
			stateOption.AssociationDefaultRouteTableId.IsNull() {
			common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
				"Invalid Configuration: When 'is_default_route_table_association' is set to true, 'association_default_route_table_id' must be specified.")
			return
		}
	}
}

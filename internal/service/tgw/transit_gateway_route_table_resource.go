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
	"time"

	"github.com/hashicorp/terraform-plugin-framework-nettypes/cidrtypes"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kakaoenterprise/kc-sdk-go/services/tgw"
)

var (
	_ resource.ResourceWithConfigure      = &transitGatewayRouteTableResource{}
	_ resource.ResourceWithImportState    = &transitGatewayRouteTableResource{}
	_ resource.ResourceWithValidateConfig = &transitGatewayRouteTableResource{}
)

func NewTransitGatewayRouteTableResource() resource.Resource {
	return &transitGatewayRouteTableResource{}
}

type transitGatewayRouteTableResource struct {
	kc *common.KakaoCloudClient
}

func (r *transitGatewayRouteTableResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_transit_gateway_route_table"
}

func (r *transitGatewayRouteTableResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: utils.MergeResourceSchemaAttributes(
			transitGatewayRouteTableResourceSchemaAttributes,
			map[string]schema.Attribute{
				"timeouts": timeouts.AttributesAll(ctx),
			},
		),
	}
}

func (r *transitGatewayRouteTableResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan transitGatewayRouteTableResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	mutex := common.LockForID(plan.TgwId.ValueString())
	mutex.Lock()
	defer mutex.Unlock()

	timeout, diags := plan.Timeouts.Create(ctx, common.DefaultCreateTimeout)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	createReq := tgw.BnsTgwV1ApiCreateTgwRouteTableModelTgwRouteTableRequestModel{
		TgwId: plan.TgwId.ValueString(),
		Name:  plan.Name.ValueString(),
	}

	body := tgw.CreateTgwRouteTableRequestModel{
		RouteTable: createReq,
	}

	respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*tgw.BnsTgwV1ApiCreateTgwRouteTableModelCreateTgwRouteTableResponseModel, *http.Response, error) {
			return r.kc.ApiClient.RouteTablesAPI.CreateTgwRouteTable(ctx).
				XAuthToken(r.kc.XAuthToken).
				CreateTgwRouteTableRequestModel(body).
				Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "CreateTgwRouteTable", err, &resp.Diagnostics)
		return
	}

	plan.Id = types.StringValue(respModel.RouteTable.Id)
	result, ok := r.pollRouteTableUntilAvailable(ctx, plan.Id.ValueString(), &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	if !plan.RequestRoutes.IsNull() {
		var requestRoutes []tgwRouteTableRequestRouteModel
		diags := plan.RequestRoutes.ElementsAs(ctx, &requestRoutes, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		for i := range requestRoutes {
			tmpPlan := requestRoutes[i]
			r.addRoute(ctx, plan.Id.ValueString(), &tmpPlan, &resp.Diagnostics)
			requestRoutes[i] = tmpPlan
		}
		elemType := types.ObjectType{AttrTypes: tgwRouteTableRequestRouteAttrType}
		plan.RequestRoutes, diags = types.SetValueFrom(ctx, elemType, requestRoutes)
		resp.Diagnostics.Append(diags...)

		result, ok = r.pollRouteTableUntilAvailable(ctx, plan.Id.ValueString(), &resp.Diagnostics)
		if !ok || resp.Diagnostics.HasError() {
			return
		}
	}

	if !plan.RequestAssociations.IsNull() {
		var requestAssociations []tgwRouteTableRequestAssociationModel
		diags := plan.RequestAssociations.ElementsAs(ctx, &requestAssociations, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		for i := range requestAssociations {
			tmpPlan := requestAssociations[i]
			r.addAssociation(ctx, plan.Id.ValueString(), &tmpPlan, &resp.Diagnostics)
			requestAssociations[i] = tmpPlan
		}
		elemType := types.ObjectType{AttrTypes: tgwRouteTableRequestAssociationAttrType}
		plan.RequestAssociations, diags = types.SetValueFrom(ctx, elemType, requestAssociations)
		resp.Diagnostics.Append(diags...)

		result, ok = r.pollRouteTableUntilAvailable(ctx, plan.Id.ValueString(), &resp.Diagnostics)
		if !ok || resp.Diagnostics.HasError() {
			return
		}
	}

	ok = mapTransitGatewayRouteTableBaseModel(ctx, &plan.transitGatewayRouteTableBaseModel, result, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *transitGatewayRouteTableResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state transitGatewayRouteTableResourceModel
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
		func() (*tgw.GetTgwRouteTableResponseModel, *http.Response, error) {
			return r.kc.ApiClient.RouteTablesAPI.GetTgwRouteTable(ctx, state.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				Execute()
		},
	)

	if httpResp != nil && httpResp.StatusCode == 404 {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "GetTgwRouteTable", err, &resp.Diagnostics)
		return
	}

	if !mapTransitGatewayRouteTableBaseModel(ctx, &state.transitGatewayRouteTableBaseModel, &respModel.TgwRouteTable, &resp.Diagnostics) {
		return
	}

	if respModel.TgwRouteTable.Routes != nil && len(respModel.TgwRouteTable.Routes) > 0 {
		var requestRoutes []tgwRouteTableRequestRouteModel

		for _, route := range respModel.TgwRouteTable.Routes {
			destinationCidrBlock := route.DestinationCidrBlock.Get()
			requestRoutes = append(requestRoutes,
				tgwRouteTableRequestRouteModel{
					Id:                   utils.ConvertNullableString(route.Id),
					DestinationCidrBlock: cidrtypes.NewIPPrefixValue(*destinationCidrBlock),
					TgwAttachmentId:      utils.ConvertNullableString(route.ResourceAttachmentId),
				})
		}
		var mapDiags diag.Diagnostics
		elemType := types.ObjectType{AttrTypes: tgwRouteTableRequestRouteAttrType}
		state.RequestRoutes, mapDiags = types.SetValueFrom(ctx, elemType, requestRoutes)
		diags.Append(mapDiags...)
		if diags.HasError() {
			return
		}
	}

	if respModel.TgwRouteTable.Associations != nil && len(respModel.TgwRouteTable.Associations) > 0 {
		var requestAssociations []tgwRouteTableRequestAssociationModel

		for _, association := range respModel.TgwRouteTable.Associations {
			requestAssociations = append(requestAssociations,
				tgwRouteTableRequestAssociationModel{
					Id:              utils.ConvertNullableString(association.Id),
					TgwAttachmentId: utils.ConvertNullableString(association.ResourceAttachmentId),
				})
		}
		var mapDiags diag.Diagnostics
		elemType := types.ObjectType{AttrTypes: tgwRouteTableRequestAssociationAttrType}
		state.RequestAssociations, mapDiags = types.SetValueFrom(ctx, elemType, requestAssociations)
		diags.Append(mapDiags...)
		if diags.HasError() {
			return
		}
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *transitGatewayRouteTableResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state transitGatewayRouteTableResourceModel
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

	mutex := common.LockForID(plan.TgwId.ValueString())
	mutex.Lock()
	defer mutex.Unlock()

	timeout, diags := plan.Timeouts.Update(ctx, common.DefaultUpdateTimeout)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var result *tgw.BnsTgwV1ApiGetTgwRouteTableModelTgwRouteTableResponseModel
	var ok bool
	if !plan.Name.Equal(state.Name) {
		updateReq := tgw.NewBnsTgwV1ApiUpdateTgwRouteTableModelTgwRouteTableRequestModel(plan.Name.ValueString())

		body := tgw.UpdateTgwRouteTableRequestModel{
			RouteTable: *updateReq,
		}

		_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
			func() (*tgw.BnsTgwV1ApiUpdateTgwRouteTableModelCreateTgwRouteTableResponseModel, *http.Response, error) {
				return r.kc.ApiClient.RouteTablesAPI.UpdateTgwRouteTable(ctx, plan.Id.ValueString()).
					XAuthToken(r.kc.XAuthToken).
					UpdateTgwRouteTableRequestModel(body).
					Execute()
			},
		)
		if err != nil {
			common.AddApiActionError(ctx, r, httpResp, "UpdateTgwRouteTable", err, &resp.Diagnostics)
			return
		}

		time.Sleep(5 * time.Second)

		result, ok = r.pollRouteTableUntilAvailable(ctx, plan.Id.ValueString(), &resp.Diagnostics)
		if !ok || resp.Diagnostics.HasError() {
			return
		}
	}

	if !plan.RequestRoutes.Equal(state.RequestRoutes) {
		var planRequestRoutes, stateRequestRoutes []tgwRouteTableRequestRouteModel
		diags := plan.RequestRoutes.ElementsAs(ctx, &planRequestRoutes, false)
		resp.Diagnostics.Append(diags...)
		diags = state.RequestRoutes.ElementsAs(ctx, &stateRequestRoutes, false)
		if resp.Diagnostics.HasError() {
			return
		}
		r.updateRouteTableRoutes(ctx, plan.Id.ValueString(), &planRequestRoutes, &stateRequestRoutes, &resp.Diagnostics)
		result, ok = r.pollRouteTableUntilAvailable(ctx, plan.Id.ValueString(), &resp.Diagnostics)
		if !ok || resp.Diagnostics.HasError() {
			return
		}
	}

	if !plan.RequestAssociations.Equal(state.RequestAssociations) {
		var planRequestAssociations, stateRequestAssociations []tgwRouteTableRequestAssociationModel
		diags := plan.RequestAssociations.ElementsAs(ctx, &planRequestAssociations, false)
		resp.Diagnostics.Append(diags...)
		diags = state.RequestAssociations.ElementsAs(ctx, &stateRequestAssociations, false)
		if resp.Diagnostics.HasError() {
			return
		}
		r.updateRouteTableAssociations(ctx, plan.Id.ValueString(), &planRequestAssociations, &stateRequestAssociations, &resp.Diagnostics)
		result, ok = r.pollRouteTableUntilAvailable(ctx, plan.Id.ValueString(), &resp.Diagnostics)
		if !ok || resp.Diagnostics.HasError() {
			return
		}
	}

	if !mapTransitGatewayRouteTableBaseModel(ctx, &plan.transitGatewayRouteTableBaseModel, result, &resp.Diagnostics) {
		return
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *transitGatewayRouteTableResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state transitGatewayRouteTableResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	mutex := common.LockForID(state.TgwId.ValueString())
	mutex.Lock()
	defer mutex.Unlock()

	timeout, diags := state.Timeouts.Delete(ctx, common.DefaultDeleteTimeout)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if !state.Routes.IsNull() && len(state.Routes.Elements()) > 0 {
		var routes []tgwRouteTableRouteNestedModel
		diags := state.Routes.ElementsAs(ctx, &routes, false)
		resp.Diagnostics.Append(diags...)
		for _, route := range routes {
			r.deleteRoute(ctx, state.Id.ValueString(), route.Id.ValueString(), &resp.Diagnostics)
		}
	}

	if !state.Associations.IsNull() && len(state.Associations.Elements()) > 0 {
		var associations []tgwRouteTableAssociationNestedModel
		diags := state.Associations.ElementsAs(ctx, &associations, false)
		resp.Diagnostics.Append(diags...)
		for _, association := range associations {
			r.deleteAssociation(ctx, state.Id.ValueString(), association.Id.ValueString(), &resp.Diagnostics)
		}
	}

	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (interface{}, *http.Response, error) {
			httpResp, err := r.kc.ApiClient.RouteTablesAPI.DeleteTgwRouteTable(ctx, state.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				Execute()
			return nil, httpResp, err
		},
	)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			return
		}
		common.AddApiActionError(ctx, r, httpResp, "DeleteTgwRouteTable", err, &resp.Diagnostics)
		return
	}

	common.PollUntilDeletion(ctx, r, 5*time.Second, &resp.Diagnostics, func(ctx context.Context) (bool, *http.Response, error) {
		_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
			func() (interface{}, *http.Response, error) {
				_, httpResp, err := r.kc.ApiClient.RouteTablesAPI.
					GetTgwRouteTable(ctx, state.Id.ValueString()).
					XAuthToken(r.kc.XAuthToken).
					Execute()
				return nil, httpResp, err
			},
		)
		return false, httpResp, err
	})
}

func (r *transitGatewayRouteTableResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *transitGatewayRouteTableResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *transitGatewayRouteTableResource) pollRouteTableUntilAvailable(
	ctx context.Context,
	routeTableId string,
	diags *diag.Diagnostics,
) (*tgw.BnsTgwV1ApiGetTgwRouteTableModelTgwRouteTableResponseModel, bool) {

	result, ok := common.PollUntilResult(
		ctx,
		r,
		5*time.Second,
		"transit gateway route table",
		routeTableId,
		[]string{
			common.TgwStatusActive,
			common.TgwStatusError,
			common.TgwStatusInUse,
			common.TgwStatusInactive,
			common.TgwStatusAvaliable,
		},
		diags,
		func(ctx context.Context) (*tgw.BnsTgwV1ApiGetTgwRouteTableModelTgwRouteTableResponseModel, *http.Response, error) {
			respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, diags,
				func() (*tgw.GetTgwRouteTableResponseModel, *http.Response, error) {
					return r.kc.ApiClient.RouteTablesAPI.
						GetTgwRouteTable(ctx, routeTableId).
						XAuthToken(r.kc.XAuthToken).
						Execute()
				},
			)
			if err != nil {
				return nil, httpResp, err
			}
			return &respModel.TgwRouteTable, httpResp, nil
		},
		func(v *tgw.BnsTgwV1ApiGetTgwRouteTableModelTgwRouteTableResponseModel) string {
			if v.ProvisioningStatus.IsSet() {
				return string(*v.ProvisioningStatus.Get())
			}
			return ""
		},
	)

	if !ok || diags.HasError() {
		return nil, false
	}

	common.CheckResourceAvailableStatus(
		ctx,
		r,
		(*string)(result.ProvisioningStatus.Get()),
		[]string{
			common.TgwStatusActive,
			common.TgwStatusInUse,
			common.TgwStatusInactive,
			common.TgwStatusAvaliable,
		},
		diags,
	)

	if diags.HasError() {
		return nil, false
	}

	return result, true
}

func (r *transitGatewayRouteTableResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config transitGatewayRouteTableResourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.validateRouteConfig(ctx, config, resp)
}

func (r *transitGatewayRouteTableResource) validateRouteConfig(ctx context.Context, config transitGatewayRouteTableResourceModel, resp *resource.ValidateConfigResponse) {
	if config.RequestRoutes.IsNull() || config.RequestRoutes.IsUnknown() {
		return
	}

	var routes []tgwRouteTableRequestRouteModel
	diags := config.RequestRoutes.ElementsAs(ctx, &routes, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if len(routes) > 0 {
		seen := make(map[string]bool)

		for _, route := range routes {

			destStr := strings.TrimSpace(route.DestinationCidrBlock.ValueString())
			if _, exists := seen[destStr]; exists {
				common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
					fmt.Sprintf("routes destinationCidrBlock '%s' is duplicated.", destStr),
				)
				return
			} else {
				seen[destStr] = true
			}
		}
	}
}

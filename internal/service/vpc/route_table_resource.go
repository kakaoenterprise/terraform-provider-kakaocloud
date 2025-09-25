// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpc

import (
	"context"
	"fmt"
	"net/http"
	"sort"
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
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/kakaoenterprise/kc-sdk-go/services/vpc"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.ResourceWithConfigure   = &routeTableResource{}
	_ resource.ResourceWithImportState = &routeTableResource{}
)

func NewRouteTableResource() resource.Resource {
	return &routeTableResource{}
}

type routeTableResource struct {
	kc *common.KakaoCloudClient
}

// Metadata returns the resource type name.
func (r *routeTableResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_route_table"
}

// Schema defines the schema for the resource.
func (r *routeTableResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Represents a route table resource.",
		Attributes: utils.MergeResourceSchemaAttributes(
			routeTableResourceSchemaAttributes,
			map[string]schema.Attribute{
				"timeouts": timeouts.AttributesAll(ctx),
			},
		),
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *routeTableResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan routeTableResourceModel
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

	createReq := vpc.CreateRouteTableModel{
		Name:  plan.Name.ValueString(),
		VpcId: plan.VpcId.ValueString(),
	}

	body := vpc.BodyCreateRouteTable{
		VpcRouteTable: createReq,
	}

	respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*vpc.BnsVpcV1ApiCreateRouteTableModelResponseRouteTableModel, *http.Response, error) {
			return r.kc.ApiClient.VPCRouteTableAPI.CreateRouteTable(ctx).
				XAuthToken(r.kc.XAuthToken).BodyCreateRouteTable(body).Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "CreateRouteTable", err, &resp.Diagnostics)
		return
	}

	plan.Id = types.StringValue(respModel.VpcRouteTable.Id)

	result, ok := r.pollRouteTableUntilStatus(
		ctx,
		plan.Id.ValueString(),
		[]string{common.VpcProvisioningStatusActive},
		&resp.Diagnostics,
	)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	// Set Main Route Table
	if !plan.IsMain.IsNull() || !plan.IsMain.IsUnknown() {
		if plan.IsMain.ValueBool() {
			r.setMainRouteTable(ctx, plan.Id.ValueString(), &resp.Diagnostics)
			if resp.Diagnostics.HasError() {
				return
			}
			result, ok = r.pollRouteTableUntilStatus(
				ctx,
				plan.Id.ValueString(),
				[]string{common.VpcProvisioningStatusActive},
				&resp.Diagnostics,
			)
			if !ok || resp.Diagnostics.HasError() {
				return
			}
		}
	}

	// Add routes
	if !plan.RequestRoutes.IsNull() {
		planList, planDiags := r.convertListToRouteTableRequestRouteModel(ctx, plan.RequestRoutes)
		resp.Diagnostics.Append(planDiags...)

		if !resp.Diagnostics.HasError() {
			for i := range planList {
				tmpPlan := planList[i]
				r.addRoute(ctx, plan.Id.ValueString(), plan.VpcId.ValueString(), &tmpPlan, &resp.Diagnostics)
				planList[i] = tmpPlan
			}
			elemType := types.ObjectType{AttrTypes: routeTableRequestRouteAttrType}
			plan.RequestRoutes, diags = types.ListValueFrom(ctx, elemType, planList)
			resp.Diagnostics.Append(diags...)

			result, ok = r.pollRouteTableUntilStatus(
				ctx,
				plan.Id.ValueString(),
				[]string{common.VpcProvisioningStatusActive},
				&resp.Diagnostics,
			)
			if !ok || resp.Diagnostics.HasError() {
				return
			}
		}
	}

	ok = mapRouteTableBaseModel(ctx, &plan.routeTableBaseModel, result, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *routeTableResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state routeTableResourceModel
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
		func() (*vpc.BnsVpcV1ApiGetRouteTableModelResponseRouteTableModel, *http.Response, error) {
			return r.kc.ApiClient.VPCRouteTableAPI.GetRouteTable(ctx, state.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).Execute()
		},
	)

	// 404 → Remove Terraform state
	if httpResp != nil && httpResp.StatusCode == 404 {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "GetRouteTable", err, &resp.Diagnostics)
		return
	}

	routeTableResult := respModel.VpcRouteTable
	ok := mapRouteTableBaseModel(ctx, &state.routeTableBaseModel, &routeTableResult, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	if state.RequestRoutes.IsNull() {
		var routes []routeTableRouteModel
		diags := state.Routes.ElementsAs(ctx, &routes, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		var filteredRoutes []routeTableRouteModel
		for _, route := range routes {
			if !route.IsLocalRoute.ValueBool() {
				filteredRoutes = append(filteredRoutes, route)
			}
		}
		routes = filteredRoutes

		sort.Slice(routes, func(i, j int) bool {
			return utils.CompareCIDRs(routes[i].Destination.ValueString(), routes[j].Destination.ValueString()) < 0
		})

		var requestRouteList []routeTableRequestRouteModel
		for _, route := range routes {
			requestRouteList = append(requestRouteList,
				routeTableRequestRouteModel{
					Id:          route.Id,
					Destination: cidrtypes.NewIPPrefixValue(route.Destination.ValueString()),
					TargetType:  types.StringValue(strings.ToLower(route.TargetType.ValueString())),
					TargetId:    route.TargetId,
				})
		}

		elemType := types.ObjectType{AttrTypes: routeTableRequestRouteAttrType}
		state.RequestRoutes, diags = types.ListValueFrom(ctx, elemType, requestRouteList)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *routeTableResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state routeTableResourceModel
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

	timeout, diags := plan.Timeouts.Update(ctx, common.DefaultUpdateTimeout)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Set Main Route Table
	if !plan.IsMain.IsNull() || !plan.IsMain.IsUnknown() {
		if plan.IsMain.ValueBool() && !state.IsMain.ValueBool() {
			r.setMainRouteTable(ctx, plan.Id.ValueString(), &resp.Diagnostics)
			if resp.Diagnostics.HasError() {
				return
			}
		}
		_, ok := r.pollRouteTableUntilStatus(
			ctx,
			plan.Id.ValueString(),
			[]string{common.VpcProvisioningStatusActive},
			&resp.Diagnostics,
		)
		if !ok || resp.Diagnostics.HasError() {
			return
		}
	}

	// Change Routes
	if !plan.RequestRoutes.Equal(state.RequestRoutes) {
		planList, planDiags := r.convertListToRouteTableRequestRouteModel(ctx, plan.RequestRoutes)
		stateList, stateDiags := r.convertListToRouteTableRequestRouteModel(ctx, state.RequestRoutes)
		resp.Diagnostics.Append(planDiags...)
		resp.Diagnostics.Append(stateDiags...)
		if resp.Diagnostics.HasError() {
			return
		}

		ok := r.updateRouteTableRoutes(ctx, plan.Id.ValueString(), plan.VpcId.ValueString(), &planList, &stateList, &resp.Diagnostics)
		if !ok || resp.Diagnostics.HasError() {
			return
		}
		elemType := types.ObjectType{AttrTypes: routeTableRequestRouteAttrType}
		plan.RequestRoutes, diags = types.ListValueFrom(ctx, elemType, planList)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	result, ok := r.pollRouteTableUntilStatus(
		ctx,
		plan.Id.ValueString(),
		[]string{common.VpcProvisioningStatusActive},
		&resp.Diagnostics,
	)
	if !plan.IsMain.IsNull() || !plan.IsMain.IsUnknown() {
		isMain := plan.IsMain.ValueBool()
		result.IsMain.Set(&isMain)
	}
	if !ok || resp.Diagnostics.HasError() {
		return
	}
	ok = mapRouteTableBaseModel(ctx, &plan.routeTableBaseModel, result, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *routeTableResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state routeTableResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.IsMain.ValueBool() {
		common.AddGeneralError(ctx, r, &resp.Diagnostics,
			fmt.Sprintf("Route table %q is the main route table and cannot be deleted.", state.Id.ValueString()),
		)
		return
	}
	if !state.Associations.IsNull() && len(state.Associations.Elements()) > 0 {
		common.AddGeneralError(ctx, r, &resp.Diagnostics,
			fmt.Sprintf("Route table %q has subnet associations.", state.Id.ValueString()),
		)
		return
	}

	timeout, diags := state.Timeouts.Delete(ctx, common.DefaultDeleteTimeout)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (interface{}, *http.Response, error) {
			httpResp, err := r.kc.ApiClient.VPCRouteTableAPI.DeleteRouteTable(ctx, state.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				Execute()
			return nil, httpResp, err
		},
	)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			return
		}
		common.AddApiActionError(ctx, r, httpResp, "DeleteRouteTable", err, &resp.Diagnostics)
		return
	}

	// Poll until resource disappears
	common.PollUntilDeletion(ctx, r, 2*time.Second, &resp.Diagnostics, func(ctx context.Context) (bool, *http.Response, error) {
		_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
			func() (*vpc.BnsVpcV1ApiGetRouteTableModelResponseRouteTableModel, *http.Response, error) {
				_, httpResp, err := r.kc.ApiClient.VPCRouteTableAPI.
					GetRouteTable(ctx, state.Id.ValueString()).
					XAuthToken(r.kc.XAuthToken).
					Execute()
				return nil, httpResp, err
			},
		)
		return false, httpResp, err
	})
}

func (r *routeTableResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Add a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*common.KakaoCloudClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *kakaocloud.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.kc = client
}

func (r *routeTableResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *routeTableResource) setMainRouteTable(ctx context.Context, routeTableId string, diag *diag.Diagnostics) {
	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, diag,
		func() (interface{}, *http.Response, error) {
			return r.kc.ApiClient.VPCRouteTableAPI.SetMainRouteTable(ctx, routeTableId).
				XAuthToken(r.kc.XAuthToken).
				Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "SetMainRouteTable", err, diag)
		return
	}
}

func (r *routeTableResource) convertListToRouteTableRequestRouteModel(
	ctx context.Context,
	list types.List,
) ([]routeTableRequestRouteModel, diag.Diagnostics) {
	var result []routeTableRequestRouteModel
	var diags diag.Diagnostics

	for _, elem := range list.Elements() {
		if obj, ok := elem.(types.Object); ok {
			var model routeTableRequestRouteModel
			elemDiags := obj.As(ctx, &model, basetypes.ObjectAsOptions{})
			diags.Append(elemDiags...)
			result = append(result, model)
		}
	}
	return result, diags
}

func (r *routeTableResource) pollRouteTableUntilStatus(
	ctx context.Context,
	routeTableId string,
	targetStatuses []string,
	resp *diag.Diagnostics,
) (*vpc.BnsVpcV1ApiGetRouteTableModelRouteTableModel, bool) {
	return common.PollUntilResult(
		ctx,
		r,
		2*time.Second,
		targetStatuses,
		resp,
		func(ctx context.Context) (*vpc.BnsVpcV1ApiGetRouteTableModelRouteTableModel, *http.Response, error) {
			respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, resp,
				func() (*vpc.BnsVpcV1ApiGetRouteTableModelResponseRouteTableModel, *http.Response, error) {
					return r.kc.ApiClient.VPCRouteTableAPI.
						GetRouteTable(ctx, routeTableId).
						XAuthToken(r.kc.XAuthToken).
						Execute()
				},
			)
			if err != nil {
				return nil, httpResp, err
			}
			return &respModel.VpcRouteTable, httpResp, nil
		},
		func(v *vpc.BnsVpcV1ApiGetRouteTableModelRouteTableModel) string {
			return string(*v.ProvisioningStatus.Get())
		},
	)
}

func (r *routeTableResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config routeTableResourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.validateRouteConfig(ctx, config, resp)
}

func (r *routeTableResource) validateRouteConfig(ctx context.Context, config routeTableResourceModel, resp *resource.ValidateConfigResponse) {
	if config.Routes.IsNull() || config.Routes.IsUnknown() {
		return
	}

	var routes []routeTableRouteModel
	diags := config.Routes.ElementsAs(ctx, &routes, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if len(routes) > 0 {
		seen := make(map[string]bool)

		for _, route := range routes {
			// Destination Dup check
			destStr := strings.TrimSpace(route.Destination.ValueString())
			if _, exists := seen[destStr]; exists {
				common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
					fmt.Sprintf("routes destination '%s' is duplicated.", destStr),
				)
				return
			} else {
				seen[destStr] = true
			}
		}
	}
}

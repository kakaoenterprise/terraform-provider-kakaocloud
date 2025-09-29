// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package vpc

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-kakaocloud/internal/common"
	"terraform-provider-kakaocloud/internal/docs"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kakaoenterprise/kc-sdk-go/services/vpc"
)

var (
	_ resource.ResourceWithConfigure   = &subnetShareResource{}
	_ resource.ResourceWithImportState = &subnetShareResource{}
)

func NewSubnetShareResource() resource.Resource {
	return &subnetShareResource{}
}

type subnetShareResource struct {
	kc *common.KakaoCloudClient
}

func (r *subnetShareResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subnet_share"
}

func (r *subnetShareResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: docs.GetResourceDescription("SubnetShare"),
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:   true,
				Validators: common.UuidValidator(),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "Subnet ID",
			},
			"projects": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "List of shared Project IDs for the Subnet",
			},
			"project_ids": schema.SetAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "List of Project IDs requesting subnet sharing",
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
					setvalidator.ValueStringsAre(common.UuidNoHyphenValidator()...),
				},
			},
			"timeouts": timeouts.AttributesAll(ctx),
		},
	}
}

func (r *subnetShareResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan subnetShareResourceModel
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

	var projectIds []string
	diags = plan.ProjectIds.ElementsAs(ctx, &projectIds, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	for _, subjectId := range projectIds {
		_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
			func() (interface{}, *http.Response, error) {
				return r.kc.ApiClient.VPCSubnetAPI.ShareSubnet(ctx, plan.Id.ValueString(), subjectId).
					XAuthToken(r.kc.XAuthToken).Execute()
			},
		)
		if err != nil {
			common.AddApiActionError(ctx, r, httpResp, "ShareSubnet", err, &resp.Diagnostics)
			return
		}
	}

	plan.Projects, diags = types.ListValueFrom(ctx, types.StringType, plan.ProjectIds.Elements())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

func (r *subnetShareResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state subnetShareResourceModel

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

	subnetShareResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*vpc.ResponseSubnetSharedProjectListModel, *http.Response, error) {
			return r.kc.ApiClient.VPCSubnetAPI.ListSubnetSharedProjects(ctx, state.Id.ValueString()).XAuthToken(r.kc.XAuthToken).Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "ListSubnetSharedProjects", err, &resp.Diagnostics)
		return
	}

	projectIDs := make([]attr.Value, 0, len(subnetShareResp.Projects))
	for _, p := range subnetShareResp.Projects {
		projectIDs = append(projectIDs, types.StringValue(p.Id))
	}

	state.Projects, diags = types.ListValue(types.StringType, projectIDs)
	resp.Diagnostics.Append(diags...)

	if state.ProjectIds.IsNull() {
		state.ProjectIds, diags = types.SetValue(types.StringType, projectIDs)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *subnetShareResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state subnetShareResourceModel
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

	var planProjectIds, stateProjectIds []string
	diags = plan.ProjectIds.ElementsAs(ctx, &planProjectIds, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	diags = state.ProjectIds.ElementsAs(ctx, &stateProjectIds, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	planSet := make(map[string]struct{})
	stateSet := make(map[string]struct{})
	for _, id := range planProjectIds {
		planSet[id] = struct{}{}
	}
	for _, id := range stateProjectIds {
		stateSet[id] = struct{}{}
	}

	for id := range planSet {
		if _, exists := stateSet[id]; !exists {
			_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
				func() (interface{}, *http.Response, error) {
					return r.kc.ApiClient.VPCSubnetAPI.ShareSubnet(ctx, plan.Id.ValueString(), id).
						XAuthToken(r.kc.XAuthToken).Execute()
				},
			)
			if err != nil {
				common.AddApiActionError(ctx, r, httpResp, "ShareSubnet", err, &resp.Diagnostics)
				return
			}
		}
	}

	for id := range stateSet {
		if _, exists := planSet[id]; !exists {
			_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
				func() (interface{}, *http.Response, error) {
					httpResp, err := r.kc.ApiClient.VPCSubnetAPI.UnshareSubnet(ctx, plan.Id.ValueString(), id).
						XAuthToken(r.kc.XAuthToken).Execute()
					return nil, httpResp, err
				},
			)
			if err != nil {
				common.AddApiActionError(ctx, r, httpResp, "UnshareSubnet", err, &resp.Diagnostics)
				return
			}
		}
	}

	plan.Projects, diags = types.ListValueFrom(ctx, types.StringType, plan.ProjectIds.Elements())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *subnetShareResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state subnetShareResourceModel
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

	var stateProjectIds []string
	diags = state.ProjectIds.ElementsAs(ctx, &stateProjectIds, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	for _, id := range stateProjectIds {
		_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
			func() (interface{}, *http.Response, error) {
				httpResp, err := r.kc.ApiClient.VPCSubnetAPI.UnshareSubnet(ctx, state.Id.ValueString(), id).
					XAuthToken(r.kc.XAuthToken).
					Execute()
				return nil, httpResp, err
			},
		)
		if err != nil {
			common.AddApiActionError(ctx, r, httpResp, "UnshareSubnet", err, &resp.Diagnostics)
			return
		}
	}
}

func (r *subnetShareResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {

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

func (r *subnetShareResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {

	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

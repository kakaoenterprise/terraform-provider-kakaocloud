// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package vpc

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-kakaocloud/internal/common"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kakaoenterprise/kc-sdk-go/v2/services/vpc"
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
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:   true,
				Validators: common.UuidValidator(),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"projects": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
			},
			"project_ids": schema.SetAttribute{
				Required:    true,
				ElementType: types.StringType,
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

	subnetShareResp, ok := r.pollUntilProjectsApplied(
		ctx,
		plan.Id.ValueString(),
		projectIds,
		nil,
		&resp.Diagnostics,
	)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	plan.Projects, diags = subnetSharedProjectsList(subnetShareResp)
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
		func() (*vpc.ListSubnetSharedProjectsResponse, *http.Response, error) {
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
	} else {
		sharedProjectIDs := make(map[string]struct{}, len(projectIDs))
		for _, projectID := range projectIDs {
			sharedProjectIDs[projectID.(types.String).ValueString()] = struct{}{}
		}

		filteredProjectIDs := make([]attr.Value, 0, len(state.ProjectIds.Elements()))
		for _, projectID := range state.ProjectIds.Elements() {
			if _, exists := sharedProjectIDs[projectID.(types.String).ValueString()]; exists {
				filteredProjectIDs = append(filteredProjectIDs, projectID)
			}
		}

		state.ProjectIds, diags = types.SetValue(types.StringType, filteredProjectIDs)
	}
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
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
	removedProjectIds := make([]string, 0)
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
			removedProjectIds = append(removedProjectIds, id)
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

	subnetShareResp, ok := r.pollUntilProjectsApplied(
		ctx,
		plan.Id.ValueString(),
		planProjectIds,
		removedProjectIds,
		&resp.Diagnostics,
	)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	plan.Projects, diags = subnetSharedProjectsList(subnetShareResp)
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

func (r *subnetShareResource) pollUntilProjectsApplied(
	ctx context.Context,
	subnetID string,
	expectedProjectIDs []string,
	removedProjectIDs []string,
	respDiags *diag.Diagnostics,
) (*vpc.ListSubnetSharedProjectsResponse, bool) {
	expected := make(map[string]struct{}, len(expectedProjectIDs))
	for _, projectID := range expectedProjectIDs {
		expected[projectID] = struct{}{}
	}

	removed := make(map[string]struct{}, len(removedProjectIDs))
	for _, projectID := range removedProjectIDs {
		removed[projectID] = struct{}{}
	}

	return common.PollUntilResult(
		ctx,
		r,
		3*time.Second,
		"subnet shared projects",
		subnetID,
		[]string{"APPLIED"},
		respDiags,
		func(ctx context.Context) (*vpc.ListSubnetSharedProjectsResponse, *http.Response, error) {
			return common.ExecuteWithRetryAndAuth(ctx, r.kc, respDiags,
				func() (*vpc.ListSubnetSharedProjectsResponse, *http.Response, error) {
					return r.kc.ApiClient.VPCSubnetAPI.
						ListSubnetSharedProjects(ctx, subnetID).
						XAuthToken(r.kc.XAuthToken).
						Execute()
				},
			)
		},
		func(result *vpc.ListSubnetSharedProjectsResponse) string {
			if result == nil {
				return "APPLYING"
			}

			actual := make(map[string]struct{}, len(result.Projects))
			for _, project := range result.Projects {
				actual[project.Id] = struct{}{}
			}

			for projectID := range expected {
				if _, exists := actual[projectID]; !exists {
					return "APPLYING"
				}
			}
			for projectID := range removed {
				if _, exists := actual[projectID]; exists {
					return "APPLYING"
				}
			}

			return "APPLIED"
		},
	)
}

func subnetSharedProjectsList(result *vpc.ListSubnetSharedProjectsResponse) (types.List, diag.Diagnostics) {
	projectIDs := make([]attr.Value, 0, len(result.Projects))
	for _, project := range result.Projects {
		projectIDs = append(projectIDs, types.StringValue(project.Id))
	}
	return types.ListValue(types.StringType, projectIDs)
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

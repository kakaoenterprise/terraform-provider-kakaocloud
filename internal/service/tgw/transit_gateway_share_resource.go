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

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kakaoenterprise/kc-sdk-go/v2/services/tgw"
)

var (
	_ resource.ResourceWithConfigure   = &transitGatewayShareResource{}
	_ resource.ResourceWithImportState = &transitGatewayShareResource{}
)

func NewTransitGatewayShareResource() resource.Resource {
	return &transitGatewayShareResource{}
}

type transitGatewayShareResource struct {
	kc *common.KakaoCloudClient
}

func (r *transitGatewayShareResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_transit_gateway_share"
}

func (r *transitGatewayShareResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: utils.MergeResourceSchemaAttributes(
			transitGatewayShareResourceSchemaAttributes,
			map[string]schema.Attribute{
				"timeouts": timeouts.AttributesAll(ctx),
			},
		),
	}
}

func (r *transitGatewayShareResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan transitGatewayShareResourceModel
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

	unlock, ok := common.AcquireLockForID(ctx, plan.TgwId.ValueString(), &resp.Diagnostics)
	if !ok {
		return
	}
	defer unlock()

	_, ok = pollTgw(ctx, r.kc, r, plan.TgwId.ValueString(), tgwAvailableAndErrorStatuses, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (interface{}, *http.Response, error) {
			return r.kc.ApiClient.TransitGatewayAPI.ShareTransitGateway(
				ctx,
				plan.TgwId.ValueString(),
				plan.TargetProjectId.ValueString(),
			).XAuthToken(r.kc.XAuthToken).Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "ShareTransitGateway", err, &resp.Diagnostics)
		return
	}

	_, ok = r.pollUntilSharedProject(
		ctx,
		plan.TgwId.ValueString(),
		plan.TargetProjectId.ValueString(),
		&resp.Diagnostics,
	)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	plan.Id = types.StringValue(fmt.Sprintf("%s/%s", plan.TgwId.ValueString(), plan.TargetProjectId.ValueString()))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *transitGatewayShareResource) pollUntilSharedProject(
	ctx context.Context,
	tgwID string,
	targetProjectID string,
	respDiags *diag.Diagnostics,
) (*tgw.ListTgwSharedProjectsResponse, bool) {
	return common.PollUntilResult(
		ctx,
		r,
		3*time.Second,
		"transit gateway shared project",
		targetProjectID,
		[]string{"APPLIED"},
		respDiags,
		func(ctx context.Context) (*tgw.ListTgwSharedProjectsResponse, *http.Response, error) {
			return common.ExecuteWithRetryAndAuth(ctx, r.kc, respDiags,
				func() (*tgw.ListTgwSharedProjectsResponse, *http.Response, error) {
					return r.kc.ApiClient.TransitGatewayAPI.
						ListTgwSharedProjects(ctx, tgwID).
						XAuthToken(r.kc.XAuthToken).
						Execute()
				},
			)
		},
		func(result *tgw.ListTgwSharedProjectsResponse) string {
			if result == nil {
				return "APPLYING"
			}

			for _, project := range result.Projects {
				if project.Id == targetProjectID {
					return "APPLIED"
				}
			}

			return "APPLYING"
		},
	)
}

func (r *transitGatewayShareResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state transitGatewayShareResourceModel
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
		func() (*tgw.ListTgwSharedProjectsResponse, *http.Response, error) {
			return r.kc.ApiClient.TransitGatewayAPI.ListTgwSharedProjects(ctx, state.TgwId.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				Execute()
		},
	)

	if httpResp != nil && httpResp.StatusCode == 404 {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "ListTgwSharedProjects", err, &resp.Diagnostics)
		return
	}

	found := false
	for _, project := range respModel.Projects {
		if project.Id == state.TargetProjectId.ValueString() {
			found = true
			break
		}
	}

	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *transitGatewayShareResource) Update(ctx context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	common.AddGeneralError(
		ctx, r, &resp.Diagnostics,
		"Updates are not supported for transit_gateway_share. Both tgw_id and target_project_id require replacement.",
	)
}

func (r *transitGatewayShareResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state transitGatewayShareResourceModel
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

	unlock, ok := common.AcquireLockForID(ctx, state.TgwId.ValueString(), &resp.Diagnostics)
	if !ok {
		return
	}
	defer unlock()

	_, ok = pollTgw(ctx, r.kc, r, state.TgwId.ValueString(), tgwAvailableAndErrorStatuses, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (interface{}, *http.Response, error) {
			httpResp, err := r.kc.ApiClient.TransitGatewayAPI.UnshareTransitGateway(
				ctx, state.TgwId.ValueString(), state.TargetProjectId.ValueString(),
			).XAuthToken(r.kc.XAuthToken).Execute()
			return nil, httpResp, err
		},
	)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			return
		}
		common.AddApiActionError(ctx, r, httpResp, "UnshareTransitGateway", err, &resp.Diagnostics)
		return
	}

	common.PollUntilDeletion(ctx, r, 10*time.Second, &resp.Diagnostics, func(ctx context.Context) (bool, *http.Response, error) {
		listResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
			func() (*tgw.ListTgwSharedProjectsResponse, *http.Response, error) {
				return r.kc.ApiClient.TransitGatewayAPI.ListTgwSharedProjects(ctx, state.TgwId.ValueString()).
					XAuthToken(r.kc.XAuthToken).
					Execute()
			},
		)

		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			return true, httpResp, nil
		}
		if err != nil {
			return false, httpResp, err
		}

		for _, project := range listResp.Projects {
			if project.Id == state.TargetProjectId.ValueString() {
				return false, httpResp, nil
			}
		}

		return true, httpResp, nil
	})
}

func (r *transitGatewayShareResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *transitGatewayShareResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {

	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		common.AddImportFormatError(ctx, r, &resp.Diagnostics,
			"Expected import ID in the format: tgw_id/target_project_id")
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("tgw_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("target_project_id"), parts[1])...)
}

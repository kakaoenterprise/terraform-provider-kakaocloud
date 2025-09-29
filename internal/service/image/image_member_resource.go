// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package image

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-kakaocloud/internal/common"
	"terraform-provider-kakaocloud/internal/docs"
	"terraform-provider-kakaocloud/internal/utils"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/kakaoenterprise/kc-sdk-go/services/image"
)

var (
	_ resource.ResourceWithConfigure   = &imageMemberResource{}
	_ resource.ResourceWithImportState = &imageMemberResource{}
)

func NewImageMemberResource() resource.Resource {
	return &imageMemberResource{}
}

type imageMemberResource struct {
	kc *common.KakaoCloudClient
}

func (r *imageMemberResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_image_member"
}

func (r *imageMemberResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *imageMemberResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: docs.GetResourceDescription("ImageMember"),
		Attributes: utils.MergeAttributes[schema.Attribute](
			imageMemberResourceSchema,
			map[string]schema.Attribute{
				"timeouts": timeouts.AttributesAll(ctx),
			},
		),
	}
}

func (r *imageMemberResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan imageMemberResourceModel
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

	imageId := plan.Id.ValueString()

	currentResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*image.ImageMemberListModel, *http.Response, error) {
			return r.kc.ApiClient.ImageAPI.
				ListImageSharedProjects(ctx, imageId).
				XAuthToken(r.kc.XAuthToken).
				Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "ListImageSharedProjects", err, &resp.Diagnostics)
		return
	}
	for _, m := range currentResp.Members {
		if m.IsShared {
			resp.Diagnostics.AddError(
				"Image already has shared members",
				fmt.Sprintf("Image %q already has shared members. Remove them first before applying.", imageId),
			)
			return
		}
	}

	var sharedMemberIds []string
	diags = plan.SharedMemberIds.ElementsAs(ctx, &sharedMemberIds, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	for _, sharedMemberId := range sharedMemberIds {
		_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
			func() (*image.ResponseImageMemberModel, *http.Response, error) {
				return r.kc.ApiClient.ImageAPI.
					AddImageShare(ctx, imageId, sharedMemberId).
					XAuthToken(r.kc.XAuthToken).
					Execute()
			},
		)
		if err != nil {
			common.AddApiActionError(ctx, r, httpResp, "AddImageShare", err, &resp.Diagnostics)
			return
		}
	}

	imageMemberResp, ok := r.pollUntilAllMembers(ctx, imageId, sharedMemberIds, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	ok = r.mapImageMemberModel(ctx, &plan, imageMemberResp.Members, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *imageMemberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state imageMemberResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	imageMemberResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*image.ImageMemberListModel, *http.Response, error) {
			return r.kc.ApiClient.ImageAPI.
				ListImageSharedProjects(ctx, state.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				Execute()
		},
	)
	if httpResp != nil && httpResp.StatusCode == 404 {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "ListImageSharedProjects", err, &resp.Diagnostics)
		return
	}

	ok := r.mapImageMemberModel(ctx, &state, imageMemberResp.Members, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *imageMemberResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state imageMemberResourceModel
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

	if !plan.SharedMemberIds.Equal(state.SharedMemberIds) {
		imageId := plan.Id.ValueString()

		var planSharedMemberIds []string
		diags = plan.SharedMemberIds.ElementsAs(ctx, &planSharedMemberIds, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		var stateSharedMemberIds []string
		diags = state.SharedMemberIds.ElementsAs(ctx, &stateSharedMemberIds, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		planMap := make(map[string]bool)
		for _, memberId := range planSharedMemberIds {
			planMap[memberId] = true
		}

		stateMap := make(map[string]bool)
		for _, memberId := range stateSharedMemberIds {
			stateMap[memberId] = true
		}

		for _, memberId := range stateSharedMemberIds {
			if !planMap[memberId] {
				_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
					func() (*http.Response, *http.Response, error) {
						httpResp, err := r.kc.ApiClient.ImageAPI.
							RemoveImageShare(ctx, imageId, memberId).
							XAuthToken(r.kc.XAuthToken).
							Execute()
						return nil, httpResp, err
					},
				)
				if err != nil {
					common.AddApiActionError(ctx, r, httpResp, "RemoveImageShare", err, &resp.Diagnostics)
					return
				}
			}
		}

		for _, memberId := range planSharedMemberIds {
			if !stateMap[memberId] {
				_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
					func() (*image.ResponseImageMemberModel, *http.Response, error) {
						return r.kc.ApiClient.ImageAPI.
							AddImageShare(ctx, imageId, memberId).
							XAuthToken(r.kc.XAuthToken).
							Execute()
					},
				)
				if err != nil {
					common.AddApiActionError(ctx, r, httpResp, "AddImageShare", err, &resp.Diagnostics)
					return
				}
			}
		}

		imageMemberResp, ok := r.pollUntilAllMembers(ctx, imageId, planSharedMemberIds, &resp.Diagnostics)
		if !ok || resp.Diagnostics.HasError() {
			return
		}

		ok = r.mapImageMemberModel(ctx, &plan, imageMemberResp.Members, &resp.Diagnostics)
		if !ok || resp.Diagnostics.HasError() {
			return
		}

		diags = resp.State.Set(ctx, &plan)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
}

func (r *imageMemberResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state imageMemberResourceModel
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

	imageId := state.Id.ValueString()

	var sharedMemberIds []string
	diags = state.SharedMemberIds.ElementsAs(ctx, &sharedMemberIds, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	for _, sharedMemberId := range sharedMemberIds {
		_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
			func() (interface{}, *http.Response, error) {
				httpResp, err := r.kc.ApiClient.ImageAPI.
					RemoveImageShare(ctx, imageId, sharedMemberId).
					XAuthToken(r.kc.XAuthToken).
					Execute()
				return nil, httpResp, err
			},
		)
		if err != nil {
			if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
				return
			}
			common.AddApiActionError(ctx, r, httpResp, "RemoveImageShare", err, &resp.Diagnostics)
			return
		}
	}
}

func (r *imageMemberResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *imageMemberResource) convertSetToMembersModel(
	ctx context.Context,
	set types.Set,
) ([]imageMemberMemberModel, diag.Diagnostics) {
	var result []imageMemberMemberModel
	var diags diag.Diagnostics

	for _, elem := range set.Elements() {
		if obj, ok := elem.(types.Object); ok {
			var model imageMemberMemberModel
			elemDiags := obj.As(ctx, &model, basetypes.ObjectAsOptions{})
			diags.Append(elemDiags...)
			result = append(result, model)
		}
	}

	return result, diags
}

func (r *imageMemberResource) pollUntilAllMembers(ctx context.Context, imageId string, memberIds []string, respDiags *diag.Diagnostics) (*image.ImageMemberListModel, bool) {
	memberCount := len(memberIds)

	for {
		imageMemberResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, respDiags,
			func() (*image.ImageMemberListModel, *http.Response, error) {
				return r.kc.ApiClient.ImageAPI.
					ListImageSharedProjects(ctx, imageId).
					XAuthToken(r.kc.XAuthToken).
					Execute()
			},
		)
		if err != nil {
			common.AddApiActionError(ctx, r, httpResp, "ListImageSharedProjects", err, respDiags)
			return nil, false
		}

		sharedCount := 0
		for _, m := range imageMemberResp.Members {
			if m.IsShared {
				sharedCount++
			}
		}
		if sharedCount != memberCount {
			continue
		}

		foundMap := make(map[string]bool)
		for _, member := range imageMemberResp.Members {
			foundMap[member.Id] = true
		}

		allFound := true
		for _, memberId := range memberIds {
			if !foundMap[memberId] {
				allFound = false
				break
			}
		}

		if allFound {
			return imageMemberResp, true
		}
		time.Sleep(2 * time.Second)
	}
}

func (r *imageMemberResource) mapImageMemberModel(
	ctx context.Context,
	model *imageMemberResourceModel,
	membersResult []image.BcsImageV1ApiListImageSharedProjectsModelImageMemberModel,
	respDiags *diag.Diagnostics,
) bool {
	ok := mapImageMemberModel(ctx, &model.imageMemberBaseModel, membersResult, respDiags)
	if !ok || respDiags.HasError() {
		return false
	}

	sharedMemberIds := make([]attr.Value, 0, len(membersResult))
	for _, member := range membersResult {
		if member.IsShared {
			sharedMemberIds = append(sharedMemberIds, types.StringValue(member.Id))
		}
	}
	val, diags := types.SetValue(types.StringType, sharedMemberIds)
	respDiags.Append(diags...)
	if respDiags.HasError() {
		return false
	}
	model.SharedMemberIds = val

	return true
}

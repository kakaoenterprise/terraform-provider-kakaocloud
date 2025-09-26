// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package bcs

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-kakaocloud/internal/common"
	"terraform-provider-kakaocloud/internal/docs"
	. "terraform-provider-kakaocloud/internal/utils"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/kakaoenterprise/kc-sdk-go/services/bcs"
)

var (
	_ resource.Resource                = &keypairResource{}
	_ resource.ResourceWithConfigure   = &keypairResource{}
	_ resource.ResourceWithImportState = &keypairResource{}
)

func NewKeypairResource() resource.Resource { return &keypairResource{} }

type keypairResource struct {
	kc *common.KakaoCloudClient
}

func (r *keypairResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_keypair"
}

func (r *keypairResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: docs.GetResourceDescription("Keypair"),
		Attributes: MergeResourceSchemaAttributes(
			keypairResourceSchemaAttributes,
			map[string]schema.Attribute{
				"timeouts": timeouts.AttributesAll(ctx),
			},
		),
	}
}

func (r *keypairResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan keypairResourceModel
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

	createReq := bcs.CreateKeypairModel{
		Name: plan.Name.ValueString(),
	}

	if !plan.PublicKey.IsNull() && !plan.PublicKey.IsUnknown() {
		createReq.SetPublicKey(plan.PublicKey.ValueString())
	}

	body := bcs.BodyCreateKeypair{
		Keypair: createReq,
	}

	respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*bcs.BcsInstanceV1ApiCreateKeypairModelResponseKeypairModel, *http.Response, error) {
			return r.kc.ApiClient.KeypairAPI.CreateKeypair(ctx).
				XAuthToken(r.kc.XAuthToken).
				BodyCreateKeypair(body).
				Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "CreateKeypair", err, &resp.Diagnostics)
		return
	}

	pollInterval := 1 * time.Second
	result, ok := common.PollUntilResultWithTimeout(
		ctx, r, pollInterval, &timeout,
		[]string{"found"}, &resp.Diagnostics,
		func(c context.Context) (*bcs.BcsInstanceV1ApiGetKeypairModelResponseKeypairModel, *http.Response, error) {
			return r.kc.ApiClient.KeypairAPI.
				GetKeypair(c, plan.Name.ValueString()).
				XAuthToken(r.kc.XAuthToken).Execute()
		},
		func(_ *bcs.BcsInstanceV1ApiGetKeypairModelResponseKeypairModel) string {
			return "found"
		},
	)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	ok = r.mapKeypair(ctx, &plan, &result.Keypair, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	plan.PrivateKey = ConvertNullableString(respModel.Keypair.PrivateKey)
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *keypairResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state keypairResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*bcs.BcsInstanceV1ApiGetKeypairModelResponseKeypairModel, *http.Response, error) {
			return r.kc.ApiClient.KeypairAPI.GetKeypair(ctx, state.Name.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				Execute()
		},
	)

	if httpResp != nil && httpResp.StatusCode == 404 {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "GetKeypair", err, &resp.Diagnostics)
		return
	}

	result := respModel.Keypair
	ok := r.mapKeypair(ctx, &state, &result, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *keypairResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

}

func (r *keypairResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state keypairResourceModel
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

	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (interface{}, *http.Response, error) {
			httpResp, err := r.kc.ApiClient.KeypairAPI.DeleteKeypair(ctx, state.Name.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				Execute()
			return nil, httpResp, err
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "DeleteKeypair", err, &resp.Diagnostics)
		return
	}

	common.PollUntilDeletion(ctx, r, 2*time.Second, &resp.Diagnostics, func(ctx context.Context) (bool, *http.Response, error) {
		_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
			func() (*bcs.BcsInstanceV1ApiGetKeypairModelResponseKeypairModel, *http.Response, error) {
				_, httpResp, err := r.kc.ApiClient.KeypairAPI.
					GetKeypair(ctx, state.Name.ValueString()).
					XAuthToken(r.kc.XAuthToken).
					Execute()
				return nil, httpResp, err
			},
		)
		return false, httpResp, err
	})
}

func (r *keypairResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *keypairResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {

	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func (r *keypairResource) mapKeypair(
	ctx context.Context,
	model *keypairResourceModel,
	keypairResult *bcs.BcsInstanceV1ApiGetKeypairModelKeypairModel,
	respDiags *diag.Diagnostics,
) bool {
	mapKeypairBaseModel(ctx, &model.keypairBaseModel, keypairResult, respDiags)

	if respDiags.HasError() {
		return false
	}
	return true
}

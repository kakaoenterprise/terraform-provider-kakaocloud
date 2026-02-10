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
	"github.com/kakaoenterprise/kc-sdk-go/services/tgw"
)

var (
	_ resource.ResourceWithConfigure   = &transitGatewayAttachmentResource{}
	_ resource.ResourceWithImportState = &transitGatewayAttachmentResource{}
)

func NewTransitGatewayAttachmentResource() resource.Resource {
	return &transitGatewayAttachmentResource{}
}

type transitGatewayAttachmentResource struct {
	kc *common.KakaoCloudClient
}

func (r *transitGatewayAttachmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_transit_gateway_attachment"
}

func (r *transitGatewayAttachmentResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: utils.MergeResourceSchemaAttributes(
			transitGatewayAttachmentResourceSchemaAttributes,
			map[string]schema.Attribute{
				"timeouts": timeouts.AttributesAll(ctx),
			},
		),
	}
}

func (r *transitGatewayAttachmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan transitGatewayAttachmentResourceModel
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

	_, ok := pollTgw(ctx, r.kc, r, plan.TgwId.ValueString(), []string{common.TgwStatusActive, common.TgwStatusError, common.TgwStatusInUse, common.TgwStatusInactive, common.TgwStatusAvaliable}, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	var subnetIds []string
	diags = plan.SubnetIds.ElementsAs(ctx, &subnetIds, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := tgw.BnsTgwV1ApiCreateTgwAttachmentModelTgwAttachmentRequestModel{
		TgwId:     plan.TgwId.ValueString(),
		VpcId:     plan.ResourceId.ValueString(),
		SubnetIds: subnetIds,
	}

	body := tgw.CreateTgwAttachmentRequestModel{
		Attachment: createReq,
	}

	respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*tgw.BnsTgwV1ApiCreateTgwAttachmentModelCreateTgwAttachmentResponseModel, *http.Response, error) {
			return r.kc.ApiClient.AttachmentsAPI.CreateTgwAttachment(ctx).
				XAuthToken(r.kc.XAuthToken).
				CreateTgwAttachmentRequestModel(body).
				Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "CreateTgwAttachment", err, &resp.Diagnostics)
		return
	}

	plan.Id = types.StringValue(respModel.Attachment.Id)

	result, ok := common.PollUntilResult(
		ctx, r, 5*time.Second, "transit gateway attachment", plan.Id.ValueString(), []string{common.TgwStatusActive, common.TgwStatusError, common.TgwStatusInUse, common.TgwStatusInactive, common.TgwStatusAvaliable, common.TgwStatusPendingApprove}, &resp.Diagnostics,
		func(ctx context.Context) (*tgw.BnsTgwV1ApiGetTgwAttachmentModelTgwAttachmentResponseModel, *http.Response, error) {
			respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
				func() (*tgw.GetTgwAttachmentResponseModel, *http.Response, error) {
					return r.kc.ApiClient.AttachmentsAPI.
						GetTgwAttachment(ctx, plan.Id.ValueString()).
						XAuthToken(r.kc.XAuthToken).
						Execute()
				},
			)
			if err != nil {
				return nil, httpResp, err
			}
			return &respModel.Attachment, httpResp, nil
		},
		func(v *tgw.BnsTgwV1ApiGetTgwAttachmentModelTgwAttachmentResponseModel) string {
			if v.ProvisioningStatus.IsSet() {
				return string(*v.ProvisioningStatus.Get())
			}
			return ""
		},
	)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	common.CheckResourceAvailableStatus(ctx, r, (*string)(result.ProvisioningStatus.Get()), []string{common.TgwStatusActive, common.TgwStatusInUse, common.TgwStatusInactive, common.TgwStatusAvaliable, common.TgwStatusPendingApprove}, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	ok = mapTransitGatewayAttachmentModelFromGet(ctx, &plan.transitGatewayAttachmentBaseModel, result, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	subnetIds = make([]string, 0, len(result.Resources))
	for _, res := range result.Resources {
		if res.Id.IsSet() && res.Id.Get() != nil {
			subnetIds = append(subnetIds, *res.Id.Get())
		}
	}
	plan.SubnetIds, diags = types.SetValueFrom(ctx, types.StringType, subnetIds)
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

func (r *transitGatewayAttachmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state transitGatewayAttachmentResourceModel
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
		func() (*tgw.GetTgwAttachmentResponseModel, *http.Response, error) {
			return r.kc.ApiClient.AttachmentsAPI.GetTgwAttachment(ctx, state.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				Execute()
		},
	)

	if httpResp != nil && httpResp.StatusCode == 404 {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "GetTgwAttachment", err, &resp.Diagnostics)
		return
	}

	ok := mapTransitGatewayAttachmentModelFromGet(ctx, &state.transitGatewayAttachmentBaseModel, &respModel.Attachment, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	subnetIds := make([]string, 0, len(respModel.Attachment.Resources))
	for _, res := range respModel.Attachment.Resources {
		if res.Id.IsSet() && res.Id.Get() != nil {
			subnetIds = append(subnetIds, *res.Id.Get())
		}
	}
	state.SubnetIds, diags = types.SetValueFrom(ctx, types.StringType, subnetIds)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.TgwId.IsNull() {
		state.TgwId = utils.ConvertNullableString(respModel.Attachment.Tgw.Id)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *transitGatewayAttachmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state transitGatewayAttachmentResourceModel
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

	_, ok := pollTgw(ctx, r.kc, r, plan.TgwId.ValueString(), []string{common.TgwStatusActive, common.TgwStatusError, common.TgwStatusInUse, common.TgwStatusInactive, common.TgwStatusAvaliable}, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	if !plan.SubnetIds.Equal(state.SubnetIds) {
		var subnetIds []string
		diags = plan.SubnetIds.ElementsAs(ctx, &subnetIds, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		updateReq := tgw.BnsTgwV1ApiUpdateTgwAttachmentModelTgwAttachmentRequestModel{
			SubnetIds: subnetIds,
		}

		body := tgw.UpdateTgwAttachmentRequestModel{
			Attachment: updateReq,
		}

		_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
			func() (*tgw.BnsTgwV1ApiUpdateTgwAttachmentModelCreateTgwAttachmentResponseModel, *http.Response, error) {
				return r.kc.ApiClient.AttachmentsAPI.UpdateTgwAttachment(ctx, plan.Id.ValueString()).
					XAuthToken(r.kc.XAuthToken).
					UpdateTgwAttachmentRequestModel(body).
					Execute()
			},
		)
		if err != nil {
			common.AddApiActionError(ctx, r, httpResp, "UpdateTgwAttachment", err, &resp.Diagnostics)
			return
		}

		time.Sleep(3 * time.Second)

		result, ok := common.PollUntilResult(
			ctx, r, 5*time.Second, "transit gateway attachment", plan.Id.ValueString(), []string{common.TgwStatusActive, common.TgwStatusError, common.TgwStatusInUse, common.TgwStatusInactive, common.TgwStatusAvaliable, common.TgwStatusPendingApprove}, &resp.Diagnostics,
			func(ctx context.Context) (*tgw.BnsTgwV1ApiGetTgwAttachmentModelTgwAttachmentResponseModel, *http.Response, error) {
				respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
					func() (*tgw.GetTgwAttachmentResponseModel, *http.Response, error) {
						return r.kc.ApiClient.AttachmentsAPI.
							GetTgwAttachment(ctx, plan.Id.ValueString()).
							XAuthToken(r.kc.XAuthToken).
							Execute()
					},
				)
				if err != nil {
					return nil, httpResp, err
				}
				return &respModel.Attachment, httpResp, nil
			},
			func(v *tgw.BnsTgwV1ApiGetTgwAttachmentModelTgwAttachmentResponseModel) string {
				if v.ProvisioningStatus.IsSet() {
					return string(*v.ProvisioningStatus.Get())
				}
				return ""
			},
		)
		if !ok || resp.Diagnostics.HasError() {
			return
		}

		common.CheckResourceAvailableStatus(ctx, r, (*string)(result.ProvisioningStatus.Get()), []string{common.TgwStatusActive, common.TgwStatusInUse, common.TgwStatusInactive, common.TgwStatusAvaliable, common.TgwStatusPendingApprove}, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}

		ok = mapTransitGatewayAttachmentModelFromGet(ctx, &state.transitGatewayAttachmentBaseModel, result, &resp.Diagnostics)
		if !ok || resp.Diagnostics.HasError() {
			return
		}

		actualSubnetIds := make([]string, 0, len(result.Resources))
		for _, res := range result.Resources {
			if res.Id.IsSet() && res.Id.Get() != nil {
				actualSubnetIds = append(actualSubnetIds, *res.Id.Get())
			}
		}
		state.SubnetIds, diags = types.SetValueFrom(ctx, types.StringType, actualSubnetIds)
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

func (r *transitGatewayAttachmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state transitGatewayAttachmentResourceModel
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

	_, ok := pollTgw(ctx, r.kc, r, state.TgwId.ValueString(), []string{common.TgwStatusActive, common.TgwStatusError, common.TgwStatusInUse, common.TgwStatusInactive, common.TgwStatusAvaliable, common.TgwStatusDeleting}, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (interface{}, *http.Response, error) {
			httpResp, err := r.kc.ApiClient.AttachmentsAPI.DeleteTgwAttachment(ctx, state.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				Execute()
			return nil, httpResp, err
		},
	)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			return
		}
		common.AddApiActionError(ctx, r, httpResp, "DeleteTgwAttachment", err, &resp.Diagnostics)
		return
	}

	common.PollUntilDeletion(ctx, r, 10*time.Second, &resp.Diagnostics, func(ctx context.Context) (bool, *http.Response, error) {
		_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
			func() (interface{}, *http.Response, error) {
				_, httpResp, err := r.kc.ApiClient.AttachmentsAPI.
					GetTgwAttachment(ctx, state.Id.ValueString()).
					XAuthToken(r.kc.XAuthToken).
					Execute()
				return nil, httpResp, err
			},
		)
		return false, httpResp, err
	})
}

func (r *transitGatewayAttachmentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *transitGatewayAttachmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

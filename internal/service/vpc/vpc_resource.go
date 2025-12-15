// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package vpc

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
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/kakaoenterprise/kc-sdk-go/services/vpc"
)

var (
	_ resource.ResourceWithConfigure   = &vpcResource{}
	_ resource.ResourceWithImportState = &vpcResource{}
	_ resource.ResourceWithModifyPlan  = &vpcResource{}
)

func NewVpcResource() resource.Resource {
	return &vpcResource{}
}

type vpcResource struct {
	kc *common.KakaoCloudClient
}

func (r *vpcResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpc"
}

func (r *vpcResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: utils.MergeResourceSchemaAttributes(
			vpcResourceSchemaAttributes,
			map[string]schema.Attribute{
				"timeouts": timeouts.AttributesAll(ctx),
			},
		),
	}
}

func (r *vpcResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan vpcResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	timeout, diags := plan.Timeouts.Create(ctx, common.LongCreateTimeout)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var subnetModel vpcSubnetModel
	diags = plan.Subnet.As(ctx, &subnetModel, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := vpc.CreateVPCModel{
		Name:      plan.Name.ValueString(),
		CidrBlock: plan.CidrBlock.ValueString(),
		Subnet: &vpc.MainSubnet{
			CidrBlock:        subnetModel.CidrBlock.ValueString(),
			AvailabilityZone: vpc.AvailabilityZone(subnetModel.AvailabilityZone.ValueString()),
		},
	}

	body := vpc.BodyCreateVpc{
		Vpc: createReq,
	}

	respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*vpc.BnsVpcV1ApiCreateVpcModelResponseVPCModel, *http.Response, error) {
			return r.kc.ApiClient.VPCAPI.CreateVpc(ctx).
				XAuthToken(r.kc.XAuthToken).BodyCreateVpc(body).Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "CreateVpc", err, &resp.Diagnostics)
		return
	}

	plan.Id = types.StringValue(respModel.Vpc.Id)

	result, ok := common.PollUntilResult(
		ctx,
		r,
		10*time.Second,
		"vpc",
		plan.Id.ValueString(),
		[]string{common.VpcProvisioningStatusActive, common.VpcProvisioningStatusError},
		&resp.Diagnostics,
		func(ctx context.Context) (*vpc.BnsVpcV1ApiGetVpcModelVpcModel, *http.Response, error) {
			respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
				func() (*vpc.BnsVpcV1ApiGetVpcModelResponseVPCModel, *http.Response, error) {
					return r.kc.ApiClient.VPCAPI.
						GetVpc(ctx, plan.Id.ValueString()).
						XAuthToken(r.kc.XAuthToken).
						Execute()
				},
			)
			if err != nil {
				return nil, httpResp, err
			}
			return &respModel.Vpc, httpResp, nil
		},
		func(v *vpc.BnsVpcV1ApiGetVpcModelVpcModel) string {
			return string(*v.ProvisioningStatus.Get())
		},
	)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	common.CheckResourceAvailableStatus(ctx, r, (*string)(result.ProvisioningStatus.Get()), []string{common.VpcProvisioningStatusActive}, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	ok = mapVpcBaseModel(ctx, &plan.vpcBaseModel, result, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

func (r *vpcResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state vpcResourceModel
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
		func() (*vpc.BnsVpcV1ApiGetVpcModelResponseVPCModel, *http.Response, error) {
			return r.kc.ApiClient.VPCAPI.GetVpc(ctx, state.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).Execute()
		},
	)

	if httpResp != nil && httpResp.StatusCode == 404 {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "GetVpc", err, &resp.Diagnostics)
		return
	}

	vpcResult := respModel.Vpc
	ok := mapVpcBaseModel(ctx, &state.vpcBaseModel, &vpcResult, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *vpcResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state vpcResourceModel
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

	if plan.Name != state.Name {
		editReq := vpc.EditVPCModel{
			Name: plan.Name.ValueString(),
		}

		body := *vpc.NewBodyPutBnsVpc(editReq)

		respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
			func() (*vpc.BnsVpcV1ApiUpdateVpcModelResponseVPCModel, *http.Response, error) {
				return r.kc.ApiClient.VPCAPI.PutBnsVpc(ctx, plan.Id.ValueString()).
					XAuthToken(r.kc.XAuthToken).
					BodyPutBnsVpc(body).
					Execute()
			},
		)
		if err != nil {
			common.AddApiActionError(ctx, r, httpResp, "PutBnsVpc", err, &resp.Diagnostics)
			return
		}

		state.Name = types.StringValue(respModel.Vpc.Name)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *vpcResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state vpcResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	timeout, diags := state.Timeouts.Delete(ctx, common.LongDeleteTimeout)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (interface{}, *http.Response, error) {
			httpResp, err := r.kc.ApiClient.VPCAPI.DeleteVpc(ctx, state.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				Execute()
			return nil, httpResp, err
		},
	)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			return
		}
		common.AddApiActionError(ctx, r, httpResp, "DeleteVpc", err, &resp.Diagnostics)
		return
	}

	common.PollUntilDeletion(ctx, r, 10*time.Second, &resp.Diagnostics, func(ctx context.Context) (bool, *http.Response, error) {
		_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
			func() (interface{}, *http.Response, error) {
				_, httpResp, err := r.kc.ApiClient.VPCAPI.
					GetVpc(ctx, state.Id.ValueString()).
					XAuthToken(r.kc.XAuthToken).
					Execute()
				return nil, httpResp, err
			},
		)
		return false, httpResp, err
	})
}

func (r *vpcResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {

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

func (r *vpcResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {

	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *vpcResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config vpcResourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.validateAvailabilityZoneConfig(ctx, config, resp)
	r.validateSubnetCidrBlockConfig(ctx, config, &resp.Diagnostics)
}

func (r *vpcResource) validateAvailabilityZoneConfig(ctx context.Context, config vpcResourceModel, resp *resource.ValidateConfigResponse) {
	if !config.Subnet.IsNull() {
		var subnetModel vpcSubnetModel
		diags := config.Subnet.As(ctx, &subnetModel, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		common.ValidateAvailabilityZone(
			path.Root("availability_zone"),
			subnetModel.AvailabilityZone,
			r.kc,
			&resp.Diagnostics,
		)
	}
}

func (r *vpcResource) validateSubnetCidrBlockConfig(ctx context.Context, config vpcResourceModel, respDiags *diag.Diagnostics) {
	if !config.Subnet.IsNull() {
		var subnetModel vpcSubnetModel
		diags := config.Subnet.As(ctx, &subnetModel, basetypes.ObjectAsOptions{})
		respDiags.Append(diags...)
		if respDiags.HasError() {
			return
		}

		if !config.CidrBlock.IsUnknown() && !subnetModel.CidrBlock.IsUnknown() {
			common.CidrContainValidator(subnetModel.CidrBlock.ValueString(), config.CidrBlock.ValueString(), "subnet", "vpc", respDiags)
		}
	}
}

func (r *vpcResource) ModifyPlan(
	ctx context.Context,
	req resource.ModifyPlanRequest,
	resp *resource.ModifyPlanResponse,
) {
	var plan, state *vpcResourceModel

	planDiags := req.Plan.Get(ctx, &plan)
	stateDiags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(planDiags...)
	resp.Diagnostics.Append(stateDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if req.Plan.Raw.IsNull() {
		return
	}

	if req.State.Raw.IsNull() && !req.Plan.Raw.IsNull() {
		if plan.Subnet.IsNull() {
			common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
				"Missing required attribute 'subnet' on create.")
		}
	}
}

func checkVpcStatus(
	ctx context.Context,
	r resource.Resource,
	kc *common.KakaoCloudClient,
	vpcId string,
	respDiags *diag.Diagnostics,
) bool {
	interval := 1 * time.Second
	result, ok := common.PollUntilResult(
		ctx,
		r,
		interval,
		"vpc",
		vpcId,
		[]string{common.VpcProvisioningStatusActive, common.VpcProvisioningStatusError, common.VpcProvisioningStatusDeleting},
		respDiags,
		func(ctx context.Context) (*vpc.BnsVpcV1ApiGetVpcModelVpcModel, *http.Response, error) {
			respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, kc, respDiags,
				func() (*vpc.BnsVpcV1ApiGetVpcModelResponseVPCModel, *http.Response, error) {
					return kc.ApiClient.VPCAPI.
						GetVpc(ctx, vpcId).
						XAuthToken(kc.XAuthToken).
						Execute()
				},
			)
			if err != nil {
				return nil, httpResp, err
			}
			return &respModel.Vpc, httpResp, nil
		},
		func(v *vpc.BnsVpcV1ApiGetVpcModelVpcModel) string {
			return string(*v.ProvisioningStatus.Get())
		},
	)
	if !ok {
		common.AddGeneralError(ctx, r, respDiags,
			fmt.Sprintf("VPC did not reach the status '%v'.", common.VpcProvisioningStatusActive),
		)
		return false
	}
	status := *result.ProvisioningStatus.Get()
	if status != common.VpcProvisioningStatusActive {
		common.AddGeneralError(ctx, r, respDiags,
			fmt.Sprintf("VPC status is '%v'.", status),
		)
	}
	if respDiags.HasError() {
		return false
	}
	return true
}

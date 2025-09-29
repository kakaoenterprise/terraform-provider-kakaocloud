// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package vpc

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-kakaocloud/internal/common"
	"terraform-provider-kakaocloud/internal/docs"
	"terraform-provider-kakaocloud/internal/utils"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kakaoenterprise/kc-sdk-go/services/vpc"
)

var (
	_ resource.ResourceWithConfigure   = &subnetResource{}
	_ resource.ResourceWithImportState = &subnetResource{}
)

func NewSubnetResource() resource.Resource {
	return &subnetResource{}
}

type subnetResource struct {
	kc *common.KakaoCloudClient
}

func (r *subnetResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subnet"
}

func (r *subnetResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: docs.GetResourceDescription("Subnet"),
		Attributes: utils.MergeResourceSchemaAttributes(
			subnetResourceSchemaAttributes,
			map[string]schema.Attribute{
				"timeouts": timeouts.AttributesAll(ctx),
			},
		),
	}
}

func (r *subnetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan subnetResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	mutex := common.LockForID(plan.VpcId.ValueString())
	mutex.Lock()
	defer mutex.Unlock()

	timeout, diags := plan.Timeouts.Create(ctx, common.DefaultCreateTimeout)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	createReq := vpc.CreateSubnetModel{
		Name:             plan.Name.ValueString(),
		VpcId:            plan.VpcId.ValueString(),
		CidrBlock:        plan.CidrBlock.ValueString(),
		AvailabilityZone: vpc.AvailabilityZone(plan.AvailabilityZone.ValueString()),
	}

	body := *vpc.NewBodyCreateSubnet(createReq)

	ok := checkVpcStatus(ctx, r, r.kc, plan.VpcId.ValueString(), &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*vpc.BnsVpcV1ApiCreateSubnetModelResponseSubnetModel, *http.Response, error) {
			return r.kc.ApiClient.VPCSubnetAPI.CreateSubnet(ctx).
				XAuthToken(r.kc.XAuthToken).BodyCreateSubnet(body).Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "CreateSubnet", err, &resp.Diagnostics)
		return
	}

	plan.Id = types.StringValue(respModel.Subnet.Id)

	result, ok := common.PollUntilResult(
		ctx,
		r,
		5*time.Second,
		[]string{common.VpcProvisioningStatusActive, common.VpcProvisioningStatusError},
		&resp.Diagnostics,
		func(ctx context.Context) (*vpc.BnsVpcV1ApiGetSubnetModelSubnetModel, *http.Response, error) {
			respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
				func() (*vpc.BnsVpcV1ApiGetSubnetModelResponseSubnetModel, *http.Response, error) {
					return r.kc.ApiClient.VPCSubnetAPI.
						GetSubnet(ctx, plan.Id.ValueString()).
						XAuthToken(r.kc.XAuthToken).
						Execute()
				},
			)
			if err != nil {
				return nil, httpResp, err
			}
			return &respModel.Subnet, httpResp, nil
		},
		func(v *vpc.BnsVpcV1ApiGetSubnetModelSubnetModel) string {
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

	if !plan.RouteTableId.IsNull() && !plan.RouteTableId.IsUnknown() {
		sourceRouteTable := result.RouteTableId.Get()
		result, ok = r.setAssociation(ctx, *sourceRouteTable, plan, &resp.Diagnostics)
	}

	ok = mapSubnetBaseModel(&plan.subnetBaseModel, result, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

func (r *subnetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state subnetResourceModel
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
		func() (*vpc.BnsVpcV1ApiGetSubnetModelResponseSubnetModel, *http.Response, error) {
			return r.kc.ApiClient.VPCSubnetAPI.GetSubnet(ctx, state.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).Execute()
		},
	)

	if httpResp != nil && httpResp.StatusCode == 404 {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "GetSubnet", err, &resp.Diagnostics)
		return
	}

	subnetResult := respModel.Subnet
	ok := mapSubnetBaseModel(&state.subnetBaseModel, &subnetResult, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *subnetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state subnetResourceModel
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

	mutex := common.LockForID(plan.VpcId.ValueString())
	mutex.Lock()
	defer mutex.Unlock()

	timeout, diags := plan.Timeouts.Update(ctx, common.DefaultUpdateTimeout)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if plan.Name != state.Name {
		editReq := vpc.EditSubnetModel{
			Name: plan.Name.ValueString(),
		}

		body := *vpc.NewBodyUpdateSubnet(editReq)

		_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
			func() (*vpc.BnsVpcV1ApiUpdateSubnetModelResponseSubnetModel, *http.Response, error) {
				return r.kc.ApiClient.VPCSubnetAPI.UpdateSubnet(ctx, plan.Id.ValueString()).
					XAuthToken(r.kc.XAuthToken).
					BodyUpdateSubnet(body).
					Execute()
			},
		)
		if err != nil {
			common.AddApiActionError(ctx, r, httpResp, "UpdateSubnet", err, &resp.Diagnostics)
			return
		}
	}

	if !plan.RouteTableId.IsUnknown() && !plan.RouteTableId.Equal(state.RouteTableId) {
		sourceRouteTable := state.RouteTableId.ValueString()
		result, ok := r.setAssociation(ctx, sourceRouteTable, plan, &resp.Diagnostics)
		if !ok || resp.Diagnostics.HasError() {
			return
		}
		ok = mapSubnetBaseModel(&plan.subnetBaseModel, result, &resp.Diagnostics)
		if !ok || resp.Diagnostics.HasError() {
			return
		}
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *subnetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state subnetResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	mutex := common.LockForID(state.VpcId.ValueString())
	mutex.Lock()
	defer mutex.Unlock()

	timeout, diags := state.Timeouts.Delete(ctx, common.DefaultDeleteTimeout)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ok := checkVpcStatus(ctx, r, r.kc, state.VpcId.ValueString(), &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	ok = checkSubnetStatus(ctx, r, r.kc, state.Id.ValueString(), &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (interface{}, *http.Response, error) {
			httpResp, err := r.kc.ApiClient.VPCSubnetAPI.DeleteSubnet(ctx, state.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				Execute()
			return nil, httpResp, err
		},
	)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			return
		}
		common.AddApiActionError(ctx, r, httpResp, "DeleteSubnet", err, &resp.Diagnostics)
		return
	}

	common.PollUntilDeletion(ctx, r, 5*time.Second, &resp.Diagnostics, func(ctx context.Context) (bool, *http.Response, error) {
		_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
			func() (*vpc.BnsVpcV1ApiGetSubnetModelResponseSubnetModel, *http.Response, error) {
				_, httpResp, err := r.kc.ApiClient.VPCSubnetAPI.
					GetSubnet(ctx, state.Id.ValueString()).
					XAuthToken(r.kc.XAuthToken).
					Execute()
				return nil, httpResp, err
			},
		)
		return false, httpResp, err
	})
}

func (r *subnetResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {

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

func (r *subnetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {

	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *subnetResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config subnetResourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.validateAvailabilityZoneConfig(config, resp)
}

func (r *subnetResource) validateAvailabilityZoneConfig(config subnetResourceModel, resp *resource.ValidateConfigResponse) {
	common.ValidateAvailabilityZone(
		path.Root("availability_zone"),
		config.AvailabilityZone,
		r.kc,
		&resp.Diagnostics,
	)
}

func (r *subnetResource) setAssociation(
	ctx context.Context,
	sourceRouteTableId string,
	plan subnetResourceModel,
	respDiags *diag.Diagnostics,
) (*vpc.BnsVpcV1ApiGetSubnetModelSubnetModel, bool) {
	subnetId := plan.Id.ValueString()

	routeTableResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, respDiags,
		func() (*vpc.BnsVpcV1ApiGetRouteTableModelResponseRouteTableModel, *http.Response, error) {
			return r.kc.ApiClient.VPCRouteTableAPI.GetRouteTable(ctx, sourceRouteTableId).
				XAuthToken(r.kc.XAuthToken).Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "GetRouteTable", err, respDiags)
		return nil, false
	}

	associations := routeTableResp.VpcRouteTable.Associations

	ok := checkVpcStatus(ctx, r, r.kc, plan.VpcId.ValueString(), respDiags)
	if !ok || respDiags.HasError() {
		return nil, false
	}

	ok = checkSubnetStatus(ctx, r, r.kc, subnetId, respDiags)
	if !ok || respDiags.HasError() {
		return nil, false
	}

	var associationId *string

	for _, association := range associations {
		if association.SubnetId.Get() != nil && *association.SubnetId.Get() == subnetId {
			associationId = &association.Id
			break
		}
	}

	if associationId == nil {
		common.AddGeneralError(ctx, r, respDiags,
			fmt.Sprintf("Source route table does not have a subnet_id '%v'.", subnetId))
		return nil, false
	}

	req := vpc.EditAssociationModel{
		TargetRouteTableId: plan.RouteTableId.ValueString(),
	}

	body := *vpc.NewBodyUpdateRouteTableAssociation(req)

	_, httpResp, err = common.ExecuteWithRetryAndAuth(ctx, r.kc, respDiags,
		func() (*vpc.ResponseRouteTableAssociationModel, *http.Response, error) {
			return r.kc.ApiClient.VPCRouteTableAssociationAPI.UpdateRouteTableAssociation(ctx, sourceRouteTableId, *associationId).
				XAuthToken(r.kc.XAuthToken).BodyUpdateRouteTableAssociation(body).Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "UpdateRouteTableAssociation", err, respDiags)
		return nil, false
	}

	result, ok := common.PollUntilResult(
		ctx,
		r,
		5*time.Second,
		[]string{common.VpcProvisioningStatusActive, common.VpcProvisioningStatusError},
		respDiags,
		func(ctx context.Context) (*vpc.BnsVpcV1ApiGetSubnetModelSubnetModel, *http.Response, error) {
			respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, respDiags,
				func() (*vpc.BnsVpcV1ApiGetSubnetModelResponseSubnetModel, *http.Response, error) {
					return r.kc.ApiClient.VPCSubnetAPI.
						GetSubnet(ctx, plan.Id.ValueString()).
						XAuthToken(r.kc.XAuthToken).
						Execute()
				},
			)
			if err != nil {
				return nil, httpResp, err
			}
			return &respModel.Subnet, httpResp, nil
		},
		func(v *vpc.BnsVpcV1ApiGetSubnetModelSubnetModel) string {
			return string(*v.ProvisioningStatus.Get())
		},
	)
	if !ok || respDiags.HasError() {
		return nil, false
	}
	common.CheckResourceAvailableStatus(ctx, r, (*string)(result.ProvisioningStatus.Get()), []string{common.VpcProvisioningStatusActive}, respDiags)
	if respDiags.HasError() {
		return nil, false
	}

	return result, true
}

func checkSubnetStatus(
	ctx context.Context,
	r resource.Resource,
	kc *common.KakaoCloudClient,
	subnetId string,
	respDiags *diag.Diagnostics,
) bool {
	interval := 1 * time.Second
	result, ok := common.PollUntilResult(
		ctx,
		r,
		interval,
		[]string{common.VpcProvisioningStatusActive, common.VpcProvisioningStatusError},
		respDiags,
		func(ctx context.Context) (*vpc.BnsVpcV1ApiGetSubnetModelSubnetModel, *http.Response, error) {
			respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, kc, respDiags,
				func() (*vpc.BnsVpcV1ApiGetSubnetModelResponseSubnetModel, *http.Response, error) {
					return kc.ApiClient.VPCSubnetAPI.
						GetSubnet(ctx, subnetId).
						XAuthToken(kc.XAuthToken).
						Execute()
				},
			)
			if err != nil {
				return nil, httpResp, err
			}
			return &respModel.Subnet, httpResp, nil
		},
		func(v *vpc.BnsVpcV1ApiGetSubnetModelSubnetModel) string {
			return string(*v.ProvisioningStatus.Get())
		},
	)
	if !ok {
		common.AddGeneralError(ctx, r, respDiags,
			fmt.Sprintf("Subnet did not reach the status '%v'.", common.VpcProvisioningStatusActive),
		)
		return false
	}
	status := *result.ProvisioningStatus.Get()
	if status == common.VpcProvisioningStatusError {
		common.AddGeneralError(ctx, r, respDiags,
			fmt.Sprintf("Subnet status is '%v'.", common.VpcProvisioningStatusActive),
		)
	}
	if respDiags.HasError() {
		return false
	}
	return true
}

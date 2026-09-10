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
	"github.com/kakaoenterprise/kc-sdk-go/v2/services/vpc"
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

	timeout, diags := plan.Timeouts.Create(ctx, common.DefaultCreateTimeout)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	unlock, ok := common.AcquireLockForID(ctx, plan.VpcId.ValueString(), &resp.Diagnostics)
	if !ok {
		return
	}
	defer unlock()

	createReq := vpc.CreateSubnet{
		Name:             plan.Name.ValueString(),
		VpcId:            plan.VpcId.ValueString(),
		CidrBlock:        plan.CidrBlock.ValueString(),
		AvailabilityZone: vpc.AvailabilityZone(plan.AvailabilityZone.ValueString()),
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		createReq.SetDescription(plan.Description.ValueString())
	}

	body := *vpc.NewCreateSubnetRequest(createReq)

	ok = pollVpcStatus(ctx, r, r.kc, plan.VpcId.ValueString(), &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*vpc.CreateSubnetResponse, *http.Response, error) {
			return r.kc.ApiClient.VPCSubnetAPI.CreateSubnet(ctx).
				XAuthToken(r.kc.XAuthToken).CreateSubnetRequest(body).Execute()
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
		"subnet",
		plan.Id.ValueString(),
		[]string{common.VpcProvisioningStatusActive, common.VpcProvisioningStatusError},
		&resp.Diagnostics,
		func(ctx context.Context) (*vpc.Subnet, *http.Response, error) {
			respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
				func() (*vpc.GetSubnetResponse, *http.Response, error) {
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
		func(v *vpc.Subnet) string {
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
		sourceRouteTableID := *result.RouteTableId.Get()

		if sourceRouteTableID != plan.RouteTableId.ValueString() {
			result, ok = r.setAssociation(
				ctx,
				sourceRouteTableID,
				plan,
				&resp.Diagnostics,
			)
			if !ok || resp.Diagnostics.HasError() {
				subnetID := plan.Id.ValueString()

				rollbackOK := r.rollbackCreatedSubnet(ctx, subnetID, &resp.Diagnostics)
				if !rollbackOK {
					resp.Diagnostics.AddError(
						"Subnet rollback failed",
						fmt.Sprintf(
							"Subnet %q was created, but its route table association "+
								"failed and the provider could not delete the subnet. "+
								"Delete the subnet manually or import it into Terraform.",
							subnetID,
						),
					)
				}

				return
			}
		}
	}

	ok = mapSubnetModel(&plan.subnetBaseModel, result, &resp.Diagnostics)
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
		func() (*vpc.GetSubnetResponse, *http.Response, error) {
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

	ok := mapSubnetModel(&state.subnetBaseModel, new(respModel.Subnet), &resp.Diagnostics)
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

	timeout, diags := plan.Timeouts.Update(ctx, common.DefaultUpdateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	unlock, ok := common.AcquireLockForID(ctx, plan.VpcId.ValueString(), &resp.Diagnostics)
	if !ok {
		return
	}
	defer unlock()

	planDescription := subnetDescriptionValue(plan.Description)
	stateDescription := subnetDescriptionValue(state.Description)
	descriptionChanged := planDescription != stateDescription
	if !plan.Name.Equal(state.Name) || descriptionChanged {
		editReq := vpc.UpdateSubnet{}

		if !plan.Name.Equal(state.Name) {
			editReq.SetName(plan.Name.ValueString())
		}
		if descriptionChanged {
			editReq.SetDescription(planDescription)
		}

		body := *vpc.NewUpdateSubnetRequest(editReq)

		respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
			func() (*vpc.UpdateSubnetResponse, *http.Response, error) {
				return r.kc.ApiClient.VPCSubnetAPI.UpdateSubnet(ctx, plan.Id.ValueString()).
					XAuthToken(r.kc.XAuthToken).
					UpdateSubnetRequest(body).
					Execute()
			},
		)
		if err != nil {
			common.AddApiActionError(ctx, r, httpResp, "UpdateSubnet", err, &resp.Diagnostics)
			return
		}
		state.Name = types.StringValue(respModel.Subnet.Name)
		state.Description = types.StringValue(respModel.Subnet.Description)
	}

	if !plan.RouteTableId.IsUnknown() && !plan.RouteTableId.Equal(state.RouteTableId) {
		sourceRouteTable := state.RouteTableId.ValueString()
		result, ok := r.setAssociation(ctx, sourceRouteTable, plan, &resp.Diagnostics)
		if !ok || resp.Diagnostics.HasError() {
			return
		}
		ok = mapSubnetModel(&plan.subnetBaseModel, result, &resp.Diagnostics)
		if !ok || resp.Diagnostics.HasError() {
			return
		}

		state.RouteTableId = plan.RouteTableId
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func subnetDescriptionValue(value types.String) string {
	if value.IsNull() || value.IsUnknown() {
		return ""
	}
	return value.ValueString()
}

func (r *subnetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state subnetResourceModel
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

	unlock, ok := common.AcquireLockForID(ctx, state.VpcId.ValueString(), &resp.Diagnostics)
	if !ok {
		return
	}
	defer unlock()

	ok = pollVpcStatus(ctx, r, r.kc, state.VpcId.ValueString(), &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	ok = pollSubnetStatus(ctx, r, r.kc, state.Id.ValueString(), &resp.Diagnostics)
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
			func() (*vpc.GetSubnetResponse, *http.Response, error) {
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
		common.ServiceVPC,
		r.kc,
		&resp.Diagnostics,
	)
}

func (r *subnetResource) setAssociation(
	ctx context.Context,
	sourceRouteTableId string,
	plan subnetResourceModel,
	respDiags *diag.Diagnostics,
) (*vpc.Subnet, bool) {
	subnetId := plan.Id.ValueString()

	ok := pollVpcStatus(ctx, r, r.kc, plan.VpcId.ValueString(), respDiags)
	if !ok || respDiags.HasError() {
		return nil, false
	}

	ok = pollSubnetStatus(ctx, r, r.kc, subnetId, respDiags)
	if !ok || respDiags.HasError() {
		return nil, false
	}

	routeTableResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, respDiags,
		func() (*vpc.GetRouteTableResponse, *http.Response, error) {
			return r.kc.ApiClient.VPCRouteTableAPI.GetRouteTable(ctx, sourceRouteTableId).
				XAuthToken(r.kc.XAuthToken).Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "GetRouteTable", err, respDiags)
		return nil, false
	}

	associations := routeTableResp.VpcRouteTable.Associations

	var associationId *string

	for _, association := range associations {
		if association.SubnetId.Get() != nil && *association.SubnetId.Get() == subnetId {
			if association.Id.Get() != nil {
				associationId = new(*association.Id.Get())
			}
			break
		}
	}

	if associationId == nil {
		common.AddGeneralError(ctx, r, respDiags,
			fmt.Sprintf("Source route table does not have a subnet_id '%v'.", subnetId))
		return nil, false
	}

	req := vpc.UpdateRouteTableAssociation{
		TargetRouteTableId: plan.RouteTableId.ValueString(),
	}

	body := *vpc.NewUpdateRouteTableAssociationRequest(req)

	_, httpResp, err = common.ExecuteWithRetryAndAuth(ctx, r.kc, respDiags,
		func() (interface{}, *http.Response, error) {
			return r.kc.ApiClient.VPCRouteTableAPI.UpdateRouteTableAssociation(ctx, sourceRouteTableId, *associationId).
				XAuthToken(r.kc.XAuthToken).UpdateRouteTableAssociationRequest(body).Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "UpdateRouteTableAssociation", err, respDiags)
		return nil, false
	}

	const subnetAssociationApplied = "ASSOCIATION_APPLIED"

	result, ok := common.PollUntilResult(
		ctx,
		r,
		5*time.Second,
		"subnet",
		plan.Id.ValueString(),
		[]string{subnetAssociationApplied, common.VpcProvisioningStatusError},
		respDiags,
		func(ctx context.Context) (*vpc.Subnet, *http.Response, error) {
			respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, respDiags,
				func() (*vpc.GetSubnetResponse, *http.Response, error) {
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
		func(subnet *vpc.Subnet) string {
			status := string(*subnet.ProvisioningStatus.Get())

			if status == common.VpcProvisioningStatusError {
				return status
			}

			if status == common.VpcProvisioningStatusActive &&
				subnet.RouteTableId.Get() != nil &&
				*subnet.RouteTableId.Get() ==
					plan.RouteTableId.ValueString() {
				return subnetAssociationApplied
			}

			return status
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

func (r *subnetResource) rollbackCreatedSubnet(
	ctx context.Context,
	subnetID string,
	respDiags *diag.Diagnostics,
) bool {
	const rollbackTimeout = 2 * time.Minute

	rollbackCtx, cancel := context.WithTimeout(
		context.WithoutCancel(ctx),
		rollbackTimeout,
	)
	defer cancel()

	var rollbackDiags diag.Diagnostics

	_, httpResp, err := common.ExecuteWithRetryAndAuth(
		rollbackCtx,
		r.kc,
		&rollbackDiags,
		func() (interface{}, *http.Response, error) {
			httpResp, err := r.kc.ApiClient.VPCSubnetAPI.DeleteSubnet(rollbackCtx, subnetID).XAuthToken(r.kc.XAuthToken).Execute()
			return nil, httpResp, err
		},
	)
	if err != nil {

		if httpResp != nil &&
			httpResp.StatusCode == http.StatusNotFound {
			return true
		}

		respDiags.AddError(
			"Failed to rollback created subnet",
			fmt.Sprintf(
				"Could not delete subnet %q after route table association failure: %v",
				subnetID,
				err,
			),
		)
		return false
	}

	common.PollUntilDeletion(
		rollbackCtx,
		r,
		5*time.Second,
		&rollbackDiags,
		func(ctx context.Context) (bool, *http.Response, error) {
			_, httpResp, err :=
				common.ExecuteWithRetryAndAuth(
					ctx,
					r.kc,
					&rollbackDiags,
					func() (*vpc.GetSubnetResponse, *http.Response, error) {
						return r.kc.ApiClient.
							VPCSubnetAPI.
							GetSubnet(ctx, subnetID).
							XAuthToken(r.kc.XAuthToken).
							Execute()
					},
				)

			return false, httpResp, err
		},
	)

	if rollbackDiags.HasError() {
		for _, rollbackDiag := range rollbackDiags.Errors() {
			respDiags.AddError(
				"Failed to confirm subnet rollback",
				fmt.Sprintf(
					"Could not confirm deletion of subnet %q: %s",
					subnetID,
					rollbackDiag.Detail(),
				),
			)
		}

		return false
	}

	return true
}

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpc

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-kakaocloud/internal/common"
	. "terraform-provider-kakaocloud/internal/utils"
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
	"github.com/kakaoenterprise/kc-sdk-go/services/vpc"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.ResourceWithConfigure   = &routeTableAssociationResource{}
	_ resource.ResourceWithImportState = &routeTableAssociationResource{}
)

func NewRouteTableAssociationResource() resource.Resource {
	return &routeTableAssociationResource{}
}

type routeTableAssociationResource struct {
	kc *common.KakaoCloudClient
}

func (r *routeTableAssociationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_route_table_association"
}

func (r *routeTableAssociationResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "kakaocloud 특정 Route Table에 연결된 서브넷 목록 관리",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the Route Table to associate the subnets with.",
				Validators:  common.UuidValidator(),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"subnet_ids": schema.SetAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "Association Subnet ID 목록",
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
					setvalidator.ValueStringsAre(common.UuidValidator()...),
				},
			},
			"associations": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed: true,
						},
						"provisioning_status": schema.StringAttribute{
							Computed: true,
						},
						"vpc_id": schema.StringAttribute{
							Computed: true,
						},
						"vpc_name": schema.StringAttribute{
							Computed: true,
						},
						"subnet_id": schema.StringAttribute{
							Computed: true,
						},
						"subnet_name": schema.StringAttribute{
							Computed: true,
						},
						"subnet_cidr_block": schema.StringAttribute{
							Computed: true,
						},
						"availability_zone": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
			"timeouts": timeouts.AttributesAll(ctx),
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *routeTableAssociationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan routeTableAssociationResourceModel
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

	var subnetIds []string
	diags = plan.SubnetIds.ElementsAs(ctx, &subnetIds, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	for _, subnetId := range subnetIds {
		ok := r.setAssociation(ctx, plan.Id.ValueString(), subnetId, &resp.Diagnostics)
		if !ok || resp.Diagnostics.HasError() {
			return
		}
	}

	associationResp, ok := r.readPollingUntilAllActive(ctx, plan.Id.ValueString(), &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	ok = r.mapRouteTableAssociationModel(ctx, &plan, associationResp.Associations, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *routeTableAssociationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state routeTableAssociationResourceModel

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

	associationResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*vpc.RouteTableAssociationListModel, *http.Response, error) {
			return r.kc.ApiClient.VPCRouteTableAssociationAPI.ListRouteTableAssociations(ctx, state.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).Limit(1000).Execute()
		},
	)
	if httpResp != nil && httpResp.StatusCode == 404 {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "ListRouteTableAssociations", err, &resp.Diagnostics)
		return
	}

	ok := r.mapRouteTableAssociationModel(ctx, &state, associationResp.Associations, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	if state.SubnetIds.IsNull() {
		subnetIds := make([]attr.Value, 0, len(associationResp.Associations))
		for _, association := range associationResp.Associations {
			subnetIds = append(subnetIds, ConvertNullableString(association.SubnetId))
		}
		state.SubnetIds, diags = types.SetValue(types.StringType, subnetIds)
		if diags.HasError() {
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
func (r *routeTableAssociationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state routeTableAssociationResourceModel
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

	var planSubnetIds, stateSubnetIds []string
	diags = plan.SubnetIds.ElementsAs(ctx, &planSubnetIds, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = state.SubnetIds.ElementsAs(ctx, &stateSubnetIds, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	stateSet := make(map[string]struct{})
	for _, id := range stateSubnetIds {
		stateSet[id] = struct{}{}
	}

	// Add Association
	for _, subnetId := range planSubnetIds {
		if _, exists := stateSet[subnetId]; !exists {
			ok := r.setAssociation(ctx, plan.Id.ValueString(), subnetId, &resp.Diagnostics)
			if !ok || resp.Diagnostics.HasError() {
				return
			}
		}
	}

	associationResp, err := r.readPollingUntilAllActive(ctx, plan.Id.ValueString(), &resp.Diagnostics)
	if err || resp.Diagnostics.HasError() {
		return
	}

	ok := r.mapRouteTableAssociationModel(ctx, &plan, associationResp.Associations, &resp.Diagnostics)
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
func (r *routeTableAssociationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state routeTableAssociationResourceModel
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

	var associationList []associationModel
	diags = state.Associations.ElementsAs(ctx, &associationList, false)
	if diags.HasError() {
		return
	}

	if len(associationList) > 0 {
		vpcId := associationList[0].VpcId.ValueString()

		if resp.Diagnostics.HasError() {
			return
		}
		routeTableResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
			func() (*vpc.RouteTableListModel, *http.Response, error) {
				return r.kc.ApiClient.VPCRouteTableAPI.ListRouteTables(ctx).VpcId(vpcId).Limit(1000).XAuthToken(r.kc.XAuthToken).Execute()
			},
		)
		if err != nil {
			common.AddApiActionError(ctx, r, httpResp, "ListRouteTables", err, &resp.Diagnostics)
			return
		}

		for _, rt := range routeTableResp.VpcRouteTables {
			isMain := rt.IsMain.Get()
			if *isMain {
				if rt.Id == state.Id.ValueString() {
					common.AddGeneralError(ctx, r, &resp.Diagnostics,
						fmt.Sprintf("The main route table cannot have its associations deleted. id: %v", state.Id.ValueString()))
				} else {
					var subnetIds []string
					diags = state.SubnetIds.ElementsAs(ctx, &subnetIds, false)
					resp.Diagnostics.Append(diags...)
					if resp.Diagnostics.HasError() {
						return
					}

					for _, subnetId := range subnetIds {
						ok := r.setAssociation(ctx, rt.Id, subnetId, &resp.Diagnostics)
						if !ok || resp.Diagnostics.HasError() {
							return
						}
					}

					_, err := r.readPollingUntilAllActive(ctx, rt.Id, &resp.Diagnostics)
					if err || resp.Diagnostics.HasError() {
						return
					}
				}
			}
		}
	}
}

func (r *routeTableAssociationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *routeTableAssociationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *routeTableAssociationResource) mapRouteTableAssociationModel(
	ctx context.Context,
	base *routeTableAssociationResourceModel,
	result []vpc.BnsVpcV1ApiListRouteTableAssociationsModelAssociationModel,
	respDiags *diag.Diagnostics,
) bool {
	associationList, diags := ConvertListFromModel(
		ctx,
		result,
		associationAttrType,
		func(src vpc.BnsVpcV1ApiListRouteTableAssociationsModelAssociationModel) any {
			return routeTableAssociationModel{
				Id:                 types.StringValue(src.Id),
				ProvisioningStatus: ConvertNullableString(src.ProvisioningStatus),
				VpcId:              ConvertNullableString(src.VpcId),
				VpcName:            ConvertNullableString(src.VpcName),
				SubnetId:           ConvertNullableString(src.SubnetId),
				SubnetName:         ConvertNullableString(src.SubnetName),
				SubnetCidrBlock:    ConvertNullableString(src.SubnetCidrBlock),
				AvailabilityZone:   ConvertNullableString(src.AvailabilityZone),
			}
		},
	)
	respDiags.Append(diags...)

	base.Associations = associationList

	return !diags.HasError()
}

func (r *routeTableAssociationResource) setAssociation(
	ctx context.Context,
	routeTableId string,
	subnetId string,
	respDiags *diag.Diagnostics,
) bool {
	routeTableResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, respDiags,
		func() (*vpc.RouteTableListModel, *http.Response, error) {
			return r.kc.ApiClient.VPCRouteTableAPI.ListRouteTables(ctx).XAuthToken(r.kc.XAuthToken).SubnetId(subnetId).Limit(1000).Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "ListRouteTables", err, respDiags)
		return false
	}

	sourceRouteTableId := routeTableResp.VpcRouteTables[0].Id
	associations := routeTableResp.VpcRouteTables[0].Associations
	vpcId := routeTableResp.VpcRouteTables[0].VpcId.Get()

	ok := checkVpcStatus(ctx, r, r.kc, *vpcId, respDiags)
	if !ok || respDiags.HasError() {
		return false
	}

	ok = checkSubnetStatus(ctx, r, r.kc, subnetId, respDiags)
	if !ok || respDiags.HasError() {
		return false
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
			fmt.Sprintf("subnet_id '%v' does not exist.", subnetId))
		return false
	}

	req := vpc.EditAssociationModel{
		TargetRouteTableId: routeTableId,
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
		return false
	}

	return true
}

func (r *routeTableAssociationResource) readPollingUntilAllActive(ctx context.Context, routeTableId string, respDiags *diag.Diagnostics,
) (*vpc.RouteTableAssociationListModel, bool) {
	for {
		isAllActive := true
		associationResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, respDiags,
			func() (*vpc.RouteTableAssociationListModel, *http.Response, error) {
				return r.kc.ApiClient.VPCRouteTableAssociationAPI.ListRouteTableAssociations(ctx, routeTableId).
					XAuthToken(r.kc.XAuthToken).Limit(1000).Execute()
			},
		)
		if err != nil {
			common.AddApiActionError(ctx, r, httpResp, "ListRouteTableAssociations", err, respDiags)
			return nil, false
		}

		for _, association := range associationResp.Associations {
			if string(*association.ProvisioningStatus.Get()) != common.VpcProvisioningStatusActive {
				isAllActive = false
				break
			}
		}

		if isAllActive {
			return associationResp, true
		}
		time.Sleep(2 * time.Second)
	}
}

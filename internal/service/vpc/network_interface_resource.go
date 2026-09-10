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
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kakaoenterprise/kc-sdk-go/v2/services/network"
	"github.com/kakaoenterprise/kc-sdk-go/v2/services/vpc"
)

const networkInterfaceSynchronizedStatus = "synchronized"

type networkInterfaceExpectedState struct {
	name                  *types.String
	description           *types.String
	securityGroupIds      map[string]struct{}
	allowedAddressPairIps map[string]struct{}
}

var (
	_ resource.ResourceWithConfigure   = &networkInterfaceResource{}
	_ resource.ResourceWithImportState = &networkInterfaceResource{}
	_ resource.ResourceWithModifyPlan  = &networkInterfaceResource{}
)

func NewNetworkInterfaceResource() resource.Resource {
	return &networkInterfaceResource{}
}

type networkInterfaceResource struct {
	kc *common.KakaoCloudClient
}

func (r *networkInterfaceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network_interface"
}

func (r *networkInterfaceResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: MergeResourceSchemaAttributes(
			networkInterfaceResourceSchema,
			map[string]schema.Attribute{
				"timeouts": timeouts.AttributesAll(ctx),
			},
		),
	}
}

func (r *networkInterfaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan networkInterfaceResourceModel
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

	createReq := vpc.CreateNetworkInterface{
		Name:     plan.Name.ValueString(),
		SubnetId: plan.SubnetId.ValueString(),
	}

	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		createReq.SetDescription(plan.Description.ValueString())
	}

	if !plan.PrivateIp.IsNull() && !plan.PrivateIp.IsUnknown() {
		createReq.SetPrivateIp(plan.PrivateIp.ValueString())
	}

	var sgList []securityGroupModel
	diags = plan.SecurityGroups.ElementsAs(ctx, &sgList, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	var sgIds []string
	for _, sg := range sgList {
		if !sg.Id.IsNull() && !sg.Id.IsUnknown() {
			sgIds = append(sgIds, sg.Id.ValueString())
		}
	}
	createReq.SetSecurityGroups(sgIds)

	body := *vpc.NewCreateNetworkInterfaceRequest(createReq)

	respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*vpc.CreateNetworkInterfaceResponse, *http.Response, error) {
			return r.kc.ApiClient.VPCNetworkInterfaceAPI.CreateNetworkInterface(ctx).
				XAuthToken(r.kc.XAuthToken).CreateNetworkInterfaceRequest(body).Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "CreateNetworkInterface", err, &resp.Diagnostics)
		return
	}

	plan.Id = ConvertNullableString(respModel.NetworkInterface.Id)
	expected := &networkInterfaceExpectedState{
		name:             &plan.Name,
		securityGroupIds: stringSet(sgIds),
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		expected.description = &plan.Description
	}

	result, ok := r.pollNetworkInterfaceUntilStatus(
		ctx,
		plan.Id.ValueString(),
		[]string{common.NetworkInterfaceStatusAvailable},
		expected,
		&resp.Diagnostics,
	)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	if !plan.AllowedAddressPairs.IsNull() && !plan.AllowedAddressPairs.IsUnknown() {
		if !r.updateAllowedAddressPairs(ctx, plan, &resp.Diagnostics) {
			return
		}
		expected, ok := buildAllowedAddressPairsExpectedState(ctx, plan, &resp.Diagnostics)
		if !ok {
			return
		}

		result, ok = r.pollNetworkInterfaceUntilStatus(
			ctx,
			plan.Id.ValueString(),
			[]string{common.NetworkInterfaceStatusAvailable},
			expected,
			&resp.Diagnostics,
		)
		if !ok || resp.Diagnostics.HasError() {
			return
		}
	}

	ok = mapNetworkInterfaceBaseModel(ctx, &plan.networkInterfaceBaseModel, result, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

func (r *networkInterfaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state networkInterfaceResourceModel
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
		func() (*vpc.GetNetworkInterfaceResponse, *http.Response, error) {
			return r.kc.ApiClient.VPCNetworkInterfaceAPI.GetNetworkInterface(ctx, state.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).Execute()
		},
	)

	if httpResp != nil && httpResp.StatusCode == 404 {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "GetNetworkInterface", err, &resp.Diagnostics)
		return
	}

	ok := mapNetworkInterfaceBaseModel(ctx, &state.networkInterfaceBaseModel, &respModel.NetworkInterface, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *networkInterfaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state networkInterfaceResourceModel
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

	var result *vpc.NetworkInterface
	var ok bool

	if !plan.Name.Equal(state.Name) || !plan.Description.Equal(state.Description) || !plan.SecurityGroups.Equal(state.SecurityGroups) {
		var editReq vpc.UpdateNetworkInterface
		expected := &networkInterfaceExpectedState{}

		if !plan.Name.Equal(state.Name) {
			editReq.SetName(plan.Name.ValueString())
			expected.name = &plan.Name
		}
		if !plan.Description.Equal(state.Description) {
			if plan.Description.IsNull() {
				editReq.SetDescriptionNil()
				expected.description = &plan.Description
			} else if !plan.Description.IsUnknown() {
				editReq.SetDescription(plan.Description.ValueString())
				expected.description = &plan.Description
			}
		}
		if !plan.SecurityGroups.IsNull() {
			var sgList []securityGroupModel
			diags := plan.SecurityGroups.ElementsAs(ctx, &sgList, false)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			var sgIds []string
			for _, sg := range sgList {
				if !sg.Id.IsNull() && !sg.Id.IsUnknown() {
					sgIds = append(sgIds, sg.Id.ValueString())
				}
			}
			editReq.SetSecurityGroups(sgIds)
			expected.securityGroupIds = stringSet(sgIds)
		}

		body := *vpc.NewUpdateNetworkInterfaceRequest(editReq)

		_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
			func() (*vpc.UpdateNetworkInterfaceResponse, *http.Response, error) {
				return r.kc.ApiClient.VPCNetworkInterfaceAPI.UpdateNetworkInterface(ctx, plan.Id.ValueString()).
					XAuthToken(r.kc.XAuthToken).
					UpdateNetworkInterfaceRequest(body).
					Execute()
			},
		)
		if err != nil {
			common.AddApiActionError(ctx, r, httpResp, "UpdateNetworkInterface", err, &resp.Diagnostics)
			return
		}

		result, ok = r.pollNetworkInterfaceUntilStatus(
			ctx,
			plan.Id.ValueString(),
			[]string{common.NetworkInterfaceStatusAvailable},
			expected,
			&resp.Diagnostics,
		)
		if !ok || resp.Diagnostics.HasError() {
			return
		}
	}

	if (plan.AllowedAddressPairs.IsUnknown() && len(state.AllowedAddressPairs.Elements()) != 0) ||
		(!plan.AllowedAddressPairs.IsUnknown() && !plan.AllowedAddressPairs.Equal(state.AllowedAddressPairs)) {

		if !r.updateAllowedAddressPairs(ctx, plan, &resp.Diagnostics) {
			return
		}
		expected, expectedOk := buildAllowedAddressPairsExpectedState(ctx, plan, &resp.Diagnostics)
		if !expectedOk {
			return
		}

		result, ok = r.pollNetworkInterfaceUntilStatus(
			ctx,
			plan.Id.ValueString(),
			[]string{common.NetworkInterfaceStatusAvailable},
			expected,
			&resp.Diagnostics,
		)
		if !ok || resp.Diagnostics.HasError() {
			return
		}
	}

	if result != nil {
		ok = mapNetworkInterfaceBaseModel(ctx, &plan.networkInterfaceBaseModel, result, &resp.Diagnostics)
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

func (r *networkInterfaceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state networkInterfaceResourceModel
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
			httpResp, err := r.kc.ApiClient.VPCNetworkInterfaceAPI.DeleteNetworkInterface(ctx, state.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				Execute()
			return nil, httpResp, err
		},
	)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			return
		}
		common.AddApiActionError(ctx, r, httpResp, "DeleteNetworkInterface", err, &resp.Diagnostics)
		return
	}

	common.PollUntilDeletion(ctx, r, 2*time.Second, &resp.Diagnostics, func(ctx context.Context) (bool, *http.Response, error) {
		_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
			func() (*vpc.GetNetworkInterfaceResponse, *http.Response, error) {
				_, httpResp, err := r.kc.ApiClient.VPCNetworkInterfaceAPI.
					GetNetworkInterface(ctx, state.Id.ValueString()).
					XAuthToken(r.kc.XAuthToken).
					Execute()
				return nil, httpResp, err
			},
		)
		return false, httpResp, err
	})
}

func (r *networkInterfaceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {

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

func (r *networkInterfaceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {

	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *networkInterfaceResource) pollNetworkInterfaceUntilStatus(
	ctx context.Context,
	networkInterfaceId string,
	targetStatuses []string,
	expected *networkInterfaceExpectedState,
	resp *diag.Diagnostics,
) (*vpc.NetworkInterface, bool) {
	pollingTargetStatuses := targetStatuses
	if expected != nil {
		pollingTargetStatuses = []string{networkInterfaceSynchronizedStatus}
	}

	return common.PollUntilResult(
		ctx,
		r,
		2*time.Second,
		"network interfaces",
		networkInterfaceId,
		pollingTargetStatuses,
		resp,
		func(ctx context.Context) (*vpc.NetworkInterface, *http.Response, error) {
			respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, resp,
				func() (*vpc.GetNetworkInterfaceResponse, *http.Response, error) {
					return r.kc.ApiClient.VPCNetworkInterfaceAPI.
						GetNetworkInterface(ctx, networkInterfaceId).
						XAuthToken(r.kc.XAuthToken).
						Execute()
				},
			)
			if err != nil {
				return nil, httpResp, err
			}
			return &respModel.NetworkInterface, httpResp, nil
		},
		func(v *vpc.NetworkInterface) string {
			if v.Status.Get() == nil {
				return ""
			}

			status := string(*v.Status.Get())
			if expected != nil && containsString(targetStatuses, status) && networkInterfaceMatchesExpected(v, expected) {
				return networkInterfaceSynchronizedStatus
			}
			return status
		},
	)
}

func buildAllowedAddressPairsExpectedState(
	ctx context.Context,
	plan networkInterfaceResourceModel,
	resp *diag.Diagnostics,
) (*networkInterfaceExpectedState, bool) {
	expected := &networkInterfaceExpectedState{
		allowedAddressPairIps: make(map[string]struct{}),
	}
	if plan.AllowedAddressPairs.IsNull() || plan.AllowedAddressPairs.IsUnknown() {
		return expected, true
	}

	var allowedAddressPairs []allowedAddressPairModel
	diags := plan.AllowedAddressPairs.ElementsAs(ctx, &allowedAddressPairs, false)
	resp.Append(diags...)
	if resp.HasError() {
		return nil, false
	}
	for _, pair := range allowedAddressPairs {
		if !pair.IpAddress.IsNull() && !pair.IpAddress.IsUnknown() {
			expected.allowedAddressPairIps[pair.IpAddress.ValueString()] = struct{}{}
		}
	}
	return expected, true
}

func networkInterfaceMatchesExpected(networkInterface *vpc.NetworkInterface, expected *networkInterfaceExpectedState) bool {
	if expected.name != nil && !ConvertNullableString(networkInterface.Name).Equal(*expected.name) {
		return false
	}
	if expected.description != nil && !ConvertNullableString(networkInterface.Description).Equal(*expected.description) {
		return false
	}
	if expected.securityGroupIds != nil {
		actual := make(map[string]struct{}, len(networkInterface.SecurityGroups))
		for _, securityGroup := range networkInterface.SecurityGroups {
			id := ConvertNullableString(securityGroup.Id)
			if !id.IsNull() && !id.IsUnknown() {
				actual[id.ValueString()] = struct{}{}
			}
		}
		if !equalStringSets(actual, expected.securityGroupIds) {
			return false
		}
	}
	if expected.allowedAddressPairIps != nil {
		actual := make(map[string]struct{}, len(networkInterface.AllowedAddressPairs))
		for _, pair := range networkInterface.AllowedAddressPairs {
			ipAddress := ConvertNullableString(pair.IpAddress)
			if !ipAddress.IsNull() && !ipAddress.IsUnknown() {
				actual[ipAddress.ValueString()] = struct{}{}
			}
		}
		if !equalStringSets(actual, expected.allowedAddressPairIps) {
			return false
		}
	}
	return true
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func stringSet(values []string) map[string]struct{} {
	result := make(map[string]struct{}, len(values))
	for _, value := range values {
		result[value] = struct{}{}
	}
	return result
}

func equalStringSets(left, right map[string]struct{}) bool {
	if len(left) != len(right) {
		return false
	}
	for value := range left {
		if _, ok := right[value]; !ok {
			return false
		}
	}
	return true
}

func (r *networkInterfaceResource) updateAllowedAddressPairs(ctx context.Context, plan networkInterfaceResourceModel, resp *diag.Diagnostics) bool {
	var editReq []vpc.UpdateNetworkInterfaceAllowedAddresses

	if plan.AllowedAddressPairs.IsUnknown() || len(plan.AllowedAddressPairs.Elements()) == 0 {
		editReq = []vpc.UpdateNetworkInterfaceAllowedAddresses{}
	} else {
		var allowedAddressPairs []allowedAddressPairModel
		diags := plan.AllowedAddressPairs.ElementsAs(ctx, &allowedAddressPairs, false)
		resp.Append(diags...)
		if resp.HasError() {
			return false
		}

		for _, allowedAddressPair := range allowedAddressPairs {
			var tmpEditReq vpc.UpdateNetworkInterfaceAllowedAddresses
			tmpEditReq.SetIpAddress(allowedAddressPair.IpAddress.ValueString())
			editReq = append(editReq, tmpEditReq)
		}
	}

	body := *vpc.NewUpdateNetworkInterfaceAllowedAddressesRequest(editReq)

	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, resp,
		func() (interface{}, *http.Response, error) {
			return r.kc.ApiClient.VPCNetworkInterfaceAPI.UpdateNetworkInterfaceAllowedAddresses(ctx, plan.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				UpdateNetworkInterfaceAllowedAddressesRequest(body).
				Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "UpdateNetworkInterfaceAllowedAddresses", err, resp)
		return false
	}
	return true
}

func (r *networkInterfaceResource) ModifyPlan(
	ctx context.Context,
	req resource.ModifyPlanRequest,
	resp *resource.ModifyPlanResponse,
) {

	if req.Plan.Raw.IsNull() {
		return
	}

	var plan networkInterfaceResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() || plan.SecurityGroups.IsNull() {
		return
	}

	var securityGroups []securityGroupModel
	resp.Diagnostics.Append(plan.SecurityGroups.ElementsAs(ctx, &securityGroups, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	for i := range securityGroups {
		if !securityGroups[i].Name.IsUnknown() {
			continue
		}

		if securityGroups[i].Id.IsUnknown() {
			continue
		}

		respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
			func() (*network.GetSecurityGroupResponse, *http.Response, error) {
				return r.kc.ApiClient.NetworkSecurityGroupAPI.
					GetSecurityGroup(ctx, securityGroups[i].Id.ValueString()).
					XAuthToken(r.kc.XAuthToken).
					Execute()
			},
		)
		if err != nil {
			common.AddApiActionError(ctx, r, httpResp, "GetSecurityGroup", err, &resp.Diagnostics)
			return
		}

		securityGroups[i].Name = ConvertNullableString(respModel.SecurityGroup.Name)
	}

	securityGroupSet, diags := types.SetValueFrom(
		ctx,
		types.ObjectType{AttrTypes: securityGroupAttrType},
		securityGroups,
	)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(
		resp.Plan.SetAttribute(ctx, path.Root("security_groups"), securityGroupSet)...)
}

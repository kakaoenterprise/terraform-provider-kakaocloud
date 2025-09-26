// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package vpc

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
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kakaoenterprise/kc-sdk-go/services/vpc"
)

var (
	_ resource.ResourceWithConfigure   = &networkInterfaceResource{}
	_ resource.ResourceWithImportState = &networkInterfaceResource{}
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
		Description: docs.GetResourceDescription("NetworkInterface"),
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

	createReq := vpc.CreateNetworkInterfaceModel{
		Name:     plan.Name.ValueString(),
		SubnetId: plan.SubnetId.ValueString(),
	}

	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		createReq.SetDescription(plan.Description.ValueString())
	}

	if !plan.PrivateIp.IsNull() && !plan.PrivateIp.IsUnknown() {
		createReq.SetPrivateIp(plan.PrivateIp.ValueString())
	}

	if !plan.SecurityGroups.IsNull() && !plan.SecurityGroups.IsUnknown() {
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
		createReq.SetSecurityGroups(sgIds)
	}

	body := *vpc.NewBodyCreateNetworkInterface(createReq)

	respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*vpc.BnsVpcV1ApiCreateNetworkInterfaceModelResponseNetworkInterfaceModel, *http.Response, error) {
			return r.kc.ApiClient.NetworkInterfaceAPI.CreateNetworkInterface(ctx).
				XAuthToken(r.kc.XAuthToken).BodyCreateNetworkInterface(body).Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "CreateNetworkInterface", err, &resp.Diagnostics)
		return
	}

	plan.Id = types.StringValue(respModel.NetworkInterface.Id)

	result, ok := r.pollNetworkInterfaceUntilStatus(
		ctx,
		plan.Id.ValueString(),
		[]string{common.NetworkInterfaceStatusAvailable},
		&resp.Diagnostics,
	)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	if !plan.AllowedAddressPairs.IsNull() && !plan.AllowedAddressPairs.IsUnknown() {
		if !r.updateAllowedAddressPairs(ctx, plan, &resp.Diagnostics) {
			return
		}

		result, ok = r.pollNetworkInterfaceUntilStatus(
			ctx,
			plan.Id.ValueString(),
			[]string{common.NetworkInterfaceStatusAvailable},
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
		func() (*vpc.BnsVpcV1ApiGetNetworkInterfaceModelResponseNetworkInterfaceModel, *http.Response, error) {
			return r.kc.ApiClient.NetworkInterfaceAPI.GetNetworkInterface(ctx, state.Id.ValueString()).
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

	var result *vpc.BnsVpcV1ApiGetNetworkInterfaceModelNetworkInterfaceModel
	var ok bool

	if !plan.Name.Equal(state.Name) || !plan.Description.Equal(state.Description) || !plan.SecurityGroups.Equal(state.SecurityGroups) {
		var editReq vpc.EditNetworkInterfaceModel

		if !plan.Name.Equal(state.Name) {
			editReq.SetName(plan.Name.ValueString())
		}
		if !plan.Description.Equal(state.Description) && !plan.Description.IsNull() && !plan.Description.IsUnknown() {
			editReq.SetDescription(plan.Description.ValueString())
		}
		if !plan.SecurityGroups.IsNull() && !plan.SecurityGroups.IsUnknown() {
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
		}

		body := *vpc.NewBodyUpdateNetworkInterface(editReq)

		_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
			func() (*vpc.BnsVpcV1ApiUpdateNetworkInterfaceModelResponseNetworkInterfaceModel, *http.Response, error) {
				return r.kc.ApiClient.NetworkInterfaceAPI.UpdateNetworkInterface(ctx, plan.Id.ValueString()).
					XAuthToken(r.kc.XAuthToken).
					BodyUpdateNetworkInterface(body).
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
			&resp.Diagnostics,
		)
		if !ok || resp.Diagnostics.HasError() {
			return
		}
	}

	if !plan.AllowedAddressPairs.IsNull() && !plan.AllowedAddressPairs.IsUnknown() && !plan.AllowedAddressPairs.Equal(state.AllowedAddressPairs) {
		if !r.updateAllowedAddressPairs(ctx, plan, &resp.Diagnostics) {
			return
		}

		result, ok = r.pollNetworkInterfaceUntilStatus(
			ctx,
			plan.Id.ValueString(),
			[]string{common.NetworkInterfaceStatusAvailable},
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
			httpResp, err := r.kc.ApiClient.NetworkInterfaceAPI.DeleteNetworkInterface(ctx, state.Id.ValueString()).
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
			func() (*vpc.BnsVpcV1ApiGetNetworkInterfaceModelResponseNetworkInterfaceModel, *http.Response, error) {
				_, httpResp, err := r.kc.ApiClient.NetworkInterfaceAPI.
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
	resp *diag.Diagnostics,
) (*vpc.BnsVpcV1ApiGetNetworkInterfaceModelNetworkInterfaceModel, bool) {
	return common.PollUntilResult(
		ctx,
		r,
		2*time.Second,
		targetStatuses,
		resp,
		func(ctx context.Context) (*vpc.BnsVpcV1ApiGetNetworkInterfaceModelNetworkInterfaceModel, *http.Response, error) {
			respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, resp,
				func() (*vpc.BnsVpcV1ApiGetNetworkInterfaceModelResponseNetworkInterfaceModel, *http.Response, error) {
					return r.kc.ApiClient.NetworkInterfaceAPI.
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
		func(v *vpc.BnsVpcV1ApiGetNetworkInterfaceModelNetworkInterfaceModel) string {
			return string(*v.Status.Get())
		},
	)
}

func (r *networkInterfaceResource) updateAllowedAddressPairs(ctx context.Context, plan networkInterfaceResourceModel, resp *diag.Diagnostics) bool {
	var editReq []vpc.EditAllowedAddressPairModel

	var allowedAddressPairs []allowedAddressPairModel
	diags := plan.AllowedAddressPairs.ElementsAs(ctx, &allowedAddressPairs, false)
	resp.Append(diags...)
	if resp.HasError() {
		return false
	}

	for _, allowedAddressPair := range allowedAddressPairs {
		if !allowedAddressPair.IpAddress.IsNull() && !allowedAddressPair.IpAddress.IsUnknown() {
			var tmpEditReq vpc.EditAllowedAddressPairModel
			tmpEditReq.SetIpAddress(allowedAddressPair.IpAddress.ValueString())
			editReq = append(editReq, tmpEditReq)
		}
	}

	body := *vpc.NewBodyUpdateNetworkInterfaceAllowedAddresses(editReq)

	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, resp,
		func() (interface{}, *http.Response, error) {
			return r.kc.ApiClient.NetworkInterfaceAPI.UpdateNetworkInterfaceAllowedAddresses(ctx, plan.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				BodyUpdateNetworkInterfaceAllowedAddresses(body).
				Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "UpdateNetworkInterfaceAllowedAddresses", err, resp)
		return false
	}
	return true
}

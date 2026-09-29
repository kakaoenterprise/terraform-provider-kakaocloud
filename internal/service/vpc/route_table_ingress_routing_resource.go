// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package vpc

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"terraform-provider-kakaocloud/internal/common"
	"terraform-provider-kakaocloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kakaoenterprise/kc-sdk-go/v2/services/vpc"
)

var (
	_ resource.ResourceWithConfigure   = &routeTableIngressRoutingResource{}
	_ resource.ResourceWithImportState = &routeTableIngressRoutingResource{}
)

func NewRouteTableIngressRoutingResource() resource.Resource {
	return &routeTableIngressRoutingResource{}
}

type routeTableIngressRoutingResource struct {
	kc *common.KakaoCloudClient
}

func (r *routeTableIngressRoutingResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_route_table_ingress_routing"
}

func (r *routeTableIngressRoutingResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: utils.MergeResourceSchemaAttributes(
			routeTableIngressRoutingResourceSchemaAttributes,
			map[string]schema.Attribute{
				"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
					Create: true,
					Read:   true,
					Delete: true,
				}),
			},
		),
	}
}

func (r *routeTableIngressRoutingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan routeTableIngressRoutingResourceModel
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

	operationCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	unlock, ok := common.AcquireLockForID(operationCtx, plan.InternetGatewayId.ValueString(), &resp.Diagnostics)
	if !ok {
		return
	}
	defer unlock()

	unlock2, ok := common.AcquireLockForID(operationCtx, plan.VpcId.ValueString(), &resp.Diagnostics)
	if !ok {
		return
	}
	defer unlock2()

	table, httpResp, err := r.getRouteTable(operationCtx, plan.RouteTableId.ValueString(), &resp.Diagnostics)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "GetRouteTable", err, &resp.Diagnostics)
		return
	}

	if table.GetVpcId() != plan.VpcId.ValueString() {
		resp.Diagnostics.AddError(
			"VPC Mismatch",
			"The route table does not belong to the configured vpc_id.",
		)
		return
	}

	if table.IngressRouting.Get() != nil {
		resp.Diagnostics.AddError(
			"Ingress Routing Already Exists",
			fmt.Sprintf("Route table %s already has an ingress routing connection. Import it using the route table ID to manage it with Terraform.", plan.RouteTableId.ValueString()),
		)
		return
	}

	ok = pollVpcStatus(operationCtx, r, r.kc, plan.VpcId.ValueString(), &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	ingressRouting := vpc.NewEnableIngressRouting(plan.VpcId.ValueString(), plan.InternetGatewayId.ValueString())
	body := vpc.NewEnableIngressRoutingRequest(*ingressRouting)

	result, httpResp, err := common.ExecuteWithRetryAndAuth(operationCtx, r.kc, &resp.Diagnostics,
		func() (*vpc.EnableIngressRoutingResponse, *http.Response, error) {
			return r.kc.ApiClient.VPCRouteTableAPI.
				EnableIngressRouting(operationCtx, plan.RouteTableId.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				EnableIngressRoutingRequest(*body).
				Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "EnableIngressRouting", err, &resp.Diagnostics)
		return
	}

	plan.Id = types.StringValue(result.IngressRouting.Id)
	plan.ProvisioningStatus = types.StringValue(string(result.IngressRouting.ProvisioningStatus))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	table, ok = r.pollUntilActive(
		operationCtx,
		plan.RouteTableId.ValueString(),
		plan.Id.ValueString(),
		plan.InternetGatewayId.ValueString(),
		&resp.Diagnostics,
	)
	if !ok {
		resp.Diagnostics.AddError(
			"Ingress Routing Creation Incomplete",
			fmt.Sprintf("Connection %s was created and recorded in Terraform state, but readiness could not be confirmed. Check route table %s before retrying. If state recovery is needed, use terraform import with route table ID %s.", plan.Id.ValueString(), plan.RouteTableId.ValueString(), plan.RouteTableId.ValueString()),
		)
		return
	}

	connection := table.IngressRouting.Get()
	plan.ProvisioningStatus = utils.ConvertNullableString(connection.ProvisioningStatus)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if connection.GetProvisioningStatus() != vpc.PROVISIONINGSTATUS_ACTIVE {
		resp.Diagnostics.AddError(
			"Ingress Routing Creation Failed",
			fmt.Sprintf("Connection %s reached status %s. It remains recorded in Terraform state.", connection.Id, connection.GetProvisioningStatus()),
		)
	}
}

func (r *routeTableIngressRoutingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state routeTableIngressRoutingResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
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

	table, httpResp, err := r.getRouteTable(ctx, state.RouteTableId.ValueString(), &resp.Diagnostics)
	if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "GetRouteTable", err, &resp.Diagnostics)
		return
	}

	connection := table.IngressRouting.Get()
	if connection == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	if !state.Id.IsNull() && !state.Id.IsUnknown() && state.Id.ValueString() != connection.Id {
		resp.State.RemoveResource(ctx)
		return
	}
	state.Id = types.StringValue(connection.Id)
	state.ProvisioningStatus = utils.ConvertNullableString(connection.ProvisioningStatus)

	if table.VpcId.Get() != nil {
		state.VpcId = utils.ConvertNullableString(table.VpcId)
	}

	if connection.IgwId.Get() != nil {
		state.InternetGatewayId = utils.ConvertNullableString(connection.IgwId)
	}

	if state.VpcId.IsNull() || state.VpcId.IsUnknown() || state.InternetGatewayId.IsNull() || state.InternetGatewayId.IsUnknown() {
		resp.Diagnostics.AddError(
			"Incomplete Ingress Routing Response",
			"GetRouteTable did not return the VPC and internet gateway IDs required to import the connection. Retry when provisioning has completed.",
		)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *routeTableIngressRoutingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

	var plan routeTableIngressRoutingResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state routeTableIngressRoutingResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ProvisioningStatus = state.ProvisioningStatus
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *routeTableIngressRoutingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state routeTableIngressRoutingResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
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

	unlock, ok := common.AcquireLockForID(ctx, state.InternetGatewayId.ValueString(), &resp.Diagnostics)
	if !ok {
		return
	}
	defer unlock()

	unlock2, ok := common.AcquireLockForID(ctx, state.VpcId.ValueString(), &resp.Diagnostics)
	if !ok {
		return
	}
	defer unlock2()

	table, httpResp, err := r.getRouteTable(ctx, state.RouteTableId.ValueString(), &resp.Diagnostics)
	if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
		return
	}
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "GetRouteTable", err, &resp.Diagnostics)
		return
	}

	connection := table.IngressRouting.Get()

	if connection == nil || connection.Id != state.Id.ValueString() {
		return
	}

	ok = pollVpcStatus(ctx, r, r.kc, state.VpcId.ValueString(), &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	table, httpResp, err = r.getRouteTable(ctx, state.RouteTableId.ValueString(), &resp.Diagnostics)
	if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
		return
	}
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "GetRouteTable", err, &resp.Diagnostics)
		return
	}
	connection = table.IngressRouting.Get()
	if connection == nil || connection.Id != state.Id.ValueString() {
		return
	}

	_, httpResp, err = common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (struct{}, *http.Response, error) {
			httpResp, err := r.kc.ApiClient.VPCRouteTableAPI.
				DisableIngressRouting(ctx, state.RouteTableId.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				Execute()
			return struct{}{}, httpResp, err
		},
	)
	if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
		return
	}
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "DisableIngressRouting", err, &resp.Diagnostics)
		return
	}

	common.PollUntilDeletion(ctx, r, 2*time.Second, &resp.Diagnostics,
		func(ctx context.Context) (bool, *http.Response, error) {
			table, httpResp, err := r.getRouteTable(ctx, state.RouteTableId.ValueString(), &resp.Diagnostics)
			if err != nil {
				return false, httpResp, err
			}
			connection := table.IngressRouting.Get()
			return connection == nil || connection.Id != state.Id.ValueString(), httpResp, nil
		},
	)
}

func (r *routeTableIngressRoutingResource) getRouteTable(ctx context.Context, id string, diags *diag.Diagnostics) (*vpc.VpcRouteTable, *http.Response, error) {
	result, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, diags,
		func() (*vpc.GetRouteTableResponse, *http.Response, error) {
			return r.kc.ApiClient.VPCRouteTableAPI.
				GetRouteTable(ctx, id).
				XAuthToken(r.kc.XAuthToken).
				Execute()
		},
	)
	if err != nil {
		return nil, httpResp, err
	}
	return &result.VpcRouteTable, httpResp, nil
}

func (r *routeTableIngressRoutingResource) pollUntilActive(
	ctx context.Context,
	routeTableID string,
	connectionID string,
	internetGatewayID string,
	diags *diag.Diagnostics,
) (*vpc.VpcRouteTable, bool) {
	return common.PollUntilResult(
		ctx,
		r,
		2*time.Second,
		"ingress routing",
		connectionID,
		[]string{common.VpcProvisioningStatusActive, common.VpcProvisioningStatusError},
		diags,
		func(ctx context.Context) (*vpc.VpcRouteTable, *http.Response, error) {
			return r.getRouteTable(ctx, routeTableID, diags)
		},
		func(table *vpc.VpcRouteTable) string {
			connection := table.IngressRouting.Get()
			if connection == nil || connection.Id != connectionID {
				return "PENDING_CREATE"
			}
			if connection.GetProvisioningStatus() == vpc.PROVISIONINGSTATUS_ERROR {
				return common.VpcProvisioningStatusError
			}
			if connection.GetIgwId() != internetGatewayID {
				return "PENDING_CREATE"
			}
			return string(connection.GetProvisioningStatus())
		},
	)
}

func (r *routeTableIngressRoutingResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *routeTableIngressRoutingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	for _, v := range common.UuidValidator() {
		var validation validator.StringResponse
		v.ValidateString(ctx, validator.StringRequest{
			Path:        path.Root("route_table_id"),
			ConfigValue: types.StringValue(req.ID),
		}, &validation)
		resp.Diagnostics.Append(validation.Diagnostics...)
	}
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("route_table_id"), req.ID)...)
}

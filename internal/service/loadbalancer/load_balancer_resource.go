// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package loadbalancer

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
	"github.com/kakaoenterprise/kc-sdk-go/services/loadbalancer"
)

var (
	_ resource.Resource                = &loadBalancerResource{}
	_ resource.ResourceWithConfigure   = &loadBalancerResource{}
	_ resource.ResourceWithImportState = &loadBalancerResource{}
)

func NewLoadBalancerResource() resource.Resource {
	return &loadBalancerResource{}
}

type loadBalancerResource struct {
	kc *common.KakaoCloudClient
}

func (r *loadBalancerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *loadBalancerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_load_balancer"
}

func (r *loadBalancerResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a KakaoCloud Load Balancer.",
		Attributes: utils.MergeResourceSchemaAttributes(
			loadBalancerResourceSchemaAttributes,
			map[string]schema.Attribute{
				"timeouts": timeouts.AttributesAll(ctx),
			},
		),
	}
}

func (r *loadBalancerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan loadBalancerResourceModel
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

	// Check name uniqueness before creating
	if err := r.checkLoadBalancerNameExists(ctx, plan.Name.ValueString(), "", &resp.Diagnostics); err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("name"),
			"Load Balancer Name Conflict",
			fmt.Sprintf("A load balancer with name '%s' already exists. Please choose a different name.", plan.Name.ValueString()),
		)
		return
	}

	// Access logs can now be set during creation (write-only field)

	createReq := loadbalancer.CreateLoadBalancerModel{
		Name:             plan.Name.ValueString(),
		SubnetId:         plan.SubnetId.ValueString(),
		AvailabilityZone: loadbalancer.AvailabilityZone(plan.AvailabilityZone.ValueString()),
		FlavorId:         plan.FlavorId.ValueString(),
	}

	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		createReq.SetDescription(plan.Description.ValueString())
	}

	// Based on the pattern from vpc_resource.go
	body := loadbalancer.BodyCreateLoadBalancer{LoadBalancer: createReq}

	// Create Load Balancer
	lb, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*loadbalancer.BnsLoadBalancerV1ApiCreateLoadBalancerModelResponseLoadBalancerModel, *http.Response, error) {
			return r.kc.ApiClient.LoadBalancerAPI.CreateLoadBalancer(ctx).XAuthToken(r.kc.XAuthToken).BodyCreateLoadBalancer(body).Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "CreateLoadBalancer", err, &resp.Diagnostics)
		return
	}

	plan.Id = types.StringValue(lb.LoadBalancer.Id)

	result, ok := r.pollLoadBalancerUntilStatus(
		ctx,
		plan.Id.ValueString(),
		[]string{ProvisioningStatusActive, ProvisioningStatusError},
		&resp.Diagnostics,
	)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	common.CheckResourceAvailableStatus(ctx, r, (*string)(result.ProvisioningStatus.Get()), []string{ProvisioningStatusActive}, &resp.Diagnostics)

	ok = mapLoadBalancer(ctx, &plan.loadBalancerBaseModel, result, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	// If access logs are specified, update them after creation
	if !plan.AccessLogs.IsNull() && !plan.AccessLogs.IsUnknown() {
		// Update access logs using the separate API endpoint
		var accessLog accessLogModel
		diags := plan.AccessLogs.As(ctx, &accessLog, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		// Validate that all required fields are provided
		if accessLog.Bucket.IsNull() || accessLog.Bucket.IsUnknown() || accessLog.Bucket.ValueString() == "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("access_logs").AtName("bucket"),
				"Missing Required Field",
				"Bucket is required for access logs configuration",
			)
			return
		}

		if accessLog.AccessKey.IsNull() || accessLog.AccessKey.IsUnknown() || accessLog.AccessKey.ValueString() == "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("access_logs").AtName("access_key"),
				"Missing Required Field",
				"Access key is required for access logs configuration",
			)
			return
		}

		if accessLog.SecretKey.IsNull() || accessLog.SecretKey.IsUnknown() || accessLog.SecretKey.ValueString() == "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("access_logs").AtName("secret_key"),
				"Missing Required Field",
				"Secret key is required for access logs configuration",
			)
			return
		}

		accessLogReq := loadbalancer.EditLoadBalancerAccessLogModel{
			Bucket:    accessLog.Bucket.ValueString(),
			AccessKey: accessLog.AccessKey.ValueString(),
			SecretKey: accessLog.SecretKey.ValueString(),
		}

		body := loadbalancer.BodyUpdateAccessLog{AccessLogs: accessLogReq}

		_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
			func() (*loadbalancer.BnsLoadBalancerV1ApiUpdateAccessLogModelResponseLoadBalancerModel, *http.Response, error) {
				return r.kc.ApiClient.LoadBalancerAPI.UpdateAccessLog(ctx, plan.Id.ValueString()).XAuthToken(r.kc.XAuthToken).BodyUpdateAccessLog(body).Execute()
			},
		)

		if err != nil {
			common.AddApiActionError(ctx, r, httpResp, "UpdateLoadBalancerAccessLog", err, &resp.Diagnostics)
			return
		}

		// Wait for the load balancer to become active again
		finalResult, ok := r.pollLoadBalancerUntilStatus(
			ctx,
			plan.Id.ValueString(),
			[]string{ProvisioningStatusActive, ProvisioningStatusError},
			&resp.Diagnostics,
		)
		if !ok || resp.Diagnostics.HasError() {
			return
		}

		common.CheckResourceAvailableStatus(ctx, r, (*string)(finalResult.ProvisioningStatus.Get()), []string{ProvisioningStatusActive}, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}

		// Update the state with the final result, but preserve access logs from plan
		ok = mapLoadBalancer(ctx, &plan.loadBalancerBaseModel, finalResult, &resp.Diagnostics)
		if !ok || resp.Diagnostics.HasError() {
			return
		}

		// Preserve the access logs from the plan (write-only field)
		// Access logs are already set in plan, no need to reassign
	} else {
		// If access logs are not specified, ensure they are set to null
		plan.AccessLogs = types.ObjectNull(accessLogAttrType)
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *loadBalancerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state loadBalancerResourceModel
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
		func() (*loadbalancer.BnsLoadBalancerV1ApiGetLoadBalancerModelResponseLoadBalancerModel, *http.Response, error) {
			return r.kc.ApiClient.LoadBalancerAPI.
				GetLoadBalancer(ctx, state.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				Execute()
		},
	)

	if httpResp != nil && httpResp.StatusCode == 404 {
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "GetLoadBalancer", err, &resp.Diagnostics)
		return
	}
	// Preserve access logs from state before mapping API response (write-only field)
	preservedAccessLogs := state.AccessLogs

	loadBalancerResult := respModel.LoadBalancer
	ok := mapLoadBalancer(ctx, &state.loadBalancerBaseModel, &loadBalancerResult, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	// Restore access logs from state (write-only field)
	state.AccessLogs = preservedAccessLogs

	if state.FlavorId.IsNull() {
		lbfs, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
			func() (*loadbalancer.FlavorListModel, *http.Response, error) {
				resp, httpResp, err := r.kc.ApiClient.LoadBalancerEtcAPI.ListLoadBalancerTypes(ctx).XAuthToken(r.kc.XAuthToken).Execute()
				return resp, httpResp, err
			},
		)
		if err != nil {
			common.AddApiActionError(ctx, r, httpResp, "ListLoadBalancerTypes", err, &resp.Diagnostics)
			return
		}

		for _, lbf := range lbfs.Flavors {
			if lbf.Name.Get() != nil && *lbf.Name.Get() == state.Type.ValueString() {
				state.FlavorId = types.StringValue(lbf.Id)
				break
			}
		}
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *loadBalancerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state loadBalancerResourceModel
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

	// Only proceed if one of the updatable attributes has changed
	if !plan.Name.Equal(state.Name) || !plan.Description.Equal(state.Description) {
		// Check name uniqueness if name is being changed
		if !plan.Name.Equal(state.Name) {
			if err := r.checkLoadBalancerNameExists(ctx, plan.Name.ValueString(), state.Id.ValueString(), &resp.Diagnostics); err != nil {
				resp.Diagnostics.AddAttributeError(
					path.Root("name"),
					"Load Balancer Name Conflict",
					fmt.Sprintf("A load balancer with name '%s' already exists. Please choose a different name.", plan.Name.ValueString()),
				)
				return
			}
		}

		timeout, diags := plan.Timeouts.Update(ctx, common.DefaultUpdateTimeout)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		// Create an empty request model
		editReq := loadbalancer.EditLoadBalancerModel{}

		// Conditionally set Name ONLY if it has changed
		if !plan.Name.Equal(state.Name) {
			if !plan.Name.IsNull() && !plan.Name.IsUnknown() {
				editReq.SetName(plan.Name.ValueString())
			}
		}

		// Conditionally set Description ONLY if it has changed
		if !plan.Description.Equal(state.Description) {
			if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
				editReq.SetDescription(plan.Description.ValueString())
			} else {
				// Handle case where description is being cleared
				editReq.SetDescription("")
			}
		}

		body := *loadbalancer.NewBodyUpdateLoadBalancer(editReq)

		_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
			func() (*loadbalancer.BnsLoadBalancerV1ApiUpdateLoadBalancerModelResponseLoadBalancerModel, *http.Response, error) {
				return r.kc.ApiClient.LoadBalancerAPI.
					UpdateLoadBalancer(ctx, state.Id.ValueString()).
					XAuthToken(r.kc.XAuthToken).
					BodyUpdateLoadBalancer(body).
					Execute()
			},
		)

		if err != nil {
			common.AddApiActionError(ctx, r, httpResp, "UpdateLoadBalancer", err, &resp.Diagnostics)
			return
		}

		// Wait for the load balancer to become active again
		result, ok := r.pollLoadBalancerUntilStatus(
			ctx,
			state.Id.ValueString(),
			[]string{ProvisioningStatusActive, ProvisioningStatusError},
			&resp.Diagnostics,
		)
		if !ok || resp.Diagnostics.HasError() {
			return
		}

		common.CheckResourceAvailableStatus(ctx, r, (*string)(result.ProvisioningStatus.Get()), []string{ProvisioningStatusActive}, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}

		ok = mapLoadBalancer(ctx, &state.loadBalancerBaseModel, result, &resp.Diagnostics)
		if !ok || resp.Diagnostics.HasError() {
			return
		}
	}

	// Handle access logs update if changed
	if !plan.AccessLogs.Equal(state.AccessLogs) {
		if !plan.AccessLogs.IsNull() && !plan.AccessLogs.IsUnknown() {
			// Update access logs using the separate API endpoint
			var accessLog accessLogModel
			diags := plan.AccessLogs.As(ctx, &accessLog, basetypes.ObjectAsOptions{})
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}

			// Validate that all required fields are provided
			if accessLog.Bucket.IsNull() || accessLog.Bucket.IsUnknown() || accessLog.Bucket.ValueString() == "" {
				resp.Diagnostics.AddAttributeError(
					path.Root("access_logs").AtName("bucket"),
					"Missing Required Field",
					"Bucket is required for access logs configuration",
				)
				return
			}

			if accessLog.AccessKey.IsNull() || accessLog.AccessKey.IsUnknown() || accessLog.AccessKey.ValueString() == "" {
				resp.Diagnostics.AddAttributeError(
					path.Root("access_logs").AtName("access_key"),
					"Missing Required Field",
					"Access key is required for access logs configuration",
				)
				return
			}

			if accessLog.SecretKey.IsNull() || accessLog.SecretKey.IsUnknown() || accessLog.SecretKey.ValueString() == "" {
				resp.Diagnostics.AddAttributeError(
					path.Root("access_logs").AtName("secret_key"),
					"Missing Required Field",
					"Secret key is required for access logs configuration",
				)
				return
			}

			accessLogReq := loadbalancer.EditLoadBalancerAccessLogModel{
				Bucket:    accessLog.Bucket.ValueString(),
				AccessKey: accessLog.AccessKey.ValueString(),
				SecretKey: accessLog.SecretKey.ValueString(),
			}

			body := loadbalancer.BodyUpdateAccessLog{AccessLogs: accessLogReq}

			_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
				func() (*loadbalancer.BnsLoadBalancerV1ApiUpdateAccessLogModelResponseLoadBalancerModel, *http.Response, error) {
					return r.kc.ApiClient.LoadBalancerAPI.UpdateAccessLog(ctx, state.Id.ValueString()).XAuthToken(r.kc.XAuthToken).BodyUpdateAccessLog(body).Execute()
				},
			)

			if err != nil {
				common.AddApiActionError(ctx, r, httpResp, "UpdateLoadBalancerAccessLog", err, &resp.Diagnostics)
				return
			}

			// Wait for the load balancer to become active again
			result, ok := r.pollLoadBalancerUntilStatus(
				ctx,
				state.Id.ValueString(),
				[]string{ProvisioningStatusActive, ProvisioningStatusError},
				&resp.Diagnostics,
			)
			if !ok || resp.Diagnostics.HasError() {
				return
			}

			common.CheckResourceAvailableStatus(ctx, r, (*string)(result.ProvisioningStatus.Get()), []string{ProvisioningStatusActive}, &resp.Diagnostics)
			if resp.Diagnostics.HasError() {
				return
			}

			// Update the state with the new result, but preserve the access logs from the plan
			ok = mapLoadBalancer(ctx, &state.loadBalancerBaseModel, result, &resp.Diagnostics)
			if !ok || resp.Diagnostics.HasError() {
				return
			}

			// Set the access logs in the state to match what was provided in the plan
			state.AccessLogs = plan.AccessLogs
		} else if plan.AccessLogs.IsNull() {
			// Handle case where access_logs is explicitly set to null (disable access logs)
			// Try to disable access logs by sending an empty request body
			// This might work if the API supports it (similar to DetachPublicIpFromLoadBalancer)

			// Try to disable access logs by sending empty values
			// This should work with PATCH API to explicitly remove access logs
			accessLogReq := loadbalancer.EditLoadBalancerAccessLogModel{
				Bucket:    "",
				AccessKey: "",
				SecretKey: "",
			}
			body := loadbalancer.BodyUpdateAccessLog{AccessLogs: accessLogReq}

			_, _, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
				func() (*loadbalancer.BnsLoadBalancerV1ApiUpdateAccessLogModelResponseLoadBalancerModel, *http.Response, error) {
					return r.kc.ApiClient.LoadBalancerAPI.UpdateAccessLog(ctx, state.Id.ValueString()).XAuthToken(r.kc.XAuthToken).BodyUpdateAccessLog(body).Execute()
				},
			)

			if err != nil {
				// If the API doesn't support disabling access logs, add a warning but don't fail
				resp.Diagnostics.AddWarning(
					"Access Logs Disable Not Supported",
					"The API does not support disabling access logs. Access logs will remain configured.",
				)
				// Don't return here - continue with the update
			} else {
				// Wait for the load balancer to become active again
				result, ok := r.pollLoadBalancerUntilStatus(
					ctx,
					state.Id.ValueString(),
					[]string{ProvisioningStatusActive, ProvisioningStatusError},
					&resp.Diagnostics,
				)
				if !ok || resp.Diagnostics.HasError() {
					return
				}

				common.CheckResourceAvailableStatus(ctx, r, (*string)(result.ProvisioningStatus.Get()), []string{ProvisioningStatusActive}, &resp.Diagnostics)
				if resp.Diagnostics.HasError() {
					return
				}

				// Update the state with the new result
				ok = mapLoadBalancer(ctx, &state.loadBalancerBaseModel, result, &resp.Diagnostics)
				if !ok || resp.Diagnostics.HasError() {
					return
				}
			}

			// Set the access logs in the state to null
			state.AccessLogs = plan.AccessLogs
		}
		// Note: If plan.AccessLogs.IsUnknown(), we preserve the existing state (handled by plan modifier)
	}

	state.Timeouts = plan.Timeouts
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *loadBalancerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state loadBalancerResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (struct{}, *http.Response, error) {
			httpResp, err := r.kc.ApiClient.LoadBalancerAPI.
				DeleteLoadBalancer(ctx, state.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				Execute()
			return struct{}{}, httpResp, err
		},
	)

	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "DeleteLoadBalancer", err, &resp.Diagnostics)
		return
	}

	common.PollUntilDeletion(ctx, r, 5*time.Second, &resp.Diagnostics, func(ctx context.Context) (bool, *http.Response, error) {
		_, httpResp, err := r.kc.ApiClient.LoadBalancerAPI.
			GetLoadBalancer(ctx, state.Id.ValueString()).
			XAuthToken(r.kc.XAuthToken).
			Execute()
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			return true, httpResp, nil
		}
		return false, httpResp, err
	})
}

func (r *loadBalancerResource) pollLoadBalancerUntilStatus(
	ctx context.Context,
	loadBalancerId string,
	targetStatuses []string,
	resp *diag.Diagnostics,
) (*loadbalancer.BnsLoadBalancerV1ApiGetLoadBalancerModelLoadBalancerModel, bool) {
	return common.PollUntilResult(
		ctx,
		r,
		5*time.Second,
		targetStatuses,
		resp,
		func(ctx context.Context) (*loadbalancer.BnsLoadBalancerV1ApiGetLoadBalancerModelLoadBalancerModel, *http.Response, error) {
			respModel, httpResp, err := r.kc.ApiClient.LoadBalancerAPI.
				GetLoadBalancer(ctx, loadBalancerId).
				XAuthToken(r.kc.XAuthToken).
				Execute()
			if err != nil {
				return nil, httpResp, err
			}
			return &respModel.LoadBalancer, httpResp, nil
		},
		func(lb *loadbalancer.BnsLoadBalancerV1ApiGetLoadBalancerModelLoadBalancerModel) string {
			return string(*lb.ProvisioningStatus.Get())
		},
	)
}

func (r *loadBalancerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *loadBalancerResource) checkLoadBalancerNameExists(ctx context.Context, name string, currentId string, diags *diag.Diagnostics) error {
	// List all load balancers and check if name exists
	lbs, _, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, diags,
		func() (*loadbalancer.LoadBalancerListModel, *http.Response, error) {
			return r.kc.ApiClient.LoadBalancerAPI.ListLoadBalancers(ctx).XAuthToken(r.kc.XAuthToken).Execute()
		},
	)
	if err != nil {
		return err
	}

	for _, lb := range lbs.LoadBalancers {
		if lb.Name.IsSet() && *lb.Name.Get() == name {
			// If this is an update operation and the name belongs to the current resource, skip it
			if currentId != "" && lb.Id == currentId {
				continue
			}
			return fmt.Errorf("load balancer with name '%s' already exists", name)
		}
	}

	return nil
}

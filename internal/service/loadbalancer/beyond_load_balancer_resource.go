// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package loadbalancer

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"sync"
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
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/kakaoenterprise/kc-sdk-go/services/loadbalancer"
)

var (
	_ resource.ResourceWithConfigure   = &beyondLoadBalancerResource{}
	_ resource.ResourceWithImportState = &beyondLoadBalancerResource{}
)

func NewBeyondLoadBalancerResource() resource.Resource {
	return &beyondLoadBalancerResource{}
}

type beyondLoadBalancerResource struct {
	kc *common.KakaoCloudClient
}

func (r *beyondLoadBalancerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_beyond_load_balancer"
}

func (r *beyondLoadBalancerResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: docs.GetResourceDescription("BeyondLoadBalancer"),
		Attributes: utils.MergeResourceSchemaAttributes(
			beyondLoadBalancerResourceSchema,
			map[string]schema.Attribute{
				"timeouts": timeouts.AttributesAll(ctx),
			},
		),
	}
}

func (r *beyondLoadBalancerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan beyondLoadBalancerResourceModel
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

	createReq := loadbalancer.BnsLoadBalancerV1ApiCreateHaGroupModelCreateBeyondLoadBalancerModel{
		Name:   plan.Name.ValueString(),
		TypeId: plan.TypeId.ValueString(),
		Scheme: loadbalancer.Scheme(plan.Scheme.ValueString()),
		VpcId:  plan.VpcId.ValueString(),
	}

	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		createReq.SetDescription(plan.Description.ValueString())
	}

	var mutexes []*sync.Mutex
	if !plan.AttachedLoadBalancers.IsNull() && !plan.AttachedLoadBalancers.IsUnknown() {
		lbList, planDiags := r.convertSetToBlbLoadBalancerModel(ctx, plan.AttachedLoadBalancers)
		resp.Diagnostics.Append(planDiags...)
		if resp.Diagnostics.HasError() {
			return
		}

		for _, lb := range lbList {
			mutex := common.LockForID(lb.Id.ValueString())
			mutex.Lock()
			mutexes = append(mutexes, mutex)
		}
		defer func() {

			for i := len(mutexes) - 1; i >= 0; i-- {
				mutexes[i].Unlock()
			}
		}()

		for _, lb := range lbList {
			lbResult, lbOk := r.pollLoadBalancerUntilStatus(
				ctx,
				lb.Id.ValueString(),
				[]string{ProvisioningStatusActive, ProvisioningStatusError},
				&resp.Diagnostics,
			)
			if !lbOk || resp.Diagnostics.HasError() {
				return
			}
			if lbResult != nil && string(*lbResult.ProvisioningStatus.Get()) == ProvisioningStatusError {
				resp.Diagnostics.AddError("Load Balancer Error", fmt.Sprintf("Load balancer %s is in error state", lb.Id.ValueString()))
				return
			}
		}

		var lbs []loadbalancer.BnsLoadBalancerV1ApiCreateHaGroupModelSubnetModel
		for _, lb := range lbList {
			tmpLb := loadbalancer.BnsLoadBalancerV1ApiCreateHaGroupModelSubnetModel{
				LoadBalancerId:   lb.Id.ValueString(),
				AvailabilityZone: loadbalancer.AvailabilityZone(lb.AvailabilityZone.ValueString()),
			}
			lbs = append(lbs, tmpLb)
		}
		createReq.SetSubnets(lbs)
	}

	body := *loadbalancer.NewBodyCreateHaGroup(createReq)

	respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*loadbalancer.BnsLoadBalancerV1ApiCreateHaGroupModelResponseBeyondLoadBalancerModel, *http.Response, error) {
			return r.kc.ApiClient.BeyondLoadBalancerAPI.CreateHaGroup(ctx).
				XAuthToken(r.kc.XAuthToken).BodyCreateHaGroup(body).Execute()
		},
	)

	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "CreateHaGroup", err, &resp.Diagnostics)
		return
	}

	plan.Id = types.StringValue(respModel.BeyondLoadBalancer.Id)

	result, ok := r.pollBeyondLoadBalancerUntilStatus(
		ctx,
		plan.Id.ValueString(),
		[]string{ProvisioningStatusActive, ProvisioningStatusError},
		&resp.Diagnostics,
	)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	if result != nil && result.LoadBalancers != nil {
		for _, lb := range result.LoadBalancers {
			lbResult, lbOk := r.pollLoadBalancerUntilStatus(
				ctx,
				lb.Id,
				[]string{ProvisioningStatusActive, ProvisioningStatusError},
				&resp.Diagnostics,
			)
			if !lbOk || resp.Diagnostics.HasError() {
				return
			}
			if lbResult != nil && string(*lbResult.ProvisioningStatus.Get()) == ProvisioningStatusError {
				resp.Diagnostics.AddError("Load Balancer Error", fmt.Sprintf("Load balancer %s is in error state", lb.Id))
				return
			}
		}
	}

	common.CheckResourceAvailableStatus(ctx, r, (*string)(result.ProvisioningStatus.Get()), []string{ProvisioningStatusActive}, &resp.Diagnostics)

	ok = mapBeyondLoadBalancerBaseModel(ctx, &plan.beyondLoadBalancerBaseModel, result, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *beyondLoadBalancerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state beyondLoadBalancerResourceModel
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
		func() (*loadbalancer.BnsLoadBalancerV1ApiGetHaGroupModelResponseBeyondLoadBalancerModel, *http.Response, error) {
			return r.kc.ApiClient.BeyondLoadBalancerAPI.GetHaGroup(ctx, state.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).Execute()
		},
	)

	if httpResp != nil && httpResp.StatusCode == 404 {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "GetHaGroup", err, &resp.Diagnostics)
		return
	}

	ok := mapBeyondLoadBalancerBaseModel(ctx, &state.beyondLoadBalancerBaseModel, &respModel.BeyondLoadBalancer, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	if state.AttachedLoadBalancers.IsNull() && !state.LoadBalancers.IsNull() {
		var lbs []blbLoadBalancerModel
		diags := state.LoadBalancers.ElementsAs(ctx, &lbs, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		sort.Slice(lbs, func(i, j int) bool {
			if lbs[i].AvailabilityZone.ValueString() < lbs[j].AvailabilityZone.ValueString() {
				return true
			}
			if lbs[i].AvailabilityZone.ValueString() > lbs[j].AvailabilityZone.ValueString() {
				return false
			}
			return lbs[i].AvailabilityZone.ValueString() < lbs[j].AvailabilityZone.ValueString()
		})

		var attachedLoadBalancers []attachedLoadBalancerModel
		for _, lb := range lbs {
			attachedLoadBalancers = append(attachedLoadBalancers,
				attachedLoadBalancerModel{
					Id:               lb.Id,
					AvailabilityZone: lb.AvailabilityZone,
				})
		}

		elemType := types.ObjectType{AttrTypes: attachedLoadBalancerAttrType}
		state.AttachedLoadBalancers, diags = types.SetValueFrom(ctx, elemType, attachedLoadBalancers)
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

func (r *beyondLoadBalancerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state beyondLoadBalancerResourceModel
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

	if !plan.Description.Equal(state.Description) && !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		var editReq loadbalancer.EditBeyondLoadBalancerModel
		editReq.SetDescription(plan.Description.ValueString())

		body := *loadbalancer.NewBodyUpdateHaGroup(editReq)

		_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
			func() (interface{}, *http.Response, error) {
				return r.kc.ApiClient.BeyondLoadBalancerAPI.UpdateHaGroup(ctx, plan.Id.ValueString()).
					XAuthToken(r.kc.XAuthToken).
					BodyUpdateHaGroup(body).
					Execute()
			},
		)

		if err != nil {
			common.AddApiActionError(ctx, r, httpResp, "UpdateHaGroup", err, &resp.Diagnostics)
			return
		}

		result, ok := r.pollBeyondLoadBalancerUntilStatus(
			ctx,
			plan.Id.ValueString(),
			[]string{ProvisioningStatusActive, ProvisioningStatusError},
			&resp.Diagnostics,
		)
		if !ok || resp.Diagnostics.HasError() {
			return
		}

		if result != nil && result.LoadBalancers != nil {
			for _, lb := range result.LoadBalancers {
				lbResult, lbOk := r.pollLoadBalancerUntilStatus(
					ctx,
					lb.Id,
					[]string{ProvisioningStatusActive, ProvisioningStatusError},
					&resp.Diagnostics,
				)
				if !lbOk || resp.Diagnostics.HasError() {
					return
				}
				if lbResult != nil && string(*lbResult.ProvisioningStatus.Get()) == ProvisioningStatusError {
					resp.Diagnostics.AddError("Load Balancer Error", fmt.Sprintf("Load balancer %s is in error state", lb.Id))
					return
				}
			}
		}

		common.CheckResourceAvailableStatus(ctx, r, (*string)(result.ProvisioningStatus.Get()), []string{ProvisioningStatusActive}, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if !plan.AttachedLoadBalancers.Equal(state.AttachedLoadBalancers) {
		planList, planDiags := r.convertSetToBlbLoadBalancerModel(ctx, plan.AttachedLoadBalancers)
		resp.Diagnostics.Append(planDiags...)
		if resp.Diagnostics.HasError() {
			return
		}

		var mutexes []*sync.Mutex
		for _, lb := range planList {
			mutex := common.LockForID(lb.Id.ValueString())
			mutex.Lock()
			mutexes = append(mutexes, mutex)
		}
		defer func() {

			for i := len(mutexes) - 1; i >= 0; i-- {
				mutexes[i].Unlock()
			}
		}()

		for _, lb := range planList {
			lbResult, lbOk := r.pollLoadBalancerUntilStatus(
				ctx,
				lb.Id.ValueString(),
				[]string{ProvisioningStatusActive, ProvisioningStatusError},
				&resp.Diagnostics,
			)
			if !lbOk || resp.Diagnostics.HasError() {
				return
			}
			if lbResult != nil && string(*lbResult.ProvisioningStatus.Get()) == ProvisioningStatusError {
				resp.Diagnostics.AddError("Load Balancer Error", fmt.Sprintf("Load balancer %s is in error state", lb.Id.ValueString()))
				return
			}
		}

		if !r.updateLoadBalancers(ctx, plan.Id.ValueString(), &planList, &resp.Diagnostics) {
			return
		}

		result, ok := r.pollBeyondLoadBalancerUntilStatus(
			ctx,
			plan.Id.ValueString(),
			[]string{ProvisioningStatusActive, ProvisioningStatusError},
			&resp.Diagnostics,
		)
		if !ok || resp.Diagnostics.HasError() {
			return
		}

		if result != nil && result.LoadBalancers != nil {
			for _, lb := range result.LoadBalancers {
				lbResult, lbOk := r.pollLoadBalancerUntilStatus(
					ctx,
					lb.Id,
					[]string{ProvisioningStatusActive, ProvisioningStatusError},
					&resp.Diagnostics,
				)
				if !lbOk || resp.Diagnostics.HasError() {
					return
				}
				if lbResult != nil && string(*lbResult.ProvisioningStatus.Get()) == ProvisioningStatusError {
					resp.Diagnostics.AddError("Load Balancer Error", fmt.Sprintf("Load balancer %s is in error state", lb.Id))
					return
				}
			}
		}

		common.CheckResourceAvailableStatus(ctx, r, (*string)(result.ProvisioningStatus.Get()), []string{ProvisioningStatusActive}, &resp.Diagnostics)

		ok = mapBeyondLoadBalancerBaseModel(ctx, &plan.beyondLoadBalancerBaseModel, result, &resp.Diagnostics)
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

func (r *beyondLoadBalancerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state beyondLoadBalancerResourceModel
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

	mutex := common.LockForID(state.Id.ValueString())
	mutex.Lock()
	defer mutex.Unlock()

	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (interface{}, *http.Response, error) {
			httpResp, err := r.kc.ApiClient.BeyondLoadBalancerAPI.DeleteHaGroup(ctx, state.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				Execute()
			return nil, httpResp, err
		},
	)

	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			return
		}
		common.AddApiActionError(ctx, r, httpResp, "DeleteHaGroup", err, &resp.Diagnostics)
		return
	}

	common.PollUntilDeletion(ctx, r, 2*time.Second, &resp.Diagnostics, func(ctx context.Context) (bool, *http.Response, error) {
		_, httpResp, err := r.kc.ApiClient.BeyondLoadBalancerAPI.
			GetHaGroup(ctx, state.Id.ValueString()).
			XAuthToken(r.kc.XAuthToken).
			Execute()

		if httpResp != nil && httpResp.StatusCode == 404 {
			return true, httpResp, nil
		}
		return false, httpResp, err
	})
}

func (r *beyondLoadBalancerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {

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

func (r *beyondLoadBalancerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {

	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *beyondLoadBalancerResource) pollBeyondLoadBalancerUntilStatus(
	ctx context.Context,
	blbId string,
	targetStatuses []string,
	resp *diag.Diagnostics,
) (*loadbalancer.BnsLoadBalancerV1ApiGetHaGroupModelBeyondLoadBalancerModel, bool) {
	return common.PollUntilResult(
		ctx,
		r,
		2*time.Second,
		targetStatuses,
		resp,
		func(ctx context.Context) (*loadbalancer.BnsLoadBalancerV1ApiGetHaGroupModelBeyondLoadBalancerModel, *http.Response, error) {
			respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, resp,
				func() (*loadbalancer.BnsLoadBalancerV1ApiGetHaGroupModelResponseBeyondLoadBalancerModel, *http.Response, error) {
					return r.kc.ApiClient.BeyondLoadBalancerAPI.
						GetHaGroup(ctx, blbId).
						XAuthToken(r.kc.XAuthToken).
						Execute()
				},
			)
			if err != nil {
				return nil, httpResp, err
			}
			return &respModel.BeyondLoadBalancer, httpResp, nil
		},
		func(v *loadbalancer.BnsLoadBalancerV1ApiGetHaGroupModelBeyondLoadBalancerModel) string {
			return string(*v.ProvisioningStatus.Get())
		},
	)
}

func (r *beyondLoadBalancerResource) pollLoadBalancerUntilStatus(
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
			respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, resp,
				func() (*loadbalancer.BnsLoadBalancerV1ApiGetLoadBalancerModelResponseLoadBalancerModel, *http.Response, error) {
					return r.kc.ApiClient.LoadBalancerAPI.
						GetLoadBalancer(ctx, loadBalancerId).
						XAuthToken(r.kc.XAuthToken).
						Execute()
				},
			)
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

func (r *beyondLoadBalancerResource) convertSetToBlbLoadBalancerModel(
	ctx context.Context,
	set types.Set,
) ([]attachedLoadBalancerModel, diag.Diagnostics) {
	var result []attachedLoadBalancerModel
	var diags diag.Diagnostics

	for _, elem := range set.Elements() {
		if obj, ok := elem.(types.Object); ok {
			var model attachedLoadBalancerModel
			elemDiags := obj.As(ctx, &model, basetypes.ObjectAsOptions{})
			diags.Append(elemDiags...)
			result = append(result, model)
		}
	}
	return result, diags
}

func (r *beyondLoadBalancerResource) updateLoadBalancers(
	ctx context.Context,
	blbId string,
	planList *[]attachedLoadBalancerModel,
	diags *diag.Diagnostics,
) bool {
	var lbs []loadbalancer.BnsLoadBalancerV1ApiUpdateHaGroupLoadBalancerModelSubnetModel
	for _, lb := range *planList {
		tmpLb := loadbalancer.BnsLoadBalancerV1ApiUpdateHaGroupLoadBalancerModelSubnetModel{
			LoadBalancerId:   lb.Id.ValueString(),
			AvailabilityZone: loadbalancer.AvailabilityZone(lb.AvailabilityZone.ValueString()),
		}
		lbs = append(lbs, tmpLb)
	}

	body := loadbalancer.BodyUpdateHaGroupLoadBalancer{
		BeyondLoadBalancer: *loadbalancer.NewBnsLoadBalancerV1ApiUpdateHaGroupLoadBalancerModelCreateBeyondLoadBalancerModel(lbs),
	}

	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, diags,
		func() (interface{}, *http.Response, error) {
			return r.kc.ApiClient.BeyondLoadBalancerAPI.
				UpdateHaGroupLoadBalancer(ctx, blbId).
				XAuthToken(r.kc.XAuthToken).
				BodyUpdateHaGroupLoadBalancer(body).
				Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "UpdateHaGroupLoadBalancer", err, diags)
		return false
	}
	return true
}

func (r *beyondLoadBalancerResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config beyondLoadBalancerResourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	configLbList, planDiags := r.convertSetToBlbLoadBalancerModel(ctx, config.AttachedLoadBalancers)
	resp.Diagnostics.Append(planDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.validateAvailabilityZoneConfig(ctx, configLbList, resp)
	r.validateAttachedLoadBalancersConfig(ctx, configLbList, resp)
}

func (r *beyondLoadBalancerResource) validateAvailabilityZoneConfig(ctx context.Context, configLbList []attachedLoadBalancerModel, resp *resource.ValidateConfigResponse) {
	for _, configLb := range configLbList {
		common.ValidateAvailabilityZone(
			path.Root("availability_zone"),
			configLb.AvailabilityZone,
			r.kc,
			&resp.Diagnostics,
		)
	}
}

func (r *beyondLoadBalancerResource) validateAttachedLoadBalancersConfig(ctx context.Context, configLbList []attachedLoadBalancerModel, resp *resource.ValidateConfigResponse) {
	zoneMap := make(map[string]bool)

	for _, configLb := range configLbList {
		if !configLb.AvailabilityZone.IsUnknown() {
			if zoneMap[configLb.AvailabilityZone.ValueString()] {
				common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
					fmt.Sprintf("Duplicate Availability Zone: %s", configLb.AvailabilityZone))
				return
			}
			zoneMap[configLb.AvailabilityZone.ValueString()] = true
		}
	}
}

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package bcs

import (
	"fmt"
	"terraform-provider-kakaocloud/internal/common"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"golang.org/x/net/context"
)

func (r *instanceResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config instanceResourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.validateAvailabilityZoneConfig(config, resp)
	r.validateVolumesConfig(ctx, config, resp)
	r.validateSubnetsConfig(ctx, config, resp)
}

func (r *instanceResource) validateAvailabilityZoneConfig(config instanceResourceModel, resp *resource.ValidateConfigResponse) {
	common.ValidateAvailabilityZone(
		path.Root("availability_zone"),
		config.AvailabilityZone,
		r.kc,
		&resp.Diagnostics,
	)
}

func (r *instanceResource) validateVolumesConfig(ctx context.Context, config instanceResourceModel, resp *resource.ValidateConfigResponse) {
	configList, planDiags := r.convertListToInstanceVolumeModel(ctx, config.Volumes)
	resp.Diagnostics.Append(planDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	for _, tfv := range configList {
		if !tfv.Id.IsNull() {
			if !tfv.EncryptionSecretId.IsNull() || !tfv.TypeId.IsNull() || !tfv.Size.IsNull() || !tfv.ImageId.IsNull() {
				common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
					"Invalid Configuration: If the request volume ID is set, only is_delete_on_termination field can be specified.",
				)
			}
		} else {
			if tfv.Size.IsNull() {
				common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
					"Invalid Configuration: Volume size must be specified when no volume ID is provided.",
				)
			}
		}
	}
}

func (r *instanceResource) validateSubnetsConfig(ctx context.Context, config instanceResourceModel, resp *resource.ValidateConfigResponse) {
	configList, planDiags := r.convertListToInstanceSubnetModel(ctx, config.Subnets)
	resp.Diagnostics.Append(planDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if len(configList) == 0 {
		return
	}
	config1st := configList[0]
	if !config1st.NetworkInterfaceId.IsNull() {
		if !config.InitialSecurityGroups.IsNull() {
			common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
				"Invalid Configuration: First Subnet Network interface ID cannot be specified when security groups are specified.",
			)
		}
	}

	for _, config := range configList {
		if !config.NetworkInterfaceId.IsNull() && !config.PrivateIp.IsNull() {
			common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
				"Invalid Configuration: Network interface ID and private IP cannot be specified together.",
			)
		}
	}
}

func (r *instanceResource) validateVolumesUpdate(ctx context.Context, plan *instanceResourceModel, state *instanceResourceModel, resp *resource.ModifyPlanResponse) {
	if plan == nil || state == nil {
		return
	}
	if plan.Volumes.Equal(state.Volumes) || state.Volumes.IsNull() || plan.Volumes.IsNull() {
		return
	}

	planList, planDiags := r.convertListToInstanceVolumeModel(ctx, plan.Volumes)
	stateList, stateDiags := r.convertListToInstanceVolumeModel(ctx, state.Volumes)

	if len(planList) == 0 || len(stateList) == 0 {
		return
	}

	resp.Diagnostics.Append(planDiags...)
	resp.Diagnostics.Append(stateDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	stateMap := make(map[string]instanceVolumeModel)
	for _, s := range stateList {
		if !s.Id.IsNull() && !s.Id.IsUnknown() {
			stateMap[s.Id.ValueString()] = s
		}
	}

	plan1st := planList[0]
	state1st := stateList[0]

	if !plan1st.Id.Equal(state1st.Id) ||
		!plan1st.TypeId.Equal(state1st.TypeId) ||
		!plan1st.EncryptionSecretId.Equal(state1st.EncryptionSecretId) {
		common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
			"Invalid Configuration: Root volumes cannot be updated except for size and delete_on_termination.")
	}

	for _, planVol := range planList {

		if !planVol.Size.IsNull() && stateMap[planVol.Id.ValueString()].Size.IsNull() {
			common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
				"Invalid Configuration: Attached volumes cannot be updated size.")
			break
		}

		if planVol.Id.IsNull() {
			common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
				"Invalid Configuration: Adding volumes requires a volume ID.")
			break
		}
	}
}

func (r *instanceResource) validateSubnetsAndSecurityGroupsUpdate(ctx context.Context, plan *instanceResourceModel, state *instanceResourceModel, resp *resource.ModifyPlanResponse) {
	if plan == nil || state == nil {
		return
	}

	if !plan.InitialSecurityGroups.Equal(state.InitialSecurityGroups) {
		common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
			"Invalid Configuration: Changing the initial security group is not allowed.")
	}

	planList, planDiags := r.convertListToInstanceSubnetModel(ctx, plan.Subnets)
	stateList, stateDiags := r.convertListToInstanceSubnetModel(ctx, state.Subnets)
	resp.Diagnostics.Append(planDiags...)
	resp.Diagnostics.Append(stateDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	for _, pl := range planList {
		if pl.NetworkInterfaceId.IsNull() {
			common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
				"Invalid Configuration: Changing the subnet requires specifying a Network Interface ID.")
		}
	}

	if !planList[0].Id.Equal(stateList[0].Id) || !planList[0].NetworkInterfaceId.Equal(stateList[0].NetworkInterfaceId) ||
		!planList[0].PrivateIp.Equal(stateList[0].PrivateIp) {
		common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
			"Invalid Configuration: Changing the default subnet is not allowed.")
	}
}

func (r *instanceResource) validateBMUpdateRestrictions(ctx context.Context, plan *instanceResourceModel, state *instanceResourceModel, resp *resource.ModifyPlanResponse) {
	if plan == nil || state == nil {
		return
	}
	if plan.InstanceType.ValueString() != common.InstanceTypeBM {
		return
	}
	if !plan.Volumes.Equal(state.Volumes) {
		common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
			"Invalid Configuration: BM instances cannot modify attached volumes.")
	}
	if !plan.Subnets.Equal(state.Subnets) {
		common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
			"Invalid Configuration: BM instances cannot modify subnets.")
	}
	if !plan.FlavorId.Equal(state.FlavorId) {
		common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
			"Invalid Configuration: BM instances cannot change the flavor.")
	}
}

func (r *instanceResource) ModifyPlan(
	ctx context.Context,
	req resource.ModifyPlanRequest,
	resp *resource.ModifyPlanResponse,
) {
	var plan, state *instanceResourceModel

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

	if !plan.InstanceType.IsNull() && !plan.InstanceType.IsUnknown() && !plan.Status.IsNull() && !plan.Status.IsUnknown() {
		instanceType := plan.InstanceType.ValueString()
		status := plan.Status.ValueString()

		if instanceType == common.InstanceTypeBM && status == common.InstanceStatusShelved {
			common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
				fmt.Sprintf("Invalid Configuration: BM instances do not support the '%s' state.", common.InstanceStatusShelved),
			)
		}
	}

	if req.State.Raw.IsNull() && !req.Plan.Raw.IsNull() {
		return
	}

	if !req.State.Raw.IsNull() && !req.Plan.Raw.IsNull() {

		r.validateVolumesUpdate(ctx, plan, state, resp)

		r.validateSubnetsAndSecurityGroupsUpdate(ctx, plan, state, resp)

		r.validateBMUpdateRestrictions(ctx, plan, state, resp)
	}
}

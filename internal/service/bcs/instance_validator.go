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
					fmt.Sprintf("Invalid Configuration: If the request volume ID is set, only is_delete_on_termination field can be specified."),
				)
			}
		}
	}
}

// NetworkInterfaceId check
func (r *instanceResource) validateSubnetsConfig(ctx context.Context, config instanceResourceModel, resp *resource.ValidateConfigResponse) {
	configList, planDiags := r.convertListToInstanceSubnetModel(ctx, config.Subnets)
	resp.Diagnostics.Append(planDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	config1st := configList[0]
	if !config1st.NetworkInterfaceId.IsNull() {
		if !config.SecurityGroups.IsNull() {
			common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
				fmt.Sprintf("Invalid Configuration: First Subnet Network interface ID cannot be specified when security groups are specified."),
			)
		}
	}

	for i := 1; i < len(configList); i++ {
		tfv := configList[i]
		if tfv.NetworkInterfaceId.IsNull() {
			common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
				fmt.Sprintf("Invalid Configuration: Network interface ID is mandatory from the second configured subnet."),
			)
		}
	}

	for _, config := range configList {
		if !config.NetworkInterfaceId.IsNull() && !config.PrivateIp.IsNull() {
			common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
				fmt.Sprintf("Invalid Configuration: Network interface ID and private IP cannot be specified together."),
			)
		}
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

	// Delete: pass
	if req.Plan.Raw.IsNull() {
		return
	}

	// instance_type, status
	if !plan.InstanceType.IsNull() && !plan.InstanceType.IsUnknown() && !plan.Status.IsNull() && !plan.Status.IsUnknown() {
		instanceType := plan.InstanceType.ValueString()
		status := plan.Status.ValueString()

		if instanceType == common.InstanceTypeBM && status == common.InstanceStatusShelved {
			common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
				fmt.Sprintf("Invalid Configuration: BM instances do not support the '%s' state.", common.InstanceStatusShelved),
			)
		}
	}

	// Update
	if !req.State.Raw.IsNull() && !req.Plan.Raw.IsNull() {
		// Volumes
		if !plan.Volumes.Equal(state.Volumes) && !state.Volumes.IsNull() {
			planList, planDiags := r.convertListToInstanceVolumeModel(ctx, plan.Volumes)
			stateList, stateDiags := r.convertListToInstanceVolumeModel(ctx, state.Volumes)
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

			// Root Volumes: only size, delete_on_termination can be updated.
			if !plan1st.Id.Equal(state1st.Id) ||
				!plan1st.TypeId.Equal(state1st.TypeId) ||
				!plan1st.EncryptionSecretId.Equal(state1st.EncryptionSecretId) {
				common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
					fmt.Sprintf("Invalid Configuration: Root volumes cannot be updated except for size and delete_on_termination."))
			}

			for _, plan := range planList {
				// Attached Volumes: size can not be updated.
				if !plan.Size.IsNull() && stateMap[plan.Id.ValueString()].Size.IsNull() {
					common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
						fmt.Sprintf("Invalid Configuration: Attached volumes cannot be updated size."))
					break
				}
				// Add Volumes: only provide id and isDeleteOnTermination as inputs.
				if plan.Id.IsNull() {
					common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
						fmt.Sprintf("Invalid Configuration: Adding volumes requires a volume ID."))
					break
				}
			}
		}

		// SecurityGroups
		if !plan.SecurityGroups.IsUnknown() && !plan.SecurityGroups.Equal(state.SecurityGroups) {
			common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
				fmt.Sprintf("Invalid Configuration: Changing the security group is not allowed."))
		}
	}
}

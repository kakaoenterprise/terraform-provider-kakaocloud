// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package mysql

import (
	"context"
	"terraform-provider-kakaocloud/internal/common"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	mysqlsdk "github.com/kakaoenterprise/kc-sdk-go/services/mysql"
)

var customParameterOverrideSchemaAttributes = map[string]schema.Attribute{
	"key": schema.StringAttribute{
		Required: true,
		Validators: []validator.String{
			stringvalidator.LengthAtLeast(1),
		},
	},
	"value": schema.StringAttribute{
		Optional: true,
	},
}

var customParameterOverrideAttrTypes = map[string]attr.Type{
	"key":   types.StringType,
	"value": types.StringType,
}

var customParameterGroupResourceSchemaAttributes = map[string]schema.Attribute{
	"id": schema.StringAttribute{
		Computed: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	},
	"name": schema.StringAttribute{
		Required:   true,
		Validators: mysqlCustomParameterGroupNameValidator(),
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	},
	"source_parameter_group_id": schema.StringAttribute{
		Required:   true,
		Validators: common.UuidValidator(),
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	},
	"source_parameter_group_type": schema.StringAttribute{
		Required: true,
		Validators: []validator.String{
			stringvalidator.OneOf(
				string(mysqlsdk.PARAMETERGROUPTYPE_DEFAULT),
				string(mysqlsdk.PARAMETERGROUPTYPE_CUSTOM),
			),
		},
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	},
	"default_parameter_group_id": schema.StringAttribute{
		Computed: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	},
	"description": schema.StringAttribute{
		Optional:   true,
		Computed:   true,
		Validators: []validator.String{stringvalidator.UTF8LengthAtMost(100)},
	},
	"apply_mode": schema.StringAttribute{
		Optional: true,
		Validators: []validator.String{
			stringvalidator.OneOf("SEQUENTIAL", "PARALLEL"),
		},
	},

	"parameter_overrides": schema.SetNestedAttribute{
		Optional: true,
		Computed: true,
		PlanModifiers: []planmodifier.Set{
			setplanmodifier.UseStateForUnknown(),
		},
		NestedObject: schema.NestedAttributeObject{
			Attributes: customParameterOverrideSchemaAttributes,
		},
	},
	"engine_version": schema.StringAttribute{
		Computed: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	},
	"exist_error_sync": schema.BoolAttribute{
		Computed: true,
		PlanModifiers: []planmodifier.Bool{
			customParameterGroupUseStateForUnknownWithoutManagedChangeBool(),
		},
	},
	"instance_group_count": schema.Int32Attribute{
		Computed: true,
		PlanModifiers: []planmodifier.Int32{
			customParameterGroupUseStateForUnknownWithoutManagedChangeInt32(),
		},
	},
	"is_rollback_possible": schema.BoolAttribute{
		Computed: true,
		PlanModifiers: []planmodifier.Bool{
			customParameterGroupUseStateForUnknownWithoutManagedChangeBool(),
		},
	},
}

type customParameterGroupUseStateForUnknownWithoutManagedChangeBoolModifier struct{}

func customParameterGroupUseStateForUnknownWithoutManagedChangeBool() planmodifier.Bool {
	return customParameterGroupUseStateForUnknownWithoutManagedChangeBoolModifier{}
}

func (m customParameterGroupUseStateForUnknownWithoutManagedChangeBoolModifier) Description(_ context.Context) string {
	return "Use prior state for unknown values when no managed custom parameter group fields changed."
}

func (m customParameterGroupUseStateForUnknownWithoutManagedChangeBoolModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m customParameterGroupUseStateForUnknownWithoutManagedChangeBoolModifier) PlanModifyBool(
	ctx context.Context,
	req planmodifier.BoolRequest,
	resp *planmodifier.BoolResponse,
) {
	if req.State.Raw.IsNull() || !req.PlanValue.IsUnknown() || req.ConfigValue.IsUnknown() {
		return
	}
	if customParameterGroupHasManagedPlanChanges(ctx, req.Config, req.Plan, req.State, &resp.Diagnostics) {
		return
	}
	resp.PlanValue = req.StateValue
}

type customParameterGroupUseStateForUnknownWithoutManagedChangeInt32Modifier struct{}

func customParameterGroupUseStateForUnknownWithoutManagedChangeInt32() planmodifier.Int32 {
	return customParameterGroupUseStateForUnknownWithoutManagedChangeInt32Modifier{}
}

func (m customParameterGroupUseStateForUnknownWithoutManagedChangeInt32Modifier) Description(_ context.Context) string {
	return "Use prior state for unknown values when no managed custom parameter group fields changed."
}

func (m customParameterGroupUseStateForUnknownWithoutManagedChangeInt32Modifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m customParameterGroupUseStateForUnknownWithoutManagedChangeInt32Modifier) PlanModifyInt32(
	ctx context.Context,
	req planmodifier.Int32Request,
	resp *planmodifier.Int32Response,
) {
	if req.State.Raw.IsNull() || !req.PlanValue.IsUnknown() || req.ConfigValue.IsUnknown() {
		return
	}
	if customParameterGroupHasManagedPlanChanges(ctx, req.Config, req.Plan, req.State, &resp.Diagnostics) {
		return
	}
	resp.PlanValue = req.StateValue
}

func customParameterGroupHasManagedPlanChanges(ctx context.Context, config tfsdk.Config, plan tfsdk.Plan, state tfsdk.State, respDiags *diag.Diagnostics) bool {
	var configModel customParameterGroupResourceModel
	respDiags.Append(config.Get(ctx, &configModel)...)
	var planModel customParameterGroupResourceModel
	respDiags.Append(plan.Get(ctx, &planModel)...)
	var stateModel customParameterGroupResourceModel
	respDiags.Append(state.Get(ctx, &stateModel)...)
	if respDiags.HasError() {
		return true
	}

	return !planModel.Name.Equal(stateModel.Name) ||
		!planModel.SourceParameterGroupId.Equal(stateModel.SourceParameterGroupId) ||
		!planModel.SourceParameterGroupType.Equal(stateModel.SourceParameterGroupType) ||
		!planModel.Description.Equal(stateModel.Description) ||
		!customParameterGroupEquivalentParameterOverrides(configModel.ParameterOverrides, planModel.ParameterOverrides, stateModel.ParameterOverrides)
}

func customParameterGroupEquivalentParameterOverrides(config types.Set, plan types.Set, state types.Set) bool {

	if config.IsNull() {
		return true
	}

	if config.IsUnknown() || plan.IsUnknown() || state.IsUnknown() {
		return false
	}

	if isNullOrEmptySet(plan) && isNullOrEmptySet(state) {
		return true
	}
	return plan.Equal(state)
}

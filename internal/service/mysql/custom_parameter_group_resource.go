// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package mysql

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-kakaocloud/internal/common"
	"terraform-provider-kakaocloud/internal/utils"
	"time"

	resourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	mysqlsdk "github.com/kakaoenterprise/kc-sdk-go/services/mysql"
)

var (
	_ resource.Resource                = &customParameterGroupResource{}
	_ resource.ResourceWithConfigure   = &customParameterGroupResource{}
	_ resource.ResourceWithImportState = &customParameterGroupResource{}
)

func NewCustomParameterGroupResource() resource.Resource { return &customParameterGroupResource{} }

type customParameterGroupResource struct {
	kc *common.KakaoCloudClient
}

func (r *customParameterGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mysql_custom_parameter_group"
}

func (r *customParameterGroupResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":                          customParameterGroupResourceSchemaAttributes["id"],
			"name":                        customParameterGroupResourceSchemaAttributes["name"],
			"source_parameter_group_id":   customParameterGroupResourceSchemaAttributes["source_parameter_group_id"],
			"source_parameter_group_type": customParameterGroupResourceSchemaAttributes["source_parameter_group_type"],
			"default_parameter_group_id":  customParameterGroupResourceSchemaAttributes["default_parameter_group_id"],
			"description":                 customParameterGroupResourceSchemaAttributes["description"],
			"apply_mode":                  customParameterGroupResourceSchemaAttributes["apply_mode"],
			"parameter_overrides":         customParameterGroupResourceSchemaAttributes["parameter_overrides"],
			"engine_version":              customParameterGroupResourceSchemaAttributes["engine_version"],
			"exist_error_sync":            customParameterGroupResourceSchemaAttributes["exist_error_sync"],
			"instance_group_count":        customParameterGroupResourceSchemaAttributes["instance_group_count"],
			"is_rollback_possible":        customParameterGroupResourceSchemaAttributes["is_rollback_possible"],
			"timeouts":                    resourceTimeouts.AttributesAll(ctx),
		},
	}
}

func (r *customParameterGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *customParameterGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan customParameterGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
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

	state, ok := r.createCustomParameterGroup(ctx, plan, &resp.Diagnostics)
	if !ok {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *customParameterGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state customParameterGroupResourceModel
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

	refreshed, found, ok := r.readCustomParameterGroupState(ctx, state, &resp.Diagnostics)
	if !ok {
		return
	}
	if !found {
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &refreshed)...)
}

func (r *customParameterGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan customParameterGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	var state customParameterGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
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

	parameterUpdatesRequired := !plan.ParameterOverrides.Equal(state.ParameterOverrides)
	var parameterMap map[string]mysqlParameterModel
	if parameterUpdatesRequired {
		var ok bool
		parameterMap, ok = r.readCustomParameterGroupParameterMap(ctx, plan.Id.ValueString(), &resp.Diagnostics)
		if !ok {
			return
		}
	}

	resetOverrides, ok := r.buildResetParamsForDeletedParameterOverrides(ctx, state, plan, parameterMap, &resp.Diagnostics)
	if !ok {
		return
	}

	hasParameterUpdates := hasCustomParameterGroupParameterUpdates(state.ParameterOverrides, plan.ParameterOverrides, resetOverrides)
	descriptionChanged := !plan.Description.Equal(state.Description)
	if descriptionChanged || hasParameterUpdates {

		if !r.updateCustomParameterGroup(ctx, plan, parameterMap, resetOverrides, descriptionChanged, hasParameterUpdates, &resp.Diagnostics) {
			return
		}
	}

	refreshed, found, ok := r.readCustomParameterGroupState(ctx, plan, &resp.Diagnostics)
	if !ok {
		return
	}
	if !found {
		common.AddGeneralError(ctx, r, &resp.Diagnostics, "custom parameter group was not found after update")
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &refreshed)...)
}

func (r *customParameterGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state customParameterGroupResourceModel
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

	if !r.deleteCustomParameterGroup(ctx, state.Id.ValueString(), &resp.Diagnostics) {
		return
	}
}

func (r *customParameterGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *customParameterGroupResource) createCustomParameterGroup(
	ctx context.Context,
	plan customParameterGroupResourceModel,
	respDiags *diag.Diagnostics,
) (customParameterGroupResourceModel, bool) {
	group := mysqlsdk.NewMysqlV1ApiCreateMysqlCustomParameterGroupModelCustomParameterGroupRequestModel(
		plan.Name.ValueString(),
		plan.SourceParameterGroupId.ValueString(),
		mysqlsdk.ParameterGroupType(plan.SourceParameterGroupType.ValueString()),
	)
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		group.SetDescription(plan.Description.ValueString())
	}

	request := mysqlsdk.NewBodyCreateMysqlCustomParameterGroup(*group)
	modelResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, respDiags,
		func() (*mysqlsdk.CreateMySQLCustomParameterGroupResponseModel, *http.Response, error) {
			return r.kc.ApiClient.MySQLCustomParameterGroupsAPI.
				CreateMysqlCustomParameterGroup(ctx).
				XAuthToken(r.kc.XAuthToken).
				BodyCreateMysqlCustomParameterGroup(*request).
				Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "CreateMysqlCustomParameterGroup", err, respDiags)
		return customParameterGroupResourceModel{}, false
	}

	current := plan
	current.Id = utils.ConvertNullableString(modelResp.CustomParameterGroup.Id)

	hasParameterUpdates, ok := r.hasConfiguredParameterOverrides(ctx, plan.ParameterOverrides, respDiags)
	if !ok {
		r.rollbackCreatedCustomParameterGroup(ctx, current.Id.ValueString(), respDiags)
		return customParameterGroupResourceModel{}, false
	}
	if hasParameterUpdates {
		parameterMap, ok := r.readCustomParameterGroupParameterMap(ctx, current.Id.ValueString(), respDiags)
		if !ok {
			r.rollbackCreatedCustomParameterGroup(ctx, current.Id.ValueString(), respDiags)
			return customParameterGroupResourceModel{}, false
		}
		if !r.updateCustomParameterGroup(ctx, current, parameterMap, nil, false, true, respDiags) {
			r.rollbackCreatedCustomParameterGroup(ctx, current.Id.ValueString(), respDiags)
			return customParameterGroupResourceModel{}, false
		}
	}

	state, found, ok := r.readCustomParameterGroupState(ctx, current, respDiags)
	if !ok || !found {
		return customParameterGroupResourceModel{}, false
	}
	return state, true
}

func (r *customParameterGroupResource) readCustomParameterGroupState(
	ctx context.Context,
	current customParameterGroupResourceModel,
	respDiags *diag.Diagnostics,
) (customParameterGroupResourceModel, bool, bool) {
	modelResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, respDiags,
		func() (*mysqlsdk.GetMySQLCustomParameterGroupResponseModel, *http.Response, error) {
			return r.kc.ApiClient.MySQLCustomParameterGroupsAPI.
				GetMysqlCustomParameterGroup(ctx, current.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				Execute()
		},
	)
	if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
		return customParameterGroupResourceModel{}, false, true
	}
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "GetMysqlCustomParameterGroup", err, respDiags)
		return customParameterGroupResourceModel{}, false, false
	}

	mapped, ok := toCustomParameterGroupSingleModel(ctx, modelResp.CustomParameterGroup, respDiags)
	if !ok {
		return customParameterGroupResourceModel{}, false, false
	}

	sourceParameterGroupId := current.SourceParameterGroupId
	sourceParameterGroupType := current.SourceParameterGroupType
	if sourceParameterGroupId.IsNull() || sourceParameterGroupId.IsUnknown() {
		sourceParameterGroupId = mapped.DefaultParameterGroupId
	}
	if sourceParameterGroupType.IsNull() || sourceParameterGroupType.IsUnknown() {
		sourceParameterGroupType = types.StringValue(string(mysqlsdk.PARAMETERGROUPTYPE_DEFAULT))
	}

	parameterOverrides, ok := r.buildParameterOverridesFromParameters(ctx, mapped.Parameters, current.ParameterOverrides, respDiags)
	if !ok {
		return customParameterGroupResourceModel{}, false, false
	}

	return customParameterGroupResourceModel{
		Id:                       mapped.Id,
		Name:                     mapped.Name,
		SourceParameterGroupId:   sourceParameterGroupId,
		SourceParameterGroupType: sourceParameterGroupType,
		DefaultParameterGroupId:  mapped.DefaultParameterGroupId,
		Description:              mapped.Description,
		ApplyMode:                current.ApplyMode,
		ParameterOverrides:       parameterOverrides,
		EngineVersion:            mapped.EngineVersion,
		ExistErrorSync:           mapped.ExistErrorSync,
		InstanceGroupCount:       mapped.InstanceGroupCount,
		IsRollbackPossible:       mapped.IsRollbackPossible,
		Timeouts:                 current.Timeouts,
	}, true, true
}

func (r *customParameterGroupResource) readCustomParameterGroupParameterMap(
	ctx context.Context,
	id string,
	respDiags *diag.Diagnostics,
) (map[string]mysqlParameterModel, bool) {
	modelResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, respDiags,
		func() (*mysqlsdk.GetMySQLCustomParameterGroupResponseModel, *http.Response, error) {
			return r.kc.ApiClient.MySQLCustomParameterGroupsAPI.
				GetMysqlCustomParameterGroup(ctx, id).
				XAuthToken(r.kc.XAuthToken).
				Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "GetMysqlCustomParameterGroup", err, respDiags)
		return nil, false
	}

	parameters, ok := customParameterGroupParametersValue(ctx, modelResp.CustomParameterGroup.Parameters, respDiags)
	if !ok {
		return nil, false
	}
	return r.buildParameterMapByKey(ctx, parameters, respDiags)
}

func (r *customParameterGroupResource) updateCustomParameterGroup(
	ctx context.Context,
	plan customParameterGroupResourceModel,
	parameters map[string]mysqlParameterModel,
	resetOverrides []mysqlsdk.DataUpdateParameterRequestModel,
	includeDescription bool,
	includeApplyMode bool,
	respDiags *diag.Diagnostics,
) bool {
	group := mysqlsdk.NewMysqlV1ApiUpdateMysqlCustomParameterGroupModelCustomParameterGroupRequestModel()

	if includeApplyMode && !plan.ApplyMode.IsNull() && !plan.ApplyMode.IsUnknown() && plan.ApplyMode.ValueString() != "" {
		applyMode, err := mysqlsdk.NewApplyModeFromValue(plan.ApplyMode.ValueString())
		if err != nil {
			common.AddGeneralError(ctx, r, respDiags, err.Error())
			return false
		}
		group.SetApplyMode(*applyMode)
	}
	if includeDescription && !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		group.SetDescription(plan.Description.ValueString())
	}

	if includeApplyMode {
		if !r.setCustomParameterGroupUpdateParameters(ctx, group, plan.ParameterOverrides, parameters, resetOverrides, respDiags) {
			return false
		}
	}

	request := mysqlsdk.NewBodyUpdateMysqlCustomParameterGroup(*group)
	_, httpResp, err := common.ExecuteWithRetryAndAuth[struct{}](ctx, r.kc, respDiags,
		func() (struct{}, *http.Response, error) {
			httpResp, err := r.kc.ApiClient.MySQLCustomParameterGroupsAPI.
				UpdateMysqlCustomParameterGroup(ctx, plan.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				BodyUpdateMysqlCustomParameterGroup(*request).
				Execute()
			return struct{}{}, httpResp, err
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "UpdateMysqlCustomParameterGroup", err, respDiags)
		return false
	}
	return true
}

func (r *customParameterGroupResource) rollbackCreatedCustomParameterGroup(ctx context.Context, id string, respDiags *diag.Diagnostics) {
	if id == "" {
		return
	}
	if r.deleteCustomParameterGroup(ctx, id, respDiags) {
		respDiags.AddWarning(
			"Rolled back MySQL custom parameter group creation",
			fmt.Sprintf("Custom parameter group %q was created, but a follow-up update failed. The provider deleted it to avoid leaving an unmanaged resource.", id),
		)
	}
}

func (r *customParameterGroupResource) setCustomParameterGroupUpdateParameters(
	ctx context.Context,
	group *mysqlsdk.MysqlV1ApiUpdateMysqlCustomParameterGroupModelCustomParameterGroupRequestModel,
	parameterOverrides types.Set,
	parameters map[string]mysqlParameterModel,
	resetOverrides []mysqlsdk.DataUpdateParameterRequestModel,
	respDiags *diag.Diagnostics,
) bool {
	overrides, ok := r.parameterOverridesFromSet(ctx, parameterOverrides, parameters, respDiags)
	if !ok {
		return false
	}
	if len(resetOverrides) > 0 {
		overrides = append(resetOverrides, overrides...)
	}
	if len(overrides) > 0 {
		group.SetParameters(overrides)
	}
	return true
}

func hasCustomParameterGroupParameterUpdates(
	stateParameterOverrides types.Set,
	planParameterOverrides types.Set,
	resetOverrides []mysqlsdk.DataUpdateParameterRequestModel,
) bool {
	if len(resetOverrides) > 0 {
		return true
	}

	if isNullOrEmptySet(stateParameterOverrides) && isNullOrEmptySet(planParameterOverrides) {
		return false
	}
	return !planParameterOverrides.Equal(stateParameterOverrides)
}

func isNullOrEmptySet(value types.Set) bool {
	return value.IsNull() || (!value.IsUnknown() && len(value.Elements()) == 0)
}

func (r *customParameterGroupResource) hasConfiguredParameterOverrides(
	ctx context.Context,
	parameterOverrides types.Set,
	respDiags *diag.Diagnostics,
) (bool, bool) {
	overrides, ok := r.parameterOverrideModelsFromSet(ctx, parameterOverrides, respDiags)
	if !ok {
		return false, false
	}
	return len(overrides) > 0, true
}

func (r *customParameterGroupResource) deleteCustomParameterGroup(ctx context.Context, id string, respDiags *diag.Diagnostics) bool {
	_, httpResp, err := common.ExecuteWithRetryAndAuth[struct{}](ctx, r.kc, respDiags,
		func() (struct{}, *http.Response, error) {
			httpResp, err := r.kc.ApiClient.MySQLCustomParameterGroupsAPI.
				DeleteMysqlCustomParameterGroup(ctx, id).
				XAuthToken(r.kc.XAuthToken).
				Execute()
			return struct{}{}, httpResp, err
		},
	)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			return true
		}
		common.AddApiActionError(ctx, r, httpResp, "DeleteMysqlCustomParameterGroup", err, respDiags)
		return false
	}

	return r.pollCustomParameterGroupUntilDeleted(ctx, id, respDiags)
}

func (r *customParameterGroupResource) pollCustomParameterGroupUntilDeleted(ctx context.Context, id string, respDiags *diag.Diagnostics) bool {
	ticker := time.NewTicker(mysqlPollInterval)
	defer ticker.Stop()

	for {
		_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, respDiags,
			func() (*mysqlsdk.GetMySQLCustomParameterGroupResponseModel, *http.Response, error) {
				return r.kc.ApiClient.MySQLCustomParameterGroupsAPI.
					GetMysqlCustomParameterGroup(ctx, id).
					XAuthToken(r.kc.XAuthToken).
					Execute()
			},
		)
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			return true
		}
		if err != nil {
			common.AddApiActionError(ctx, r, httpResp, "GetMysqlCustomParameterGroup", err, respDiags)
			return false
		}

		select {
		case <-ctx.Done():
			common.AddGeneralError(ctx, r, respDiags, "context deadline exceeded")
			return false
		case <-ticker.C:
		}
	}
}

func (r *customParameterGroupResource) parameterOverridesFromSet(
	ctx context.Context,
	value types.Set,
	parameters map[string]mysqlParameterModel,
	respDiags *diag.Diagnostics,
) ([]mysqlsdk.DataUpdateParameterRequestModel, bool) {
	models, ok := r.parameterOverrideModelsFromSet(ctx, value, respDiags)
	if !ok {
		return nil, false
	}

	result := make([]mysqlsdk.DataUpdateParameterRequestModel, 0, len(models))
	for _, item := range models {
		param := mysqlsdk.NewDataUpdateParameterRequestModel(item.Key.ValueString())
		if item.Value.IsNull() {
			if !r.setResetParameterValue(param, parameters, respDiags) {
				return nil, false
			}
		} else if !item.Value.IsUnknown() {
			param.SetValue(item.Value.ValueString())
		}
		result = append(result, *param)
	}
	return result, true
}

func (r *customParameterGroupResource) parameterOverrideModelsFromSet(
	ctx context.Context,
	value types.Set,
	respDiags *diag.Diagnostics,
) ([]customParameterOverrideModel, bool) {
	if value.IsNull() || value.IsUnknown() {
		return nil, true
	}
	var models []customParameterOverrideModel
	respDiags.Append(value.ElementsAs(ctx, &models, false)...)
	if respDiags.HasError() {
		return nil, false
	}
	return models, true
}

func (r *customParameterGroupResource) buildParameterOverridesFromParameters(
	ctx context.Context,
	value types.List,
	current types.Set,
	respDiags *diag.Diagnostics,
) (types.Set, bool) {
	if value.IsNull() || value.IsUnknown() {
		return types.SetNull(types.ObjectType{AttrTypes: customParameterOverrideAttrTypes}), true
	}

	var parameters []mysqlParameterModel
	respDiags.Append(value.ElementsAs(ctx, &parameters, false)...)
	if respDiags.HasError() {
		return types.SetNull(types.ObjectType{AttrTypes: customParameterOverrideAttrTypes}), false
	}

	currentOverrides, ok := r.buildParameterOverrideMapByKey(ctx, current, respDiags)
	if !ok {
		return types.SetNull(types.ObjectType{AttrTypes: customParameterOverrideAttrTypes}), false
	}

	overrides := make([]customParameterOverrideModel, 0)
	for _, parameter := range parameters {
		currentOverride, isCurrentOverride := currentOverrides[parameter.Key.ValueString()]
		if isCurrentOverride && currentOverride.Value.IsNull() {

			overrides = append(overrides, customParameterOverrideModel{
				Key:   parameter.Key,
				Value: types.StringNull(),
			})
			continue
		}
		if !isParameterOverride(parameter, isCurrentOverride) {
			continue
		}
		overrides = append(overrides, customParameterOverrideModel{
			Key:   parameter.Key,
			Value: parameter.Value,
		})
	}
	if len(overrides) == 0 {
		if !current.IsNull() && !current.IsUnknown() && len(current.Elements()) == 0 {
			return current, true
		}
		return types.SetNull(types.ObjectType{AttrTypes: customParameterOverrideAttrTypes}), true
	}

	result, diags := utils.ConvertSetFromModel(ctx, overrides, customParameterOverrideAttrTypes, func(override customParameterOverrideModel) any {
		return override
	})
	respDiags.Append(diags...)
	if respDiags.HasError() {
		return types.SetNull(types.ObjectType{AttrTypes: customParameterOverrideAttrTypes}), false
	}
	return result, true
}

func (r *customParameterGroupResource) buildParameterOverrideMapByKey(
	ctx context.Context,
	value types.Set,
	respDiags *diag.Diagnostics,
) (map[string]customParameterOverrideModel, bool) {
	result := map[string]customParameterOverrideModel{}
	if value.IsNull() || value.IsUnknown() {
		return result, true
	}

	var overrides []customParameterOverrideModel
	respDiags.Append(value.ElementsAs(ctx, &overrides, false)...)
	if respDiags.HasError() {
		return nil, false
	}
	for _, override := range overrides {
		result[override.Key.ValueString()] = override
	}
	return result, true
}

func isParameterOverride(parameter mysqlParameterModel, isCurrentOverride bool) bool {
	if isCurrentOverride {
		return true
	}
	if parameter.Value.IsUnknown() || parameter.DefaultParameterValue.IsUnknown() {
		return false
	}
	return !sameStringValue(parameter.Value, parameter.DefaultParameterValue)
}

func (r *customParameterGroupResource) buildResetParamsForDeletedParameterOverrides(
	ctx context.Context,
	state customParameterGroupResourceModel,
	plan customParameterGroupResourceModel,
	parameters map[string]mysqlParameterModel,
	respDiags *diag.Diagnostics,
) ([]mysqlsdk.DataUpdateParameterRequestModel, bool) {
	if state.ParameterOverrides.IsNull() || state.ParameterOverrides.IsUnknown() {
		return nil, true
	}

	var stateOverrides []customParameterOverrideModel
	respDiags.Append(state.ParameterOverrides.ElementsAs(ctx, &stateOverrides, false)...)
	if respDiags.HasError() {
		return nil, false
	}

	planKeys := map[string]struct{}{}
	if !plan.ParameterOverrides.IsNull() && !plan.ParameterOverrides.IsUnknown() {
		var planOverrides []customParameterOverrideModel
		respDiags.Append(plan.ParameterOverrides.ElementsAs(ctx, &planOverrides, false)...)
		if respDiags.HasError() {
			return nil, false
		}
		for _, item := range planOverrides {
			planKeys[item.Key.ValueString()] = struct{}{}
		}
	}

	result := make([]mysqlsdk.DataUpdateParameterRequestModel, 0)
	for _, item := range stateOverrides {
		if item.Value.IsNull() {
			continue
		}
		key := item.Key.ValueString()
		if _, ok := planKeys[key]; ok {
			continue
		}

		param := mysqlsdk.NewDataUpdateParameterRequestModel(key)
		if !r.setResetParameterValue(param, parameters, respDiags) {
			return nil, false
		}
		result = append(result, *param)
	}
	return result, true
}

func (r *customParameterGroupResource) setResetParameterValue(
	param *mysqlsdk.DataUpdateParameterRequestModel,
	parameters map[string]mysqlParameterModel,
	respDiags *diag.Diagnostics,
) bool {
	key := param.GetKey()
	parameter, ok := parameters[key]
	if !ok {
		respDiags.AddAttributeError(
			path.Root("parameter_overrides"),
			"Unable to reset MySQL parameter override",
			fmt.Sprintf("Parameter %q was marked for reset, but its metadata was not found. Refresh the resource and try again.", key),
		)
		return false
	}

	if parameter.IsRequired.ValueBool() {
		if parameter.DefaultParameterValue.IsNull() || parameter.DefaultParameterValue.IsUnknown() {
			respDiags.AddAttributeError(
				path.Root("parameter_overrides"),
				"Unable to reset MySQL parameter override",
				fmt.Sprintf("Parameter %q is required, but default_parameter_value is not available. The provider cannot reset it safely.", key),
			)
			return false
		}
		param.SetValue(parameter.DefaultParameterValue.ValueString())
	} else {
		param.SetValueNil()
	}
	return true
}

func (r *customParameterGroupResource) buildParameterMapByKey(
	ctx context.Context,
	value types.List,
	respDiags *diag.Diagnostics,
) (map[string]mysqlParameterModel, bool) {
	if value.IsNull() || value.IsUnknown() {
		return map[string]mysqlParameterModel{}, true
	}

	var models []mysqlParameterModel
	respDiags.Append(value.ElementsAs(ctx, &models, false)...)
	if respDiags.HasError() {
		return nil, false
	}

	result := make(map[string]mysqlParameterModel, len(models))
	for _, item := range models {
		result[item.Key.ValueString()] = item
	}
	return result, true
}

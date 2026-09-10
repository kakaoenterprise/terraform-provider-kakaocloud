// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package network

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"terraform-provider-kakaocloud/internal/common"
	"terraform-provider-kakaocloud/internal/utils"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kakaoenterprise/kc-sdk-go/v2/services/network"
)

var (
	_ resource.Resource                   = &securityGroupResource{}
	_ resource.ResourceWithConfigure      = &securityGroupResource{}
	_ resource.ResourceWithImportState    = &securityGroupResource{}
	_ resource.ResourceWithValidateConfig = &securityGroupResource{}
)

func NewSecurityGroupResource() resource.Resource {
	return &securityGroupResource{}
}

type securityGroupResource struct {
	kc *common.KakaoCloudClient
}

func (r *securityGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_security_group"
}

func (r *securityGroupResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: utils.MergeAttributes(
			securityGroupResourceSchemaAttributes,
			map[string]schema.Attribute{
				"timeouts": timeouts.AttributesAll(ctx),
			},
		),
	}
}

func (r *securityGroupResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {

	if req.Config.Raw.IsNull() {
		return
	}

	var cfg securityGroupResourceModel
	d := req.Config.Get(ctx, &cfg)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	if cfg.Rules.IsNull() || cfg.Rules.IsUnknown() {
		return
	}

	var rules []securityGroupRuleModel
	d2 := cfg.Rules.ElementsAs(ctx, &rules, false)
	resp.Diagnostics.Append(d2...)
	if resp.Diagnostics.HasError() {
		return
	}

	for _, rule := range rules {
		proto := strings.ToUpper(rule.Protocol.ValueString())

		remoteIpSet := !rule.RemoteIpPrefix.IsNull()
		remoteGroupSet := !rule.RemoteGroupId.IsNull()

		onlyOneRemoteFieldExists := (remoteIpSet) != (remoteGroupSet)
		if !onlyOneRemoteFieldExists {
			common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
				"Exactly one of remote_ip_prefix or remote_group_id must be provided (not both or neither)",
			)
			continue
		}

		switch proto {
		case string(network.RULEPROTOCOL_TCP), string(network.RULEPROTOCOL_UDP):
			if rule.PortRangeMin.IsNull() {
				common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
					fmt.Sprintf("port_range_min is required when protocol is '%s'", proto),
				)
			}
			if rule.PortRangeMax.IsNull() {
				common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
					fmt.Sprintf("port_range_max is required when protocol is '%s'", proto),
				)
			}

			if rule.PortRangeMin.IsNull() || rule.PortRangeMin.IsUnknown() ||
				rule.PortRangeMax.IsNull() || rule.PortRangeMax.IsUnknown() {
				continue
			}

			if rule.PortRangeMin.ValueInt32() > rule.PortRangeMax.ValueInt32() {
				common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
					fmt.Sprintf("port_range_max must be greater than or equal to port_range_min when protocol is '%s'", proto),
				)
			}
		case string(network.RULEPROTOCOL_ALL), string(network.RULEPROTOCOL_ICMP):
			if !rule.PortRangeMin.IsNull() || !rule.PortRangeMax.IsNull() {
				common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
					fmt.Sprintf("port_range_min and port_range_max must be null when protocol is '%s'", proto),
				)
			}
		}
	}
}

func (r *securityGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state securityGroupResourceModel
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
		func() (*network.GetSecurityGroupResponse, *http.Response, error) {
			return r.kc.ApiClient.NetworkSecurityGroupAPI.
				GetSecurityGroup(ctx, state.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				Execute()
		},
	)

	if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "GetSecurityGroup", err, &resp.Diagnostics)
		return
	}

	ok := mapSecurityGroupModel(ctx, &state.securityGroupBaseModel, &respModel.SecurityGroup, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *securityGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan securityGroupResourceModel
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

	createReq := *network.NewCreateSecurityGroup(plan.Name.ValueString())
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		createReq.SetDescription(plan.Description.ValueString())
	}

	body := *network.NewCreateSecurityGroupRequest(createReq)
	respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*network.CreateSecurityGroupResponse, *http.Response, error) {
			return r.kc.ApiClient.NetworkSecurityGroupAPI.CreateSecurityGroup(ctx).
				XAuthToken(r.kc.XAuthToken).
				CreateSecurityGroupRequest(body).
				Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "CreateSecurityGroup", err, &resp.Diagnostics)
		return
	}

	desiredRuleIDs := []string{}
	if !plan.Rules.IsNull() && !plan.Rules.IsUnknown() {
		var desired []securityGroupRuleModel
		dDiags := plan.Rules.ElementsAs(ctx, &desired, false)
		resp.Diagnostics.Append(dDiags...)
		if resp.Diagnostics.HasError() {
			_ = r.cleanupCreateFailureSecurityGroup(ctx, respModel.SecurityGroup.Id, &resp.Diagnostics)
			return
		}

		for i := range desired {
			dr := &desired[i]
			if ok := r.addSecurityGroupRule(ctx, respModel.SecurityGroup.Id, dr, &resp.Diagnostics); !ok {
				_ = r.cleanupCreateFailureSecurityGroup(ctx, respModel.SecurityGroup.Id, &resp.Diagnostics)
				return
			}
		}

		var idErr error
		desiredRuleIDs, idErr = securityGroupRuleModelIDs(desired)
		if idErr != nil {
			common.AddGeneralError(ctx, r, &resp.Diagnostics, idErr.Error())
			_ = r.cleanupCreateFailureSecurityGroup(ctx, respModel.SecurityGroup.Id, &resp.Diagnostics)
			return
		}
	}

	finalSg, ok := r.waitForRulesToPropagate(ctx, respModel.SecurityGroup.Id, desiredRuleIDs, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		_ = r.cleanupCreateFailureSecurityGroup(ctx, respModel.SecurityGroup.Id, &resp.Diagnostics)
		return
	}

	ok = mapSecurityGroupModel(ctx, &plan.securityGroupBaseModel, finalSg, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		_ = r.cleanupCreateFailureSecurityGroup(ctx, respModel.SecurityGroup.Id, &resp.Diagnostics)
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		_ = r.cleanupCreateFailureSecurityGroup(ctx, respModel.SecurityGroup.Id, &resp.Diagnostics)
		return
	}
}

func (r *securityGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state securityGroupResourceModel
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

	var newName, newDescription types.String
	if !plan.Name.Equal(state.Name) || !plan.Description.Equal(state.Description) && !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		editReq := *network.NewUpdateSecurityGroup()
		if !plan.Name.Equal(state.Name) {
			editReq.SetName(plan.Name.ValueString())
			newName = plan.Name
		}

		if !plan.Description.Equal(state.Description) && !plan.Description.IsNull() && !plan.Description.IsUnknown() {
			editReq.SetDescription(plan.Description.ValueString())
			newDescription = plan.Description
		}

		body := *network.NewUpdateSecurityGroupRequest(editReq)
		_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
			func() (interface{}, *http.Response, error) {
				return r.kc.ApiClient.NetworkSecurityGroupAPI.
					UpdateSecurityGroup(ctx, state.Id.ValueString()).
					XAuthToken(r.kc.XAuthToken).
					UpdateSecurityGroupRequest(body).
					Execute()
			},
		)

		if err != nil {
			if httpResp != nil && httpResp.StatusCode >= 200 && httpResp.StatusCode < 300 {
				if strings.Contains(err.Error(), "json: cannot unmarshal number into Go struct field") &&
					(strings.Contains(err.Error(), "port_range_min") || strings.Contains(err.Error(), "port_range_max")) {
				} else {
					common.AddApiActionError(ctx, r, httpResp, "UpdateSecurityGroup", err, &resp.Diagnostics)
					return
				}
			} else {
				common.AddApiActionError(ctx, r, httpResp, "UpdateSecurityGroup", err, &resp.Diagnostics)
				return
			}
		}
	}

	planRules, pDiags := expandSecurityGroupRules(ctx, plan.Rules)
	resp.Diagnostics.Append(pDiags...)
	stateRules, sDiags := expandSecurityGroupRules(ctx, state.Rules)
	resp.Diagnostics.Append(sDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Rules.Equal(state.Rules) {

		planRulesMap := make(map[string]securityGroupRuleModel)
		for _, rule := range planRules {
			planRulesMap[rule.Id.ValueString()] = rule
		}
		stateRulesMap := make(map[string]securityGroupRuleModel)
		for _, rule := range stateRules {
			stateRulesMap[rule.Id.ValueString()] = rule
		}

		for i := range planRules {
			rule := &planRules[i]
			if _, exists := stateRulesMap[rule.Id.ValueString()]; !exists {
				if ok := r.addSecurityGroupRule(ctx, state.Id.ValueString(), rule, &resp.Diagnostics); !ok {
					return
				}
			}
		}

		for key, rule := range stateRulesMap {
			if _, exists := planRulesMap[key]; !exists {
				if ok := r.deleteSecurityGroupRule(ctx, state.Id.ValueString(), rule.Id.ValueString(), &resp.Diagnostics); !ok {
					return
				}
			}
		}
	}

	desiredRuleIDs, err := securityGroupRuleModelIDs(planRules)
	if err != nil {
		common.AddGeneralError(ctx, r, &resp.Diagnostics, err.Error())
		return
	}

	finalSg, ok := r.waitForRulesToPropagate(ctx, state.Id.ValueString(), desiredRuleIDs, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	ok = mapSecurityGroupModel(ctx, &plan.securityGroupBaseModel, finalSg, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if !newName.IsNull() {
		plan.Name = newName
	}
	if !newDescription.IsNull() && !newDescription.IsUnknown() {
		plan.Description = newDescription
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *securityGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state securityGroupResourceModel
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
		func() (struct{}, *http.Response, error) {
			httpResp, err := r.kc.ApiClient.NetworkSecurityGroupAPI.
				DeleteSecurityGroup(ctx, state.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				Execute()
			return struct{}{}, httpResp, err
		},
	)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			return
		}
		common.AddApiActionError(ctx, r, httpResp, "DeleteSecurityGroup", err, &resp.Diagnostics)
		return
	}
}

func (r *securityGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*common.KakaoCloudClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *common.KakaoCloudClient, got: %T.", req.ProviderData),
		)
		return
	}
	r.kc = client
}

func (r *securityGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *securityGroupResource) waitForRulesToPropagate(
	ctx context.Context,
	securityGroupId string,
	desiredRuleIDs []string,
	respDiags *diag.Diagnostics,
) (*network.SecurityGroup, bool) {

	targetStatus := securityGroupRuleIDStatus(desiredRuleIDs)

	result, ok := common.PollUntilResult(
		ctx,
		r,
		3*time.Second,
		"security group",
		securityGroupId,
		[]string{targetStatus},
		respDiags,
		func(ctx context.Context) (*network.SecurityGroup, *http.Response, error) {
			respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, respDiags,
				func() (*network.GetSecurityGroupResponse, *http.Response, error) {
					return r.kc.ApiClient.NetworkSecurityGroupAPI.
						GetSecurityGroup(ctx, securityGroupId).
						XAuthToken(r.kc.XAuthToken).
						Execute()
				},
			)
			if err != nil {
				return nil, httpResp, err
			}
			return &respModel.SecurityGroup, httpResp, nil
		},

		func(sg *network.SecurityGroup) string {
			return securityGroupRuleAPIIDStatus(sg.Rules)
		},
	)

	return result, ok
}

func securityGroupRuleModelIDs(rules []securityGroupRuleModel) ([]string, error) {
	ids := make([]string, 0, len(rules))
	seen := make(map[string]struct{}, len(rules))
	for i, rule := range rules {
		if rule.Id.IsNull() || rule.Id.IsUnknown() {
			return nil, fmt.Errorf("security group rule ID at index %d is null or unknown", i)
		}

		id := rule.Id.ValueString()
		if strings.TrimSpace(id) == "" {
			return nil, fmt.Errorf("security group rule ID at index %d is empty", i)
		}
		if _, exists := seen[id]; exists {
			return nil, fmt.Errorf("duplicate security group rule ID %q", id)
		}

		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	return ids, nil
}

func securityGroupRuleIDStatus(ids []string) string {
	sortedIds := append([]string(nil), ids...)
	sort.Strings(sortedIds)
	return fmt.Sprintf("%q", sortedIds)
}

const invalidSecurityGroupRuleIDStatus = "invalid security group rule IDs"

func securityGroupRuleAPIIDStatus(rules []network.Rule) string {
	ids := make([]string, 0, len(rules))
	seen := make(map[string]struct{}, len(rules))
	for i := range rules {
		id, ok := rules[i].GetIdOk()
		if !ok || id == nil || strings.TrimSpace(*id) == "" {
			return invalidSecurityGroupRuleIDStatus
		}
		if _, exists := seen[*id]; exists {
			return invalidSecurityGroupRuleIDStatus
		}

		seen[*id] = struct{}{}
		ids = append(ids, *id)
	}
	return securityGroupRuleIDStatus(ids)
}

func (r *securityGroupResource) addSecurityGroupRule(ctx context.Context, sgId string, rule *securityGroupRuleModel, diags *diag.Diagnostics) bool {
	creq := *network.NewCreateSecurityGroupRule(
		network.RuleDirection(rule.Direction.ValueString()),
		network.RuleProtocol(rule.Protocol.ValueString()),
	)
	if !rule.Description.IsNull() && !rule.Description.IsUnknown() {
		creq.SetDescription(rule.Description.ValueString())
	}
	if !rule.PortRangeMin.IsNull() && !rule.PortRangeMin.IsUnknown() {
		creq.SetPortRangeMin(rule.PortRangeMin.ValueInt32())
	}
	if !rule.PortRangeMax.IsNull() && !rule.PortRangeMax.IsUnknown() {
		creq.SetPortRangeMax(rule.PortRangeMax.ValueInt32())
	}
	if !rule.RemoteIpPrefix.IsNull() && !rule.RemoteIpPrefix.IsUnknown() {
		creq.SetRemoteIpPrefix(rule.RemoteIpPrefix.ValueString())
	}
	if !rule.RemoteGroupId.IsNull() && !rule.RemoteGroupId.IsUnknown() {
		creq.SetRemoteGroupId(rule.RemoteGroupId.ValueString())
	}

	crResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, diags,
		func() (*network.CreateSecurityGroupRuleResponse, *http.Response, error) {
			return r.kc.ApiClient.NetworkSecurityGroupAPI.
				CreateSecurityGroupRule(ctx, sgId).
				XAuthToken(r.kc.XAuthToken).
				CreateSecurityGroupRuleRequest(*network.NewCreateSecurityGroupRuleRequest(creq)).
				Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "CreateSecurityGroupRule", err, diags)
		return false
	}
	if crResp == nil || strings.TrimSpace(crResp.SecurityGroupRule.Id) == "" {
		common.AddGeneralError(ctx, r, diags, "CreateSecurityGroupRule returned an empty security group rule ID")
		return false
	}
	rule.Id = types.StringValue(crResp.SecurityGroupRule.Id)
	return true
}

func (r *securityGroupResource) cleanupCreateFailureSecurityGroup(
	ctx context.Context,
	securityGroupId string,
	diags *diag.Diagnostics,
) error {
	cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), common.DefaultDeleteTimeout)
	defer cancel()

	_, httpResp, err := common.ExecuteWithRetryAndAuth[struct{}](cleanupCtx, r.kc, diags,
		func() (struct{}, *http.Response, error) {
			httpResp, err := r.kc.ApiClient.NetworkSecurityGroupAPI.
				DeleteSecurityGroup(cleanupCtx, securityGroupId).
				XAuthToken(r.kc.XAuthToken).
				Execute()
			return struct{}{}, httpResp, err
		},
	)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			return nil
		}
		common.AddApiActionError(cleanupCtx, r, httpResp, "DeleteSecurityGroup", err, diags)
		return err
	}

	return nil
}

func (r *securityGroupResource) deleteSecurityGroupRule(ctx context.Context, sgId string, ruleId string, diags *diag.Diagnostics) bool {
	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, diags,
		func() (interface{}, *http.Response, error) {
			httpResp, err := r.kc.ApiClient.NetworkSecurityGroupAPI.
				DeleteSecurityGroupRule(ctx, sgId, ruleId).
				XAuthToken(r.kc.XAuthToken).
				Execute()
			return nil, httpResp, err
		},
	)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			return true
		}
		common.AddApiActionError(ctx, r, httpResp, "DeleteSecurityGroupRule", err, diags)
		return false
	}
	return true
}

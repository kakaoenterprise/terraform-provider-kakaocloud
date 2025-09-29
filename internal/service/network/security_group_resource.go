// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package network

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
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
	"github.com/kakaoenterprise/kc-sdk-go/services/network"
)

var (
	_ resource.Resource                   = &securityGroupResource{}
	_ resource.ResourceWithConfigure      = &securityGroupResource{}
	_ resource.ResourceWithImportState    = &securityGroupResource{}
	_ resource.ResourceWithValidateConfig = &securityGroupResource{}
	_ resource.ResourceWithModifyPlan     = &securityGroupResource{}
)

func NewSecurityGroupResource() resource.Resource {
	return &securityGroupResource{}
}

func (r *securityGroupResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {

	if req.Plan.Raw.IsNull() {
		return
	}

	var plan, state securityGroupResourceModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !req.State.Raw.IsNull() {
		_ = req.State.Get(ctx, &state)
	}

	if plan.Rules.IsNull() && !state.Rules.IsNull() && !state.Rules.IsUnknown() {
		resp.Diagnostics.AddWarning(
			"Optional 'rules' attribute not set",
			"'rules' is optional, but to avoid unexpected update you should explicitly set it to an empty array: rules = []",
		)
	}

	_ = resp.Plan.Set(ctx, &plan)
}

type securityGroupResource struct {
	kc *common.KakaoCloudClient
}

func (r *securityGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_security_group"
}

func (r *securityGroupResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: docs.GetResourceDescription("SecurityGroup"),
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

		remoteIpKnown := !rule.RemoteIpPrefix.IsUnknown()
		remoteGrpKnown := !rule.RemoteGroupId.IsUnknown()
		remoteIpSet := remoteIpKnown && !rule.RemoteIpPrefix.IsNull() && rule.RemoteIpPrefix.ValueString() != ""
		remoteGroupSet := remoteGrpKnown && !rule.RemoteGroupId.IsNull() && rule.RemoteGroupId.ValueString() != ""

		if remoteIpKnown && remoteGrpKnown {

			onlyOneRemoteFieldExists := (remoteIpSet) != (remoteGroupSet)
			if !onlyOneRemoteFieldExists {
				resp.Diagnostics.AddError(
					"Invalid security group rule",
					"Exactly one of remote_ip_prefix or remote_group_id must be provided (not both or neither)",
				)
				continue
			}
		}

		minProvided := !(rule.PortRangeMin.IsNull() || rule.PortRangeMin.IsUnknown() || rule.PortRangeMin.ValueString() == "")
		maxProvided := !(rule.PortRangeMax.IsNull() || rule.PortRangeMax.IsUnknown() || rule.PortRangeMax.ValueString() == "")

		switch proto {
		case string(network.SECURITYGROUPRULEPROTOCOL_TCP), string(network.SECURITYGROUPRULEPROTOCOL_UDP):
			if !minProvided {
				resp.Diagnostics.AddError("Invalid security group rule", fmt.Sprintf("port_range_min is required when protocol is %s", proto))
			}
			if !maxProvided {
				resp.Diagnostics.AddError("Invalid security group rule", fmt.Sprintf("port_range_max is required when protocol is %s", proto))
			}
			if minProvided && maxProvided {
				minVal, err1 := strconv.Atoi(rule.PortRangeMin.ValueString())
				maxVal, err2 := strconv.Atoi(rule.PortRangeMax.ValueString())
				minValid := err1 == nil && minVal >= 1 && minVal <= 65535
				maxValid := err2 == nil && maxVal >= 1 && maxVal <= 65535
				if !minValid || !maxValid {
					msg := fmt.Sprintf("port_range_min and port_range_max must be between 1 and 65535 when protocol is %s.", proto)
					if !minValid {
						msg += fmt.Sprintf(" port_range_min=%s is out of range.", rule.PortRangeMin.ValueString())
					}
					if !maxValid {
						msg += fmt.Sprintf(" port_range_max=%s is out of range.", rule.PortRangeMax.ValueString())
					}
					resp.Diagnostics.AddError("Invalid security group rule", msg)
				}
			}
		case string(network.SECURITYGROUPRULEPROTOCOL_ALL), string(network.SECURITYGROUPRULEPROTOCOL_ICMP):
			if minProvided || maxProvided {
				resp.Diagnostics.AddError("Invalid security group rule", fmt.Sprintf("port_range_min and port_range_max must be null when protocol is %s", proto))
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
		func() (*network.BnsNetworkV1ApiGetSecurityGroupModelResponseSecurityGroupModel, *http.Response, error) {
			return r.kc.ApiClient.SecurityGroupAPI.
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

	ok := mapSecurityGroupBaseModel(ctx, &state.securityGroupBaseModel, &respModel.SecurityGroup, &resp.Diagnostics)
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

	createReq := network.CreateSecurityGroupModel{
		Name: plan.Name.ValueString(),
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		createReq.SetDescription(plan.Description.ValueString())
	}

	body := network.BodyCreateSecurityGroup{SecurityGroup: createReq}
	respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*network.BnsNetworkV1ApiCreateSecurityGroupModelResponseSecurityGroupModel, *http.Response, error) {
			return r.kc.ApiClient.SecurityGroupAPI.CreateSecurityGroup(ctx).
				XAuthToken(r.kc.XAuthToken).
				BodyCreateSecurityGroup(body).
				Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "CreateSecurityGroup", err, &resp.Diagnostics)
		return
	}

	desiredRuleCount := 0
	if !plan.Rules.IsNull() && !plan.Rules.IsUnknown() {
		var desired []securityGroupRuleModel
		dDiags := plan.Rules.ElementsAs(ctx, &desired, false)
		resp.Diagnostics.Append(dDiags...)

		if !resp.Diagnostics.HasError() {
			for i := range desired {
				dr := &desired[i]
				if ok := r.addSecurityGroupRule(ctx, respModel.SecurityGroup.Id, dr, &resp.Diagnostics); !ok {
					return
				}
			}
		}

		desiredRuleCount = len(desired)

	}

	finalSg, ok := r.waitForRulesToPropagate(ctx, respModel.SecurityGroup.Id, desiredRuleCount, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	originalRulesNull := plan.Rules.IsNull() || plan.Rules.IsUnknown()

	ok = mapSecurityGroupBaseModel(ctx, &plan.securityGroupBaseModel, finalSg, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}
	if originalRulesNull {
		plan.Rules = types.SetNull(types.ObjectType{AttrTypes: securityGroupRuleAttrType})
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
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

	if !plan.Name.Equal(state.Name) || !plan.Description.Equal(state.Description) && !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		editReq := network.EditSecurityGroupModel{}
		if !plan.Name.Equal(state.Name) {
			editReq.SetName(plan.Name.ValueString())
		}

		if !plan.Description.Equal(state.Description) && !plan.Description.IsNull() && !plan.Description.IsUnknown() {
			editReq.SetDescription(plan.Description.ValueString())
		}

		body := network.BodyUpdateSecurityGroup{SecurityGroup: editReq}
		_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
			func() (interface{}, *http.Response, error) {
				return r.kc.ApiClient.SecurityGroupAPI.
					UpdateSecurityGroup(ctx, state.Id.ValueString()).
					XAuthToken(r.kc.XAuthToken).
					BodyUpdateSecurityGroup(body).
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

	if !plan.Rules.Equal(state.Rules) {

		planRules, pDiags := expandSecurityGroupRules(ctx, plan.Rules)
		resp.Diagnostics.Append(pDiags...)
		stateRules, sDiags := expandSecurityGroupRules(ctx, state.Rules)
		resp.Diagnostics.Append(sDiags...)
		if resp.Diagnostics.HasError() {
			return
		}

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

	desiredRuleCount := 0
	if !plan.Rules.IsNull() && !plan.Rules.IsUnknown() {
		desiredRuleCount = len(plan.Rules.Elements())
	}

	finalSg, ok := r.waitForRulesToPropagate(ctx, state.Id.ValueString(), desiredRuleCount, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	originalRulesNull := plan.Rules.IsNull() || plan.Rules.IsUnknown()

	ok = mapSecurityGroupBaseModel(ctx, &plan.securityGroupBaseModel, finalSg, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if originalRulesNull {
		plan.Rules = types.SetNull(types.ObjectType{AttrTypes: securityGroupRuleAttrType})
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
			httpResp, err := r.kc.ApiClient.SecurityGroupAPI.
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
	desiredRuleCount int,
	respDiags *diag.Diagnostics,
) (*network.BnsNetworkV1ApiGetSecurityGroupModelSecurityGroupModel, bool) {

	targetStatus := fmt.Sprintf("%d rules", desiredRuleCount)

	result, ok := common.PollUntilResult(
		ctx,
		r,
		3*time.Second,
		[]string{targetStatus},
		respDiags,
		func(ctx context.Context) (*network.BnsNetworkV1ApiGetSecurityGroupModelSecurityGroupModel, *http.Response, error) {
			respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, respDiags,
				func() (*network.BnsNetworkV1ApiGetSecurityGroupModelResponseSecurityGroupModel, *http.Response, error) {
					return r.kc.ApiClient.SecurityGroupAPI.
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

		func(sg *network.BnsNetworkV1ApiGetSecurityGroupModelSecurityGroupModel) string {
			return fmt.Sprintf("%d rules", len(sg.Rules))
		},
	)

	return result, ok
}

func (r *securityGroupResource) addSecurityGroupRule(ctx context.Context, sgId string, rule *securityGroupRuleModel, diags *diag.Diagnostics) bool {
	creq := network.CreateSecurityGroupRuleModel{
		Direction: network.SecurityGroupRuleDirection(rule.Direction.ValueString()),
		Protocol:  network.SecurityGroupRuleProtocol(rule.Protocol.ValueString()),
	}
	if !rule.Description.IsNull() && !rule.Description.IsUnknown() {
		creq.SetDescription(rule.Description.ValueString())
	}
	if !rule.PortRangeMin.IsNull() && !rule.PortRangeMin.IsUnknown() {
		if v, err := strconv.Atoi(rule.PortRangeMin.ValueString()); err == nil {
			creq.SetPortRangeMin(int32(v))
		}
	}
	if !rule.PortRangeMax.IsNull() && !rule.PortRangeMax.IsUnknown() {
		if v, err := strconv.Atoi(rule.PortRangeMax.ValueString()); err == nil {
			creq.SetPortRangeMax(int32(v))
		}
	}
	if !rule.RemoteIpPrefix.IsNull() && !rule.RemoteIpPrefix.IsUnknown() {
		creq.SetRemoteIpPrefix(rule.RemoteIpPrefix.ValueString())
	}
	if !rule.RemoteGroupId.IsNull() && !rule.RemoteGroupId.IsUnknown() {
		creq.SetRemoteGroupId(rule.RemoteGroupId.ValueString())
	}

	crResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, diags,
		func() (*network.ResponseSecurityGroupRuleModel, *http.Response, error) {
			return r.kc.ApiClient.SecurityGroupAPI.
				CreateSecurityGroupRule(ctx, sgId).
				XAuthToken(r.kc.XAuthToken).
				BodyCreateSecurityGroupRule(network.BodyCreateSecurityGroupRule{SecurityGroupRule: creq}).
				Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "CreateSecurityGroupRule", err, diags)
		return false
	}
	rule.Id = types.StringValue(crResp.SecurityGroupRule.Id)
	return true
}

func (r *securityGroupResource) deleteSecurityGroupRule(ctx context.Context, sgId string, ruleId string, diags *diag.Diagnostics) bool {
	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, diags,
		func() (interface{}, *http.Response, error) {
			httpResp, err := r.kc.ApiClient.SecurityGroupAPI.
				DeleteSecurityGroupRule(ctx, sgId, ruleId).
				XAuthToken(r.kc.XAuthToken).
				Execute()
			return nil, httpResp, err
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "DeleteSecurityGroupRule", err, diags)
		return false
	}
	return true
}

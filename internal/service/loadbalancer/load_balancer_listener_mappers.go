// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package loadbalancer

import (
	"terraform-provider-kakaocloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/kakaoenterprise/kc-sdk-go/services/loadbalancer"
	"golang.org/x/net/context"
)

func mapLoadBalancerListenerBaseModel(
	ctx context.Context,
	base *loadBalancerListenerBaseModel,
	src *loadbalancer.BnsLoadBalancerV1ApiGetListenerModelListenerModel,
	diags *diag.Diagnostics,
) bool {

	secrets, secretsDiags := utils.ConvertListFromModel(ctx, src.Secrets, loadBalancerListenerSecretAttrType, func(secret loadbalancer.BnsLoadBalancerV1ApiGetListenerModelSecretModel) any {
		return loadBalancerListenerSecretsModel{
			Id:         types.StringValue(secret.Id),
			Name:       utils.ConvertNullableString(secret.Name),
			Expiration: utils.ConvertNullableString(secret.Expiration),
			Status:     utils.ConvertNullableString(secret.Status),
			SecretType: utils.ConvertNullableString(secret.SecretType),
			CreatorId:  utils.ConvertNullableString(secret.CreatorId),
			IsDefault:  utils.ConvertNullableBool(secret.IsDefault),
		}
	})

	diags.Append(secretsDiags...)

	l7Policies, l7PolicyDiags := utils.ConvertListFromModel(ctx, src.L7Policies, loadBalancerListenerL7PolicyAttrType, func(policy loadbalancer.BnsLoadBalancerV1ApiGetListenerModelL7PolicyModel) any {
		if policy.Id == "" {

			return nil
		}
		rules, ruleDiags := utils.ConvertListFromModel(ctx, policy.Rules, loadBalancerListenerL7PolicyRuleAttrType, func(rule loadbalancer.BnsLoadBalancerV1ApiGetListenerModelRuleModel) any {
			return loadBalancerListenerL7PolicyRuleModel{
				Id:                 types.StringValue(rule.Id),
				CompareType:        utils.ConvertNullableString(rule.CompareType),
				IsInverted:         types.BoolValue(rule.IsInverted),
				Key:                utils.ConvertNullableString(rule.Key),
				Value:              utils.ConvertNullableString(rule.Value),
				ProvisioningStatus: utils.ConvertNullableString(rule.ProvisioningStatus),
				OperatingStatus:    utils.ConvertNullableString(rule.OperatingStatus),
				ProjectId:          types.StringValue(rule.ProjectId),
				Type:               utils.ConvertNullableString(rule.Type),
			}
		})
		diags.Append(ruleDiags...)

		return loadBalancerListenerL7PolicyModel{
			Id:                    types.StringValue(policy.Id),
			Name:                  utils.ConvertNullableString(policy.Name),
			Description:           utils.ConvertNullableString(policy.Description),
			ProvisioningStatus:    utils.ConvertNullableString(policy.ProvisioningStatus),
			OperatingStatus:       utils.ConvertNullableString(policy.OperatingStatus),
			ProjectId:             utils.ConvertNullableString(policy.ProjectId),
			Action:                utils.ConvertNullableString(policy.Action),
			Position:              utils.ConvertNullableInt32ToInt64(policy.Position),
			Rules:                 rules,
			RedirectTargetGroupId: utils.ConvertNullableString(policy.RedirectTargetGroupId),
			RedirectUrl:           utils.ConvertNullableString(policy.RedirectUrl),
			RedirectPrefix:        utils.ConvertNullableString(policy.RedirectPrefix),
			RedirectHttpCode:      utils.ConvertNullableInt32ToInt64(policy.RedirectHttpCode)}
	})
	diags.Append(l7PolicyDiags...)

	insertHeadersAttrTypes := map[string]attr.Type{
		"x_forwarded_for":   types.StringType,
		"x_forwarded_proto": types.StringType,
		"x_forwarded_port":  types.StringType,
	}

	var insertHeadersObject types.Object

	if len(src.InsertHeaders) == 0 {
		insertHeadersObject = types.ObjectNull(insertHeadersAttrTypes)
	} else {
		headerValues := make(map[string]attr.Value)
		mapHeader := func(key string) attr.Value {
			if val, ok := src.InsertHeaders[key]; ok && val.String != nil {
				return types.StringValue(*val.String)
			}
			return types.StringNull()
		}
		headerValues["x_forwarded_for"] = mapHeader("X-Forwarded-For")
		headerValues["x_forwarded_proto"] = mapHeader("X-Forwarded-Proto")
		headerValues["x_forwarded_port"] = mapHeader("X-Forwarded-Port")

		obj, d := types.ObjectValue(insertHeadersAttrTypes, headerValues)
		diags.Append(d...)
		insertHeadersObject = obj
	}

	base.InsertHeaders = insertHeadersObject

	base.Id = types.StringValue(src.Id)
	base.Name = utils.ConvertNullableString(src.Name)
	base.Description = utils.ConvertNullableString(src.Description)
	base.Protocol = utils.ConvertNullableString(src.Protocol)
	base.IsEnabled = utils.ConvertNullableBool(src.IsEnabled)
	base.Secrets = secrets
	base.L7Policies = l7Policies
	base.TlsCiphers = utils.ConvertNullableString(src.TlsCiphers)

	if len(src.TlsVersions) > 0 {
		tlsVersionStrings := make([]string, len(src.TlsVersions))
		for i, version := range src.TlsVersions {
			tlsVersionStrings[i] = string(version)
		}
		base.TlsVersions = utils.ConvertNullableStringList(tlsVersionStrings)
	} else {
		base.TlsVersions = types.ListNull(types.StringType)
	}
	base.AlpnProtocols = utils.ConvertNullableStringList(src.AlpnProtocols)
	base.ProjectId = utils.ConvertNullableString(src.ProjectId)
	base.ProtocolPort = utils.ConvertNullableInt32ToInt64(src.ProtocolPort)
	base.ConnectionLimit = utils.ConvertNullableInt32ToInt64(src.ConnectionLimit)
	base.LoadBalancerId = utils.ConvertNullableString(src.LoadBalancerId)
	base.TlsCertificateId = utils.ConvertNullableString(src.TlsCertificateId)
	base.ProvisioningStatus = utils.ConvertNullableString(src.ProvisioningStatus)
	base.OperatingStatus = utils.ConvertNullableString(src.OperatingStatus)

	base.CreatedAt = utils.ConvertNullableTime(src.CreatedAt)
	base.UpdatedAt = utils.ConvertNullableTime(src.UpdatedAt)
	base.TimeoutClientData = utils.ConvertNullableInt32ToInt64(src.TimeoutClientData)
	base.DefaultTargetGroupName = utils.ConvertNullableString(src.DefaultTargetGroupName)
	base.DefaultTargetGroupId = utils.ConvertNullableString(src.DefaultTargetGroupId)
	base.LoadBalancerType = utils.ConvertNullableString(src.LoadBalancerType)

	return !diags.HasError()
}

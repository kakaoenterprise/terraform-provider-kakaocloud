// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package network

import (
	"context"
	"fmt"

	. "terraform-provider-kakaocloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jinzhu/copier"
	"github.com/kakaoenterprise/kc-sdk-go/services/network"
)

func mapSecurityGroupBaseModel(
	ctx context.Context,
	base *securityGroupBaseModel,
	result *network.BnsNetworkV1ApiGetSecurityGroupModelSecurityGroupModel,
	respDiags *diag.Diagnostics,
) bool {
	rules, ruleDiags := ConvertSetFromModel(
		ctx,
		result.Rules,
		securityGroupRuleAttrType,
		func(src network.BnsNetworkV1ApiGetSecurityGroupModelSecurityGroupRuleModel) any {

			portMin := ConvertNullableString(src.PortRangeMin)
			portMax := ConvertNullableString(src.PortRangeMax)

			if !portMin.IsNull() && !portMin.IsUnknown() &&
				!portMax.IsNull() && !portMax.IsUnknown() {

				if portMin.ValueString() == "ALL" && portMax.ValueString() == "ALL" {
					portMin = types.StringValue("1")
					portMax = types.StringValue("65535")
				}
			}

			return securityGroupRuleModel{
				Id:              types.StringValue(src.Id),
				Description:     ConvertNullableString(src.Description),
				RemoteGroupId:   ConvertNullableString(src.RemoteGroupId),
				RemoteGroupName: ConvertNullableString(src.RemoteGroupName),
				Direction:       ConvertNullableString(src.Direction),
				Protocol:        types.StringValue(string(src.Protocol)),
				PortRangeMin:    portMin,
				PortRangeMax:    portMax,
				RemoteIpPrefix:  ConvertNullableString(src.RemoteIpPrefix),
				CreatedAt:       ConvertNullableTime(src.CreatedAt),
				UpdatedAt:       ConvertNullableTime(src.UpdatedAt),
			}
		},
	)
	respDiags.Append(ruleDiags...)

	base.Id = types.StringValue(result.Id)
	base.Name = ConvertNullableString(result.Name)
	base.Description = ConvertNullableString(result.Description)
	base.ProjectId = ConvertNullableString(result.ProjectId)
	base.ProjectName = ConvertNullableString(result.ProjectName)
	base.IsStateful = ConvertNullableBool(result.IsStateful)
	base.CreatedAt = ConvertNullableTime(result.CreatedAt)
	base.UpdatedAt = ConvertNullableTime(result.UpdatedAt)
	base.Rules = rules

	if respDiags.HasError() {
		return false
	}
	return true
}

func mapSecurityGroupBaseModelFromList(
	ctx context.Context,
	base *securityGroupBaseModel,
	sgResult *network.BnsNetworkV1ApiListSecurityGroupsModelSecurityGroupModel,
	respDiags *diag.Diagnostics,
) {
	var getLike network.BnsNetworkV1ApiGetSecurityGroupModelSecurityGroupModel
	if err := copier.Copy(&getLike, sgResult); err != nil {
		respDiags.AddError("Mapping failed", fmt.Sprintf("copier.Copy failed: %v", err))
		return
	}
	mapSecurityGroupBaseModel(ctx, base, &getLike, respDiags)
}

func (d *securityGroupDataSource) mapSecurityGroup(
	ctx context.Context,
	model *securityGroupDataSourceModel,
	sgResult *network.BnsNetworkV1ApiGetSecurityGroupModelSecurityGroupModel,
	respDiags *diag.Diagnostics,
) bool {
	mapSecurityGroupBaseModel(ctx, &model.securityGroupBaseModel, sgResult, respDiags)

	if respDiags.HasError() {
		return false
	}
	return true
}

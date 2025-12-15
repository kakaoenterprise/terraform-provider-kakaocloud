// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package network

import (
	"context"
	"fmt"
	"strconv"
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

	if result.Rules != nil && len(result.Rules) != 0 {
		rules, ruleDiags := ConvertSetFromModel(
			ctx,
			result.Rules,
			securityGroupRuleAttrType,
			func(src network.BnsNetworkV1ApiGetSecurityGroupModelSecurityGroupRuleModel) any {

				portMin := ConvertNullableString(src.PortRangeMin)
				portMax := ConvertNullableString(src.PortRangeMax)

				portMinInt := types.Int32Null()
				if !portMin.IsNull() && !portMin.IsUnknown() {
					if portMin.ValueString() == "ALL" {
						portMinInt = types.Int32Value(1)
					} else {
						if v, err := strconv.Atoi(portMin.ValueString()); err == nil {
							portMinInt = types.Int32Value(int32(v))
						}
					}
				}
				portMaxInt := types.Int32Null()
				if !portMax.IsNull() && !portMax.IsUnknown() {
					if portMax.ValueString() == "ALL" {
						portMaxInt = types.Int32Value(65535)
					} else {
						if v, err := strconv.Atoi(portMax.ValueString()); err == nil {
							portMaxInt = types.Int32Value(int32(v))
						}
					}
				}

				return securityGroupRuleModel{
					Id:              types.StringValue(src.Id),
					Description:     ConvertNullableString(src.Description),
					RemoteGroupId:   ConvertNullableString(src.RemoteGroupId),
					RemoteGroupName: ConvertNullableString(src.RemoteGroupName),
					Direction:       ConvertNullableString(src.Direction),
					Protocol:        types.StringValue(string(src.Protocol)),
					PortRangeMin:    portMinInt,
					PortRangeMax:    portMaxInt,
					RemoteIpPrefix:  ConvertNullableIPPrefix(src.RemoteIpPrefix),
					CreatedAt:       ConvertNullableTime(src.CreatedAt),
					UpdatedAt:       ConvertNullableTime(src.UpdatedAt),
				}
			},
		)
		respDiags.Append(ruleDiags...)
		base.Rules = rules
	} else {
		base.Rules = types.SetNull(
			types.ObjectType{AttrTypes: securityGroupRuleAttrType},
		)
	}

	base.Id = types.StringValue(result.Id)
	base.Name = ConvertNullableString(result.Name)
	base.Description = ConvertNullableString(result.Description)
	base.ProjectId = ConvertNullableString(result.ProjectId)
	base.ProjectName = ConvertNullableString(result.ProjectName)
	base.IsStateful = ConvertNullableBool(result.IsStateful)
	base.CreatedAt = ConvertNullableTime(result.CreatedAt)
	base.UpdatedAt = ConvertNullableTime(result.UpdatedAt)

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

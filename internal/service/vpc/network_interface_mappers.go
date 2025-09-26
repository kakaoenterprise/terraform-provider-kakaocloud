// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package vpc

import (
	. "terraform-provider-kakaocloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework-nettypes/iptypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kakaoenterprise/kc-sdk-go/services/vpc"
	"golang.org/x/net/context"
)

func mapNetworkInterfaceBaseModel(
	ctx context.Context,
	base *networkInterfaceBaseModel,
	niResult *vpc.BnsVpcV1ApiGetNetworkInterfaceModelNetworkInterfaceModel,
	respDiags *diag.Diagnostics,
) bool {
	allowedPairs, aadDiags := ConvertSetFromModel(
		ctx,
		niResult.AllowedAddressPairs,
		allowedAddressPairAttrType,
		func(src vpc.BnsVpcV1ApiGetNetworkInterfaceModelAllowedAddressPairModel) any {
			return allowedAddressPairModel{
				MacAddress: ConvertNullableString(src.MacAddress),
				IpAddress:  ConvertNullableString(src.IpAddress),
			}
		},
	)
	respDiags.Append(aadDiags...)

	securityGroups, sgDiags := ConvertSetFromModel(
		ctx,
		niResult.SecurityGroups,
		securityGroupAttrType,
		func(src vpc.SecurityGroupModel) any {
			return securityGroupModel{
				Id:   types.StringValue(src.Id),
				Name: ConvertNullableString(src.Name),
			}
		},
	)
	respDiags.Append(sgDiags...)

	secondaryIps, secIpDiags := types.ListValueFrom(ctx, types.StringType, niResult.SecondaryIps)
	respDiags.Append(secIpDiags...)

	base.Id = types.StringValue(niResult.Id)
	base.Name = ConvertNullableString(niResult.Name)
	base.Status = ConvertNullableString(niResult.Status)
	base.Description = ConvertNullableString(niResult.Description)
	base.ProjectId = ConvertNullableString(niResult.ProjectId)
	base.VpcId = ConvertNullableString(niResult.VpcId)
	base.SubnetId = ConvertNullableString(niResult.SubnetId)
	base.MacAddress = ConvertNullableString(niResult.MacAddress)
	base.DeviceId = ConvertNullableString(niResult.DeviceId)
	base.DeviceOwner = ConvertNullableString(niResult.DeviceOwner)
	base.ProjectName = ConvertNullableString(niResult.ProjectName)
	base.SecondaryIps = secondaryIps
	base.PublicIp = ConvertNullableString(niResult.PublicIp)
	base.PrivateIp = iptypes.NewIPAddressValue(*niResult.PrivateIp.Get())
	base.IsNetworkInterfaceSecurityEnabled = ConvertNullableBool(niResult.IsNetworkInterfaceSecurityEnabled)
	base.AllowedAddressPairs = allowedPairs
	base.SecurityGroups = securityGroups
	base.CreatedAt = ConvertNullableTime(niResult.CreatedAt)
	base.UpdatedAt = ConvertNullableTime(niResult.UpdatedAt)

	if respDiags.HasError() {
		return false
	}
	return true
}

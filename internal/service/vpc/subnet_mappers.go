// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package vpc

import (
	. "terraform-provider-kakaocloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework-nettypes/cidrtypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/kakaoenterprise/kc-sdk-go/v2/services/vpc"
)

func mapSubnetModel(
	base *subnetBaseModel,
	result *vpc.Subnet,
	respDiags *diag.Diagnostics,
) bool {
	base.Id = ConvertNullableString(result.Id)
	base.Name = ConvertNullableString(result.Name)
	base.Description = ConvertNullableString(result.Description)
	base.IsShared = ConvertNullableBool(result.IsShared)
	base.AvailabilityZone = ConvertNullableString(result.AvailabilityZone)
	base.CidrBlock = cidrtypes.NewIPPrefixValue(*result.CidrBlock.Get())
	base.ProjectId = ConvertNullableString(result.ProjectId)
	base.ProvisioningStatus = ConvertNullableString(result.ProvisioningStatus)
	base.VpcId = ConvertNullableString(result.VpcId)
	base.VpcName = ConvertNullableString(result.VpcName)
	base.ProjectName = ConvertNullableString(result.ProjectName)
	base.OwnerProjectId = ConvertNullableString(result.OwnerProjectId)
	base.RouteTableId = ConvertNullableString(result.RouteTableId)
	base.RouteTableName = ConvertNullableString(result.RouteTableName)
	base.CreatedAt = ConvertNullableTime(result.CreatedAt)
	base.UpdatedAt = ConvertNullableTime(result.UpdatedAt)

	return !respDiags.HasError()
}

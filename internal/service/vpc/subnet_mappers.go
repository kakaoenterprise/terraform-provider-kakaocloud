// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package vpc

import (
	. "terraform-provider-kakaocloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework-nettypes/cidrtypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kakaoenterprise/kc-sdk-go/services/vpc"
)

func mapSubnetBaseModel(
	base *subnetBaseModel,
	result *vpc.BnsVpcV1ApiGetSubnetModelSubnetModel,
	respDiags *diag.Diagnostics,
) bool {
	base.Id = types.StringValue(result.Id)
	base.Name = ConvertNullableString(result.Name)
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

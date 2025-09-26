// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package network

import (
	"context"
	"strings"
	. "terraform-provider-kakaocloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kakaoenterprise/kc-sdk-go/services/network"
)

func mapPublicIpBaseModel(
	ctx context.Context,
	base *publicIpBaseModel,
	publicIpResult *network.BnsNetworkV1ApiGetPublicIpModelFloatingIpModel,
	respDiags *diag.Diagnostics,
) bool {
	base.Id = types.StringValue(publicIpResult.Id)
	base.Status = ConvertNullableString(publicIpResult.Status)
	base.Description = ConvertNullableString(publicIpResult.Description)
	base.ProjectId = ConvertNullableString(publicIpResult.ProjectId)
	base.PublicIp = ConvertNullableString(publicIpResult.PublicIp)
	base.PrivateIp = ConvertNullableString(publicIpResult.PrivateIp)
	base.CreatedAt = ConvertNullableTime(publicIpResult.CreatedAt)
	base.UpdatedAt = ConvertNullableTime(publicIpResult.UpdatedAt)

	relatedResourceObj, relatedResourceDiags := ConvertObjectFromModel(
		ctx, publicIpResult.RelatedResource, relatedResourceAttrType,
		func(src network.BnsNetworkV1ApiGetPublicIpModelRelatedResourceInfoModel) any {
			dt := types.StringNull()
			if dt.IsNull() || dt.IsUnknown() {
				owner := ""
				if src.DeviceOwner.Get() != nil {
					owner = *src.DeviceOwner.Get()
				}
				switch {
				case strings.HasPrefix(owner, "compute:"):
					dt = types.StringValue("instance")
				case owner == "network:f5listener":
					dt = types.StringValue("load-balancer")
				}
			}

			return resourceModel{
				Id:          types.StringValue(src.Id),
				Name:        ConvertNullableString(src.Name),
				Status:      ConvertNullableString(src.Status),
				Type:        ConvertNullableString(src.Type),
				DeviceId:    ConvertNullableString(src.DeviceId),
				DeviceOwner: ConvertNullableString(src.DeviceOwner),
				DeviceType:  dt,
				SubnetId:    ConvertNullableString(src.SubnetId),
				SubnetName:  ConvertNullableString(src.SubnetName),
				SubnetCIDR:  ConvertNullableString(src.SubnetCidr),
				VpcId:       ConvertNullableString(src.VpcId),
				VpcName:     ConvertNullableString(src.VpcName),
			}
		},
	)

	respDiags.Append(relatedResourceDiags...)
	if respDiags.HasError() {
		return false
	}

	base.RelatedResource = relatedResourceObj

	return true
}

func (d *publicIpsDataSource) mapPublicIps(
	ctx context.Context,
	base *publicIpBaseModel,
	publicIpResult *network.BnsNetworkV1ApiListPublicIpsModelFloatingIpModel,
	respDiags *diag.Diagnostics,
) bool {
	base.Id = types.StringValue(publicIpResult.Id)
	base.Status = ConvertNullableString(publicIpResult.Status)
	base.Description = ConvertNullableString(publicIpResult.Description)
	base.ProjectId = ConvertNullableString(publicIpResult.ProjectId)
	base.PublicIp = ConvertNullableString(publicIpResult.PublicIp)
	base.PrivateIp = ConvertNullableString(publicIpResult.PrivateIp)
	base.CreatedAt = ConvertNullableTime(publicIpResult.CreatedAt)
	base.UpdatedAt = ConvertNullableTime(publicIpResult.UpdatedAt)

	relatedResourceObj, relatedResourceDiags := ConvertObjectFromModel(ctx, publicIpResult.RelatedResource, relatedResourceAttrType,
		func(src network.BnsNetworkV1ApiListPublicIpsModelRelatedResourceInfoModel) any {
			dt := types.StringNull()
			if dt.IsNull() || dt.IsUnknown() {
				owner := ""
				if src.DeviceOwner.Get() != nil {
					owner = *src.DeviceOwner.Get()
				}
				switch {
				case strings.HasPrefix(owner, "compute:"):
					dt = types.StringValue("instance")
				case owner == "network:f5listener":
					dt = types.StringValue("load-balancer")
				}
			}

			return resourceModel{
				Id:          types.StringValue(src.Id),
				Name:        ConvertNullableString(src.Name),
				Status:      ConvertNullableString(src.Status),
				Type:        ConvertNullableString(src.Type),
				DeviceId:    ConvertNullableString(src.DeviceId),
				DeviceOwner: ConvertNullableString(src.DeviceOwner),
				DeviceType:  dt,
				SubnetId:    ConvertNullableString(src.SubnetId),
				SubnetName:  ConvertNullableString(src.SubnetName),
				SubnetCIDR:  ConvertNullableString(src.SubnetCidr),
				VpcId:       ConvertNullableString(src.VpcId),
				VpcName:     ConvertNullableString(src.VpcName),
			}
		},
	)

	respDiags.Append(relatedResourceDiags...)

	if respDiags.HasError() {
		return false
	}

	base.RelatedResource = relatedResourceObj

	return true
}

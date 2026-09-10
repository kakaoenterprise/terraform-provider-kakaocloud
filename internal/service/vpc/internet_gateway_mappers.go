// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package vpc

import (
	"context"

	. "terraform-provider-kakaocloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kakaoenterprise/kc-sdk-go/v2/services/vpc"
)

func mapInternetGatewayModel(
	ctx context.Context,
	base *internetGatewayBaseModel,
	igwResult *vpc.Igw,
	respDiags *diag.Diagnostics,
) bool {
	if base == nil || igwResult == nil {
		return false
	}

	vpcObject, vpcDiags := ConvertObjectFromModel(ctx, igwResult.Vpc, igwVpcAttrType, func(src vpc.InternetGatewayVpc) any {
		return igwVpcModel{
			Id:        types.StringValue(src.Id),
			Name:      ConvertNullableString(src.Name),
			CidrBlock: ConvertNullableString(src.CidrBlock),
		}
	})
	respDiags.Append(vpcDiags...)

	attachment, attachmentDiags := ConvertObjectFromModel(ctx, igwResult.Attachment, igwAttachmentAttrType, func(src vpc.Attachment) any {
		return igwAttachmentModel{
			Id:                 types.StringValue(src.Id),
			ProvisioningStatus: ConvertNullableString(src.ProvisioningStatus),
		}
	})
	respDiags.Append(attachmentDiags...)

	if respDiags.HasError() {
		return false
	}

	base.Id = types.StringValue(igwResult.Id)
	base.Name = types.StringValue(igwResult.Name)
	base.Description = ConvertNullableString(igwResult.Description)
	base.Vpc = vpcObject
	base.Attachment = attachment
	base.AttachmentStatus = types.StringValue(igwResult.AttachmentStatus)
	base.NatIp = ConvertNullableString(igwResult.NatIp)
	base.ProjectId = ConvertNullableString(igwResult.ProjectId)
	base.ProjectName = ConvertNullableString(igwResult.ProjectName)
	base.DomainId = ConvertNullableString(igwResult.DomainId)
	base.ProvisioningStatus = ConvertNullableString(igwResult.ProvisioningStatus)
	base.CreatedAt = ConvertNullableTime(igwResult.CreatedAt)
	base.UpdatedAt = ConvertNullableTime(igwResult.UpdatedAt)

	return true
}

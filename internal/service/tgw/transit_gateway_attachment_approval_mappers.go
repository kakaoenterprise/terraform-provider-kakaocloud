// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package tgw

import (
	"context"
	"terraform-provider-kakaocloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kakaoenterprise/kc-sdk-go/services/tgw"
)

func mapAttachmentApprovalToModel(
	ctx context.Context,
	model *transitGatewayAttachmentApprovalResourceModel,
	attachment *tgw.BnsTgwV1ApiGetTgwAttachmentModelTgwAttachmentResponseModel,
	respDiags *diag.Diagnostics,
) bool {
	model.Id = types.StringValue(attachment.Id)
	model.TgwId = utils.ConvertNullableString(attachment.Tgw.Id)
	model.VpcId = utils.ConvertNullableString(attachment.ResourceId)
	model.TgwProjectId = utils.ConvertNullableString(attachment.Tgw.ProjectId)
	model.VpcName = utils.ConvertNullableString(attachment.ResourceName)
	model.CidrBlock = utils.ConvertNullableString(attachment.ResourceCidrBlock)
	model.ProjectId = utils.ConvertNullableString(attachment.ProjectId)
	model.CreatedAt = utils.ConvertNullableTime(attachment.CreatedAt)
	model.UpdatedAt = utils.ConvertNullableTime(attachment.UpdatedAt)
	model.ProvisioningStatus = utils.ConvertNullableString(attachment.ProvisioningStatus)

	return !respDiags.HasError()
}

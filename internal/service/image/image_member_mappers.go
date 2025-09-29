// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package image

import (
	"context"
	. "terraform-provider-kakaocloud/internal/utils"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kakaoenterprise/kc-sdk-go/services/image"
)

func mapImageMemberModel(
	ctx context.Context,
	base *imageMemberBaseModel,
	membersResult []image.BcsImageV1ApiListImageSharedProjectsModelImageMemberModel,
	respDiags *diag.Diagnostics,
) bool {
	members, membersDiags := ConvertListFromModel(ctx, membersResult, imageMemberMembersAttrType, func(src image.BcsImageV1ApiListImageSharedProjectsModelImageMemberModel) any {
		return imageMemberMemberModel{
			Id:        types.StringValue(src.Id),
			CreatedAt: types.StringValue(src.CreatedAt.Format(time.RFC3339)),
			UpdatedAt: types.StringValue(src.UpdatedAt.Format(time.RFC3339)),
			ImageId:   types.StringValue(src.ImageId),
			Status:    ConvertNullableString(src.Status),
			IsShared:  types.BoolValue(src.IsShared),
		}
	})
	respDiags.Append(membersDiags...)
	base.Members = members

	if respDiags.HasError() {
		return false
	}
	return true
}

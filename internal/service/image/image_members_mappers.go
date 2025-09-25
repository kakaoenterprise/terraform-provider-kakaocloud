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
	imageMemberResult *image.BcsImageV1ApiListImageSharedProjectsModelImageMemberModel,
	respDiags *diag.Diagnostics,
) bool {
	base.Id = types.StringValue(imageMemberResult.Id)
	base.CreatedAt = types.StringValue(imageMemberResult.CreatedAt.Format(time.RFC3339))
	base.UpdatedAt = types.StringValue(imageMemberResult.UpdatedAt.Format(time.RFC3339))
	base.ImageId = types.StringValue(imageMemberResult.ImageId)
	base.Status = ConvertNullableString(imageMemberResult.Status)
	base.IsShared = types.BoolValue(imageMemberResult.IsShared)

	if respDiags.HasError() {
		return false
	}
	return true
}

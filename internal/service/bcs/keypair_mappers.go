// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package bcs

import (
	"context"
	. "terraform-provider-kakaocloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kakaoenterprise/kc-sdk-go/services/bcs"
)

func mapKeypairBaseModel(
	ctx context.Context,
	base *keypairBaseModel,
	keypairResult *bcs.BcsInstanceV1ApiGetKeypairModelKeypairModel,
	respDiags *diag.Diagnostics,
) bool {
	base.Id = types.StringValue(keypairResult.GetId())
	base.Name = ConvertNullableString(keypairResult.Name)
	base.Fingerprint = ConvertNullableString(keypairResult.Fingerprint)
	base.PublicKey = ConvertNullableString(keypairResult.PublicKey)
	base.UserId = ConvertNullableString(keypairResult.UserId)
	base.Type = ConvertNullableString(keypairResult.Type)
	base.CreatedAt = ConvertNullableTime(keypairResult.CreatedAt)

	if respDiags.HasError() {
		return false
	}
	return true
}

func (d *keypairDataSource) mapKeypair(
	ctx context.Context,
	model *keypairDataSourceModel,
	keypairResult *bcs.BcsInstanceV1ApiGetKeypairModelKeypairModel,
	respDiags *diag.Diagnostics,
) bool {
	mapKeypairBaseModel(ctx, &model.keypairBaseModel, keypairResult, respDiags)

	if respDiags.HasError() {
		return false
	}

	return true
}

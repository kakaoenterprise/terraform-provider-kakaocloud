// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package bcs

import (
	. "terraform-provider-kakaocloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/kakaoenterprise/kc-sdk-go/v2/services/bcs"
)

func mapKeypairBaseModel(
	base *keypairBaseModel,
	keypairResult *bcs.Keypair,
	respDiags *diag.Diagnostics,
) bool {
	base.Id = ConvertNullableString(keypairResult.Id)
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

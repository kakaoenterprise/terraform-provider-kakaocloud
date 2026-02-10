// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package iam

import (
	"context"
	. "terraform-provider-kakaocloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kakaoenterprise/kc-sdk-go/services/iam"
)

func mapProjectDataSourceModel(
	ctx context.Context,
	base *projectDataSourceModel,
	result *iam.TokenProject,
	respDiags *diag.Diagnostics,
) bool {
	domainObj, imageDiags := ConvertNonNullableObjectFromModel(ctx, result.Domain, projectDomainAttrType, func(src *iam.TokenDomain) any {
		return projectDomainModel{
			Id:   types.StringValue(src.GetId()),
			Name: types.StringValue(src.GetName()),
		}
	})
	respDiags.Append(imageDiags...)
	base.Domain = domainObj

	base.Id = types.StringValue(result.GetId())
	base.Name = types.StringValue(result.GetName())

	if respDiags.HasError() {
		return false
	}
	return true
}

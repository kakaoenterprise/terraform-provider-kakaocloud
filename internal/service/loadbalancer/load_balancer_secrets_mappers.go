// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package loadbalancer

import (
	"terraform-provider-kakaocloud/internal/utils"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kakaoenterprise/kc-sdk-go/services/loadbalancer"
	"golang.org/x/net/context"
)

func mapLoadBalancerSecretsBaseModel(
	ctx context.Context,
	base *loadBalancerSecretBaseModel,
	src *loadbalancer.BnsLoadBalancerV1ApiListTlsCertificatesModelSecretModel,
	diags *diag.Diagnostics,
) bool {

	contentTypes, lbDiags := utils.ConvertNonNullableObjectFromModel(ctx, src.ContentTypes, lbSecretsContentTypeAttrType, func(lb loadbalancer.ContentType) any {
		return ContentTypeModel{
			Default: types.StringValue(lb.Default),
		}
	})
	diags.Append(lbDiags...)

	base.CreatedAt = types.StringValue(src.CreatedAt.Format(time.RFC3339))
	base.UpdatedAt = types.StringValue(src.UpdatedAt.Format(time.RFC3339))
	base.Status = types.StringValue(src.Status)
	base.Name = types.StringValue(src.Name)
	base.SecretType = types.StringValue(src.SecretType)
	base.Expiration = types.StringValue(src.Expiration)
	base.CreatorId = types.StringValue(src.CreatorId)
	base.SecretRef = types.StringValue(src.SecretRef)
	base.ContentTypes = contentTypes

	return !diags.HasError()
}

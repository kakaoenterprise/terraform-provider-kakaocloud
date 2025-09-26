// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package loadbalancer

import (
	"context"
	"terraform-provider-kakaocloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/kakaoenterprise/kc-sdk-go/services/loadbalancer"
)

func mapLoadBalancerFlavor(
	ctx context.Context,
	base *loadBalancerFlavorBaseModel,
	lbf *loadbalancer.FlavorModel,
	diags *diag.Diagnostics,
) bool {

	if lbf == nil {
		diags.AddError("Mapping Error", "Load balancer flavor source is nil")
		return false
	}

	base.Id = types.StringValue(lbf.Id)

	base.Name = utils.ConvertNullableString(lbf.Name)
	base.Description = utils.ConvertNullableString(lbf.Description)
	base.IsEnabled = utils.ConvertNullableBool(lbf.IsEnabled)

	return !diags.HasError()
}

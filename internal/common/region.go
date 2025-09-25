// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"strings"
)

const (
	RegionKR1 = "kr-central-1"
	RegionKR2 = "kr-central-2"
	RegionKR3 = "kr-central-3"
)

var RegionAll = []string{
	RegionKR1,
	RegionKR2,
	RegionKR3,
}

func RegionValidators() []validator.String {
	return []validator.String{
		regionCustomValidator{},
	}
}

type regionCustomValidator struct{}

func (v regionCustomValidator) Description(ctx context.Context) string {
	return "Region must be one of: " + strings.Join(RegionAll, ", ")
}

func (v regionCustomValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v regionCustomValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	val := req.ConfigValue.ValueString()
	for _, allowed := range RegionAll {
		if val == allowed {
			return
		}
	}

	resp.Diagnostics.AddAttributeError(
		req.Path,
		"Invalid Region",
		fmt.Sprintf("The region '%s' is not valid. Must be one of: %s", val, strings.Join(RegionAll, ", ")),
	)
}

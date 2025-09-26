package // Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
common

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type nullToUnknownString struct{}

func (m nullToUnknownString) Description(context.Context) string {
	return "Convert null to unknown so computed values can be filled by the API"
}
func (m nullToUnknownString) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}
func (m nullToUnknownString) PlanModifyString(
	_ context.Context,
	req planmodifier.StringRequest,
	resp *planmodifier.StringResponse,
) {
	if req.PlanValue.IsNull() {
		resp.PlanValue = types.StringUnknown()
	}
}

func NullToUnknownString() planmodifier.String {
	return nullToUnknownString{}
}

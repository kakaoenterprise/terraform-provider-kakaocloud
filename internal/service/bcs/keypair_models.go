// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package bcs

import (
	"terraform-provider-kakaocloud/internal/common"

	datasourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	resourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type keypairBaseModel struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	UserId      types.String `tfsdk:"user_id"`
	Fingerprint types.String `tfsdk:"fingerprint"`
	PublicKey   types.String `tfsdk:"public_key"`
	Type        types.String `tfsdk:"type"`
	CreatedAt   types.String `tfsdk:"created_at"`
}

type keypairsDataSourceModel struct {
	Keypairs []keypairBaseModel       `tfsdk:"keypairs"`
	Filter   []common.FilterModel     `tfsdk:"filter"`
	Timeouts datasourceTimeouts.Value `tfsdk:"timeouts"`
}

type keypairDataSourceModel struct {
	keypairBaseModel
	Timeouts datasourceTimeouts.Value `tfsdk:"timeouts"`
}

type keypairResourceModel struct {
	keypairBaseModel
	PrivateKey types.String           `tfsdk:"private_key"`
	Timeouts   resourceTimeouts.Value `tfsdk:"timeouts"`
}

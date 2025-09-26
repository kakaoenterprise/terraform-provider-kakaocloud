// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package loadbalancer

import (
	datasourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-kakaocloud/internal/common"
)

type ContentTypeModel struct {
	Default types.String `tfsdk:"default"`
}

type loadBalancerSecretBaseModel struct {
	Name         types.String `tfsdk:"name"`
	Expiration   types.String `tfsdk:"expiration"`
	Status       types.String `tfsdk:"status"`
	SecretType   types.String `tfsdk:"secret_type"`
	CreatorId    types.String `tfsdk:"creator_id"`
	CreatedAt    types.String `tfsdk:"created_at"`
	UpdatedAt    types.String `tfsdk:"updated_at"`
	ContentTypes types.Object `tfsdk:"content_types"`
	SecretRef    types.String `tfsdk:"secret_ref"`
}

type loadBalancerListenerSecretsModel struct {
	Id         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Expiration types.String `tfsdk:"expiration"`
	Status     types.String `tfsdk:"status"`
	SecretType types.String `tfsdk:"secret_type"`
	CreatorId  types.String `tfsdk:"creator_id"`
	IsDefault  types.Bool   `tfsdk:"is_default"`
}

type loadBalancerSecretsDataSourceModel struct {
	Filter              []common.FilterModel          `tfsdk:"filter"`
	LoadBalancerSecrets []loadBalancerSecretBaseModel `tfsdk:"secrets"`
	Timeouts            datasourceTimeouts.Value      `tfsdk:"timeouts"`
}

var lbSecretsContentTypeAttrType = map[string]attr.Type{
	"default": types.StringType,
}

var loadBalancerListenerSecretAttrType = map[string]attr.Type{
	"id":          types.StringType,
	"name":        types.StringType,
	"expiration":  types.StringType,
	"status":      types.StringType,
	"secret_type": types.StringType,
	"is_default":  types.BoolType,
	"creator_id":  types.StringType,
}

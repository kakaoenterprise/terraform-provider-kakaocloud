// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kubernetesengine

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kakaoenterprise/kc-sdk-go/services/kubernetesengine"
)

func (d *kubernetesVersionsDataSource) mapVersions(
	ctx context.Context,
	base *versionBaseModel,
	result *kubernetesengine.KubernetesEngineV1ApiListAvailableKubernetesVersionsModelUpgradableVersionResponseModel,
	respDiags *diag.Diagnostics,
) bool {
	base.IsDeprecated = types.BoolValue(result.IsDeprecated)
	base.MinorVersion = types.StringValue(result.MinorVersion)
	base.PatchVersion = types.StringValue(result.PatchVersion)
	base.Eol = types.StringValue(result.Eol)
	base.NextVersion = types.StringValue(result.NextVersion)

	return true
}

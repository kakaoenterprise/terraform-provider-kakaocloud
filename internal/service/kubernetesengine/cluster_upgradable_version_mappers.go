// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package kubernetesengine

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kakaoenterprise/kc-sdk-go/services/kubernetesengine"
)

func (d *upgradableVersionDataSource) mapUpgradableVersions(
	ctx context.Context,
	base *upgradableVersionsDataSourceModel,
	result *kubernetesengine.UpgradeResponseModel,
	respDiags *diag.Diagnostics,
) bool {
	currentVal, currentDiags := types.ObjectValueFrom(ctx, versionAttrTypes, versionBaseModel{
		IsDeprecated: types.BoolValue(result.Current.IsDeprecated),
		Eol:          types.StringValue(result.Current.Eol),
		MinorVersion: types.StringValue(result.Current.MinorVersion),
		NextVersion:  types.StringValue(result.Current.NextVersion),
		PatchVersion: types.StringValue(result.Current.PatchVersion),
	})
	respDiags.Append(currentDiags...)
	base.Current = currentVal
	if respDiags.HasError() {
		return false
	}

	base.Upgradable = make([]versionBaseModel, 0, len(result.Upgradable))
	for _, v := range result.Upgradable {
		base.Upgradable = append(base.Upgradable, versionBaseModel{
			IsDeprecated: types.BoolValue(v.IsDeprecated),
			Eol:          types.StringValue(v.Eol),
			MinorVersion: types.StringValue(v.MinorVersion),
			NextVersion:  types.StringValue(v.NextVersion),
			PatchVersion: types.StringValue(v.PatchVersion),
		})
	}
	if respDiags.HasError() {
		return false
	}

	return true
}

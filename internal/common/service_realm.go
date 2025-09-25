// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

const (
	ServiceRealmStage  = "stage"
	ServiceRealmPublic = "public"
	ServiceRealmGov    = "gov"
)

var ServiceRealmAll = []string{
	ServiceRealmStage,
	ServiceRealmPublic,
	ServiceRealmGov,
}

func ServiceRealmValidators() []validator.String {
	return []validator.String{
		stringvalidator.OneOf(ServiceRealmAll...),
	}
}

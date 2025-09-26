// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package vpc

import (
	"terraform-provider-kakaocloud/internal/common"

	"github.com/hashicorp/terraform-plugin-framework-nettypes/cidrtypes"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var subnetDataSourceSchemaAttributes = map[string]dschema.Attribute{
	"name":              dschema.StringAttribute{Computed: true},
	"is_shared":         dschema.BoolAttribute{Computed: true},
	"availability_zone": dschema.StringAttribute{Computed: true},
	"cidr_block": dschema.StringAttribute{
		Computed:   true,
		CustomType: cidrtypes.IPPrefixType{},
	},
	"project_id":          dschema.StringAttribute{Computed: true},
	"provisioning_status": dschema.StringAttribute{Computed: true},
	"vpc_id":              dschema.StringAttribute{Computed: true},
	"vpc_name":            dschema.StringAttribute{Computed: true},
	"project_name":        dschema.StringAttribute{Computed: true},
	"owner_project_id":    dschema.StringAttribute{Computed: true},
	"route_table_id":      dschema.StringAttribute{Computed: true},
	"route_table_name":    dschema.StringAttribute{Computed: true},
	"created_at":          dschema.StringAttribute{Computed: true},
	"updated_at":          dschema.StringAttribute{Computed: true},
}

var subnetResourceSchemaAttributes = map[string]rschema.Attribute{
	"id": rschema.StringAttribute{
		Computed: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	},
	"name": rschema.StringAttribute{
		Required:   true,
		Validators: common.NameValidator(200),
	},
	"is_shared": rschema.BoolAttribute{Computed: true},
	"availability_zone": rschema.StringAttribute{
		Required: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	},
	"cidr_block": rschema.StringAttribute{
		Required:   true,
		CustomType: cidrtypes.IPPrefixType{},
		Validators: []validator.String{
			common.NewCIDRPrefixLengthValidator(20, 26),
		},
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	},
	"project_id":          rschema.StringAttribute{Computed: true},
	"provisioning_status": rschema.StringAttribute{Computed: true},
	"vpc_id": rschema.StringAttribute{
		Required:   true,
		Validators: common.UuidValidator(),
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	},
	"vpc_name":         rschema.StringAttribute{Computed: true},
	"project_name":     rschema.StringAttribute{Computed: true},
	"owner_project_id": rschema.StringAttribute{Computed: true},
	"route_table_id":   rschema.StringAttribute{Computed: true},
	"route_table_name": rschema.StringAttribute{Computed: true},
	"created_at":       rschema.StringAttribute{Computed: true},
	"updated_at":       rschema.StringAttribute{Computed: true},
}

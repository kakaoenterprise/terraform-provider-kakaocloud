// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package vpc

import (
	resourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type routeTableAssociationResourceModel struct {
	Id           types.String           `tfsdk:"id"`
	SubnetIds    types.Set              `tfsdk:"subnet_ids"`
	Associations types.List             `tfsdk:"associations"`
	Timeouts     resourceTimeouts.Value `tfsdk:"timeouts"`
}

type associationModel struct {
	Id                 types.String `tfsdk:"id"`
	ProvisioningStatus types.String `tfsdk:"provisioning_status"`
	VpcId              types.String `tfsdk:"vpc_id"`
	VpcName            types.String `tfsdk:"vpc_name"`
	SubnetId           types.String `tfsdk:"subnet_id"`
	SubnetName         types.String `tfsdk:"subnet_name"`
	SubnetCidrBlock    types.String `tfsdk:"subnet_cidr_block"`
	AvailabilityZone   types.String `tfsdk:"availability_zone"`
}

var associationAttrType = map[string]attr.Type{
	"id":                  types.StringType,
	"provisioning_status": types.StringType,
	"vpc_id":              types.StringType,
	"vpc_name":            types.StringType,
	"subnet_id":           types.StringType,
	"subnet_name":         types.StringType,
	"subnet_cidr_block":   types.StringType,
	"availability_zone":   types.StringType,
}

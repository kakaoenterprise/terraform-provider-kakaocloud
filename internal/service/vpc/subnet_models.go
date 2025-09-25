// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpc

import (
	"github.com/hashicorp/terraform-plugin-framework-nettypes/cidrtypes"
	datasourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	resourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-kakaocloud/internal/common"
)

type subnetBaseModel struct {
	Id                 types.String       `tfsdk:"id"`
	Name               types.String       `tfsdk:"name"`
	IsShared           types.Bool         `tfsdk:"is_shared"`
	AvailabilityZone   types.String       `tfsdk:"availability_zone"`
	CidrBlock          cidrtypes.IPPrefix `tfsdk:"cidr_block"`
	ProjectId          types.String       `tfsdk:"project_id"`
	ProvisioningStatus types.String       `tfsdk:"provisioning_status"`
	VpcId              types.String       `tfsdk:"vpc_id"`
	VpcName            types.String       `tfsdk:"vpc_name"`
	ProjectName        types.String       `tfsdk:"project_name"`
	OwnerProjectId     types.String       `tfsdk:"owner_project_id"`
	RouteTableId       types.String       `tfsdk:"route_table_id"`
	RouteTableName     types.String       `tfsdk:"route_table_name"`
	CreatedAt          types.String       `tfsdk:"created_at"`
	UpdatedAt          types.String       `tfsdk:"updated_at"`
}

type subnetResourceModel struct {
	subnetBaseModel
	Timeouts resourceTimeouts.Value `tfsdk:"timeouts"`
}

type subnetDataSourceModel struct {
	subnetBaseModel
	Timeouts datasourceTimeouts.Value `tfsdk:"timeouts"`
}

type subnetsDataSourceModel struct {
	Filter   []common.FilterModel     `tfsdk:"filter"`
	Subnets  []subnetBaseModel        `tfsdk:"subnets"`
	Timeouts datasourceTimeouts.Value `tfsdk:"timeouts"`
}

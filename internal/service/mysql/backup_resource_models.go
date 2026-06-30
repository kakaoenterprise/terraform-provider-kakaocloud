// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package mysql

import (
	resourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type backupResourceModel struct {
	Id                types.String           `tfsdk:"id"`
	Name              types.String           `tfsdk:"name"`
	CreatedAt         types.String           `tfsdk:"created_at"`
	CreatorName       types.String           `tfsdk:"creator_name"`
	Description       types.String           `tfsdk:"description"`
	DiskSize          types.Int32            `tfsdk:"disk_size"`
	ExpireAt          types.String           `tfsdk:"expire_at"`
	ExpiryDuration    types.Int32            `tfsdk:"expiry_duration"`
	ExtraInfo         types.Object           `tfsdk:"extra_info"`
	InstanceGroupId   types.String           `tfsdk:"instance_group_id"`
	InstanceGroupName types.String           `tfsdk:"instance_group_name"`
	ProjectId         types.String           `tfsdk:"project_id"`
	Size              types.Int64            `tfsdk:"size"`
	Status            types.String           `tfsdk:"status"`
	Type              types.String           `tfsdk:"type"`
	StartedAt         types.String           `tfsdk:"started_at"`
	UpdatedAt         types.String           `tfsdk:"updated_at"`
	EngineVersion     types.String           `tfsdk:"engine_version"`
	Timeouts          resourceTimeouts.Value `tfsdk:"timeouts"`
}

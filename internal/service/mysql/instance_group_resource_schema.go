// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package mysql

import (
	"terraform-provider-kakaocloud/internal/common"

	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	mysqlsdk "github.com/kakaoenterprise/kc-sdk-go/services/mysql"
)

var instanceGroupResourceDesiredSubnetInfoSchemaAttributes = map[string]schema.Attribute{
	"replicas": schema.Int32Attribute{
		Required:   true,
		Validators: []validator.Int32{int32validator.Between(1, 6)},
	},
	"subnet_id": schema.StringAttribute{
		Required:   true,
		Validators: common.UuidValidator(),
	},
}

var instanceGroupResourceDesiredNetworkInfoSchemaAttributes = map[string]schema.Attribute{
	"primary_subnet_info": schema.SingleNestedAttribute{
		Required:   true,
		Attributes: instanceGroupResourceDesiredSubnetInfoSchemaAttributes,
	},
	"standby_subnet_info": schema.SetNestedAttribute{
		Optional: true,
		Validators: []validator.Set{
			setvalidator.SizeAtLeast(1),
		},
		NestedObject: schema.NestedAttributeObject{
			Attributes: instanceGroupResourceDesiredSubnetInfoSchemaAttributes,
		},
	},
	"security_group_ids": schema.SetAttribute{
		Required: true,
		Validators: []validator.Set{
			setvalidator.SizeAtLeast(1),
			setvalidator.ValueStringsAre(common.UuidValidator()...),
		},
		ElementType: types.StringType,
	},
}

var instanceGroupResourceSubnetInfoSchemaAttributes = map[string]schema.Attribute{
	"replicas":          schema.Int32Attribute{Computed: true},
	"availability_zone": schema.StringAttribute{Computed: true},
	"subnet_id":         schema.StringAttribute{Computed: true},
}

var instanceGroupResourceNetworkInfoSchemaAttributes = map[string]schema.Attribute{
	"primary_subnet_info": schema.SingleNestedAttribute{
		Computed:   true,
		Attributes: instanceGroupResourceSubnetInfoSchemaAttributes,
	},
	"standby_subnet_info": schema.SetNestedAttribute{
		Computed: true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: instanceGroupResourceSubnetInfoSchemaAttributes,
		},
	},
	"security_group_ids": schema.SetAttribute{
		Computed:    true,
		ElementType: types.StringType,
	},
}

var instanceGroupResourceSpecContentSchemaAttributes = map[string]schema.Attribute{
	"database_user_name": schema.StringAttribute{
		Required:   true,
		Validators: mysqlDatabaseUsernameValidator(),
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	},
	"database_user_password": schema.StringAttribute{
		Optional:   true,
		Sensitive:  true,
		WriteOnly:  true,
		Validators: mysqlDatabasePasswordValidator(),
	},
	"primary_port": schema.Int32Attribute{
		Required:   true,
		Validators: mysqlPortValidator(),
		PlanModifiers: []planmodifier.Int32{
			int32planmodifier.RequiresReplace(),
		},
	},
	"standby_port": schema.Int32Attribute{
		Optional:   true,
		Validators: mysqlPortValidator(),
	},
	"engine_version": schema.StringAttribute{
		Required: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	},
	"flavor_id": schema.StringAttribute{
		Required:   true,
		Validators: common.UuidValidator(),
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	},
	"vcpu": schema.Int32Attribute{
		Computed: true,
	},
	"memory": schema.Int32Attribute{
		Computed: true,
	},
	"log_disk_size": schema.Int32Attribute{
		Required:   true,
		Validators: mysqlDiskSizeValidator(),
		PlanModifiers: []planmodifier.Int32{
			common.PreventShrinkModifier[int32]{
				TypeName:        "MySQL log disk size",
				DescriptionText: "Prevents reducing the MySQL log disk size",
			},
		},
	},
	"data_disk_size": schema.Int32Attribute{
		Required:   true,
		Validators: mysqlDiskSizeValidator(),
		PlanModifiers: []planmodifier.Int32{
			common.PreventShrinkModifier[int32]{
				TypeName:        "MySQL data disk size",
				DescriptionText: "Prevents reducing the MySQL data disk size",
			},
		},
	},
	"instance_group_type": schema.StringAttribute{
		Computed: true,
	},
	"node_size": schema.Int32Attribute{
		Computed: true,
	},
}

var instanceGroupResourceBackupScheduleSchemaAttributes = map[string]schema.Attribute{
	"id": schema.StringAttribute{
		Computed: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	},
	"type": schema.StringAttribute{
		Optional: true,
		Computed: true,
		Validators: []validator.String{
			stringvalidator.OneOf(string(mysqlsdk.BACKUPSCHEDULETYPE_DAY)),
		},
	},
	"start_time": schema.StringAttribute{
		Optional:   true,
		Computed:   true,
		Validators: mysqlBackupScheduleStartTimeValidator(),
	},
	"expiry_duration": schema.Int32Attribute{
		Optional:   true,
		Computed:   true,
		Validators: mysqlBackupExpiryDurationValidator(),
	},
	"enabled": schema.BoolAttribute{
		Required: true,
	},
}

var instanceGroupResourceRestoreSourceSchemaAttributes = map[string]schema.Attribute{
	"type": schema.StringAttribute{
		Required: true,
		Validators: []validator.String{
			stringvalidator.OneOf(
				string(mysqlsdk.RESTORESOURCETYPE_BACKUP),
				string(mysqlsdk.RESTORESOURCETYPE_INSTANCE_GROUP),
			),
		},
	},
	"id": schema.StringAttribute{
		Required:   true,
		Validators: common.UuidValidator(),
	},
	"time": schema.StringAttribute{
		Optional: true,
	},
}

var instanceGroupResourceParameterGroupSchemaAttributes = map[string]schema.Attribute{
	"id": schema.StringAttribute{
		Required:   true,
		Validators: common.UuidValidator(),
	},
	"type": schema.StringAttribute{
		Required: true,
		Validators: []validator.String{
			stringvalidator.OneOf(
				string(mysqlsdk.PARAMETERGROUPTYPE_DEFAULT),
				string(mysqlsdk.PARAMETERGROUPTYPE_CUSTOM),
			),
		},
	},
	"apply_status": schema.StringAttribute{
		Computed: true,
	},
	"engine_version": schema.StringAttribute{
		Computed: true,
	},
	"is_engine_version_mismatch": schema.BoolAttribute{
		Computed: true,
	},
}

var instanceGroupResourceExtraInfoSchemaAttributes = map[string]schema.Attribute{
	"use_case_sensitive_table_names": schema.BoolAttribute{
		Optional: true,
		Computed: true,
	},
}

var instanceGroupResourceInstanceNodeSchemaAttributes = map[string]schema.Attribute{
	"instance_id":       schema.StringAttribute{Computed: true},
	"subnet_id":         schema.StringAttribute{Computed: true},
	"availability_zone": schema.StringAttribute{Computed: true},
}

var instanceGroupResourceInstancesSchemaAttributes = map[string]schema.Attribute{
	"primary": schema.SingleNestedAttribute{
		Computed:   true,
		Attributes: instanceGroupResourceInstanceNodeSchemaAttributes,
	},
	"standby": schema.ListNestedAttribute{
		Computed: true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: instanceGroupResourceInstanceNodeSchemaAttributes,
		},
	},
}

var instanceGroupResourceSchemaAttributes = map[string]schema.Attribute{
	"id": schema.StringAttribute{
		Computed: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	},
	"created_at": schema.StringAttribute{
		Computed: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	},
	"updated_at": schema.StringAttribute{
		Computed: true,
	},
	"license": schema.StringAttribute{
		Computed: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	},
	"name": schema.StringAttribute{
		Required:   true,
		Validators: mysqlInstanceGroupNameValidator(),
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	},
	"project_id": schema.StringAttribute{
		Computed: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	},
	"description": schema.StringAttribute{
		Optional:   true,
		Computed:   true,
		Validators: common.DescriptionValidator(),
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
			stringplanmodifier.RequiresReplace(),
		},
	},
	"creator": schema.StringAttribute{
		Computed: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	},
	"source_backup_id": schema.StringAttribute{
		Computed: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	},
	"is_multi_az": schema.BoolAttribute{
		Computed: true,
	},
	"endpoint": schema.ListAttribute{
		Computed:    true,
		ElementType: types.StringType,
	},
	"status": schema.StringAttribute{
		Computed: true,
	},
	"desired_network_info": schema.SingleNestedAttribute{
		Required:   true,
		Attributes: instanceGroupResourceDesiredNetworkInfoSchemaAttributes,
	},
	"network_info": schema.SingleNestedAttribute{
		Computed:   true,
		Attributes: instanceGroupResourceNetworkInfoSchemaAttributes,
	},
	"spec_content": schema.SingleNestedAttribute{
		Required:   true,
		Attributes: instanceGroupResourceSpecContentSchemaAttributes,
	},
	"source": schema.SingleNestedAttribute{
		Optional:   true,
		Attributes: instanceGroupResourceRestoreSourceSchemaAttributes,
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.RequiresReplace(),
		},
	},
	"backup_schedule": schema.SingleNestedAttribute{
		Required:   true,
		Attributes: instanceGroupResourceBackupScheduleSchemaAttributes,
	},
	"parameter_group": schema.SingleNestedAttribute{
		Required:   true,
		Attributes: instanceGroupResourceParameterGroupSchemaAttributes,
	},
	"extra_info": schema.SingleNestedAttribute{
		Optional:   true,
		Computed:   true,
		Attributes: instanceGroupResourceExtraInfoSchemaAttributes,
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.RequiresReplace(),
		},
	},
	"instances": schema.SingleNestedAttribute{
		Computed:   true,
		Attributes: instanceGroupResourceInstancesSchemaAttributes,
	},
}

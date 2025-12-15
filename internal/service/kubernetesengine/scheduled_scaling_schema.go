// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package kubernetesengine

import (
	"regexp"
	"terraform-provider-kakaocloud/internal/common"

	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/kakaoenterprise/kc-sdk-go/services/kubernetesengine"
)

func getScheduledScalingDataSourceSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"created_at": schema.StringAttribute{
			Computed: true,
		},
		"desired_nodes": schema.Int32Attribute{
			Computed: true,
		},
		"schedule": schema.StringAttribute{
			Computed: true,
		},
		"schedule_type": schema.StringAttribute{
			Computed: true,
		},
		"start_time": schema.StringAttribute{
			Computed: true,
		},
		"status": schema.SingleNestedAttribute{
			Computed:   true,
			Attributes: getScheduledScalingStatusDataSourceSchemaAttributes(),
		},
	}
}

func getScheduledScalingStatusDataSourceSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"histories": schema.ListNestedAttribute{
			Computed: true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: getScheduledScalingHistoryDataSourceSchemaAttributes(),
			},
		},
	}
}

func getScheduledScalingHistoryDataSourceSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"description": schema.StringAttribute{
			Computed: true,
		},
		"occurred_time": schema.StringAttribute{
			Computed: true,
		},
		"state": schema.StringAttribute{
			Computed: true,
		},
	}
}

func getScheduledScalingResourceSchema() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"cluster_name": rschema.StringAttribute{
			Required:   true,
			Validators: common.NameValidator(20),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"node_pool_name": rschema.StringAttribute{
			Required:   true,
			Validators: common.NameValidator(20),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"created_at": rschema.StringAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"desired_nodes": rschema.Int32Attribute{
			Required: true,
			Validators: []validator.Int32{
				int32validator.Between(0, 100),
			},
			PlanModifiers: []planmodifier.Int32{
				int32planmodifier.RequiresReplace(),
			},
		},
		"name": rschema.StringAttribute{
			Required:   true,
			Validators: common.NameValidator(20),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"schedule": rschema.StringAttribute{
			Optional:   true,
			Validators: common.CronScheduleValidator(),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"schedule_type": rschema.StringAttribute{
			Required: true,
			Validators: []validator.String{
				stringvalidator.OneOf(
					string(kubernetesengine.SCHEDULINGTYPE_ONCE),
					string(kubernetesengine.SCHEDULINGTYPE_CRON),
				),
			},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"start_time": rschema.StringAttribute{
			Required: true,
			Validators: []validator.String{
				stringvalidator.RegexMatches(
					regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:00Z$`),
					"start_time must be like 2025-09-22T16:00:00Z (seconds must be :00Z)",
				),
			},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"status": rschema.SingleNestedAttribute{
			Computed:   true,
			Attributes: getScheduledScalingStatusResourceSchemaAttributes(),
		},
	}
}

func getScheduledScalingStatusResourceSchemaAttributes() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"histories": rschema.ListNestedAttribute{
			Computed: true,
			NestedObject: rschema.NestedAttributeObject{
				Attributes: getScheduledScalingHistoryResourceSchemaAttributes(),
			},
		},
	}
}

func getScheduledScalingHistoryResourceSchemaAttributes() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"description": rschema.StringAttribute{
			Computed: true,
		},
		"occurred_time": rschema.StringAttribute{
			Computed: true,
		},
		"state": rschema.StringAttribute{
			Computed: true,
		},
	}
}

var scheduledScalingDataSourceSchemaAttributes = getScheduledScalingDataSourceSchema()
var scheduledScalingResourceSchema = getScheduledScalingResourceSchema()

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package kubernetesengine

import (
	"regexp"
	"terraform-provider-kakaocloud/internal/docs"

	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func getScheduledScalingDataSourceSchema() map[string]schema.Attribute {
	desc := docs.Kubernetesengine("ScheduledScaleResponseModel")

	return map[string]schema.Attribute{
		"created_at": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("created_at"),
		},
		"desired_nodes": schema.Int32Attribute{
			Computed:    true,
			Description: desc.String("desired_nodes"),
		},
		"schedule": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("schedule"),
		},
		"schedule_type": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("scheduling_type"),
		},
		"start_time": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("start_time"),
		},
		"status": schema.SingleNestedAttribute{
			Computed:    true,
			Description: desc.String("status"),
			Attributes:  getScheduledScalingStatusDataSourceSchemaAttributes(),
		},
	}
}

func getScheduledScalingStatusDataSourceSchemaAttributes() map[string]schema.Attribute {
	desc := docs.Kubernetesengine("ScalingStatusResponseModel")

	return map[string]schema.Attribute{
		"histories": schema.ListNestedAttribute{
			Computed:    true,
			Description: desc.String("histories"),
			NestedObject: schema.NestedAttributeObject{
				Attributes: getScheduledScalingHistoryDataSourceSchemaAttributes(),
			},
		},
	}
}

func getScheduledScalingHistoryDataSourceSchemaAttributes() map[string]schema.Attribute {
	desc := docs.Kubernetesengine("ScalingHistoryResponseModel")

	return map[string]schema.Attribute{
		"description": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("description"),
		},
		"occurred_time": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("occurred_time"),
		},
		"state": schema.StringAttribute{
			Computed:    true,
			Description: desc.String("state"),
		},
	}
}

func getScheduledScalingResourceSchema() map[string]rschema.Attribute {
	desc := docs.Kubernetesengine("ScheduledScaleResponseModel")

	return map[string]rschema.Attribute{
		"cluster_name": rschema.StringAttribute{
			Required:    true,
			Description: docs.ParameterDescription("kubernetesengine", "create_node_pool_scheduled_scaling", "path_cluster_name"),
			Validators: []validator.String{
				stringvalidator.LengthBetween(4, 20),
			},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"node_pool_name": rschema.StringAttribute{
			Required:    true,
			Description: docs.ParameterDescription("kubernetesengine", "create_node_pool_scheduled_scaling", "path_node_pool_name"),
			Validators: []validator.String{
				stringvalidator.LengthBetween(4, 20),
			},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"created_at": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("created_at"),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"desired_nodes": rschema.Int32Attribute{
			Required:    true,
			Description: desc.String("desired_nodes"),
			Validators: []validator.Int32{
				int32validator.Between(0, 100),
			},
			PlanModifiers: []planmodifier.Int32{
				int32planmodifier.RequiresReplace(),
			},
		},
		"name": rschema.StringAttribute{
			Required:    true,
			Description: desc.String("name"),
			Validators: []validator.String{
				stringvalidator.LengthBetween(4, 20),
			},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"schedule": rschema.StringAttribute{
			Optional:    true,
			Description: desc.String("schedule"),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"schedule_type": rschema.StringAttribute{
			Required:    true,
			Description: desc.String("schedule_type"),
			Validators: []validator.String{
				stringvalidator.OneOf("cron", "once"),
			},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"start_time": rschema.StringAttribute{
			Required:    true,
			Description: desc.String("start_time"),
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
			Computed:    true,
			Description: desc.String("status"),
			PlanModifiers: []planmodifier.Object{
				objectplanmodifier.UseStateForUnknown(),
			},
			Attributes: getScheduledScalingStatusResourceSchemaAttributes(),
		},
	}
}

func getScheduledScalingStatusResourceSchemaAttributes() map[string]rschema.Attribute {
	desc := docs.Kubernetesengine("ScalingStatusResponseModel")

	return map[string]rschema.Attribute{
		"histories": rschema.ListNestedAttribute{
			Computed:    true,
			Description: desc.String("histories"),
			PlanModifiers: []planmodifier.List{
				listplanmodifier.UseStateForUnknown(),
			},
			NestedObject: rschema.NestedAttributeObject{
				Attributes: getScheduledScalingHistoryResourceSchemaAttributes(),
			},
		},
	}
}

func getScheduledScalingHistoryResourceSchemaAttributes() map[string]rschema.Attribute {
	desc := docs.Kubernetesengine("ScalingHistoryResponseModel")

	return map[string]rschema.Attribute{
		"description": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("description"),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"occurred_time": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("occurred_time"),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"state": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("state"),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
	}
}

var scheduledScalingDataSourceSchemaAttributes = getScheduledScalingDataSourceSchema()
var scheduledScalingResourceSchema = getScheduledScalingResourceSchema()

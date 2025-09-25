// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package loadbalancer

import (
	"regexp"

	"terraform-provider-kakaocloud/internal/common"
	"terraform-provider-kakaocloud/internal/docs"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var (
	// Health monitor type validator - based on Octavia DTO
	healthMonitorTypeValidator = stringvalidator.OneOf(
		"HTTP", "HTTPS", "TCP", "UDP", "PING",
	)

	// HTTP method validator - based on Octavia DTO
	httpMethodValidator = stringvalidator.OneOf(
		"CONNECT", "GET", "POST", "DELETE", "PATCH", "PUT", "HEAD", "OPTIONS", "TRACE",
	)

	// HTTP version validator - based on SDK
	httpVersionValidator = stringvalidator.OneOf(
		"1.0", "1.1",
	)

	// URL path validator - based on Octavia DTO: ^/(.{0,120})$
	urlPathValidator = stringvalidator.RegexMatches(
		regexp.MustCompile(`^/(.{0,120})$`),
		"Must start with a forward slash (/) and be at most 120 characters long excluding the slash",
	)

	// Expected codes validator - HTTP status codes (supports ranges and comma-separated)
	expectedCodesValidator = stringvalidator.RegexMatches(
		regexp.MustCompile(`^([0-9]{3}(-[0-9]{3})?)(,([0-9]{3}(-[0-9]{3})?))*$`),
		"must be comma-separated HTTP status codes or ranges (e.g., 200,201,202 or 200-399)",
	)

	// Delay validator - based on Octavia DTO: @Max(3600) @Min(0)
	delayValidator = int64validator.Between(0, 3600)

	// Timeout validator - based on Octavia DTO: @Max(900) @Min(0)
	timeoutValidator = int64validator.Between(0, 900)

	// Max retries validator - no specific constraints in Octavia DTO, using reasonable defaults
	maxRetriesValidator = int64validator.Between(1, 10)

	// Max retries down validator - no specific constraints in Octavia DTO, using reasonable defaults
	maxRetriesDownValidator = int64validator.Between(1, 10)
)

func getHealthMonitorDescriptions() map[string]string {
	desc := docs.Loadbalancer("bns_load_balancer__v1__api__create_health_monitor__model__HealthMonitorModel")
	createDesc := docs.Loadbalancer("CreateHealthMonitor")
	targetGroupDesc := docs.Loadbalancer("bns_load_balancer__v1__api__get_target_group__model__TargetGroupModel")

	return map[string]string{
		"id":                     desc.String("id"),
		"name":                   desc.String("name"),
		"type":                   desc.String("type"),
		"delay":                  desc.String("delay"),
		"timeout":                desc.String("timeout"),
		"max_retries":            desc.String("max_retries"),
		"max_retries_down":       desc.String("max_retries_down"),
		"http_method":            desc.String("http_method"),
		"http_version":           desc.String("http_version"),
		"url_path":               desc.String("url_path"),
		"expected_codes":         desc.String("expected_codes"),
		"project_id":             desc.String("project_id"),
		"target_group_id":        createDesc.String("target_group_id"),
		"target_groups":          desc.String("target_groups"),
		"target_group_id_nested": targetGroupDesc.String("id"),
		"provisioning_status":    desc.String("provisioning_status"),
		"operating_status":       desc.String("operating_status"),
		"created_at":             desc.String("created_at"),
		"updated_at":             desc.String("updated_at"),
	}
}

func getHealthMonitorResourceSchema() map[string]rschema.Attribute {
	descriptions := getHealthMonitorDescriptions()

	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			Computed:    true,
			Description: descriptions["id"],
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"name": rschema.StringAttribute{
			Computed:    true,
			Description: descriptions["name"],
		},
		"type": rschema.StringAttribute{
			Required:    true,
			Description: descriptions["type"],
			Validators:  []validator.String{healthMonitorTypeValidator},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"delay": rschema.Int64Attribute{
			Required:    true,
			Description: descriptions["delay"],
			Validators:  []validator.Int64{delayValidator},
		},
		"timeout": rschema.Int64Attribute{
			Required:    true,
			Description: descriptions["timeout"],
			Validators:  []validator.Int64{timeoutValidator},
		},
		"max_retries": rschema.Int64Attribute{
			Required:    true,
			Description: descriptions["max_retries"],
			Validators:  []validator.Int64{maxRetriesValidator},
		},
		"max_retries_down": rschema.Int64Attribute{
			Required:    true,
			Description: descriptions["max_retries_down"],
			Validators:  []validator.Int64{maxRetriesDownValidator},
		},
		"http_method": rschema.StringAttribute{
			Optional:    true,
			Description: descriptions["http_method"],
			Validators:  []validator.String{httpMethodValidator},
		},
		"http_version": rschema.StringAttribute{
			Optional:    true,
			Description: descriptions["http_version"],
			Validators:  []validator.String{httpVersionValidator},
		},
		"url_path": rschema.StringAttribute{
			Optional:    true,
			Description: descriptions["url_path"],
			Validators:  []validator.String{urlPathValidator},
		},
		"expected_codes": rschema.StringAttribute{
			Optional:    true,
			Description: descriptions["expected_codes"],
			Validators:  []validator.String{expectedCodesValidator},
		},
		"project_id": rschema.StringAttribute{
			Computed:    true,
			Description: descriptions["project_id"],
		},
		"target_group_id": rschema.StringAttribute{
			Required:    true,
			Description: descriptions["target_group_id"],
			Validators:  common.UuidValidator(),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"target_groups": rschema.ListNestedAttribute{
			Computed:    true,
			Description: descriptions["target_groups"],
			NestedObject: rschema.NestedAttributeObject{
				Attributes: map[string]rschema.Attribute{
					"id": rschema.StringAttribute{
						Computed:    true,
						Description: descriptions["target_group_id_nested"],
					},
				},
			},
		},
		"provisioning_status": rschema.StringAttribute{
			Computed:    true,
			Description: descriptions["provisioning_status"],
		},
		"operating_status": rschema.StringAttribute{
			Computed:    true,
			Description: descriptions["operating_status"],
		},
		"created_at": rschema.StringAttribute{
			Computed:    true,
			Description: descriptions["created_at"],
		},
		"updated_at": rschema.StringAttribute{
			Computed:    true,
			Description: descriptions["updated_at"],
		},
	}
}

func getHealthMonitorDataSourceSchema() map[string]dschema.Attribute {
	descriptions := getHealthMonitorDescriptions()

	return map[string]dschema.Attribute{
		"id": dschema.StringAttribute{
			Required:    true,
			Description: descriptions["id"],
		},
		"name": dschema.StringAttribute{
			Computed:    true,
			Description: descriptions["name"],
		},
		"type": dschema.StringAttribute{
			Computed:    true,
			Description: descriptions["type"],
		},
		"delay": dschema.Int64Attribute{
			Computed:    true,
			Description: descriptions["delay"],
		},
		"timeout": dschema.Int64Attribute{
			Computed:    true,
			Description: descriptions["timeout"],
		},
		"max_retries": dschema.Int64Attribute{
			Computed:    true,
			Description: descriptions["max_retries"],
		},
		"max_retries_down": dschema.Int64Attribute{
			Computed:    true,
			Description: descriptions["max_retries_down"],
		},
		"http_method": dschema.StringAttribute{
			Computed:    true,
			Description: descriptions["http_method"],
		},
		"http_version": dschema.StringAttribute{
			Computed:    true,
			Description: descriptions["http_version"],
		},
		"url_path": dschema.StringAttribute{
			Computed:    true,
			Description: descriptions["url_path"],
		},
		"expected_codes": dschema.StringAttribute{
			Computed:    true,
			Description: descriptions["expected_codes"],
		},
		"project_id": dschema.StringAttribute{
			Computed:    true,
			Description: descriptions["project_id"],
		},
		"target_group_id": dschema.StringAttribute{
			Computed:    true,
			Description: descriptions["target_group_id"],
		},
		"target_groups": dschema.ListNestedAttribute{
			Computed:    true,
			Description: descriptions["target_groups"],
			NestedObject: dschema.NestedAttributeObject{
				Attributes: map[string]dschema.Attribute{
					"id": dschema.StringAttribute{
						Computed:    true,
						Description: descriptions["target_group_id_nested"],
					},
				},
			},
		},
		"provisioning_status": dschema.StringAttribute{
			Computed:    true,
			Description: descriptions["provisioning_status"],
		},
		"operating_status": dschema.StringAttribute{
			Computed:    true,
			Description: descriptions["operating_status"],
		},
		"created_at": dschema.StringAttribute{
			Computed:    true,
			Description: descriptions["created_at"],
		},
		"updated_at": dschema.StringAttribute{
			Computed:    true,
			Description: descriptions["updated_at"],
		},
	}
}

// Resource schema attributes
var loadBalancerHealthMonitorResourceSchema = getHealthMonitorResourceSchema()

// Data source schema with computed attributes
var loadBalancerHealthMonitorDataSourceSchema = getHealthMonitorDataSourceSchema()

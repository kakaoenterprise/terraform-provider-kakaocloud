// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package loadbalancer

import (
	"regexp"

	"terraform-provider-kakaocloud/internal/common"

	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/kakaoenterprise/kc-sdk-go/services/loadbalancer"
)

var (
	healthMonitorTypeValidator = stringvalidator.OneOf(
		string(loadbalancer.HEALTHMONITORTYPE_HTTP),
		string(loadbalancer.HEALTHMONITORTYPE_HTTPS),
		string(loadbalancer.HEALTHMONITORTYPE_TCP),
		string(loadbalancer.HEALTHMONITORTYPE_PING),
	)

	httpMethodValidator = stringvalidator.OneOf(
		string(loadbalancer.HEALTHMONITORMETHOD_CONNECT),
		string(loadbalancer.HEALTHMONITORMETHOD_GET),
		string(loadbalancer.HEALTHMONITORMETHOD_POST),
		string(loadbalancer.HEALTHMONITORMETHOD_DELETE),
		string(loadbalancer.HEALTHMONITORMETHOD_PATCH),
		string(loadbalancer.HEALTHMONITORMETHOD_PUT),
		string(loadbalancer.HEALTHMONITORMETHOD_HEAD),
		string(loadbalancer.HEALTHMONITORMETHOD_OPTIONS),
		string(loadbalancer.HEALTHMONITORMETHOD_TRACE),
	)

	httpVersionValidator = stringvalidator.OneOf(
		string(loadbalancer.HEALTHMONITORHTTPVERSION__1_0),
		string(loadbalancer.HEALTHMONITORHTTPVERSION__1_1),
	)

	urlPathValidator = stringvalidator.RegexMatches(
		regexp.MustCompile(`^/(.{0,120})$`),
		"Must start with a forward slash (/) and be at most 120 characters long excluding the slash",
	)

	expectedCodesValidator = stringvalidator.RegexMatches(
		regexp.MustCompile(`^([0-9]{3}(-[0-9]{3})?)(,([0-9]{3}(-[0-9]{3})?))*$`),
		"must be comma-separated HTTP status codes or ranges (e.g., 200,201,202 or 200-399)",
	)

	delayValidator = int32validator.Between(0, 3600)

	timeoutValidator = int32validator.Between(0, 900)

	maxRetriesValidator = int32validator.Between(1, 10)

	maxRetriesDownValidator = int32validator.Between(1, 10)
)

func getHealthMonitorResourceSchema() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"name": rschema.StringAttribute{
			Computed: true,
		},
		"type": rschema.StringAttribute{
			Required:   true,
			Validators: []validator.String{healthMonitorTypeValidator},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"delay": rschema.Int32Attribute{
			Required:   true,
			Validators: []validator.Int32{delayValidator},
		},
		"timeout": rschema.Int32Attribute{
			Required:   true,
			Validators: []validator.Int32{timeoutValidator},
		},
		"max_retries": rschema.Int32Attribute{
			Required:   true,
			Validators: []validator.Int32{maxRetriesValidator},
		},
		"max_retries_down": rschema.Int32Attribute{
			Required:   true,
			Validators: []validator.Int32{maxRetriesDownValidator},
		},
		"http_method": rschema.StringAttribute{
			Optional:   true,
			Validators: []validator.String{httpMethodValidator},
		},
		"http_version": rschema.StringAttribute{
			Optional:   true,
			Validators: []validator.String{httpVersionValidator},
		},
		"url_path": rschema.StringAttribute{
			Optional:   true,
			Validators: []validator.String{urlPathValidator},
		},
		"expected_codes": rschema.StringAttribute{
			Optional:   true,
			Validators: []validator.String{expectedCodesValidator},
		},
		"project_id": rschema.StringAttribute{
			Computed: true,
		},
		"target_group_id": rschema.StringAttribute{
			Required:   true,
			Validators: common.UuidValidator(),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"target_groups": rschema.ListNestedAttribute{
			Computed: true,
			NestedObject: rschema.NestedAttributeObject{
				Attributes: map[string]rschema.Attribute{
					"id": rschema.StringAttribute{
						Computed: true,
					},
				},
			},
		},
		"provisioning_status": rschema.StringAttribute{
			Computed: true,
		},
		"operating_status": rschema.StringAttribute{
			Computed: true,
		},
		"created_at": rschema.StringAttribute{
			Computed: true,
		},
		"updated_at": rschema.StringAttribute{
			Computed: true,
		},
	}
}

func getHealthMonitorDataSourceSchema() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"id": dschema.StringAttribute{
			Required: true,
		},
		"name": dschema.StringAttribute{
			Computed: true,
		},
		"type": dschema.StringAttribute{
			Computed: true,
		},
		"delay": dschema.Int32Attribute{
			Computed: true,
		},
		"timeout": dschema.Int32Attribute{
			Computed: true,
		},
		"max_retries": dschema.Int32Attribute{
			Computed: true,
		},
		"max_retries_down": dschema.Int32Attribute{
			Computed: true,
		},
		"http_method": dschema.StringAttribute{
			Computed: true,
		},
		"http_version": dschema.StringAttribute{
			Computed: true,
		},
		"url_path": dschema.StringAttribute{
			Computed: true,
		},
		"expected_codes": dschema.StringAttribute{
			Computed: true,
		},
		"project_id": dschema.StringAttribute{
			Computed: true,
		},
		"target_group_id": dschema.StringAttribute{
			Computed: true,
		},
		"target_groups": dschema.ListNestedAttribute{
			Computed: true,
			NestedObject: dschema.NestedAttributeObject{
				Attributes: map[string]dschema.Attribute{
					"id": dschema.StringAttribute{
						Computed: true,
					},
				},
			},
		},
		"provisioning_status": dschema.StringAttribute{
			Computed: true,
		},
		"operating_status": dschema.StringAttribute{
			Computed: true,
		},
		"created_at": dschema.StringAttribute{
			Computed: true,
		},
		"updated_at": dschema.StringAttribute{
			Computed: true,
		},
	}
}

var loadBalancerHealthMonitorResourceSchema = getHealthMonitorResourceSchema()

var loadBalancerHealthMonitorDataSourceSchema = getHealthMonitorDataSourceSchema()

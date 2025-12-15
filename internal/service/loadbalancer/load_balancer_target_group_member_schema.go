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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	ipv4Validator = stringvalidator.RegexMatches(
		regexp.MustCompile(`^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`),
		"must be a valid IPv4 address",
	)
	weightValidator = int32validator.Between(0, 256)
)

func getBaseMemberAttributes() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			Computed: true,
		},
		"name": rschema.StringAttribute{
			Optional:   true,
			Computed:   true,
			Validators: common.NameValidator(255),
		},
		"address": rschema.StringAttribute{
			Required:   true,
			Validators: []validator.String{ipv4Validator},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"protocol_port": rschema.Int32Attribute{
			Required:   true,
			Validators: common.PortValidator(),
			PlanModifiers: []planmodifier.Int32{
				int32planmodifier.RequiresReplace(),
			},
		},
		"subnet_id": rschema.StringAttribute{
			Required:   true,
			Validators: common.UuidValidator(),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"weight": rschema.Int32Attribute{
			Optional:   true,
			Computed:   true,
			Validators: []validator.Int32{weightValidator},
		},
		"monitor_port": rschema.Int32Attribute{
			Optional:   true,
			Computed:   true,
			Validators: common.PortValidator(),
		},
		"operating_status": rschema.StringAttribute{
			Computed: true,
		},
		"provisioning_status": rschema.StringAttribute{
			Computed: true,
		},
		"is_backup": rschema.BoolAttribute{
			Computed: true,
		},
		"project_id": rschema.StringAttribute{
			Computed: true,
		},
		"created_at": rschema.StringAttribute{
			Computed: true,
		},
		"updated_at": rschema.StringAttribute{
			Computed: true,
		},
		"network_interface_id": rschema.StringAttribute{
			Computed: true,
		},
		"instance_id": rschema.StringAttribute{
			Computed: true,
		},
		"instance_name": rschema.StringAttribute{
			Computed: true,
		},
		"vpc_id": rschema.StringAttribute{
			Computed: true,
		},
		"subnet": rschema.SingleNestedAttribute{
			Computed: true,
			Attributes: map[string]rschema.Attribute{
				"id": rschema.StringAttribute{
					Computed: true,
				},
				"name": rschema.StringAttribute{
					Computed: true,
				},
				"cidr_block": rschema.StringAttribute{
					Computed: true,
				},
				"availability_zone": rschema.StringAttribute{
					Computed: true,
				},
				"health_check_ips": rschema.ListAttribute{
					Computed:    true,
					ElementType: types.StringType,
				},
			},
		},
		"security_groups": rschema.ListNestedAttribute{
			Computed: true,
			NestedObject: rschema.NestedAttributeObject{
				Attributes: map[string]rschema.Attribute{
					"id": rschema.StringAttribute{
						Computed: true,
					},
					"name": rschema.StringAttribute{
						Computed: true,
					},
				},
			},
		},
	}
}

func getLoadBalancerTargetGroupMemberResourceSchema() map[string]rschema.Attribute {
	baseAttrs := getBaseMemberAttributes()

	return map[string]rschema.Attribute{
		"id": baseAttrs["id"],
		"target_group_id": rschema.StringAttribute{
			Required:   true,
			Validators: common.UuidValidator(),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"name":                 baseAttrs["name"],
		"address":              baseAttrs["address"],
		"protocol_port":        baseAttrs["protocol_port"],
		"subnet_id":            baseAttrs["subnet_id"],
		"weight":               baseAttrs["weight"],
		"monitor_port":         baseAttrs["monitor_port"],
		"operating_status":     baseAttrs["operating_status"],
		"provisioning_status":  baseAttrs["provisioning_status"],
		"is_backup":            baseAttrs["is_backup"],
		"project_id":           baseAttrs["project_id"],
		"created_at":           baseAttrs["created_at"],
		"updated_at":           baseAttrs["updated_at"],
		"network_interface_id": baseAttrs["network_interface_id"],
		"instance_id":          baseAttrs["instance_id"],
		"instance_name":        baseAttrs["instance_name"],
		"vpc_id":               baseAttrs["vpc_id"],
		"subnet":               baseAttrs["subnet"],
		"security_groups":      baseAttrs["security_groups"],
	}
}

func getLoadBalancerTargetGroupMemberDataSourceSchema() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"id": dschema.StringAttribute{
			Required:   true,
			Validators: common.UuidValidator(),
		},
		"target_group_id": dschema.StringAttribute{
			Required:   true,
			Validators: common.UuidValidator(),
		},
		"name": dschema.StringAttribute{
			Computed: true,
		},
		"address": dschema.StringAttribute{
			Computed: true,
		},
		"protocol_port": dschema.Int32Attribute{
			Computed: true,
		},
		"subnet_id": dschema.StringAttribute{
			Computed: true,
		},
		"weight": dschema.Int32Attribute{
			Computed: true,
		},
		"monitor_port": dschema.Int32Attribute{
			Computed: true,
		},
		"operating_status": dschema.StringAttribute{
			Computed: true,
		},
		"provisioning_status": dschema.StringAttribute{
			Computed: true,
		},
		"is_backup": dschema.BoolAttribute{
			Computed: true,
		},
		"project_id": dschema.StringAttribute{
			Computed: true,
		},
		"created_at": dschema.StringAttribute{
			Computed: true,
		},
		"updated_at": dschema.StringAttribute{
			Computed: true,
		},
		"network_interface_id": dschema.StringAttribute{
			Computed: true,
		},
		"instance_id": dschema.StringAttribute{
			Computed: true,
		},
		"instance_name": dschema.StringAttribute{
			Computed: true,
		},
		"vpc_id": dschema.StringAttribute{
			Computed: true,
		},
		"subnet": dschema.SingleNestedAttribute{
			Computed: true,
			Attributes: map[string]dschema.Attribute{
				"id": dschema.StringAttribute{
					Computed: true,
				},
				"name": dschema.StringAttribute{
					Computed: true,
				},
				"cidr_block": dschema.StringAttribute{
					Computed: true,
				},
				"availability_zone": dschema.StringAttribute{
					Computed: true,
				},
				"health_check_ips": dschema.ListAttribute{
					Computed:    true,
					ElementType: types.StringType,
				},
			},
		},
		"security_groups": dschema.ListNestedAttribute{
			Computed: true,
			NestedObject: dschema.NestedAttributeObject{
				Attributes: map[string]dschema.Attribute{
					"id": dschema.StringAttribute{
						Computed: true,
					},
					"name": dschema.StringAttribute{
						Computed: true,
					},
				},
			},
		},
	}
}

func getLoadBalancerTargetGroupMemberListDataSourceSchema() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"target_group_id": dschema.StringAttribute{
			Required:   true,
			Validators: common.UuidValidator(),
		},
		"members": dschema.ListNestedAttribute{
			Computed: true,
			NestedObject: dschema.NestedAttributeObject{
				Attributes: map[string]dschema.Attribute{
					"id": dschema.StringAttribute{
						Computed: true,
					},
					"target_group_id": dschema.StringAttribute{
						Computed: true,
					},
					"name": dschema.StringAttribute{
						Computed: true,
					},
					"address": dschema.StringAttribute{
						Computed: true,
					},
					"protocol_port": dschema.Int32Attribute{
						Computed: true,
					},
					"subnet_id": dschema.StringAttribute{
						Computed: true,
					},
					"weight": dschema.Int32Attribute{
						Computed: true,
					},
					"monitor_port": dschema.Int32Attribute{
						Computed: true,
					},
					"operating_status": dschema.StringAttribute{
						Computed: true,
					},
					"provisioning_status": dschema.StringAttribute{
						Computed: true,
					},
					"is_backup": dschema.BoolAttribute{
						Computed: true,
					},
					"project_id": dschema.StringAttribute{
						Computed: true,
					},
					"created_at": dschema.StringAttribute{
						Computed: true,
					},
					"updated_at": dschema.StringAttribute{
						Computed: true,
					},
					"network_interface_id": dschema.StringAttribute{
						Computed: true,
					},
					"instance_id": dschema.StringAttribute{
						Computed: true,
					},
					"instance_name": dschema.StringAttribute{
						Computed: true,
					},
					"vpc_id": dschema.StringAttribute{
						Computed: true,
					},
					"subnet": dschema.SingleNestedAttribute{
						Computed: true,
						Attributes: map[string]dschema.Attribute{
							"id": dschema.StringAttribute{
								Computed: true,
							},
							"name": dschema.StringAttribute{
								Computed: true,
							},
							"cidr_block": dschema.StringAttribute{
								Computed: true,
							},
							"availability_zone": dschema.StringAttribute{
								Computed: true,
							},
							"health_check_ips": dschema.ListAttribute{
								Computed:    true,
								ElementType: types.StringType,
							},
						},
					},
					"security_groups": dschema.ListNestedAttribute{
						Computed: true,
						NestedObject: dschema.NestedAttributeObject{
							Attributes: map[string]dschema.Attribute{
								"id": dschema.StringAttribute{
									Computed: true,
								},
								"name": dschema.StringAttribute{
									Computed: true,
								},
							},
						},
					},
				},
			},
		},
	}
}

var loadBalancerTargetGroupMemberResourceSchema = getLoadBalancerTargetGroupMemberResourceSchema()

var loadBalancerTargetGroupMemberDataSourceSchema = getLoadBalancerTargetGroupMemberDataSourceSchema()

var loadBalancerTargetGroupMemberListDataSourceSchema = getLoadBalancerTargetGroupMemberListDataSourceSchema()

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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
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
	weightValidator = int64validator.Between(0, 256)
)

func getBaseMemberAttributes() map[string]rschema.Attribute {
	desc := docs.Loadbalancer("bns_load_balancer__v1__api__add_target__model__TargetGroupMemberModel")
	listDesc := docs.Loadbalancer("bns_load_balancer__v1__api__list_targets_in_target_group__model__TargetGroupMemberModel")

	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("id"),
		},
		"name": rschema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: desc.String("name"),
			Validators:  common.NameValidator(255),
		},
		"address": rschema.StringAttribute{
			Required:    true,
			Description: desc.String("address"),
			Validators:  []validator.String{ipv4Validator},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"protocol_port": rschema.Int64Attribute{
			Required:    true,
			Description: desc.String("protocol_port"),
			Validators:  common.PortValidator(),
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.RequiresReplace(),
			},
		},
		"subnet_id": rschema.StringAttribute{
			Required:    true,
			Description: desc.String("subnet_id"),
			Validators:  common.UuidValidator(),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"weight": rschema.Int64Attribute{
			Optional:    true,
			Computed:    true,
			Description: desc.String("weight"),
			Validators:  []validator.Int64{weightValidator},
		},
		"monitor_port": rschema.Int64Attribute{
			Optional:    true,
			Computed:    true,
			Description: desc.String("monitor_port"),
			Validators:  common.PortValidator(),
		},
		"operating_status": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("operating_status"),
		},
		"provisioning_status": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("provisioning_status"),
		},
		"is_backup": rschema.BoolAttribute{
			Computed:    true,
			Description: desc.String("is_backup"),
		},
		"project_id": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("project_id"),
		},
		"created_at": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("created_at"),
		},
		"updated_at": rschema.StringAttribute{
			Computed:    true,
			Description: desc.String("updated_at"),
		},
		"network_interface_id": rschema.StringAttribute{
			Computed:    true,
			Description: listDesc.String("network_interface_id"),
		},
		"instance_id": rschema.StringAttribute{
			Computed:    true,
			Description: listDesc.String("instance_id"),
		},
		"instance_name": rschema.StringAttribute{
			Computed:    true,
			Description: listDesc.String("instance_name"),
		},
		"vpc_id": rschema.StringAttribute{
			Computed:    true,
			Description: listDesc.String("vpc_id"),
		},
		"subnet": rschema.SingleNestedAttribute{
			Computed:    true,
			Description: listDesc.String("subnet"),
			Attributes: map[string]rschema.Attribute{
				"id": rschema.StringAttribute{
					Computed:    true,
					Description: docs.Loadbalancer("HealthCheckSubnetModel").String("id"),
				},
				"name": rschema.StringAttribute{
					Computed:    true,
					Description: docs.Loadbalancer("HealthCheckSubnetModel").String("name"),
				},
				"cidr_block": rschema.StringAttribute{
					Computed:    true,
					Description: docs.Loadbalancer("HealthCheckSubnetModel").String("cidr_block"),
				},
				"availability_zone": rschema.StringAttribute{
					Computed:    true,
					Description: docs.Loadbalancer("HealthCheckSubnetModel").String("availability_zone"),
				},
				"health_check_ips": rschema.ListAttribute{
					Computed:    true,
					Description: docs.Loadbalancer("HealthCheckSubnetModel").String("health_check_ips"),
					ElementType: types.StringType,
				},
			},
		},
		"security_groups": rschema.ListNestedAttribute{
			Computed:    true,
			Description: listDesc.String("security_groups"),
			NestedObject: rschema.NestedAttributeObject{
				Attributes: map[string]rschema.Attribute{
					"id": rschema.StringAttribute{
						Computed:    true,
						Description: docs.Loadbalancer("SecurityGroupModel").String("id"),
					},
					"name": rschema.StringAttribute{
						Computed:    true,
						Description: docs.Loadbalancer("SecurityGroupModel").String("name"),
					},
				},
			},
		},
	}
}

func getLoadBalancerTargetGroupMemberResourceSchema() map[string]rschema.Attribute {
	baseAttrs := getBaseMemberAttributes()
	desc := docs.Loadbalancer("bns_load_balancer__v1__api__list_targets_in_target_group__model__TargetGroupMemberModel")

	return map[string]rschema.Attribute{
		"id": baseAttrs["id"],
		"target_group_id": rschema.StringAttribute{
			Required:    true,
			Description: desc.String("target_group_id"),
			Validators:  common.UuidValidator(),
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
	desc := docs.Loadbalancer("bns_load_balancer__v1__api__list_targets_in_target_group__model__TargetGroupMemberModel")

	return map[string]dschema.Attribute{
		"id": dschema.StringAttribute{
			Required:    true,
			Description: desc.String("id"),
			Validators:  common.UuidValidator(),
		},
		"target_group_id": dschema.StringAttribute{
			Required:    true,
			Description: desc.String("target_group_id"),
			Validators:  common.UuidValidator(),
		},
		"name": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("name"),
		},
		"address": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("ip_address"),
		},
		"protocol_port": dschema.Int64Attribute{
			Computed:    true,
			Description: desc.String("protocol_port"),
		},
		"subnet_id": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("subnet_id"),
		},
		"weight": dschema.Int64Attribute{
			Computed:    true,
			Description: desc.String("weight"),
		},
		"monitor_port": dschema.Int64Attribute{
			Computed:    true,
			Description: desc.String("monitor_port"),
		},
		"operating_status": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("operating_status"),
		},
		"provisioning_status": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("provisioning_status"),
		},
		"is_backup": dschema.BoolAttribute{
			Computed:    true,
			Description: desc.String("is_backup"),
		},
		"project_id": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("project_id"),
		},
		"created_at": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("created_at"),
		},
		"updated_at": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("updated_at"),
		},
		"network_interface_id": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("network_interface_id"),
		},
		"instance_id": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("instance_id"),
		},
		"instance_name": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("instance_name"),
		},
		"vpc_id": dschema.StringAttribute{
			Computed:    true,
			Description: desc.String("vpc_id"),
		},
		"subnet": dschema.SingleNestedAttribute{
			Computed:    true,
			Description: desc.String("subnet"),
			Attributes: map[string]dschema.Attribute{
				"id": dschema.StringAttribute{
					Computed:    true,
					Description: docs.Loadbalancer("HealthCheckSubnetModel").String("id"),
				},
				"name": dschema.StringAttribute{
					Computed:    true,
					Description: docs.Loadbalancer("HealthCheckSubnetModel").String("name"),
				},
				"cidr_block": dschema.StringAttribute{
					Computed:    true,
					Description: docs.Loadbalancer("HealthCheckSubnetModel").String("cidr_block"),
				},
				"availability_zone": dschema.StringAttribute{
					Computed:    true,
					Description: docs.Loadbalancer("HealthCheckSubnetModel").String("availability_zone"),
				},
				"health_check_ips": dschema.ListAttribute{
					Computed:    true,
					Description: docs.Loadbalancer("HealthCheckSubnetModel").String("health_check_ips"),
					ElementType: types.StringType,
				},
			},
		},
		"security_groups": dschema.ListNestedAttribute{
			Computed:    true,
			Description: desc.String("security_groups"),
			NestedObject: dschema.NestedAttributeObject{
				Attributes: map[string]dschema.Attribute{
					"id": dschema.StringAttribute{
						Computed:    true,
						Description: docs.Loadbalancer("SecurityGroupModel").String("id"),
					},
					"name": dschema.StringAttribute{
						Computed:    true,
						Description: docs.Loadbalancer("SecurityGroupModel").String("name"),
					},
				},
			},
		},
	}
}

func getLoadBalancerTargetGroupMemberListDataSourceSchema() map[string]dschema.Attribute {
	desc := docs.Loadbalancer("bns_load_balancer__v1__api__list_targets_in_target_group__model__TargetGroupMemberModel")

	return map[string]dschema.Attribute{
		"target_group_id": dschema.StringAttribute{
			Required:    true,
			Description: desc.String("target_group_id"),
			Validators:  common.UuidValidator(),
		},
		"members": dschema.ListNestedAttribute{
			Computed:    true,
			Description: docs.Loadbalancer("bns_load_balancer__v1__api__create_target_group__model__TargetGroupModel").String("members"),
			NestedObject: dschema.NestedAttributeObject{
				Attributes: map[string]dschema.Attribute{
					"id": dschema.StringAttribute{
						Computed:    true,
						Description: desc.String("id"),
					},
					"target_group_id": dschema.StringAttribute{
						Computed:    true,
						Description: desc.String("target_group_id"),
					},
					"name": dschema.StringAttribute{
						Computed:    true,
						Description: desc.String("name"),
					},
					"address": dschema.StringAttribute{
						Computed:    true,
						Description: desc.String("ip_address"),
					},
					"protocol_port": dschema.Int64Attribute{
						Computed:    true,
						Description: desc.String("protocol_port"),
					},
					"subnet_id": dschema.StringAttribute{
						Computed:    true,
						Description: desc.String("subnet_id"),
					},
					"weight": dschema.Int64Attribute{
						Computed:    true,
						Description: desc.String("weight"),
					},
					"monitor_port": dschema.Int64Attribute{
						Computed:    true,
						Description: desc.String("monitor_port"),
					},
					"operating_status": dschema.StringAttribute{
						Computed:    true,
						Description: desc.String("operating_status"),
					},
					"provisioning_status": dschema.StringAttribute{
						Computed:    true,
						Description: desc.String("provisioning_status"),
					},
					"is_backup": dschema.BoolAttribute{
						Computed:    true,
						Description: desc.String("is_backup"),
					},
					"project_id": dschema.StringAttribute{
						Computed:    true,
						Description: desc.String("project_id"),
					},
					"created_at": dschema.StringAttribute{
						Computed:    true,
						Description: desc.String("created_at"),
					},
					"updated_at": dschema.StringAttribute{
						Computed:    true,
						Description: desc.String("updated_at"),
					},
					"network_interface_id": dschema.StringAttribute{
						Computed:    true,
						Description: desc.String("network_interface_id"),
					},
					"instance_id": dschema.StringAttribute{
						Computed:    true,
						Description: desc.String("instance_id"),
					},
					"instance_name": dschema.StringAttribute{
						Computed:    true,
						Description: desc.String("instance_name"),
					},
					"vpc_id": dschema.StringAttribute{
						Computed:    true,
						Description: desc.String("vpc_id"),
					},
					"subnet": dschema.SingleNestedAttribute{
						Computed:    true,
						Description: desc.String("subnet"),
						Attributes: map[string]dschema.Attribute{
							"id": dschema.StringAttribute{
								Computed:    true,
								Description: docs.Loadbalancer("HealthCheckSubnetModel").String("id"),
							},
							"name": dschema.StringAttribute{
								Computed:    true,
								Description: docs.Loadbalancer("HealthCheckSubnetModel").String("name"),
							},
							"cidr_block": dschema.StringAttribute{
								Computed:    true,
								Description: docs.Loadbalancer("HealthCheckSubnetModel").String("cidr_block"),
							},
							"availability_zone": dschema.StringAttribute{
								Computed:    true,
								Description: docs.Loadbalancer("HealthCheckSubnetModel").String("availability_zone"),
							},
							"health_check_ips": dschema.ListAttribute{
								Computed:    true,
								Description: docs.Loadbalancer("HealthCheckSubnetModel").String("health_check_ips"),
								ElementType: types.StringType,
							},
						},
					},
					"security_groups": dschema.ListNestedAttribute{
						Computed:    true,
						Description: desc.String("security_groups"),
						NestedObject: dschema.NestedAttributeObject{
							Attributes: map[string]dschema.Attribute{
								"id": dschema.StringAttribute{
									Computed:    true,
									Description: docs.Loadbalancer("SecurityGroupModel").String("id"),
								},
								"name": dschema.StringAttribute{
									Computed:    true,
									Description: docs.Loadbalancer("SecurityGroupModel").String("name"),
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

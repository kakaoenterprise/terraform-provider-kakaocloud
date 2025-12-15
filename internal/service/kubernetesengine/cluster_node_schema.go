// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package kubernetesengine

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func getNodeDataSourceSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"is_cordon": schema.BoolAttribute{
			Computed: true,
		},
		"created_at": schema.StringAttribute{
			Computed: true,
		},
		"flavor": schema.StringAttribute{
			Computed: true,
		},
		"id": schema.StringAttribute{
			Computed: true,
		},
		"ip": schema.StringAttribute{
			Computed: true,
		},
		"name": schema.StringAttribute{
			Computed: true,
		},
		"node_pool_name": schema.StringAttribute{
			Computed: true,
		},
		"ssh_key_name": schema.StringAttribute{
			Computed: true,
		},
		"failure_message": schema.StringAttribute{
			Computed: true,
		},
		"updated_at": schema.StringAttribute{
			Computed: true,
		},
		"version": schema.StringAttribute{
			Computed: true,
		},
		"volume_size": schema.Int32Attribute{
			Computed: true,
		},
		"is_hyper_threading": schema.BoolAttribute{
			Computed: true,
		},

		"image": schema.SingleNestedAttribute{
			Computed:   true,
			Attributes: getNodeImageSchemaAttributes(),
		},

		"status": schema.SingleNestedAttribute{
			Computed:   true,
			Attributes: getNodeStatusSchemaAttributes(),
		},

		"vpc_info": schema.SingleNestedAttribute{
			Computed:   true,
			Attributes: getNodeVpcInfoSchemaAttributes(),
		},
	}
}

func getNodeImageSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"architecture": schema.StringAttribute{
			Computed: true,
		},
		"is_gpu_type": schema.BoolAttribute{
			Computed: true,
		},
		"id": schema.StringAttribute{
			Computed: true,
		},
		"instance_type": schema.StringAttribute{
			Computed: true,
		},
		"kernel_version": schema.StringAttribute{
			Computed: true,
		},
		"key_package": schema.StringAttribute{
			Computed: true,
		},
		"name": schema.StringAttribute{
			Computed: true,
		},
		"os_distro": schema.StringAttribute{
			Computed: true,
		},
		"os_type": schema.StringAttribute{
			Computed: true,
		},
		"os_version": schema.StringAttribute{
			Computed: true,
		},
	}
}

func getNodeStatusSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"phase": schema.StringAttribute{
			Computed: true,
		},
	}
}

func getNodeVpcInfoSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed: true,
		},
		"subnets": schema.SetNestedAttribute{
			Computed: true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: getNodeSubnetSchemaAttributes(),
			},
		},
	}
}

func getNodeSubnetSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"availability_zone": schema.StringAttribute{
			Computed: true,
		},
		"cidr_block": schema.StringAttribute{
			Computed: true,
		},
		"id": schema.StringAttribute{
			Computed: true,
		},
	}
}

func getNodeResourceSchema() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"cluster_name": rschema.StringAttribute{
			Required: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"node_names": rschema.SetAttribute{
			ElementType: types.StringType,
			Required:    true,
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
			},
			PlanModifiers: []planmodifier.Set{
				setplanmodifier.RequiresReplace(),
			},
		},
		"is_remove": rschema.BoolAttribute{
			Optional:      true,
			PlanModifiers: []planmodifier.Bool{boolplanmodifier.RequiresReplace()},
		},
		"is_cordon": rschema.BoolAttribute{
			Optional:      true,
			PlanModifiers: []planmodifier.Bool{boolplanmodifier.RequiresReplace()},
		},
		"id": rschema.StringAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
	}
}

var nodeDataSourceSchemaAttributes = getNodeDataSourceSchema()
var nodeResourceSchemaAttributes = getNodeResourceSchema()

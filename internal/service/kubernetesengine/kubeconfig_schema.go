// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package kubernetesengine

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var kubernetesKubeconfigDataSourceSchemaAttributes = map[string]schema.Attribute{
	"cluster_name": schema.StringAttribute{
		Required: true,
	},

	"kubeconfig_yaml": schema.StringAttribute{
		Computed: true,
	},

	"api_version":     schema.StringAttribute{Computed: true},
	"kind":            schema.StringAttribute{Computed: true},
	"current_context": schema.StringAttribute{Computed: true},

	"preferences": schema.MapAttribute{
		ElementType: types.StringType,
		Computed:    true,
	},

	"clusters": schema.ListNestedAttribute{
		Computed: true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{Computed: true},
				"cluster": schema.SingleNestedAttribute{
					Computed: true,
					Attributes: map[string]schema.Attribute{
						"server":                     schema.StringAttribute{Computed: true},
						"certificate_authority_data": schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	},

	"contexts": schema.ListNestedAttribute{
		Computed: true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{Computed: true},
				"context": schema.SingleNestedAttribute{
					Computed: true,
					Attributes: map[string]schema.Attribute{
						"cluster": schema.StringAttribute{Computed: true},
						"user":    schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	},

	"users": schema.ListNestedAttribute{
		Computed: true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{Computed: true},
				"user": schema.SingleNestedAttribute{
					Computed: true,
					Attributes: map[string]schema.Attribute{
						"exec": schema.SingleNestedAttribute{
							Computed: true,
							Attributes: map[string]schema.Attribute{
								"api_version": schema.StringAttribute{Computed: true},
								"command":     schema.StringAttribute{Computed: true},
								"args": schema.ListAttribute{
									ElementType: types.StringType,
									Computed:    true,
								},
								"env": schema.ListNestedAttribute{
									Computed: true,
									NestedObject: schema.NestedAttributeObject{
										Attributes: map[string]schema.Attribute{
											"name":  schema.StringAttribute{Computed: true},
											"value": schema.StringAttribute{Computed: true},
										},
									},
								},
								"provide_cluster_info": schema.BoolAttribute{Computed: true},
							},
						},
					},
				},
			},
		},
	},
}

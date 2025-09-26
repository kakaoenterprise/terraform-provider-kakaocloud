// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package kubernetesengine

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type kubernetesKubeconfigDataSourceModel struct {
	ClusterName    types.String `tfsdk:"cluster_name"`
	KubeconfigYAML types.String `tfsdk:"kubeconfig_yaml"`

	ApiVersion     types.String `tfsdk:"api_version"`
	Kind           types.String `tfsdk:"kind"`
	CurrentContext types.String `tfsdk:"current_context"`
	Preferences    types.Map    `tfsdk:"preferences"`

	Clusters []kcfgClusterEntry `tfsdk:"clusters"`
	Contexts []kcfgContextEntry `tfsdk:"contexts"`
	Users    []kcfgUserEntry    `tfsdk:"users"`

	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

type kcfgClusterEntry struct {
	Name    types.String `tfsdk:"name"`
	Cluster kcfgCluster  `tfsdk:"cluster"`
}

type kcfgCluster struct {
	Server                   types.String `tfsdk:"server"`
	CertificateAuthorityData types.String `tfsdk:"certificate_authority_data"`
}

type kcfgContextEntry struct {
	Name    types.String `tfsdk:"name"`
	Context kcfgContext  `tfsdk:"context"`
}

type kcfgContext struct {
	Cluster types.String `tfsdk:"cluster"`
	User    types.String `tfsdk:"user"`
}

type kcfgUserEntry struct {
	Name types.String `tfsdk:"name"`
	User kcfgUser     `tfsdk:"user"`
}

type kcfgUser struct {
	Exec *kcfgExec `tfsdk:"exec"`
}

type kcfgExec struct {
	ApiVersion         types.String `tfsdk:"api_version"`
	Command            types.String `tfsdk:"command"`
	Args               types.List   `tfsdk:"args"`
	Env                types.List   `tfsdk:"env"`
	ProvideClusterInfo types.Bool   `tfsdk:"provide_cluster_info"`
}

type kcfgExecEnv struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

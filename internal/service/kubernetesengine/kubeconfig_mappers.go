// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package kubernetesengine

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gopkg.in/yaml.v3"
)

type rawKubeconfig struct {
	APIVersion     string            `yaml:"apiVersion"`
	Kind           string            `yaml:"kind"`
	CurrentContext string            `yaml:"current-context"`
	Preferences    map[string]string `yaml:"preferences"`
	Clusters       []rawClusterEntry `yaml:"clusters"`
	Contexts       []rawContextEntry `yaml:"contexts"`
	Users          []rawUserEntry    `yaml:"users"`
}

type rawClusterEntry struct {
	Name    string     `yaml:"name"`
	Cluster rawCluster `yaml:"cluster"`
}
type rawCluster struct {
	Server                   string `yaml:"server"`
	CertificateAuthorityData string `yaml:"certificate-authority-data"`
}

type rawContextEntry struct {
	Name    string     `yaml:"name"`
	Context rawContext `yaml:"context"`
}
type rawContext struct {
	Cluster string `yaml:"cluster"`
	User    string `yaml:"user"`
}

type rawUserEntry struct {
	Name string  `yaml:"name"`
	User rawUser `yaml:"user"`
}
type rawUser struct {
	Exec *rawExec `yaml:"exec"`
}
type rawExec struct {
	APIVersion         string       `yaml:"apiVersion"`
	Command            string       `yaml:"command"`
	Args               []string     `yaml:"args"`
	Env                []rawExecEnv `yaml:"env"`
	ProvideClusterInfo bool         `yaml:"provideClusterInfo"`
}
type rawExecEnv struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

func stringSliceToList(ss []string) types.List {
	if len(ss) == 0 {
		return types.ListNull(types.StringType)
	}
	items := make([]attr.Value, 0, len(ss))
	for _, s := range ss {
		items = append(items, types.StringValue(s))
	}
	lv, _ := types.ListValue(types.StringType, items)
	return lv
}

func execEnvSliceToList(envs []rawExecEnv) types.List {
	if len(envs) == 0 {
		return types.ListNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"name":  types.StringType,
				"value": types.StringType,
			},
		})
	}
	items := make([]attr.Value, 0, len(envs))
	for _, e := range envs {
		obj, _ := types.ObjectValue(map[string]attr.Type{
			"name":  types.StringType,
			"value": types.StringType,
		}, map[string]attr.Value{
			"name":  types.StringValue(e.Name),
			"value": types.StringValue(e.Value),
		})
		items = append(items, obj)
	}
	lv, _ := types.ListValue(types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"name":  types.StringType,
			"value": types.StringType,
		},
	}, items)
	return lv
}

func mapKubeconfigYAMLToModel(
	ctx context.Context,
	rawYAML string,
	out *kubernetesKubeconfigDataSourceModel,
	diags *diag.Diagnostics,
) {
	out.KubeconfigYAML = types.StringValue(rawYAML)

	var kc rawKubeconfig
	if err := yaml.Unmarshal([]byte(rawYAML), &kc); err != nil {
		diags.AddError("Failed to parse kubeconfig YAML", err.Error())
		return
	}

	out.ApiVersion = types.StringValue(kc.APIVersion)
	out.Kind = types.StringValue(kc.Kind)
	out.CurrentContext = types.StringValue(kc.CurrentContext)

	if len(kc.Preferences) == 0 {
		out.Preferences = types.MapNull(types.StringType)
	} else {
		m := make(map[string]attr.Value, len(kc.Preferences))
		for k, v := range kc.Preferences {
			m[k] = types.StringValue(v)
		}
		mv, _ := types.MapValue(types.StringType, m)
		out.Preferences = mv
	}

	out.Clusters = make([]kcfgClusterEntry, 0, len(kc.Clusters))
	for _, c := range kc.Clusters {
		out.Clusters = append(out.Clusters, kcfgClusterEntry{
			Name: types.StringValue(c.Name),
			Cluster: kcfgCluster{
				Server:                   types.StringValue(c.Cluster.Server),
				CertificateAuthorityData: types.StringValue(c.Cluster.CertificateAuthorityData),
			},
		})
	}

	out.Contexts = make([]kcfgContextEntry, 0, len(kc.Contexts))
	for _, cx := range kc.Contexts {
		out.Contexts = append(out.Contexts, kcfgContextEntry{
			Name: types.StringValue(cx.Name),
			Context: kcfgContext{
				Cluster: types.StringValue(cx.Context.Cluster),
				User:    types.StringValue(cx.Context.User),
			},
		})
	}

	out.Users = make([]kcfgUserEntry, 0, len(kc.Users))
	for _, u := range kc.Users {
		var exec *kcfgExec
		if u.User.Exec != nil {
			exec = &kcfgExec{
				ApiVersion:         types.StringValue(u.User.Exec.APIVersion),
				Command:            types.StringValue(u.User.Exec.Command),
				Args:               stringSliceToList(u.User.Exec.Args),
				Env:                execEnvSliceToList(u.User.Exec.Env),
				ProvideClusterInfo: types.BoolValue(u.User.Exec.ProvideClusterInfo),
			}
		}
		out.Users = append(out.Users, kcfgUserEntry{
			Name: types.StringValue(u.Name),
			User: kcfgUser{Exec: exec},
		})
	}
}

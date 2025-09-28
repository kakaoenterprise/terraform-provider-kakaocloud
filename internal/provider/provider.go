// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package provider

import (
	"context"
	"terraform-provider-kakaocloud/internal/common"
	"terraform-provider-kakaocloud/internal/service/bcs"

	"terraform-provider-kakaocloud/internal/service/image"
	"terraform-provider-kakaocloud/internal/service/kubernetesengine"
	"terraform-provider-kakaocloud/internal/service/loadbalancer"
	"terraform-provider-kakaocloud/internal/service/network"
	"terraform-provider-kakaocloud/internal/service/volume"
	"terraform-provider-kakaocloud/internal/service/vpc"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ provider.Provider = &kakaocloudProvider{}
)

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &kakaocloudProvider{
			version: version,
		}
	}
}

type kakaocloudProviderModel struct {
	ServiceRealm                types.String `tfsdk:"service_realm"`
	Region                      types.String `tfsdk:"region"`
	EndpointOverrides           types.Map    `tfsdk:"endpoint_overrides"`
	ApplicationCredentialId     types.String `tfsdk:"application_credential_id"`
	ApplicationCredentialSecret types.String `tfsdk:"application_credential_secret"`
}

type kakaocloudProvider struct {
	version string
}

func (p *kakaocloudProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "kakaocloud"
	resp.Version = p.version
}

func (p *kakaocloudProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Interact with kakaocloud.",
		Attributes: map[string]schema.Attribute{
			"service_realm": schema.StringAttribute{
				Optional:    true,
				Description: "public|gov",
				Validators:  common.ServiceRealmValidators(),
			},
			"region": schema.StringAttribute{
				Optional:    true,
				Description: "e.g. kr-central-2",
				Validators:  common.RegionValidators(),
			},
			"endpoint_overrides": schema.MapAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Custom endpoint URLs for services",
			},
			"application_credential_id": schema.StringAttribute{
				Optional:    true,
				Description: "Application credential ID",
				Sensitive:   true,
			},
			"application_credential_secret": schema.StringAttribute{
				Optional:    true,
				Description: "Application credential secret",
				Sensitive:   true,
			},
		},
	}
}

func (p *kakaocloudProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config kakaocloudProviderModel

	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var endpointOverrides map[string]string
	if !config.EndpointOverrides.IsNull() && !config.EndpointOverrides.IsUnknown() {
		endpointOverrides = make(map[string]string)
		elements := config.EndpointOverrides.Elements()
		for key, value := range elements {
			if strValue, ok := value.(types.String); ok && !strValue.IsNull() {
				endpointOverrides[key] = strValue.ValueString()
			}
		}
	}

	authConfig := &common.Config{
		ApplicationCredentialID:     config.ApplicationCredentialId,
		ApplicationCredentialSecret: config.ApplicationCredentialSecret,
		ServiceRealm:                config.ServiceRealm,
		Region:                      config.Region,
		EndpointOverrides:           endpointOverrides,
	}

	userAgent := "terraform-provider-kakaocloud/" + p.version
	authClient, err := common.NewClient(authConfig, userAgent)
	if err != nil {
		common.AddGeneralError(ctx, p, &resp.Diagnostics, "Failed to initialize authenticated client: "+err.Error())
		return
	}

	resp.DataSourceData = authClient
	resp.ResourceData = authClient
}

func (p *kakaocloudProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		bcs.NewInstanceDataSource,
		bcs.NewInstancesDataSource,
		bcs.NewInstanceFlavorDataSource,
		bcs.NewInstanceFlavorsDataSource,
		bcs.NewKeypairDataSource,
		bcs.NewKeypairsDataSource,

		volume.NewVolumeTypesDataSource,
		volume.NewVolumeDataSource,
		volume.NewVolumesDataSource,
		volume.NewVolumeSnapshotDataSource,
		volume.NewVolumeSnapshotsDataSource,

		image.NewImagesDataSource,
		image.NewImageDataSource,
		image.NewImageMembersDataSource,

		vpc.NewVpcDataSource,
		vpc.NewVpcsDataSource,
		vpc.NewSubnetDataSource,
		vpc.NewSubnetsDataSource,
		vpc.NewSubnetShareDataSource,
		vpc.NewRouteTableDataSource,
		vpc.NewRouteTablesDataSource,
		vpc.NewNetworkInterfaceDataSource,
		vpc.NewNetworkInterfacesDataSource,

		network.NewPublicIpDataSource,
		network.NewPublicIpsDataSource,
		network.NewSecurityGroupDataSource,
		network.NewSecurityGroupsDataSource,

		loadbalancer.NewLoadBalancerDataSource,
		loadbalancer.NewLoadBalancersDataSource,
		loadbalancer.NewBeyondLoadBalancerDataSource,
		loadbalancer.NewBeyondLoadBalancersDataSource,
		loadbalancer.NewLoadBalancerFlavorsDataSource,
		loadbalancer.NewLoadBalancerSecretsDataSource,
		loadbalancer.NewLoadBalancerListenerDataSource,
		loadbalancer.NewLoadBalancerListenersDataSource,
		loadbalancer.NewLoadBalancerL7PolicyDataSource,
		loadbalancer.NewLoadBalancerL7PoliciesDataSource,
		loadbalancer.NewLoadBalancerL7PolicyRuleDataSource,
		loadbalancer.NewLoadBalancerL7PolicyRulesDataSource,
		loadbalancer.NewLoadBalancerTargetGroupDataSource,
		loadbalancer.NewLoadBalancerTargetGroupsDataSource,
		loadbalancer.NewLoadBalancerTargetGroupMembersDataSource,
		loadbalancer.NewLoadBalancerHealthMonitorDataSource,

		kubernetesengine.NewClusterDataSource,
		kubernetesengine.NewClustersDataSource,
		kubernetesengine.NewNodePoolDataSource,
		kubernetesengine.NewNodePoolsDataSource,
		kubernetesengine.NewKubernetesImagesDataSource,
		kubernetesengine.NewScheduledScalingDataSource,
		kubernetesengine.NewClusterNodeDataSource,
		kubernetesengine.NewKubernetesVersionsDataSource,
		kubernetesengine.NewKubernetesKubeconfigDataSource,
	}
}

func (p *kakaocloudProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		bcs.NewInstanceResource,
		bcs.NewKeypairResource,

		volume.NewVolumeResource,
		volume.NewVolumeSnapshotResource,

		image.NewImageResource,
		image.NewImageMemberResource,

		vpc.NewVpcResource,
		vpc.NewSubnetResource,
		vpc.NewSubnetShareResource,
		vpc.NewRouteTableResource,
		vpc.NewNetworkInterfaceResource,

		network.NewPublicIpResource,
		network.NewSecurityGroupResource,

		loadbalancer.NewLoadBalancerResource,
		loadbalancer.NewBeyondLoadBalancerResource,
		loadbalancer.NewLoadBalancerListenerResource,
		loadbalancer.NewLoadBalancerL7PolicyResource,
		loadbalancer.NewLoadBalancerL7PolicyRuleResource,
		loadbalancer.NewLoadBalancerTargetGroupResource,
		loadbalancer.NewLoadBalancerTargetGroupMemberResource,
		loadbalancer.NewLoadBalancerHealthMonitorResource,

		kubernetesengine.NewNodePoolResource,
		kubernetesengine.NewClusterResource,
		kubernetesengine.NewScheduledScalingResource,
		kubernetesengine.NewClusterNodeResource,
	}
}

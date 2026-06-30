// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package common

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	kakaocloud "github.com/kakaoenterprise/kc-sdk-go/common"
	"github.com/kakaoenterprise/kc-sdk-go/services/bcs"
	"github.com/kakaoenterprise/kc-sdk-go/services/config"
	iam "github.com/kakaoenterprise/kc-sdk-go/services/iam"
	"github.com/kakaoenterprise/kc-sdk-go/services/image"
	"github.com/kakaoenterprise/kc-sdk-go/services/kubernetesengine"
	"github.com/kakaoenterprise/kc-sdk-go/services/loadbalancer"
	"github.com/kakaoenterprise/kc-sdk-go/services/mysql"
	"github.com/kakaoenterprise/kc-sdk-go/services/network"
	"github.com/kakaoenterprise/kc-sdk-go/services/tgw"
	"github.com/kakaoenterprise/kc-sdk-go/services/volume"
	"github.com/kakaoenterprise/kc-sdk-go/services/vpc"
	"golang.org/x/net/context"
)

func (c *KakaoCloudClient) initEndpoints() kakaocloud.Endpoints {

	endpoints := kakaocloud.Endpoints{
		IAM:    getSDKDefaultEndpoint(iam.NewConfiguration()),
		Config: getSDKDefaultEndpoint(config.NewConfiguration()),
	}

	for service, endpoint := range c.Config.EndpointOverrides {
		switch service {
		case "iam":
			endpoints.IAM = endpoint
		case "config":
			endpoints.Config = endpoint
		}
	}
	return endpoints
}

func getSDKDefaultEndpoint(cfg interface{}) string {
	v := reflect.ValueOf(cfg)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	serversField := v.FieldByName("Servers")
	if !serversField.IsValid() || serversField.IsNil() {
		return ""
	}

	servers := serversField.Interface()
	serversValue := reflect.ValueOf(servers)

	if serversValue.Kind() == reflect.Slice && serversValue.Len() > 0 {
		urlField := serversValue.Index(0).FieldByName("URL")
		if urlField.IsValid() && urlField.Kind() == reflect.String {
			return urlField.String()
		}
	}

	return ""
}

func applyClientEndpoint(endpoints *kakaocloud.Endpoints, clientEndpoint *config.ClientEndpoint) bool {
	if endpoints == nil || clientEndpoint == nil {
		return false
	}

	for service, endpoint := range clientEndpoint.ServiceEndpoints {
		switch service {
		case "iam":
			endpoints.IAM = endpoint
		case "vpc":
			endpoints.VPC = endpoint
		case "network":
			endpoints.Network = endpoint
		case "load-balancer":
			endpoints.LoadBalancer = endpoint
		case "volume":
			endpoints.Volume = endpoint
		case "image":
			endpoints.Image = endpoint
		case "bcs":
			endpoints.BCS = endpoint
		case "kubernetes-engine":
			endpoints.KubernetesEngine = endpoint
		case "tgw":
			endpoints.TGW = endpoint
		case "mysql":
			endpoints.MySQL = endpoint
		}
	}
	return true
}

func applySDKDefaultEndpoints(endpoints *kakaocloud.Endpoints) {
	if endpoints == nil {
		return
	}

	endpoints.IAM = getSDKDefaultEndpoint(iam.NewConfiguration())
	endpoints.VPC = getSDKDefaultEndpoint(vpc.NewConfiguration())
	endpoints.Network = getSDKDefaultEndpoint(network.NewConfiguration())
	endpoints.LoadBalancer = getSDKDefaultEndpoint(loadbalancer.NewConfiguration())
	endpoints.Volume = getSDKDefaultEndpoint(volume.NewConfiguration())
	endpoints.Image = getSDKDefaultEndpoint(image.NewConfiguration())
	endpoints.BCS = getSDKDefaultEndpoint(bcs.NewConfiguration())
	endpoints.KubernetesEngine = getSDKDefaultEndpoint(kubernetesengine.NewConfiguration())
	endpoints.TGW = getSDKDefaultEndpoint(tgw.NewConfiguration())
	endpoints.MySQL = getSDKDefaultEndpoint(mysql.NewConfiguration())
}

func (c *KakaoCloudClient) loadEndpointsFromConfigAPI(ctx context.Context, endpoints *kakaocloud.Endpoints) (*kakaocloud.Endpoints, error) {
	diags := &diag.Diagnostics{}

	respModel, httpResp, err := ExecuteWithRetryAndAuth(ctx, c, diags,
		func() (*config.ClientEndpointResponse, *http.Response, error) {
			return c.ApiClient.ConfigAPI.ResolveClientEndpoint(ctx).
				XAuthToken(c.XAuthToken).Execute()
		},
	)
	if err != nil {
		if c.Config.EndpointOverrides == nil || len(c.Config.EndpointOverrides) == 0 {
			tflog.Warn(ctx, "ResolveClientEndpoint failed, default Endpoints will be used")
			applySDKDefaultEndpoints(endpoints)
			return endpoints, nil
		}

		AddApiActionError(ctx, c, httpResp, "ResolveClientEndpoint", err, diags)
		return nil, err
	}

	ok := applyClientEndpoint(endpoints, respModel.Data)
	if !ok {
		return nil, fmt.Errorf("failed to apply client endpoint: %w", err)
	}
	return endpoints, nil
}

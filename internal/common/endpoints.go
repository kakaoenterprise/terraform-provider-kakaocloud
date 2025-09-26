// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package common

import (
	"fmt"
	"reflect"

	kakaocloud "github.com/kakaoenterprise/kc-sdk-go/common"
	bcs "github.com/kakaoenterprise/kc-sdk-go/services/bcs"
	iam "github.com/kakaoenterprise/kc-sdk-go/services/iam"
	image "github.com/kakaoenterprise/kc-sdk-go/services/image"
	kubernetesEngine "github.com/kakaoenterprise/kc-sdk-go/services/kubernetesengine"
	loadbalancer "github.com/kakaoenterprise/kc-sdk-go/services/loadbalancer"
	network "github.com/kakaoenterprise/kc-sdk-go/services/network"
	volume "github.com/kakaoenterprise/kc-sdk-go/services/volume"
	vpc "github.com/kakaoenterprise/kc-sdk-go/services/vpc"
)

func (c *KakaoCloudClient) buildEndpoints() kakaocloud.Endpoints {

	endpoints := kakaocloud.Endpoints{
		IAM:              getSDKDefaultEndpoint(iam.NewConfiguration()),
		VPC:              getSDKDefaultEndpoint(vpc.NewConfiguration()),
		Network:          getSDKDefaultEndpoint(network.NewConfiguration()),
		LoadBalancer:     getSDKDefaultEndpoint(loadbalancer.NewConfiguration()),
		Volume:           getSDKDefaultEndpoint(volume.NewConfiguration()),
		Image:            getSDKDefaultEndpoint(image.NewConfiguration()),
		BCS:              getSDKDefaultEndpoint(bcs.NewConfiguration()),
		KubernetesEngine: getSDKDefaultEndpoint(kubernetesEngine.NewConfiguration()),
	}

	for service, endpoint := range c.Config.EndpointOverrides {
		switch service {
		case "iam":
			endpoints.IAM = endpoint
		case "vpc":
			endpoints.VPC = endpoint
		case "network":
			endpoints.Network = endpoint
		case "load_balancer":
			endpoints.LoadBalancer = endpoint
		case "volume":
			endpoints.Volume = endpoint
		case "image":
			endpoints.Image = endpoint
		case "bcs":
			endpoints.BCS = endpoint
		case "kubernetes_engine":
			endpoints.KubernetesEngine = endpoint
		}
	}

	fmt.Printf("Endpoints: %+v\n", endpoints)

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

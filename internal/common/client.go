// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package common

import (
	"fmt"
	"os"
	"terraform-provider-kakaocloud/internal/auth"

	"github.com/hashicorp/terraform-plugin-framework/types"
	kakaocloud "github.com/kakaoenterprise/kc-sdk-go/common"
	"golang.org/x/net/context"
)

type Config struct {
	ApplicationCredentialID     types.String
	ApplicationCredentialSecret types.String
	ServiceRealm                types.String
	Region                      types.String
	EndpointOverrides           map[string]string
}

type KakaoCloudClient struct {
	Config          *Config
	TokenManager    *auth.TokenManager
	ApiClient       *kakaocloud.APIClient
	XAuthToken      string
	XApiVersion     string
	ServiceAzPolicy map[string]map[string]struct{}
}

func NewClient(ctx context.Context, config *Config, userAgent, apiVersion string) (*KakaoCloudClient, error) {
	if err := completeConfig(config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	client := &KakaoCloudClient{
		Config: config,
	}

	endpoints := client.initEndpoints()
	client.ApiClient = kakaocloud.NewAPIClient(kakaocloud.Config{
		Endpoints: endpoints,
		UserAgent: userAgent,
		Version:   apiVersion,
	})

	client.TokenManager = auth.NewTokenManager(
		client.ApiClient.IdentityAPI,
		client.Config.ApplicationCredentialID.ValueString(),
		client.Config.ApplicationCredentialSecret.ValueString(),
	)

	resolvedEndpoints, err := client.loadEndpointsFromConfigAPI(ctx, &endpoints)
	if err != nil {
		return nil, fmt.Errorf("failed to load endpoints: %w", err)
	}

	client.ApiClient = kakaocloud.NewAPIClient(kakaocloud.Config{
		Endpoints: *resolvedEndpoints,
		UserAgent: userAgent,
		Version:   apiVersion,
	})

	client.ServiceAzPolicy, err = client.loadServiceAzPolicyFromConfigAPI(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load az policy: %w", err)
	}

	return client, nil
}

func completeConfig(config *Config) error {

	if config.ApplicationCredentialID.IsNull() {
		config.ApplicationCredentialID = types.StringValue(os.Getenv("KAKAOCLOUD_APPLICATION_CREDENTIAL_ID"))
	}
	if config.ApplicationCredentialSecret.IsNull() {
		config.ApplicationCredentialSecret = types.StringValue(os.Getenv("KAKAOCLOUD_APPLICATION_CREDENTIAL_SECRET"))
	}

	if config.ApplicationCredentialID.ValueString() == "" {
		return fmt.Errorf("application_credential_id is required")
	}
	if config.ApplicationCredentialSecret.ValueString() == "" {
		return fmt.Errorf("application_credential_secret is required")
	}

	if config.ServiceRealm.IsUnknown() || config.ServiceRealm.IsNull() {
		config.ServiceRealm = types.StringValue("public")
	}
	if config.Region.IsUnknown() || config.Region.IsNull() {
		config.Region = types.StringValue("kr-central-2")
	}

	if config.EndpointOverrides == nil {
		config.EndpointOverrides = make(map[string]string)
	}

	return nil
}

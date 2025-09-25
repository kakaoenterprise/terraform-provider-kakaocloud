// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"fmt"
	"os"
	"terraform-provider-kakaocloud/internal/auth"

	"github.com/hashicorp/terraform-plugin-framework/types"
	kakaocloud "github.com/kakaoenterprise/kc-sdk-go/common"
)

type Config struct {
	ApplicationCredentialID     types.String
	ApplicationCredentialSecret types.String
	ServiceRealm                types.String
	Region                      types.String
	AvailabilityZones           []string
	EndpointOverrides           map[string]string
}

type KakaoCloudClient struct {
	Config       *Config
	TokenManager *auth.TokenManager
	ApiClient    *kakaocloud.APIClient
	XAuthToken   string
}

// NewClient creates a new authenticated KakaoCloud client
func NewClient(config *Config, userAgent string) (*KakaoCloudClient, error) {
	if err := completeConfig(config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	// Create client instance
	client := &KakaoCloudClient{
		Config: config,
	}

	// Initialize API client
	endpoints := client.buildEndpoints()
	client.ApiClient = kakaocloud.NewAPIClient(kakaocloud.Config{
		Endpoints: endpoints,
		UserAgent: userAgent,
	})

	// Initialize token manager
	client.TokenManager = auth.NewTokenManager(
		client.ApiClient.IdentityAPI,
		client.Config.ApplicationCredentialID.ValueString(),
		client.Config.ApplicationCredentialSecret.ValueString(),
	)

	return client, nil
}

func completeConfig(config *Config) error {
	// Load credentials from environment if not provided
	if config.ApplicationCredentialID.IsNull() {
		config.ApplicationCredentialID = types.StringValue(os.Getenv("APPLICATION_CREDENTIAL_ID"))
	}
	if config.ApplicationCredentialSecret.IsNull() {
		config.ApplicationCredentialSecret = types.StringValue(os.Getenv("APPLICATION_CREDENTIAL_SECRET"))
	}

	// Validate required fields
	if config.ApplicationCredentialID.ValueString() == "" {
		return fmt.Errorf("application_credential_id is required")
	}
	if config.ApplicationCredentialSecret.ValueString() == "" {
		return fmt.Errorf("application_credential_secret is required")
	}

	// Set default values
	if config.Region.IsUnknown() || config.ServiceRealm.IsNull() {
		config.ServiceRealm = types.StringValue("public")
	}
	if config.Region.IsUnknown() || config.Region.IsNull() {
		config.Region = types.StringValue("kr-central-2")
	}

	if config.EndpointOverrides == nil {
		config.EndpointOverrides = make(map[string]string)
	}

	// Validate region and service realm combination
	availabilityZones, ok := AvailabilityZonesFor(config.ServiceRealm.ValueString(), config.Region.ValueString())
	if !ok {
		return fmt.Errorf("unsupported combination: service_realm=%s, region=%s",
			config.ServiceRealm.ValueString(), config.Region.ValueString())
	}
	config.AvailabilityZones = availabilityZones

	return nil
}

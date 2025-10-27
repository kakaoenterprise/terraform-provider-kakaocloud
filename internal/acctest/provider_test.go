// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package acctest

import (
	"terraform-provider-kakaocloud/internal/provider"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

const (
	providerConfig = `
	provider "kakaocloud" {
	x_auth_token =  "your-auth-token-here"
	}
`
)

var (
	testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"kakaocloud": providerserver.NewProtocol6WithError(provider.New("test")()),
	}
)

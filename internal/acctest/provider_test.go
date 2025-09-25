// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package acctest

import (
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"terraform-provider-kakaocloud/internal/provider"
)

const (
	providerConfig = `
	provider "kakaocloud" {
	x_auth_token =  "gAAAAABohzSLs3PCvr1ytmHl9wNtVbllQtt5C8tIsXsqf2qU3sChzr2So8P6TCGnjDegpby1GAsCy0tR8Aqhy35Ip7U64s734nsEk15MbvsX9MGba08-aH2Ho8wkbpLhi7sL2k0eX6bzLCtbBu_W93g_TtCDinZZOf46l2FfzYKl1btdZSaV9IMov2o6h4QgCOzNWcAq_2t6"
	}
`
)

var (
	testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"kakaocloud": providerserver.NewProtocol6WithError(provider.New("test")()),
	}
)

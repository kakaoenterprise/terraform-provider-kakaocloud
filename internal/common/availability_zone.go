// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package common

import (
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/kakaoenterprise/kc-sdk-go/services/config"
	"golang.org/x/net/context"
)

func (c *KakaoCloudClient) loadServiceAzPolicyFromConfigAPI(ctx context.Context) (map[string]map[string]struct{}, error) {
	diags := &diag.Diagnostics{}

	result := make(map[string]map[string]struct{})

	respModel, httpResp, err := ExecuteWithRetryAndAuth(ctx, c, diags,
		func() (*config.AzPolicyResponse, *http.Response, error) {
			return c.ApiClient.ConfigAPI.ResolveAzPolicy(ctx).
				XAuthToken(c.XAuthToken).Execute()
		},
	)
	if err != nil {
		if c.Config.EndpointOverrides == nil || len(c.Config.EndpointOverrides) == 0 {
			tflog.Warn(ctx, "ResolveAzPolicy failed, availability zone validation disabled")
			return result, nil
		}

		AddApiActionError(ctx, c, httpResp, "ResolveAzPolicy", err, diags)
		return nil, err
	}

	for svc, list := range respModel.Data.Services {
		m := make(map[string]struct{}, len(list))
		for _, az := range list {
			m[az] = struct{}{}
		}
		result[svc] = m
	}

	return result, nil
}

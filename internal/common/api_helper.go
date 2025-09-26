// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package common

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-kakaocloud/internal/auth"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func ExecuteWithRetryAndAuth[T any](
	ctx context.Context,
	kc *KakaoCloudClient,
	respDiags *diag.Diagnostics,
	operation func() (T, *http.Response, error),
) (T, *http.Response, error) {
	var zero T
	maxAuthRetries := 2
	max429Retries := 1000
	interval := 100 * time.Millisecond

	action := GetCallerMethodName()

	for authAttempt := 1; authAttempt <= maxAuthRetries; authAttempt++ {
		token, err := kc.TokenManager.GetValidToken(ctx)
		if err != nil {
			return zero, nil, fmt.Errorf("failed to get valid token: %w", err)
		}
		kc.XAuthToken = token

		for i := 0; i < max429Retries; i++ {
			result, httpResp, err := operation()

			if httpResp == nil || httpResp.StatusCode != http.StatusTooManyRequests {
				if auth.IsAuthError(err) && authAttempt < maxAuthRetries {
					kc.TokenManager.InvalidateToken()
					break
				}

				return result, httpResp, err
			}

			tflog.Warn(ctx, fmt.Sprintf(
				"Received 429 Too Many Requests during %s. Retrying %d/%d after %dms...",
				action, i+1, max429Retries, interval.Milliseconds(),
			))

			select {
			case <-time.After(interval):
				continue
			case <-ctx.Done():
				respDiags.AddError(
					fmt.Sprintf("Error during %s", action),
					"Context cancelled while retrying due to 429",
				)
				return zero, httpResp, ctx.Err()
			}
		}

		if authAttempt == maxAuthRetries {
			respDiags.AddError(
				fmt.Sprintf("Error during %s", action),
				fmt.Sprintf("Exceeded max retry attempts (%d) due to 429 Too Many Requests", max429Retries),
			)
			return zero, nil, fmt.Errorf("max retries exceeded for %s", action)
		}
	}

	return zero, nil, fmt.Errorf("unexpected error: should not reach here")
}

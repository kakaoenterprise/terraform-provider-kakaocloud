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

// ExecuteWithRetryAndAuth executes operations with automatic token refresh and 429 retry handling
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

	// Outer loop: handle token refresh
	for authAttempt := 1; authAttempt <= maxAuthRetries; authAttempt++ {
		token, err := kc.TokenManager.GetValidToken(ctx)
		if err != nil {
			return zero, nil, fmt.Errorf("failed to get valid token: %w", err)
		}
		kc.XAuthToken = token

		// Inner loop: handle 429 rate limiting
		for i := 0; i < max429Retries; i++ {
			result, httpResp, err := operation()

			// Handle non-429 responses
			if httpResp == nil || httpResp.StatusCode != http.StatusTooManyRequests {
				if auth.IsAuthError(err) && authAttempt < maxAuthRetries {
					kc.TokenManager.InvalidateToken()
					break // Retry with new token
				}
				// Success, other error, or final auth attempt
				return result, httpResp, err
			}

			// Handle 429: wait and retry
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

		// 429 retries exhausted on final auth attempt
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

// Legacy code - kept for reference

// // ExecuteWithAutoRefresh handles API calls with automatic token refresh
// func ExecuteWithAutoRefresh(
// 	ctx context.Context,
// 	kc *KakaoCloudClient,
// 	sdkCall func(token string) (interface{}, *http.Response, error),
// ) (interface{}, *http.Response, error) {
// 	if kc.tokenManager == nil {
// 		return nil, nil, fmt.Errorf("token manager not initialized")
// 	}

// 	// Max 2 attempts (retry after token refresh on first failure)
// 	for attempt := 1; attempt <= 2; attempt++ {
// 		token, err := kc.tokenManager.GetValidToken(ctx)
// 		if err != nil {
// 			return nil, nil, fmt.Errorf("failed to get valid token: %w", err)
// 		}

// 		result, httpResp, err := sdkCall(token)

// 		// Return result if not auth error or final attempt
// 		if !auth.IsAuthError(err) || attempt == 2 {
// 			return result, httpResp, err
// 		}

// 		// Invalidate token on auth error
// 		kc.tokenManager.InvalidateToken()
// 	}

// 	return nil, nil, fmt.Errorf("unexpected error: should not reach here")
// }

// func RetryOnTooManyRequests[T any](
// 	ctx context.Context, obj interface{},
// 	respDiags *diag.Diagnostics,
// 	operation func() (T, *http.Response, error),
// ) (T, *http.Response, error) {
// 	var zero T
// 	maxRetries := 1000
// 	interval := 100 * time.Millisecond

// 	typeName, tfObjectType := ExtractTypeMetadata(ctx, obj)
// 	action := GetCallerMethodName()

// 	for i := 0; i < maxRetries; i++ {
// 		resp, httpResp, err := operation()
// 		if httpResp == nil || httpResp.StatusCode != http.StatusTooManyRequests {
// 			// Success or non-429 response
// 			return resp, httpResp, err
// 		}

// 		if httpResp != nil && httpResp.StatusCode == http.StatusTooManyRequests {
// 			tflog.Warn(ctx, fmt.Sprintf(
// 				"Received 429 Too Many Requests during %s %s %s. Retrying %d/%d after %d...",
// 				tfObjectType, typeName, action, i+1, maxRetries, interval,
// 			))
// 			select {
// 			case <-time.After(interval):
// 				interval *= 1
// 				continue
// 			case <-ctx.Done():
// 				respDiags.AddError(
// 					fmt.Sprintf("Error during %s %s %s", tfObjectType, typeName, action),
// 					"Context cancelled while retrying due to 429",
// 				)
// 				return zero, httpResp, ctx.Err()
// 			}
// 		}
// 		// Other errors: return immediately
// 		return resp, httpResp, err
// 	}

// 	respDiags.AddError(
// 		fmt.Sprintf("Error %s %s: %s", action, tfObjectType, typeName),
// 		fmt.Sprintf("Exceeded max retry attempts (%d) due to 429 Too Many Requests", maxRetries),
// 	)
// 	return zero, nil, fmt.Errorf("max retries exceeded for %s", action)
// }

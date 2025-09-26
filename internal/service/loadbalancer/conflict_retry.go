// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package loadbalancer

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"terraform-provider-kakaocloud/internal/common"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func ExecuteWithLoadBalancerConflictRetry[T any](
	ctx context.Context,
	kc *common.KakaoCloudClient,
	respDiags *diag.Diagnostics,
	operation func() (T, *http.Response, error),
) (T, *http.Response, error) {
	var zero T
	maxRetries := 10
	conflictInterval := 1 * time.Second

	action := common.GetCallerMethodName()

	for i := 0; i < maxRetries; i++ {
		result, httpResp, err := operation()

		if httpResp != nil && httpResp.StatusCode == http.StatusConflict {
			if shouldRetryLoadBalancerConflict(err, httpResp) {
				if i < maxRetries-1 {
					tflog.Warn(ctx, fmt.Sprintf(
						"Received retryable 409 Conflict during %s. Retrying %d/%d after %v...",
						action, i+1, maxRetries, conflictInterval,
					))

					select {
					case <-time.After(conflictInterval):
						conflictInterval *= 2
						continue
					case <-ctx.Done():
						respDiags.AddError(
							fmt.Sprintf("Error during %s", action),
							"Context cancelled while retrying due to 409 conflict",
						)
						return zero, httpResp, ctx.Err()
					}
				}

				respDiags.AddError(
					fmt.Sprintf("Error during %s", action),
					fmt.Sprintf("Exceeded max retry attempts (%d) due to 409 Conflict", maxRetries),
				)
				return zero, httpResp, err
			}

			return result, httpResp, err
		}

		return result, httpResp, err
	}

	return zero, nil, fmt.Errorf("unexpected error: should not reach here")
}

func shouldRetryLoadBalancerConflict(err error, resp *http.Response) bool {
	if resp == nil || resp.Body == nil {
		return false
	}

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return false
	}
	resp.Body = io.NopCloser(strings.NewReader(string(body)))

	errorMsg := strings.ToLower(string(body))

	retryableKeywords := []string{
		"resource is locked",
		"operation in progress",
		"load balancer is being modified",
		"concurrent modification",
		"temporarily unavailable",
		"resource busy",
		"immutable",
		"cannot modify",
		"modification in progress",
	}

	for _, keyword := range retryableKeywords {
		if strings.Contains(errorMsg, keyword) {
			tflog.Info(context.Background(), fmt.Sprintf("Retryable 409 conflict detected: %s", keyword))
			return true
		}
	}

	tflog.Info(context.Background(), fmt.Sprintf("Non-retryable 409 conflict: %s", errorMsg))
	return false
}

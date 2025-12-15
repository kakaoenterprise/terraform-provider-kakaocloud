// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package common

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func PollUntilResultWithTimeout[T any](
	ctx context.Context,
	obj interface{},
	interval time.Duration,
	timeout *time.Duration,
	targetName string,
	targetId string,
	targetStatuses []string,
	respDiags *diag.Diagnostics,
	fetch func(context.Context) (T, *http.Response, error),
	getStatus func(T) string,
) (T, bool) {
	var zero T

	var ctxWithTimeout context.Context
	var cancel context.CancelFunc
	if timeout != nil {
		ctxWithTimeout, cancel = context.WithTimeout(ctx, *timeout)
		defer cancel()
	} else {
		ctxWithTimeout = ctx
	}

	start := time.Now()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	retry404Count := 0
	maxRetries := 10

	typeName, _ := ExtractTypeMetadata(ctx, obj)

	for {
		select {
		case <-ctxWithTimeout.Done():
			AddGeneralError(ctx, obj, respDiags, "context deadline exceeded")
			return zero, false

		case <-ticker.C:
			result, httpResp, err := fetch(ctxWithTimeout)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
					retry404Count++
					if retry404Count <= maxRetries {
						tflog.Warn(ctxWithTimeout, fmt.Sprintf(
							"%s not found (404). Retrying %d/%d...",
							typeName, retry404Count, maxRetries,
						))
						continue
					}
					AddApiActionError(
						ctx,
						obj,
						httpResp,
						"PollForStatus",
						err,
						respDiags,
						fmt.Sprintf("The requested %s '%s' does not exsist or is not accessible after %d retries. Please verify the resource exists.", targetName, targetId, maxRetries),
					)
					return zero, false
				}

				AddApiActionError(ctx, obj, httpResp, "PollForStatus", err, respDiags)
				return zero, false
			}

			status := getStatus(result)
			for _, s := range targetStatuses {
				if status == s {
					return result, true
				}
			}

			elapsed := time.Since(start).Round(time.Second)
			tflog.Info(ctxWithTimeout, fmt.Sprintf("%s... [%s elapsed]", status, elapsed))
		}
	}
}

func PollUntilResult[T any](
	ctx context.Context,
	obj interface{},
	interval time.Duration,
	targetName string,
	targetId string,
	targetStatuses []string,
	respDiags *diag.Diagnostics,
	fetch func(context.Context) (T, *http.Response, error),
	getStatus func(T) string,
) (T, bool) {
	return PollUntilResultWithTimeout(
		ctx,
		obj,
		interval,
		nil,
		targetName,
		targetId,
		targetStatuses,
		respDiags,
		fetch,
		getStatus,
	)
}

func PollUntilDeletion(
	ctx context.Context,
	obj interface{},
	interval time.Duration,
	respDiags *diag.Diagnostics,
	check func(ctx context.Context) (bool, *http.Response, error),
) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			AddGeneralError(ctx, obj, respDiags, "Context cancelled while waiting for deletion")
			return

		case <-ticker.C:
			deleted, httpResp, err := check(ctx)

			if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
				return
			}

			if err != nil {
				AddApiActionError(ctx, obj, httpResp, "PollForDeletion", err, respDiags)
				return
			}

			if deleted {
				return
			}
		}
	}
}

func CheckResourceAvailableStatus(ctx context.Context, obj interface{}, statusPtr *string, expected []string, diags *diag.Diagnostics) {
	typeName, _ := ExtractTypeMetadata(ctx, obj)
	action := GetCallerMethodName()

	if statusPtr == nil {
		diags.AddError(
			fmt.Sprintf("Error %s resource: %s", action, typeName),
			"status is nil",
		)
		return
	}

	status := *statusPtr
	for _, s := range expected {
		if status == s {
			return
		}
	}

	diags.AddError(
		fmt.Sprintf("Error %s resource: %s", action, typeName),
		fmt.Sprintf("Status is %q, expected one of: %v", status, expected),
	)
}

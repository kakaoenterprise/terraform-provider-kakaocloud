package // Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
common

import (
	"context"
	"sync"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/action"
)

func SendActionProgress(send func(action.InvokeProgressEvent), message string) {
	if send == nil {
		return
	}
	send(action.InvokeProgressEvent{Message: message})
}

func StartActionProgress(ctx context.Context, send func(action.InvokeProgressEvent), message string) func() {
	if send == nil {
		return func() {}
	}

	SendActionProgress(send, message)

	done := make(chan struct{})
	var closeOnce sync.Once
	ticker := time.NewTicker(defaultActionProgressInterval)

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-done:
				return
			case <-ticker.C:
				SendActionProgress(send, message)
			}
		}
	}()

	return func() {
		closeOnce.Do(func() {
			close(done)
		})
	}
}

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package common

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"golang.org/x/net/context"
)

var idMutexes = struct {
	sync.Mutex
	m map[string]*sync.Mutex
}{m: make(map[string]*sync.Mutex)}

func LockForID(id string) *sync.Mutex {
	idMutexes.Lock()
	defer idMutexes.Unlock()

	if _, ok := idMutexes.m[id]; !ok {
		idMutexes.m[id] = &sync.Mutex{}
	}
	return idMutexes.m[id]
}

func LockForIDContext(ctx context.Context, id string) (*sync.Mutex, error) {
	mutex := LockForID(id)

	if mutex.TryLock() {
		return mutex, nil
	}

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()

		case <-ticker.C:
			if mutex.TryLock() {
				return mutex, nil
			}
		}
	}
}

func AcquireLockForID(
	ctx context.Context,
	id string,
	diags *diag.Diagnostics,
) (func(), bool) {
	mutex, err := LockForIDContext(ctx, id)
	if err != nil {
		diags.AddError(
			"Timeout While Waiting for Lock",
			fmt.Sprintf("Failed to acquire lock for %q: %s", id, err),
		)
		return nil, false
	}

	return mutex.Unlock, true
}

func AcquireLocksForIDs(
	ctx context.Context,
	ids []string,
	diags *diag.Diagnostics,
) (func(), bool) {
	uniqueIDs := make(map[string]struct{}, len(ids))
	orderedIDs := make([]string, 0, len(ids))
	for _, id := range ids {
		if _, exists := uniqueIDs[id]; exists {
			continue
		}
		uniqueIDs[id] = struct{}{}
		orderedIDs = append(orderedIDs, id)
	}
	sort.Strings(orderedIDs)

	unlockFuncs := make([]func(), 0, len(orderedIDs))
	unlockAll := func() {
		for i := len(unlockFuncs) - 1; i >= 0; i-- {
			unlockFuncs[i]()
		}
	}

	for _, id := range orderedIDs {
		unlock, ok := AcquireLockForID(ctx, id, diags)
		if !ok {
			unlockAll()
			return nil, false
		}
		unlockFuncs = append(unlockFuncs, unlock)
	}

	return unlockAll, true
}

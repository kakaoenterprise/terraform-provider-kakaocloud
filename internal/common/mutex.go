// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package common

import "sync"

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

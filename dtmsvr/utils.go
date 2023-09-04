/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package dtmsvr

import (
	"context"
	"fmt"
	"github.com/dtm-labs/dtm/client/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmsvr/config"
	"github.com/dtm-labs/dtm/dtmsvr/storage"
	"github.com/dtm-labs/dtm/dtmsvr/storage/registry"
	"github.com/lithammer/shortuuid/v3"
	"reflect"
	"time"
	"unsafe"
)

type branchStatus struct {
	gid        string
	branchID   string
	op         string
	status     string
	finishTime *time.Time
}

var e2p = dtmimp.E2P

var conf = &config.Config

// GetStore returns storage.Store
func GetStore() storage.Store {
	return registry.GetStore()
}

// TransProcessedTestChan only for test usage. when transaction processed once, write gid to this chan
var TransProcessedTestChan chan string

// GenGid generate gid, use uuid
func GenGid() string {
	return shortuuid.New()
}

// GetTransGlobal construct trans from db
func GetTransGlobal(gid string) *TransGlobal {
	trans := GetStore().FindTransGlobalStore(gid)
	//nolint:staticcheck
	dtmimp.PanicIf(trans == nil, fmt.Errorf("no TransGlobal with gid: %s found", gid))
	//nolint:staticcheck
	return &TransGlobal{TransGlobalStore: *trans}
}

type iface struct {
	itab, data uintptr
}

type valueCtx struct {
	context.Context
	key, value any
}

type cancelCtx struct {
	context.Context
}

type timerCtx struct {
	*cancelCtx
}

// CopyContext copy context with value and grpc metadata
// if raw context is nil, return nil
func CopyContext(ctx context.Context) context.Context {
	if ctx == nil {
		return ctx
	}
	newCtx := context.Background()
	kv := make(map[interface{}]interface{})
	getKeyValues(ctx, kv)
	for k, v := range kv {
		newCtx = context.WithValue(newCtx, k, v)
	}
	return newCtx
}

func getKeyValues(ctx context.Context, kv map[interface{}]interface{}) {
	if ctx == nil {
		return
	}
	ictx := *(*iface)(unsafe.Pointer(&ctx))
	if ictx.data == 0 {
		return
	}
	valCtx := (*valueCtx)(unsafe.Pointer(ictx.data))
	if valCtx.key != nil && valCtx.value != nil && reflect.TypeOf(ctx).String() == "*context.valueCtx" {
		kv[valCtx.key] = valCtx.value
	}
	if reflect.Indirect(reflect.ValueOf(ctx)).Type().String() == "context.timerCtx" {
		tCtx := (*timerCtx)(unsafe.Pointer(ictx.data))
		getKeyValues(tCtx.cancelCtx, kv)
		return
	}
	getKeyValues(valCtx.Context, kv)
}

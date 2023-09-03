/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package dtmsvr

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

func TestUtils(t *testing.T) {
	CronExpiredTrans(1)
	sleepCronTime()
}

func TestSetNextCron(t *testing.T) {
	conf.RetryInterval = 10
	tg := TransGlobal{}
	tg.NextCronInterval = conf.RetryInterval
	tg.RetryInterval = 15
	assert.Equal(t, int64(15), tg.getNextCronInterval(cronReset))
	tg.RetryInterval = 0
	assert.Equal(t, conf.RetryInterval, tg.getNextCronInterval(cronReset))
	assert.Equal(t, conf.RetryInterval*2, tg.getNextCronInterval(cronBackoff))
	tg.TimeoutToFail = 3
	assert.Equal(t, int64(3), tg.getNextCronInterval(cronReset))
}

type testContextType string

func TestCopyContext(t *testing.T) {
	var key testContextType = "key"
	var value testContextType = "value"
	ctxWithValue := context.WithValue(context.Background(), key, value)
	newCtx := CopyContext(ctxWithValue)
	assert.Equal(t, ctxWithValue.Value(key), newCtx.Value(key))

	var ctx context.Context
	newCtx = CopyContext(ctx)
	assert.Nil(t, newCtx)
}

func TestCopyContextRecursive(t *testing.T) {
	var key testContextType = "key"
	var key2 testContextType = "key2"
	var value testContextType = "value"
	var value2 testContextType = "value2"
	var nestedKey testContextType = "nested_key"
	var nestedValue testContextType = "nested_value"
	ctxWithValue := context.WithValue(context.Background(), key, value)
	nestedCtx := context.WithValue(ctxWithValue, nestedKey, nestedValue)
	timer, cancel := context.WithCancel(nestedCtx)
	defer cancel()
	context.WithValue(timer, key2, value2)
	newCtx := CopyContext(timer)

	assert.Equal(t, timer.Value(nestedKey), newCtx.Value(nestedKey))
	assert.Equal(t, timer.Value(key), newCtx.Value(key))
	assert.Equal(t, timer.Value(key2), newCtx.Value(key2))
}

func TestCopyContextWithMetadata(t *testing.T) {
	md := metadata.New(map[string]string{"key": "value"})
	ctx := metadata.NewIncomingContext(context.Background(), md)
	ctx = metadata.NewOutgoingContext(ctx, md)
	newCtx := CopyContext(ctx)

	copiedMD, ok := metadata.FromIncomingContext(newCtx)
	assert.True(t, ok)
	assert.Equal(t, 1, len(copiedMD["key"]))
	assert.Equal(t, "value", copiedMD["key"][0])
	copiedMD, ok = metadata.FromOutgoingContext(newCtx)
	assert.True(t, ok)
	assert.Equal(t, 1, len(copiedMD["key"]))
	assert.Equal(t, "value", copiedMD["key"][0])
}

func BenchmarkCopyContext(b *testing.B) {
	var key testContextType = "key"
	var value testContextType = "value"
	ctx := context.WithValue(context.Background(), key, value)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CopyContext(ctx)
	}
}

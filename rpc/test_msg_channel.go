// Copyright 2023 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package rpc

import (
	"github.com/basecomplextech/baselibrary/async"
	"github.com/basecomplextech/baselibrary/collect/chans"
	"github.com/basecomplextech/baselibrary/status"
)

// TestMessageChannelFunc is a function that returns the next message.
type TestMessageChannelFunc[T any] func(ctx async.Context) (T, status.Status)

// TestMessageChannel returns a test message channel.
func TestMessageChannel[T any](next TestMessageChannelFunc[T]) MessageChannel[T] {
	return newTestMessageChannel(next)
}

// internal

var _ MessageChannel[any] = (*testMessageChannel[any])(nil)

type testMessageChannel[T any] struct {
	next TestMessageChannelFunc[T]
}

func newTestMessageChannel[T any](next TestMessageChannelFunc[T]) *testMessageChannel[T] {
	return &testMessageChannel[T]{next: next}
}

// Receive returns the next message or blocks until a message is available.
func (ch *testMessageChannel[T]) Receive(ctx async.Context) (T, status.Status) {
	return ch.next(ctx)
}

// ReceiveAsync returns the next message or false if no message is available.
func (ch *testMessageChannel[T]) ReceiveAsync(ctx async.Context) (T, bool, status.Status) {
	msg, st := ch.next(ctx)
	if !st.OK() {
		return msg, false, st
	}
	return msg, true, status.OK
}

// ReceiveWait waits for the next message.
func (ch *testMessageChannel[T]) ReceiveWait() <-chan struct{} {
	return chans.Closed()
}

// Response receives the response.
func (ch *testMessageChannel[T]) Response(ctx async.Context) status.Status {
	return status.OK
}

// Free frees the channel.
func (ch *testMessageChannel[T]) Free() {}

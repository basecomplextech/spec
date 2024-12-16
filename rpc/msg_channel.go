// Copyright 2023 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package rpc

import (
	"github.com/basecomplextech/baselibrary/async"
	"github.com/basecomplextech/baselibrary/status"
)

// MessageChannel is a client message channel.
type MessageChannel[T any] interface {
	// Receive returns the next message or blocks until a message is available.
	Receive(ctx async.Context) (T, status.Status)

	// ReceiveAsync returns the next message or false if no message is available.
	ReceiveAsync(ctx async.Context) (T, bool, status.Status)

	// ReceiveWait waits for the next message.
	ReceiveWait() <-chan struct{}

	// Response receives the response.
	Response(ctx async.Context) status.Status

	// Free frees the channel.
	Free()
}

// Copyright 2023 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package mpx

import (
	"github.com/basecomplextech/baselibrary/status"
)

// Handler is a server channel handler.
type Handler interface {
	// HandleChannel handles an incoming channel.
	HandleChannel(ctx Context, ch Channel) status.Status
}

// HandleFunc is a type adapter to allow use of ordinary functions as channel handlers.
type HandleFunc func(ctx Context, ch Channel) status.Status

// HandleChannel handles an incoming channel.
func (f HandleFunc) HandleChannel(ctx Context, ch Channel) status.Status {
	return f(ctx, ch)
}

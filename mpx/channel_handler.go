// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package mpx

import (
	"github.com/basecomplextech/baselibrary/async"
	"github.com/basecomplextech/baselibrary/pools"
	"github.com/basecomplextech/baselibrary/status"
)

var _ async.Runner = (*channelHandler)(nil)

type channelHandler struct {
	c  *conn
	ch *channel
}

func newChannelHandler(c *conn, ch *channel) *channelHandler {
	h := acquireChannelHandler()
	h.c = c
	h.ch = ch
	return h
}

func (h *channelHandler) Run() {
	defer releaseChannelHandler(h)

	// No need to use async.Go here, because we don't need the result/cancellation,
	// and recover panics manually.
	defer func() {
		if e := recover(); e != nil {
			st := status.Recover(e)
			h.c.logger.ErrorStatus("Channel panic", st)
		}
	}()
	defer h.ch.Free()

	// Handle channel
	ctx := h.ch.Context()
	st := h.c.handler.HandleChannel(ctx, h.ch)
	switch st.Code {
	case status.CodeOK,
		status.CodeCancelled,
		status.CodeEnd,
		status.CodeClosed:
		return
	}

	// Log errors
	h.c.logger.ErrorStatus("Channel error", st)
}

// pool

var channelHandlerPool = pools.NewPoolFunc(
	func() *channelHandler {
		return &channelHandler{}
	},
)

func acquireChannelHandler() *channelHandler {
	return channelHandlerPool.New()
}

func releaseChannelHandler(h *channelHandler) {
	*h = channelHandler{}
	channelHandlerPool.Put(h)
}

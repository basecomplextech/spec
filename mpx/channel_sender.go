// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package mpx

import (
	"github.com/basecomplextech/baselibrary/alloc"
	"github.com/basecomplextech/baselibrary/async"
	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/spec/proto/pmpx"
)

type channelSender struct {
	ch   *channelState
	conn internalConn
}

func newChanSender(ch *channelState, conn internalConn) channelSender {
	return channelSender{
		ch:   ch,
		conn: conn,
	}
}

// open

func (s channelSender) sendOpen(ctx async.Context, data []byte) status.Status {
	// Build batch
	buf := alloc.AcquireBuffer()
	defer buf.Free()

	b := pmpx.NewBatchBuilder(buf)
	b, err := b.Open(s.ch.id, s.ch.initWindow)
	if err != nil {
		return mpxError(err)
	}
	b, err = b.Data(s.ch.id, data)
	if err != nil {
		return mpxError(err)
	}
	msg, err := b.Build()
	if err != nil {
		return mpxError(err)
	}

	// Send message
	if st := s.conn.send(ctx, msg); !st.OK() {
		return st
	}
	return status.OK
}

func (s channelSender) sendOpenData(ctx async.Context, data []byte) status.Status {
	// Build batch
	buf := alloc.AcquireBuffer()
	defer buf.Free()

	b := pmpx.NewBatchBuilder(buf)
	b, err := b.Open(s.ch.id, s.ch.initWindow)
	if err != nil {
		return mpxError(err)
	}
	b, err = b.Data(s.ch.id, data)
	if err != nil {
		return mpxError(err)
	}
	msg, err := b.Build()
	if err != nil {
		return mpxError(err)
	}

	// Send message
	return s.conn.send(ctx, msg)
}

func (s channelSender) sendOpenDataClose(ctx async.Context, data []byte) status.Status {
	// Build batch
	buf := alloc.AcquireBuffer()
	defer buf.Free()

	b := pmpx.NewBatchBuilder(buf)
	b, err := b.Open(s.ch.id, s.ch.initWindow)
	if err != nil {
		return mpxError(err)
	}
	b, err = b.Data(s.ch.id, data)
	if err != nil {
		return mpxError(err)
	}
	b, err = b.Close(s.ch.id)
	if err != nil {
		return mpxError(err)
	}
	msg, err := b.Build()
	if err != nil {
		return mpxError(err)
	}

	// Send message
	return s.conn.send(ctx, msg)
}

// close

func (s channelSender) sendClose(ctx async.Context) status.Status {
	// Build message
	buf := alloc.AcquireBuffer()
	defer buf.Free()

	w := pmpx.NewMessageWriterBuffer(buf)
	msg, err := pmpx.BuildChannelClose(w, s.ch.id, nil)
	if err != nil {
		return mpxError(err)
	}

	// Send message
	return s.conn.send(ctx, msg)
}

// data

func (s channelSender) sendData(ctx async.Context, data []byte) status.Status {
	// Build message
	buf := alloc.AcquireBuffer()
	defer buf.Free()

	w := pmpx.NewMessageWriterBuffer(buf)
	msg, err := pmpx.BuildChannelData(w, s.ch.id, data)
	if err != nil {
		return mpxError(err)
	}

	// Send message
	return s.conn.send(ctx, msg)
}

func (s channelSender) sendDataClose(ctx async.Context, data []byte) status.Status {
	// Build batch
	buf := alloc.AcquireBuffer()
	defer buf.Free()

	b := pmpx.NewBatchBuilder(buf)
	b, err := b.Data(s.ch.id, data)
	if err != nil {
		return mpxError(err)
	}
	b, err = b.Close(s.ch.id)
	if err != nil {
		return mpxError(err)
	}
	msg, err := b.Build()
	if err != nil {
		return mpxError(err)
	}

	// Send message
	return s.conn.send(ctx, msg)
}

// window

func (s channelSender) sendWindow(ctx async.Context, delta int32) status.Status {
	// Build message
	buf := alloc.AcquireBuffer()
	defer buf.Free()

	w := pmpx.NewMessageWriterBuffer(buf)
	msg, err := pmpx.BuildChannelWindow(w, s.ch.id, delta)
	if err != nil {
		return mpxError(err)
	}

	// Send message
	return s.conn.send(ctx, msg)
}

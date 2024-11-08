// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package mpx

import (
	"math"
	"sync"
	"sync/atomic"

	"github.com/basecomplextech/baselibrary/alloc"
	"github.com/basecomplextech/baselibrary/async"
	"github.com/basecomplextech/baselibrary/bin"
	"github.com/basecomplextech/baselibrary/pools"
	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/spec/proto/pmpx"
)

type channelState2 struct {
	id   bin.Bin128
	ctx  Context
	conn internalConn

	client     bool  // client or server
	initWindow int32 // initial window size

	opened     atomic.Bool
	closed     atomic.Bool
	closedUser atomic.Bool // close user once

	sendMu         sync.Mutex
	sender         chanSender
	sendWindow     atomic.Int32  // remaining send window, can become negative on sending large messages
	sendWindowWait chan struct{} // wait for send window increment

	recvMu     sync.Mutex
	recvQueue  alloc.ByteQueue // data queue
	recvWindow atomic.Int32    // remaining recv window, can become negative on receiving large messages
}

func newChannelState2(conn internalConn, client bool, id bin.Bin128, window int32) *channelState2 {
	s := channelStatePool.New()
	s.id = id
	s.ctx = newContext(nil) // TODO: conn
	s.conn = conn

	s.client = client
	s.initWindow = window

	s.sender = newChanSender(s, conn)
	s.sendWindow.Store(window)
	s.recvWindow.Store(window)
	return s
}

func openChannelState2(conn internalConn, client bool, msg pmpx.ChannelOpen) *channelState2 {
	id := msg.Id()
	window := msg.Window()

	s := channelStatePool.New()
	s.id = id
	s.ctx = newContext(nil) // TODO: conn
	s.conn = conn

	s.client = client
	s.initWindow = window
	s.opened.Store(true)

	s.sender = newChanSender(s, conn)
	s.sendWindow.Store(window)
	s.recvWindow.Store(window)
	return s
}

// open/close

func (s *channelState2) open() {
	if s.opened.Load() {
		return
	}
	if !s.client {
		panic("cannot open server channel")
	}

	s.opened.Store(true)
}

func (s *channelState2) close() {
	// Try to close channel
	closed := s.closed.CompareAndSwap(false, true)
	if !closed {
		return
	}

	// Cancel context, close receive queue
	s.ctx.Cancel()
	s.recvQueue.Close()
}

// receive

func (s *channelState2) receiveMessage(msg pmpx.Message, insideBatch bool) status.Status {
	code := msg.Code()

	switch code {
	case pmpx.Code_ChannelOpen:
		if !insideBatch {
			panic("open channel message must be handled by connection")
		}
		return status.OK
	case pmpx.Code_ChannelClose:
		return s.receiveClose(msg.ChannelClose())
	case pmpx.Code_ChannelData:
		return s.receiveData(msg.ChannelData())
	case pmpx.Code_ChannelWindow:
		return s.receiveWindow(msg.ChannelWindow())
	case pmpx.Code_ChannelBatch:
		if insideBatch {
			return mpxErrorf("nested channel batch messages are not allowed")
		}
		return s.receiveBatch(msg.ChannelBatch())
	}

	return mpxErrorf("unsupported channel message, code=%v", code)
}

func (s *channelState2) receiveClose(_ pmpx.ChannelClose) status.Status {
	s.close()
	return status.OK
}

func (s *channelState2) receiveData(msg pmpx.ChannelData) status.Status {
	data := msg.Data()
	_, _ = s.recvQueue.Write(data) // ignore end and false, receive queues are unbounded
	return status.OK
}

func (s *channelState2) receiveWindow(msg pmpx.ChannelWindow) status.Status {
	// Increment send window
	delta := msg.Delta()
	s.sendWindow.Add(delta)

	// Notify waiters
	select {
	case s.sendWindowWait <- struct{}{}:
	default:
	}
	return status.OK
}

func (s *channelState2) receiveBatch(msg pmpx.ChannelBatch) status.Status {
	list := msg.List()
	num := list.Len()

	for i := 0; i < num; i++ {
		m1, err := list.GetErr(i)
		if err != nil {
			return status.WrapError(err)
		}
		if st := s.receiveMessage(m1, true /* inside batch */); !st.OK() {
			return st
		}
	}
	return status.OK
}

// window

func (s *channelState2) incrementRecvWindow() int32 {
	window := s.recvWindow.Load()
	if window > s.initWindow/2 {
		return 0
	}

	// Increment recv window
	delta := s.initWindow - window
	s.recvWindow.Add(delta)
	return delta
}

func (s *channelState2) decrementSendWindow(ctx async.Context, data []byte) status.Status {
	// Check data size
	n := len(data)
	if n > math.MaxInt32 {
		return mpxErrorf("message too large, size=%d", n)
	}
	size := int32(n)

	for {
		// Decrement send window for normal small messages
		window := s.sendWindow.Load()
		if window >= int32(size) {
			s.sendWindow.Add(-size)
			return status.OK
		}

		// Decrement send window for large messages, when the remaining window
		// is greater than the half of the initial window, but the message size
		// still exceeds it.
		if window >= s.initWindow/2 {
			s.sendWindow.Add(-size)
			return status.OK
		}

		// Wait for send window increment
		select {
		case <-ctx.Wait():
			return ctx.Status()
		case <-s.ctx.Wait():
			return statusChannelClosed
		case <-s.sendWindowWait:
		}
	}
}

// reset

func (s *channelState2) reset() {
	// Free context
	if s.ctx != nil {
		s.ctx.Free()
		s.ctx = nil
	}

	// Clear wait channel
	sendWindowWait := s.sendWindowWait
	select {
	case <-sendWindowWait:
	default:
	}

	// Reset receive queue
	recvQueue := s.recvQueue
	recvQueue.Reset()

	// Reset state
	*s = channelState2{}
	s.sendWindowWait = sendWindowWait
	s.recvQueue = recvQueue
}

// pool

var channelStatePool = pools.NewPoolFunc(
	func() *channelState2 {
		return &channelState2{
			sendWindowWait: make(chan struct{}, 1),
			recvQueue:      alloc.NewByteQueue(),
		}
	},
)

func releaseChannelState2(s *channelState2) {
	s.reset()
	channelStatePool.Put(s)
}

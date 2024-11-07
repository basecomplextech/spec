// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package mpx

import (
	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/spec/proto/pmpx"
)

var _ internalChan = (*channel2)(nil)

type internalChan interface {
	// close is called by the connection to close the channel.
	close(st status.Status)

	// receive is called by the connection to receive a message.
	receive(msg pmpx.Message) status.Status

	// release is called by the connection to release removed channel.
	release()
}

// internal

// close is called by the connection to close the channel.
func (ch *channel2) close(st status.Status) {
	return
}

func (ch *channel2) receive(msg pmpx.Message) status.Status {
	ch.recvMu.Lock()
	defer ch.recvMu.Unlock()

	// Ignore messages if closed
	if ch.closed.Load() {
		return status.OK
	}

	// Receive message
	return ch.receiveMessage(msg, false /* not inside batch */)
}

// release is called by the connection to release removed channel.
func (ch *channel2) release() {

}

// private

func (ch *channel2) receiveMessage(msg pmpx.Message, insideBatch bool) status.Status {
	// ChannelOpen handled by connection
	code := msg.Code()

	switch code {
	case pmpx.Code_ChannelClose:
		return ch.receiveClose(msg.ChannelClose())
	case pmpx.Code_ChannelData:
		return ch.receiveData(msg.ChannelData())
	case pmpx.Code_ChannelWindow:
		return ch.receiveWindow(msg.ChannelWindow())
	case pmpx.Code_ChannelBatch:
		if insideBatch {
			return mpxErrorf("nested channel batch messages are not allowed")
		}
		return ch.receiveBatch(msg.ChannelBatch())
	}

	return mpxErrorf("unsupported channel message, code=%v", code)
}

func (ch *channel2) receiveClose(_ pmpx.ChannelClose) status.Status {
	closed := ch.closed.CompareAndSwap(false, true)
	if !closed {
		return status.OK
	}

	// Cancel context, close receive queue
	ch.ctx.Cancel()
	ch.recvQueue.Close()
	return status.OK
}

func (ch *channel2) receiveData(msg pmpx.ChannelData) status.Status {
	data := msg.Data()
	_, _ = ch.recvQueue.Write(data) // ignore end and false, receive queues are unbounded
	return status.OK
}

func (ch *channel2) receiveWindow(msg pmpx.ChannelWindow) status.Status {
	// Increment send window
	delta := msg.Delta()
	ch.sendWindow.Add(delta)

	// Notify waiters
	select {
	case ch.sendWindowWait <- struct{}{}:
	default:
	}
	return status.OK
}

func (ch *channel2) receiveBatch(msg pmpx.ChannelBatch) status.Status {
	list := msg.List()
	num := list.Len()

	for i := 0; i < num; i++ {
		m1, err := list.GetErr(i)
		if err != nil {
			return status.WrapError(err)
		}
		if st := ch.receiveMessage(m1, true /* inside batch */); !st.OK() {
			return st
		}
	}
	return status.OK
}

// window

func (ch *channel2) incrementRecvWindow() int32 {
	window := ch.recvWindow.Load()
	if window > ch.initWindow/2 {
		return 0
	}

	// Increment recv window
	delta := ch.initWindow - window
	ch.recvWindow.Add(delta)
	return delta
}

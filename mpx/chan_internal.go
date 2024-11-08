// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package mpx

import (
	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/spec/proto/pmpx"
)

type internalChan interface {
	// closeConn is called by the connection to close the channel.
	closeConn(st status.Status)

	// receiveConn is called by the connection to receive a message.
	receiveConn(msg pmpx.Message) status.Status

	// releaseConn is called by the connection to release removed channel.
	releaseConn()
}

// internal

var _ internalChan = (*channel2)(nil)

// closeConn is called by the connection to close the channel.
func (ch *channel2) closeConn(st status.Status) {
	ch.close()
}

// receiveConn is called by the connection to receive a message.
func (ch *channel2) receiveConn(msg pmpx.Message) status.Status {
	ch.recvMu.Lock()
	defer ch.recvMu.Unlock()

	// Ignore messages if closed
	if ch.closed.Load() {
		return status.OK
	}

	// Receive message
	return ch.receiveMessage(msg, false /* not inside batch */)
}

// releaseConn is called by the connection to release removed channel.
func (ch *channel2) releaseConn() {

}

// private

func (ch *channel2) receiveMessage(msg pmpx.Message, insideBatch bool) status.Status {
	// ChannelOpen handled by connection
	code := msg.Code()

	switch code {
	case pmpx.Code_ChannelOpen:
		if !insideBatch {
			panic("open channel message must be handled by connection")
		}
		return status.OK

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
	ch.close()
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

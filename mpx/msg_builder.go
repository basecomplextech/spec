// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package mpx

import (
	"github.com/basecomplextech/baselibrary/alloc"
	"github.com/basecomplextech/baselibrary/bin"
	"github.com/basecomplextech/spec/proto/pmpx"
)

type builder struct{}

type messageInput struct {
	id     bin.Bin128
	data   []byte
	window int32

	open  bool
	close bool
}

func newBuilder() builder {
	return builder{}
}

func (b builder) buildMessage(buf alloc.Buffer, input messageInput) (pmpx.Message, error) {
	w := pmpx.NewMessageWriterBuffer(buf)

	id := input.id
	data := input.data
	window := input.window

	open := input.open
	close := input.close
	if open && close {
		panic("open and close cannot be true at the same time")
	}

	switch {
	// Open message
	case open:
		w.Code(pmpx.Code_ChannelOpen)

		w1 := w.ChannelOpen()
		w1.Id(id)
		w1.Window(window)

		if data != nil {
			w1.Data(data)
		}

		if err := w1.End(); err != nil {
			return pmpx.Message{}, err
		}
		return w.Build()

	// Close message
	case close:
		w.Code(pmpx.Code_ChannelClose)

		w1 := w.ChannelClose()
		w1.Id(id)
		if data != nil {
			w1.Data(data)
		}

		if err := w1.End(); err != nil {
			return pmpx.Message{}, err
		}
		return w.Build()

	// Data message
	default:
		w.Code(pmpx.Code_ChannelData)

		w1 := w.ChannelData()
		w1.Id(id)
		w1.Data(data)
		if err := w1.End(); err != nil {
			return pmpx.Message{}, err
		}
		return w.Build()
	}
}

func (b builder) buildWindow(buf alloc.Buffer, id bin.Bin128, delta int32) (pmpx.Message, error) {
	w := pmpx.NewMessageWriterBuffer(buf)
	w.Code(pmpx.Code_ChannelWindow)

	w1 := w.ChannelWindow()
	w1.Id(id)
	w1.Delta(delta)
	if err := w1.End(); err != nil {
		return pmpx.Message{}, err
	}
	return w.Build()
}

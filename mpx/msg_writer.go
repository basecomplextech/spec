// Copyright 2023 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package mpx

import (
	"bufio"
	"encoding/binary"
	"io"
	"strings"

	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/spec/proto/pmpx"
	"github.com/pierrec/lz4/v4"
)

type writer struct {
	dst  *bufio.Writer
	comp *lz4.Writer // Nil when no compression

	client bool
	head   [4]byte
	writer writerFlusher // Points to dst or comp
}

type writerFlusher interface {
	io.Writer
	Flush() error
}

func newWriter(w io.Writer, client bool, bufferSize int) *writer {
	dst := bufio.NewWriterSize(w, bufferSize)
	return &writer{
		dst:    dst,
		client: client,
		writer: dst,
	}
}

func (w *writer) initLZ4() status.Status {
	if w.comp != nil {
		return status.OK
	}

	w.comp = lz4.NewWriter(w.dst)
	err := w.comp.Apply(lz4.BlockSizeOption(lz4.Block1Mb))
	if err != nil {
		return mpxError(err)
	}

	w.writer = w.comp
	return status.OK
}

// flush

func (w *writer) flush() status.Status {
	if err := w.writer.Flush(); err != nil {
		return mpxError(err)
	}
	if err := w.dst.Flush(); err != nil {
		return mpxError(err)
	}
	return status.OK
}

// write

// writeLine writes a single string, used only to write the protocol line.
func (w *writer) writeLine(s string) status.Status {
	if _, err := w.dst.WriteString(s); err != nil {
		return mpxError(err)
	}
	if debug {
		debugPrint(w.client, "-> line\t", strings.TrimSpace(s))
	}
	return status.OK
}

// write writes a message, prefixed with its size.
func (w *writer) write(msg pmpx.Message) status.Status {
	b := msg.Unwrap().Raw()
	head := w.head[:]

	// Write size
	binary.BigEndian.PutUint32(head, uint32(len(b)))
	if _, err := w.writer.Write(head); err != nil {
		return mpxError(err)
	}

	// Write message
	if _, err := w.writer.Write(b); err != nil {
		return mpxError(err)
	}

	if debug {
		code := msg.Code()
		switch code {
		case pmpx.Code_ConnectRequest:
			debugPrint(w.client, "-> connect_req")

		case pmpx.Code_ConnectResponse:
			debugPrint(w.client, "-> connect_resp")

		case pmpx.Code_Batch:
			m := msg.Batch()
			list := m.List()
			num := list.Len()
			codes := make([]string, 0, num)

			for i := 0; i < num; i++ {
				m1 := list.Get(i)
				c1 := m1.Code().String()
				codes = append(codes, c1)
			}
			debugPrint(w.client, "-> batch\t", num, codes)

		case pmpx.Code_ChannelOpen:
			m := msg.ChannelOpen()
			id := m.Id()
			data := debugString(m.Data())
			cmd := "-> channel_open\t"
			debugPrint(w.client, cmd, id, data)

		case pmpx.Code_ChannelClose:
			m := msg.ChannelClose()
			id := m.Id()
			data := debugString(m.Data())
			debugPrint(w.client, "-> channel_close\t", id, data)

		case pmpx.Code_ChannelData:
			m := msg.ChannelData()
			id := m.Id()
			data := debugString(m.Data())
			debugPrint(w.client, "-> channel_data\t", id, data)

		case pmpx.Code_ChannelWindow:
			m := msg.ChannelWindow()
			id := m.Id()
			delta := m.Delta()
			debugPrint(w.client, "-> channel_window\t", id, delta)

		default:
			debugPrint(w.client, "-> unknown", code)
		}
	}
	return status.OK
}

// writeAndFlush writes a message and flushes the buffer.
func (w *writer) writeAndFlush(msg pmpx.Message) status.Status {
	if st := w.write(msg); !st.OK() {
		return st
	}
	return w.flush()
}

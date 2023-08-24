package tcp

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"sync"

	"github.com/basecomplextech/baselibrary/alloc"
	"github.com/basecomplextech/baselibrary/bin"
	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/spec"
	"github.com/basecomplextech/spec/proto/ptcp"
)

type writer struct {
	w *bufio.Writer

	mu   sync.Mutex
	head [4]byte

	buf    *alloc.Buffer
	writer spec.Writer
}

func newWriter(w io.Writer) *writer {
	buf := alloc.NewBuffer()

	return &writer{
		w: bufio.NewWriterSize(w, writeBufferSize),

		buf:    buf,
		writer: spec.NewWriterBuffer(buf),
	}
}

func (w *writer) free() {
	w.buf.Free()
	w.buf = nil
}

func (w *writer) flush() status.Status {
	w.mu.Lock()
	defer w.mu.Unlock()

	if err := w.w.Flush(); err != nil {
		return tcpError(err)
	}
	return status.OK
}

// write

func (w *writer) writeOpenStream(id bin.Bin128) status.Status {
	w.mu.Lock()
	defer w.mu.Unlock()

	msg, st := w._openStream(id)
	if !st.OK() {
		return st
	}

	b := msg.Unwrap().Raw()
	if debug {
		fmt.Println("-> open_stream", id)
	}
	return w._write(b)
}

func (w *writer) writeCloseStream(id bin.Bin128) status.Status {
	w.mu.Lock()
	defer w.mu.Unlock()

	msg, st := w._closeStream(id)
	if !st.OK() {
		return st
	}

	if debug {
		fmt.Println("-> close_stream", id)
	}
	b := msg.Unwrap().Raw()
	return w._write(b)
}

func (w *writer) writeStreamMessage(id bin.Bin128, data []byte) status.Status {
	w.mu.Lock()
	defer w.mu.Unlock()

	msg, st := w._streamMessage(id, data)
	if !st.OK() {
		return st
	}

	if debug {
		fmt.Println("-> stream_message", id)
	}
	b := msg.Unwrap().Raw()
	return w._write(b)
}

func (w *writer) _write(msg []byte) status.Status {
	head := w.head[:]

	// Write size
	binary.BigEndian.PutUint32(head, uint32(len(msg)))
	if _, err := w.w.Write(head); err != nil {
		return tcpError(err)
	}

	// Write message
	if _, err := w.w.Write(msg); err != nil {
		return tcpError(err)
	}
	return status.OK
}

// build

func (w *writer) _openStream(id bin.Bin128) (msg ptcp.Message, st status.Status) {
	w.buf.Reset()
	w.writer.Reset(w.buf)

	w0 := ptcp.NewMessageWriterTo(w.writer.Message())
	w0.Code(ptcp.Code_OpenStream)

	w1 := w0.Open()
	w1.Id(id)
	if err := w1.End(); err != nil {
		return msg, tcpError(err)
	}

	var err error
	msg, err = w0.Build()
	if err != nil {
		return msg, tcpError(err)
	}
	return msg, status.OK
}

func (w *writer) _closeStream(id bin.Bin128) (msg ptcp.Message, st status.Status) {
	w.buf.Reset()
	w.writer.Reset(w.buf)

	w0 := ptcp.NewMessageWriterTo(w.writer.Message())
	w0.Code(ptcp.Code_CloseStream)

	w1 := w0.Open()
	w1.Id(id)
	if err := w1.End(); err != nil {
		return msg, tcpError(err)
	}

	var err error
	msg, err = w0.Build()
	if err != nil {
		return msg, tcpError(err)
	}
	return msg, status.OK
}

func (w *writer) _streamMessage(id bin.Bin128, data []byte) (msg ptcp.Message, st status.Status) {
	w.buf.Reset()
	w.writer.Reset(w.buf)

	w0 := ptcp.NewMessageWriterTo(w.writer.Message())
	w0.Code(ptcp.Code_StreamMessage)

	w1 := w0.Message()
	w1.Id(id)
	w1.Data(data)
	if err := w1.End(); err != nil {
		return msg, tcpError(err)
	}

	var err error
	msg, err = w0.Build()
	if err != nil {
		return msg, tcpError(err)
	}
	return msg, status.OK
}

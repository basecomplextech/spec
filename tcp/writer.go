package tcp

import (
	"bufio"
	"encoding/binary"
	"io"

	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/spec/proto/ptcp"
)

type writer struct {
	w      *bufio.Writer
	client bool
	head   [4]byte
}

func newWriter(w io.Writer, client bool) *writer {
	return &writer{
		w:      bufio.NewWriterSize(w, writeBufferSize),
		client: client,
	}
}

func (w *writer) flush() status.Status {
	if err := w.w.Flush(); err != nil {
		return tcpError(err)
	}
	return status.OK
}

func (w *writer) write(msg []byte) status.Status {
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

	if debug {
		m := ptcp.NewMessage(msg)
		code := m.Code()
		switch code {
		case ptcp.Code_NewStream:
			debugPrint(w.client, "-> stream\t", m.New().Id())
		case ptcp.Code_CloseStream:
			debugPrint(w.client, "-> close\t", m.Close().Id())
		case ptcp.Code_StreamMessage:
			debugPrint(w.client, "-> message\t", m.Message().Id())
		default:
			debugPrint(w.client, "-> unknown", code)
		}
	}
	return status.OK
}

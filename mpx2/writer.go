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

func (w *writer) flush() status.Status {
	if err := w.writer.Flush(); err != nil {
		return mpxError(err)
	}
	if err := w.dst.Flush(); err != nil {
		return mpxError(err)
	}
	return status.OK
}

// writeString writes a single string, used only to write the protocol line.
func (w *writer) writeString(s string) status.Status {
	if _, err := w.dst.WriteString(s); err != nil {
		return mpxError(err)
	}
	if debug {
		debugPrint(w.client, "-> line\t", strings.TrimSpace(s))
	}
	return status.OK
}

// writeRequest writes a connect request and flushes the writer.
func (w *writer) writeRequest(req pmpx.ConnectRequest) status.Status {
	msg := req.Unwrap().Raw()
	head := w.head[:]

	// Write size
	binary.BigEndian.PutUint32(head, uint32(len(msg)))
	if _, err := w.writer.Write(head); err != nil {
		return mpxError(err)
	}

	// Write message
	if _, err := w.writer.Write(msg); err != nil {
		return mpxError(err)
	}

	if debug {
		debugPrint(w.client, "-> connect req")
	}
	return w.flush()
}

// writeRequest writes a connect response and flushes the writer.
func (w *writer) writeResponse(resp pmpx.ConnectResponse) status.Status {
	msg := resp.Unwrap().Raw()
	head := w.head[:]

	// Write size
	binary.BigEndian.PutUint32(head, uint32(len(msg)))
	if _, err := w.writer.Write(head); err != nil {
		return mpxError(err)
	}

	// Write message
	if _, err := w.writer.Write(msg); err != nil {
		return mpxError(err)
	}

	if debug {
		debugPrint(w.client, "-> connect resp")
	}
	return w.flush()
}

// writeMessage writes a single message.
func (w *writer) writeMessage(msg []byte) status.Status {
	head := w.head[:]

	// Write size
	binary.BigEndian.PutUint32(head, uint32(len(msg)))
	if _, err := w.writer.Write(head); err != nil {
		return mpxError(err)
	}

	// Write message
	if _, err := w.writer.Write(msg); err != nil {
		return mpxError(err)
	}

	if debug {
		m := pmpx.NewMessage(msg)
		code := m.Code()
		switch code {
		case pmpx.Code_ChannelOpen:
			debugPrint(w.client, "-> open\t", m.Open().Id())
		case pmpx.Code_ChannelClose:
			debugPrint(w.client, "-> close\t", m.Close().Id())
		case pmpx.Code_ChannelMessage:
			debugPrint(w.client, "-> message\t", m.Message().Id())
		case pmpx.Code_ChannelWindow:
			debugPrint(w.client, "-> window\t", m.Window().Id(), m.Window().Delta())
		default:
			debugPrint(w.client, "-> unknown", code)
		}
	}
	return status.OK
}

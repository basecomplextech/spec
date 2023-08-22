package tcp

import (
	"bufio"
	"encoding/binary"
	"io"
	"sync"

	"github.com/basecomplextech/baselibrary/status"
)

const writeBufferSize = 4096

type writer struct {
	w *bufio.Writer

	mu   sync.Mutex
	head [4]byte
}

func newWriter(w io.Writer) *writer {
	return &writer{
		w: bufio.NewWriterSize(w, writeBufferSize),
	}
}

func (w *writer) flush() status.Status {
	w.mu.Lock()
	defer w.mu.Unlock()

	err := w.w.Flush()
	return tcpError(err)
}

func (w *writer) write(msg []byte) status.Status {
	w.mu.Lock()
	defer w.mu.Unlock()

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

func (w *writer) writeError() status.Status {
	return status.OK
}

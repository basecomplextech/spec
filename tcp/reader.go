package tcp

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"sync"

	"github.com/basecomplextech/baselibrary/alloc"
	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/spec/proto/ptcp"
)

const readBufferSize = 4096

type reader struct {
	r *bufio.Reader

	mu   sync.Mutex
	head [4]byte
	buf  *alloc.Buffer
}

func newReader(r io.Reader) *reader {
	return &reader{
		r:   bufio.NewReaderSize(r, readBufferSize),
		buf: alloc.NewBuffer(),
	}
}

func (r *reader) free() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.buf.Free()
	r.buf = nil
}

// read reads the next message, the message is valid until the next read.
func (r *reader) read() (ptcp.Message, status.Status) {
	r.mu.Lock()
	defer r.mu.Unlock()

	head := r.head[:]

	// Read size
	if _, err := io.ReadFull(r.r, head); err != nil {
		return ptcp.Message{}, tcpError(err)
	}
	size := binary.BigEndian.Uint32(head)

	// Read message
	r.buf.Reset()
	buf := r.buf.Grow(int(size))
	if _, err := io.ReadFull(r.r, buf); err != nil {
		return ptcp.Message{}, tcpError(err)
	}

	// Parse message
	msg, _, err := ptcp.ParseMessage(buf)
	if err != nil {
		return ptcp.Message{}, tcpError(err)
	}

	if debug {
		code := msg.Code()
		switch code {
		case ptcp.Code_OpenStream:
			fmt.Println("<- open_stream", msg.Open().Id())
		case ptcp.Code_CloseStream:
			fmt.Println("<- close_stream", msg.Close().Id())
		case ptcp.Code_StreamMessage:
			fmt.Println("<- stream_message", msg.Message().Id())
		default:
			fmt.Println("<- unknown", code)
		}
	}
	return msg, status.OK
}

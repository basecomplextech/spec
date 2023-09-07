package tcp

import (
	"bufio"
	"encoding/binary"
	"io"
	"sync"

	"github.com/basecomplextech/baselibrary/alloc"
	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/spec/proto/ptcp"
)

type reader struct {
	r      *bufio.Reader
	client bool
	freed  bool

	mu   sync.Mutex
	head [4]byte
	buf  *alloc.Buffer
}

func newReader(r io.Reader, client bool) *reader {
	return &reader{
		r:      bufio.NewReaderSize(r, readBufferSize),
		client: client,

		buf: alloc.NewBuffer(),
	}
}

func (r *reader) free() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.freed {
		return
	}
	r.freed = true

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
		case ptcp.Code_NewChannel:
			debugPrint(r.client, "<- open\t", msg.New().Id())
		case ptcp.Code_CloseChannel:
			debugPrint(r.client, "<- close\t", msg.Close().Id())
		case ptcp.Code_ChannelMessage:
			debugPrint(r.client, "<- message\t", msg.Message().Id())
		default:
			debugPrint(r.client, "<- unknown", code)
		}
	}
	return msg, status.OK
}

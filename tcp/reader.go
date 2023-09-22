package tcp

import (
	"bufio"
	"encoding/binary"
	"io"
	"strings"
	"sync"

	"github.com/basecomplextech/baselibrary/alloc"
	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/spec/proto/ptcp"
	"github.com/pierrec/lz4/v4"
)

type reader struct {
	src    *bufio.Reader
	comp   *lz4.Reader // Nil when no compression
	reader io.Reader   // Points to src or comp

	client bool
	freed  bool

	mu   sync.Mutex
	head [4]byte
	buf  *alloc.Buffer
}

func newReader(r io.Reader, client bool, bufferSize int) *reader {
	src := bufio.NewReaderSize(r, bufferSize)
	return &reader{
		src:    src,
		client: client,
		reader: src,

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

func (r *reader) initLZ4() status.Status {
	if r.comp != nil {
		return status.OK
	}

	r.comp = lz4.NewReader(r.src)
	r.reader = r.comp
	return status.OK
}

// readLine reads and returns a single line delimited by \n, includes the delimiter.
func (r *reader) readLine() (string, status.Status) {
	s, err := r.src.ReadString('\n')
	if err != nil {
		return "", tcpError(err)
	}

	if debug {
		debugPrint(r.client, "<- line\t", strings.TrimSpace(s))
	}
	return s, status.OK
}

// readRequest reads and parses a connect request, the message is valid until the next read call.
func (r *reader) readRequest() (ptcp.ConnectRequest, status.Status) {
	buf, st := r.read()
	if !st.OK() {
		return ptcp.ConnectRequest{}, st
	}

	req, _, err := ptcp.ParseConnectRequest(buf)
	if err != nil {
		return ptcp.ConnectRequest{}, tcpErrorf("failed to parse connect request: %v", err)
	}

	if debug {
		debugPrint(r.client, "<- connect req")
	}
	return req, status.OK
}

// readResponse reads and parses a connect response, the message is valid until the next read call.
func (r *reader) readResponse() (ptcp.ConnectResponse, status.Status) {
	buf, st := r.read()
	if !st.OK() {
		return ptcp.ConnectResponse{}, st
	}

	resp, _, err := ptcp.ParseConnectResponse(buf)
	if err != nil {
		return ptcp.ConnectResponse{}, tcpErrorf("failed to parse connect response: %v", err)
	}

	if debug {
		debugPrint(r.client, "<- connect resp")
	}
	return resp, status.OK
}

// readMessage reads and parses the next message, the message is valid until the next read call.
func (r *reader) readMessage() (ptcp.Message, status.Status) {
	buf, st := r.read()
	if !st.OK() {
		return ptcp.Message{}, st
	}

	// Parse message
	msg, _, err := ptcp.ParseMessage(buf)
	if err != nil {
		return ptcp.Message{}, tcpError(err)
	}

	if debug {
		code := msg.Code()
		switch code {
		case ptcp.Code_OpenChannel:
			debugPrint(r.client, "<- open\t", msg.Open().Id())
		case ptcp.Code_CloseChannel:
			debugPrint(r.client, "<- close\t", msg.Close().Id())
		case ptcp.Code_ChannelMessage:
			debugPrint(r.client, "<- message\t", msg.Message().Id())
		case ptcp.Code_ChannelWindow:
			debugPrint(r.client, "<- window\t", msg.Window().Id(), msg.Window().Delta())
		default:
			debugPrint(r.client, "<- unknown", code)
		}
	}
	return msg, status.OK
}

// read reads the next message bytes, the bytes are valid until the next read call.
func (r *reader) read() ([]byte, status.Status) {
	r.mu.Lock()
	defer r.mu.Unlock()

	head := r.head[:]

	// Read size
	if _, err := io.ReadFull(r.reader, head); err != nil {
		return nil, tcpError(err)
	}
	size := binary.BigEndian.Uint32(head)

	// Read bytes
	r.buf.Reset()
	buf := r.buf.Grow(int(size))
	if _, err := io.ReadFull(r.reader, buf); err != nil {
		return nil, tcpError(err)
	}
	return buf, status.OK
}

package mpx

import (
	"bufio"
	"encoding/binary"
	"io"
	"strings"
	"sync"

	"github.com/basecomplextech/baselibrary/alloc"
	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/spec/proto/pmpx"
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
		return "", mpxError(err)
	}

	if debug {
		debugPrint(r.client, "<- line\t", strings.TrimSpace(s))
	}
	return s, status.OK
}

// readRequest reads and parses a connect request, the message is valid until the next read call.
func (r *reader) readRequest() (pmpx.ConnectRequest, status.Status) {
	buf, st := r.read()
	if !st.OK() {
		return pmpx.ConnectRequest{}, st
	}

	req, _, err := pmpx.ParseConnectRequest(buf)
	if err != nil {
		return pmpx.ConnectRequest{}, mpxErrorf("failed to parse connect request: %v", err)
	}

	if debug {
		debugPrint(r.client, "<- connect req")
	}
	return req, status.OK
}

// readResponse reads and parses a connect response, the message is valid until the next read call.
func (r *reader) readResponse() (pmpx.ConnectResponse, status.Status) {
	buf, st := r.read()
	if !st.OK() {
		return pmpx.ConnectResponse{}, st
	}

	resp, _, err := pmpx.ParseConnectResponse(buf)
	if err != nil {
		return pmpx.ConnectResponse{}, mpxErrorf("failed to parse connect response: %v", err)
	}

	if debug {
		debugPrint(r.client, "<- connect resp")
	}
	return resp, status.OK
}

// readMessage reads and parses the next message, the message is valid until the next read call.
func (r *reader) readMessage() (pmpx.Message, status.Status) {
	buf, st := r.read()
	if !st.OK() {
		return pmpx.Message{}, st
	}

	// Parse message
	msg, _, err := pmpx.ParseMessage(buf)
	if err != nil {
		return pmpx.Message{}, mpxError(err)
	}

	if debug {
		code := msg.Code()
		switch code {
		case pmpx.Code_ChannelOpen:
			m := msg.Open()
			id := m.Id()
			data := string(m.Data())
			debugPrint(r.client, "<- open\t", id, data)

		case pmpx.Code_ChannelClose:
			m := msg.Close()
			id := m.Id()
			data := string(m.Data())
			debugPrint(r.client, "<- close\t", id, data)

		case pmpx.Code_ChannelEnd:
			m := msg.End_()
			id := m.Id()
			data := string(m.Data())
			debugPrint(r.client, "<- end\t", id, data)

		case pmpx.Code_ChannelMessage:
			m := msg.Message()
			id := m.Id()
			data := string(m.Data())
			debugPrint(r.client, "<- message\t", id, data)

		case pmpx.Code_ChannelWindow:
			m := msg.Window()
			id := m.Id()
			delta := m.Delta()
			debugPrint(r.client, "<- window\t", id, delta)

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
		return nil, mpxError(err)
	}
	size := binary.BigEndian.Uint32(head)

	// Read bytes
	r.buf.Reset()
	buf := r.buf.Grow(int(size))
	if _, err := io.ReadFull(r.reader, buf); err != nil {
		return nil, mpxError(err)
	}
	return buf, status.OK
}

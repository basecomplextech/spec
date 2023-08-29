package tcp

import (
	"sync"

	"github.com/basecomplextech/baselibrary/alloc"
	"github.com/basecomplextech/baselibrary/bin"
	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/spec"
	"github.com/basecomplextech/spec/proto/ptcp"
)

// Stream is a single stream in a TCP connection.
type Stream interface {
	// Read reads a message from the stream, the message is valid until the next iteration.
	Read(cancel <-chan struct{}) ([]byte, status.Status)

	// Write writes a message to the stream.
	Write(cancel <-chan struct{}, msg []byte) status.Status

	// Close closes the stream and sends the close message.
	Close() status.Status

	// Internal

	// Free closes the stream and releases its resources.
	Free()
}

// internal

var _ Stream = (*stream)(nil)

type stream struct {
	id     bin.Bin128
	conn   *conn
	client bool

	reader streamReader
	writer streamWriter

	mu       sync.RWMutex
	freed    bool
	closed   bool
	started  bool
	openSent bool
}

func openStream(id bin.Bin128, conn *conn) *stream {
	if debug {
		debugPrint(conn.client, "stream.open\t", id)
	}

	return &stream{
		id:     id,
		conn:   conn,
		client: conn.client,

		reader: newStreamReader(),
		writer: newStreamWriter(conn),

		started: true,
	}
}

func openedStream(id bin.Bin128, conn *conn) *stream {
	if debug {
		debugPrint(conn.client, "stream.opened\t", id)
	}

	return &stream{
		id:     id,
		conn:   conn,
		client: conn.client,

		reader: newStreamReader(),
		writer: newStreamWriter(conn),

		openSent: true,
		started:  false,
	}
}

// Read reads a message from the stream, the message is valid until the next iteration.
func (s *stream) Read(cancel <-chan struct{}) ([]byte, status.Status) {
	for {
		b, ok, st := s.reader.read()
		switch {
		case !st.OK():
			// End
			return nil, st

		case !ok:
			// Wait for next messages or end
			select {
			case <-cancel:
				return nil, status.Cancelled
			case <-s.reader.wait():
				continue
			}
		}

		if debug && debugStream {
			debugPrint(s.client, "stream.read\t", s.id, ok, st)
		}

		msg, _, err := ptcp.ParseMessage(b)
		if err != nil {
			s.close()
			return nil, tcpError(err)
		}

		code := msg.Code()
		switch code {
		case ptcp.Code_StreamMessage:
			data := msg.Message().Data()
			return data, status.OK

		case ptcp.Code_CloseStream:
			s.remoteClosed()
			return nil, status.End
		}

		return nil, tcpErrorf("unexpected message code %d", code)
	}
}

// Write writes a message to the stream.
func (s *stream) Write(cancel <-chan struct{}, msg []byte) status.Status {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return statusStreamClosed
	}

	if !s.openSent {
		if st := s.writer.writeOpen(cancel, s.id); !st.OK() {
			return st
		}
		s.openSent = true
	}

	if debug && debugStream {
		debugPrint(s.client, "stream.write\t", s.id)
	}
	return s.writer.writeMessage(cancel, s.id, msg)
}

// Close closes the stream and sends the close message.
func (s *stream) Close() status.Status {
	return s.close()
}

// Internal

// Free closes the stream and releases its resources.
func (s *stream) Free() {
	s.close()
	s.free()
}

// internal

// receive receives a message from the connection.
func (s *stream) receive(cancel <-chan struct{}, msg ptcp.Message) status.Status {
	b := msg.Unwrap().Raw()
	return s.reader.write(cancel, b)
}

// private

func (s *stream) free() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.freed {
		return
	}

	s.freed = true
	s.reader.free()
	s.writer.free()

	if debug {
		debugPrint(s.client, "stream.free\t", s.id)
	}
}

func (s *stream) close() status.Status {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return status.OK
	}

	s.closed = true
	s.reader.close()

	if debug {
		debugPrint(s.client, "stream.close\t", s.id)
	}
	return s.writer.writeClose(nil /* no cancel */, s.id)
}

func (s *stream) connClosed() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return
	}

	s.closed = true
	s.reader.close()

	if debug {
		debugPrint(s.client, "stream.conn-closed\t", s.id)
	}
}

func (s *stream) remoteClosed() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return
	}

	s.closed = true
	s.reader.close()

	if debug {
		debugPrint(s.client, "stream.remote-closed\t", s.id)
	}
}

// reader

type streamReader struct {
	mu     sync.Mutex
	queue  alloc.MQueue
	freed  bool
	closed bool
}

func newStreamReader() streamReader {
	return streamReader{
		queue: acquireQueue(),
	}
}

func (r *streamReader) free() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.freed {
		return
	}

	q := r.queue
	q.Close()

	r.queue = nil
	r.freed = true
	r.closed = true

	releaseQueue(q)
}

func (r *streamReader) close() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return
	}

	r.queue.Close()
	r.closed = true
}

func (r *streamReader) read() ([]byte, bool, status.Status) {
	q, st := r.get()
	if !st.OK() {
		return nil, false, st
	}
	return q.Read()
}

func (r *streamReader) write(cancel <-chan struct{}, msg []byte) status.Status {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return status.OK
	}

	for {
		ok, st := r.queue.Write(msg)
		switch {
		case !st.OK():
			return status.OK // ignore end
		case ok:
			return status.OK
		}

		// Wait for space
		select {
		case <-cancel:
			return status.Cancelled
		case <-r.queue.WriteWait(len(msg)):
			continue
		}
	}
}

func (r *streamReader) wait() <-chan struct{} {
	q, st := r.get()
	if !st.OK() {
		return closedChan
	}
	return q.ReadWait()
}

func (r *streamReader) get() (alloc.MQueue, status.Status) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return nil, status.End
	}

	q := r.queue
	return q, status.OK
}

// writer

type streamWriter struct {
	mu    sync.Mutex
	conn  *conn
	freed bool

	buf    *alloc.Buffer
	writer spec.Writer
}

func newStreamWriter(c *conn) streamWriter {
	buf := alloc.NewBuffer()

	return streamWriter{
		conn:   c,
		buf:    buf,
		writer: spec.NewWriterBuffer(buf),
	}
}

func (w *streamWriter) free() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.freed {
		return
	}
	w.freed = true

	w.writer.Free()
	w.writer = nil

	w.buf.Free()
	w.buf = nil
}

func (w *streamWriter) writeOpen(cancel <-chan struct{}, id bin.Bin128) status.Status {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.freed {
		return statusStreamClosed
	}

	var msg ptcp.Message
	{
		w.buf.Reset()
		w.writer.Reset(w.buf)

		w0 := ptcp.NewMessageWriterTo(w.writer.Message())
		w0.Code(ptcp.Code_OpenStream)

		w1 := w0.Open()
		w1.Id(id)
		if err := w1.End(); err != nil {
			return tcpError(err)
		}

		var err error
		msg, err = w0.Build()
		if err != nil {
			return tcpError(err)
		}
	}

	return w.conn.write(cancel, msg)
}

func (w *streamWriter) writeMessage(cancel <-chan struct{}, id bin.Bin128, data []byte) status.Status {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.freed {
		return statusStreamClosed
	}

	var msg ptcp.Message
	{
		w.buf.Reset()
		w.writer.Reset(w.buf)

		w0 := ptcp.NewMessageWriterTo(w.writer.Message())
		w0.Code(ptcp.Code_StreamMessage)

		w1 := w0.Message()
		w1.Id(id)
		w1.Data(data)
		if err := w1.End(); err != nil {
			return tcpError(err)
		}

		var err error
		msg, err = w0.Build()
		if err != nil {
			return tcpError(err)
		}
	}

	return w.conn.write(cancel, msg)
}

func (w *streamWriter) writeClose(cancel <-chan struct{}, id bin.Bin128) status.Status {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.freed {
		return statusStreamClosed
	}

	var msg ptcp.Message
	{
		w.buf.Reset()
		w.writer.Reset(w.buf)

		w0 := ptcp.NewMessageWriterTo(w.writer.Message())
		w0.Code(ptcp.Code_CloseStream)

		w1 := w0.Close()
		w1.Id(id)
		if err := w1.End(); err != nil {
			return tcpError(err)
		}

		var err error
		msg, err = w0.Build()
		if err != nil {
			return tcpError(err)
		}
	}

	return w.conn.write(cancel, msg)
}

// queue

var queuePool = &sync.Pool{}

func acquireQueue() alloc.MQueue {
	obj := queuePool.Get()
	if obj == nil {
		return alloc.NewMQueue()
	}
	return obj.(alloc.MQueue)
}

func releaseQueue(q alloc.MQueue) {
	q.Reset()
	queuePool.Put(q)
}

// closed chan

var closedChan = func() chan struct{} {
	ch := make(chan struct{})
	close(ch)
	return ch
}()

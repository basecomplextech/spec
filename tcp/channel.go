package tcp

import (
	"sync"

	"github.com/basecomplextech/baselibrary/alloc"
	"github.com/basecomplextech/baselibrary/bin"
	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/spec"
	"github.com/basecomplextech/spec/proto/ptcp"
)

// Channel is a single ch in a TCP connection.
type Channel interface {
	// Read reads a message from the ch, the message is valid until the next iteration.
	Read(cancel <-chan struct{}) ([]byte, status.Status)

	// Write writes a message to the ch.
	Write(cancel <-chan struct{}, msg []byte) status.Status

	// Close closes the ch and sends the close message.
	Close() status.Status

	// Internal

	// Free closes the ch and releases its resources.
	Free()
}

// internal

var _ Channel = (*channel)(nil)

type channel struct {
	id     bin.Bin128
	conn   *conn
	client bool

	reader channelReader
	writer channelWriter

	mu      sync.RWMutex
	freed   bool
	closed  bool
	started bool
	newSent bool
}

func openChannel(id bin.Bin128, conn *conn) *channel {
	if debug {
		debugPrint(conn.client, "channel.open\t", id)
	}

	return &channel{
		id:     id,
		conn:   conn,
		client: conn.client,

		reader: newChannelReader(),
		writer: newChannelWriter(conn),

		started: true,
	}
}

func openedChannel(id bin.Bin128, conn *conn) *channel {
	if debug {
		debugPrint(conn.client, "channel.opened\t", id)
	}

	return &channel{
		id:     id,
		conn:   conn,
		client: conn.client,

		reader: newChannelReader(),
		writer: newChannelWriter(conn),

		newSent: true,
		started: false,
	}
}

// Read reads a message from the ch, the message is valid until the next iteration.
func (ch *channel) Read(cancel <-chan struct{}) ([]byte, status.Status) {
	for {
		b, ok, st := ch.reader.read()
		switch {
		case !st.OK():
			// End
			return nil, st

		case !ok:
			// Wait for next messages or end
			select {
			case <-cancel:
				return nil, status.Cancelled
			case <-ch.reader.wait():
				continue
			}
		}

		if debug && debugChannel {
			debugPrint(ch.client, "ch.read\t", ch.id, ok, st)
		}

		msg, _, err := ptcp.ParseMessage(b)
		if err != nil {
			ch.close()
			return nil, tcpError(err)
		}

		code := msg.Code()
		switch code {
		case ptcp.Code_ChannelMessage:
			data := msg.Message().Data()
			return data, status.OK

		case ptcp.Code_CloseChannel:
			ch.remoteClosed()
			return nil, status.End
		}

		return nil, tcpErrorf("unexpected message code %d", code)
	}
}

// Write writes a message to the ch.
func (ch *channel) Write(cancel <-chan struct{}, msg []byte) status.Status {
	ch.mu.RLock()
	defer ch.mu.RUnlock()

	if ch.closed {
		return statusChannelClosed
	}

	if !ch.newSent {
		if st := ch.writer.writeNew(cancel, ch.id); !st.OK() {
			return st
		}
		ch.newSent = true
	}

	if debug && debugChannel {
		debugPrint(ch.client, "ch.write\t", ch.id)
	}
	return ch.writer.writeMessage(cancel, ch.id, msg)
}

// Close closes the ch and sends the close message.
func (ch *channel) Close() status.Status {
	return ch.close()
}

// Internal

// Free closes the ch and releases its resources.
func (ch *channel) Free() {
	ch.close()
	ch.free()
}

// internal

// receive receives a message from the connection.
func (ch *channel) receive(cancel <-chan struct{}, msg ptcp.Message) status.Status {
	ch.maybeStart()

	b := msg.Unwrap().Raw()
	return ch.reader.write(cancel, b)
}

// private

func (ch *channel) free() {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	if ch.freed {
		return
	}

	ch.freed = true
	ch.reader.free()
	ch.writer.free()

	if debug {
		debugPrint(ch.client, "ch.free\t", ch.id)
	}
}

func (ch *channel) close() status.Status {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	if ch.closed {
		return status.OK
	}

	ch.closed = true
	ch.reader.close()

	if debug {
		debugPrint(ch.client, "ch.close\t", ch.id)
	}
	return ch.writer.writeClose(nil /* no cancel */, ch.id)
}

func (ch *channel) connClosed() {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	if ch.closed {
		return
	}

	ch.closed = true
	ch.reader.close()

	if debug {
		debugPrint(ch.client, "ch.conn-closed\t", ch.id)
	}
}

func (ch *channel) remoteClosed() {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	if ch.closed {
		return
	}

	ch.closed = true
	ch.reader.close()

	if debug {
		debugPrint(ch.client, "ch.remote-closed\t", ch.id)
	}
}

// handler

func (ch *channel) maybeStart() {
	ch.mu.RLock()
	defer ch.mu.RUnlock()

	if ch.started {
		return
	}
	ch.started = true

	go ch.handleLoop()
}

func (ch *channel) handleLoop() {
	// No need to use async.Go here, because we don't need the result/cancellation,
	// and recover panics manually.
	defer func() {
		if e := recover(); e != nil {
			st, stack := status.RecoverStack(e)
			ch.conn.logger.Error("Channel panic", "status", st, "stack", string(stack))
		}
	}()
	defer ch.Free()

	// Handle ch
	st := ch.conn.handler.HandleChannel(ch)
	switch st.Code {
	case status.CodeOK,
		status.CodeCancelled,
		status.CodeEnd,
		status.CodeClosed:
		return
	}

	// Log errors
	ch.conn.logger.Error("Channel error", "status", st)
}

// reader

type channelReader struct {
	mu     sync.Mutex
	queue  alloc.MQueue
	freed  bool
	closed bool
}

func newChannelReader() channelReader {
	return channelReader{
		queue: acquireQueue(),
	}
}

func (r *channelReader) free() {
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

func (r *channelReader) close() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return
	}

	r.queue.Close()
	r.closed = true
}

func (r *channelReader) read() ([]byte, bool, status.Status) {
	q, st := r.get()
	if !st.OK() {
		return nil, false, st
	}
	return q.Read()
}

func (r *channelReader) write(cancel <-chan struct{}, msg []byte) status.Status {
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

func (r *channelReader) wait() <-chan struct{} {
	q, st := r.get()
	if !st.OK() {
		return closedChan
	}
	return q.ReadWait()
}

func (r *channelReader) get() (alloc.MQueue, status.Status) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return nil, status.End
	}

	q := r.queue
	return q, status.OK
}

// writer

type channelWriter struct {
	mu    sync.Mutex
	conn  *conn
	freed bool

	buf    *alloc.Buffer
	writer spec.Writer
}

func newChannelWriter(c *conn) channelWriter {
	buf := acquireBuffer()

	return channelWriter{
		conn:   c,
		buf:    buf,
		writer: spec.NewWriterBuffer(buf),
	}
}

func (w *channelWriter) free() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.freed {
		return
	}
	w.freed = true

	w.writer.Free()
	w.writer = nil

	releaseBuffer(w.buf)
	w.buf = nil
}

func (w *channelWriter) writeNew(cancel <-chan struct{}, id bin.Bin128) status.Status {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.freed {
		return statusChannelClosed
	}

	var msg ptcp.Message
	{
		w.buf.Reset()
		w.writer.Reset(w.buf)

		w0 := ptcp.NewMessageWriterTo(w.writer.Message())
		w0.Code(ptcp.Code_NewChannel)

		w1 := w0.New()
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

func (w *channelWriter) writeMessage(cancel <-chan struct{}, id bin.Bin128, data []byte) status.Status {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.freed {
		return statusChannelClosed
	}

	var msg ptcp.Message
	{
		w.buf.Reset()
		w.writer.Reset(w.buf)

		w0 := ptcp.NewMessageWriterTo(w.writer.Message())
		w0.Code(ptcp.Code_ChannelMessage)

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

func (w *channelWriter) writeClose(cancel <-chan struct{}, id bin.Bin128) status.Status {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.freed {
		return statusChannelClosed
	}

	var msg ptcp.Message
	{
		w.buf.Reset()
		w.writer.Reset(w.buf)

		w0 := ptcp.NewMessageWriterTo(w.writer.Message())
		w0.Code(ptcp.Code_CloseChannel)

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

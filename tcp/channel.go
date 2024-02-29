package tcp

import (
	"sync"

	"github.com/basecomplextech/baselibrary/alloc"
	"github.com/basecomplextech/baselibrary/async"
	"github.com/basecomplextech/baselibrary/bin"
	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/spec"
	"github.com/basecomplextech/spec/proto/ptcp"
)

// Channel is a single channel in a TCP connection.
type Channel interface {
	// Close closes the channel and sends the close message.
	Close() status.Status

	// Context returns the channel context.
	Context() async.Context

	// Read

	// Read reads and returns a message, or false/end.
	// The message is valid until the next call to Read.
	// The method does not block if no messages, and returns false instead.
	Read(ctx async.Context) ([]byte, bool, status.Status)

	// ReadSync reads and returns a message, or an end.
	// The message is valid until the next call to Read.
	// The method blocks until a message is received, or the channel is closed.
	ReadSync(ctx async.Context) ([]byte, status.Status)

	// ReadWait returns a channel that is notified on a new message, or a channel close.
	ReadWait() <-chan struct{}

	// Write

	// Write writes a message to the channel.
	Write(ctx async.Context, msg []byte) status.Status

	// WriteAndClose writes a close message with a payload.
	WriteAndClose(ctx async.Context, msg []byte) status.Status

	// Internal

	// Free closes the channel and releases its resources.
	Free()
}

// internal

var _ Channel = (*channel)(nil)

type channel struct {
	id     bin.Bin128
	conn   *conn
	client bool

	stateMu sync.RWMutex
	state   *channelState
}

func openChannel(conn *conn, id bin.Bin128, window int) *channel {
	if debug {
		debugPrint(conn.client, "channel.open\t", id)
	}

	s := acquireState()
	s.started = true
	s.window = window
	s.writeWindow = window

	return &channel{
		id:     id,
		conn:   conn,
		client: conn.client,

		state: s,
	}
}

func openedChannel(conn *conn, msg ptcp.OpenChannel) *channel {
	id := msg.Id()
	window := int(msg.Window())

	if debug {
		debugPrint(conn.client, "channel.opened\t", id)
	}

	s := acquireState()
	s.newSent = true
	s.started = false
	s.window = window
	s.writeWindow = window

	return &channel{
		id:     id,
		conn:   conn,
		client: conn.client,

		state: s,
	}
}

// Close closes the channel and sends the close message.
func (ch *channel) Close() status.Status {
	return ch.close(nil /* no data */)
}

// Context returns the channel context.
func (ch *channel) Context() async.Context {
	s, ok := ch.rlock()
	if !ok {
		return async.DoneContext()
	}
	defer ch.stateMu.RUnlock()

	return s.context
}

// Read

// Read reads and returns a message, or false/end.
// The message is valid until the next call to Read.
// The method does not block if no messages, and returns false instead.
func (ch *channel) Read(ctx async.Context) ([]byte, bool, status.Status) {
	s, ok := ch.rlock()
	if !ok {
		return nil, false, statusChannelClosed
	}
	defer ch.stateMu.RUnlock()

	// Read message
	b, ok, st := s.readQueue.Read()
	switch {
	case !st.OK():
		return nil, false, st
	case !ok:
		return nil, false, status.OK
	}

	if debug && debugChannel {
		debugPrint(ch.client, "channel.read\t", ch.id, ok, st)
	}

	// Parse message
	msg, _, err := ptcp.ParseMessage(b)
	if err != nil {
		ch.close(nil /* no data */)
		return nil, false, tcpError(err)
	}

	// Handle message
	code := msg.Code()
	switch code {
	case ptcp.Code_OpenChannel:
		data := msg.Open().Data()
		if st := ch.incrementReadBytes(ctx, s, len(b)); !st.OK() {
			return nil, false, st
		}
		return data, true, status.OK

	case ptcp.Code_ChannelMessage:
		data := msg.Message().Data()
		if st := ch.incrementReadBytes(ctx, s, len(b)); !st.OK() {
			return nil, false, st
		}
		return data, true, status.OK

	case ptcp.Code_ChannelWindow:
		delta := msg.Window().Delta()
		ch.incrementWriteWindow(s, int(delta))

		if debug {
			debugPrint(ch.client, "channel.increment-window\t", ch.id)
		}
		return nil, false, status.OK

	case ptcp.Code_CloseChannel:
		s.close()

		if debug {
			debugPrint(ch.client, "channel.remote-closed\t", ch.id)
		}

		data := msg.Close().Data()
		if len(data) > 0 {
			return data, true, status.OK
		}
		return nil, false, status.End
	}

	return nil, false, tcpErrorf("unexpected message code %d", code)
}

// ReadSync reads and returns a message, or an end.
// The message is valid until the next call to Read.
// The method blocks until a message is received, or the channel is closed.
func (ch *channel) ReadSync(ctx async.Context) ([]byte, status.Status) {
	for {
		msg, ok, st := ch.Read(ctx)
		switch {
		case !st.OK():
			return nil, st
		case ok:
			return msg, status.OK
		}

		select {
		case <-ctx.Wait():
			return nil, ctx.Status()
		case <-ch.ReadWait():
		}
	}
}

// ReadWait returns a channel that is notified on a new message, or a channel close.
func (ch *channel) ReadWait() <-chan struct{} {
	s, ok := ch.rlock()
	if !ok {
		return nil
	}
	defer ch.stateMu.RUnlock()

	return s.readQueue.ReadWait()
}

// Write

// Write writes a message to the channel.
func (ch *channel) Write(ctx async.Context, msg []byte) status.Status {
	s, ok := ch.rlock()
	if !ok {
		return statusChannelClosed
	}
	defer ch.stateMu.RUnlock()

	if s.closed {
		return statusChannelClosed
	}

	if !s.newSent {
		return ch.writeNew(ctx, s, msg)
	}
	return ch.writeMessage(ctx, s, msg)
}

// WriteAndClose writes a close message with a payload.
func (ch *channel) WriteAndClose(ctx async.Context, msg []byte) status.Status {
	s, ok := ch.rlock()
	if !ok {
		return statusChannelClosed
	}
	defer ch.stateMu.RUnlock()

	if s.closed {
		return statusChannelClosed
	}

	if !s.newSent {
		if st := ch.writeNew(ctx, s, msg); !st.OK() {
			return st
		}
		return ch.close(nil /* no data */)
	}
	return ch.close(msg)
}

// Internal

// Free closes the channel and releases its resources.
func (ch *channel) Free() {
	ch.close(nil /* no data */)
	ch.free()
}

// internal

func (ch *channel) connClosed() {
	s, ok := ch.rlock()
	if !ok {
		return
	}
	defer ch.stateMu.RUnlock()

	ok = s.close()
	if !ok {
		return
	}

	if debug {
		debugPrint(ch.client, "channel.conn-closed\t", ch.id)
	}
}

// receiveMessage receives a message from the connection.
func (ch *channel) receiveMessage(ctx async.Context, msg ptcp.Message) {
	s, ok := ch.rlock()
	if !ok {
		return
	}
	defer ch.stateMu.RUnlock()

	if s.start() {
		go ch.run()
	}

	b := msg.Unwrap().Raw()
	_, _ = s.readQueue.Write(b) // ignore end and false, read queues are unbounded
}

// receiveWindow receives a window increment from the connection.
func (ch *channel) receiveWindow(ctx async.Context, msg ptcp.Message) {
	s, ok := ch.rlock()
	if !ok {
		return
	}
	defer ch.stateMu.RUnlock()

	delta := int(msg.Window().Delta())
	ch.incrementWriteWindow(s, delta)
}

// private

func (ch *channel) free() {
	ch.stateMu.Lock()
	defer ch.stateMu.Unlock()

	if ch.state == nil {
		return
	}

	s := ch.state
	ch.state = nil
	releaseState(s)

	if debug {
		debugPrint(ch.client, "channel.free\t", ch.id)
	}
}

func (ch *channel) rlock() (*channelState, bool) {
	ch.stateMu.RLock()

	if ch.state == nil {
		ch.stateMu.RUnlock()
		return nil, false
	}

	return ch.state, true
}

func (ch *channel) close(data []byte) status.Status {
	s, ok := ch.rlock()
	if !ok {
		return statusChannelClosed
	}
	defer ch.stateMu.RUnlock()

	ok = s.close()
	if !ok {
		return status.OK
	}

	if debug {
		debugPrint(ch.client, "channel.close\t", ch.id)
	}
	return ch.writeClose(async.NoContext(), s, data)
}

// run

func (ch *channel) run() {
	// No need to use async.Go here, because we don't need the result/cancellation,
	// and recover panics manually.
	defer func() {
		if e := recover(); e != nil {
			st := status.Recover(e)
			ch.conn.logger.ErrorStatus("Channel panic", st)
		}
	}()
	defer ch.Free()

	// Handle channel
	ctx := ch.Context()
	st := ch.conn.handler.HandleChannel(ctx, ch)
	switch st.Code {
	case status.CodeOK,
		status.CodeCancelled,
		status.CodeEnd,
		status.CodeClosed:
		return
	}

	// Log errors
	ch.conn.logger.ErrorStatus("Channel error", st)
}

// read bytes

func (ch *channel) incrementReadBytes(ctx async.Context, s *channelState, n int) status.Status {
	s.readBytes += n
	if s.readBytes < s.window/2 {
		return status.OK
	}

	delta := s.readBytes
	s.readBytes = 0
	return ch.writeWindow(ctx, s, delta)
}

// write

func (ch *channel) writeNew(ctx async.Context, s *channelState, data []byte) status.Status {
	if s.isClosed() {
		return statusChannelClosed
	}
	s.newSent = true

	var msg ptcp.Message
	{
		s.writeBuf.Reset()
		s.writer.Reset(s.writeBuf)

		w0 := ptcp.NewMessageWriterTo(s.writer.Message())
		w0.Code(ptcp.Code_OpenChannel)

		w1 := w0.Open()
		w1.Id(ch.id)
		w1.Window(int32(s.window))
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

	n := len(msg.Unwrap().Raw())
	if st := ch.decrementWriteWindow(ctx, s, n); !st.OK() {
		return st
	}

	s.wmu.Lock()
	defer s.wmu.Unlock()

	if debug && debugChannel {
		debugPrint(ch.client, "channel.write\t", ch.id)
	}

	return ch.conn.write(ctx, msg)
}

func (ch *channel) writeMessage(ctx async.Context, s *channelState, data []byte) status.Status {
	if s.isClosed() {
		return statusChannelClosed
	}

	var msg ptcp.Message
	{
		s.writeBuf.Reset()
		s.writer.Reset(s.writeBuf)

		w0 := ptcp.NewMessageWriterTo(s.writer.Message())
		w0.Code(ptcp.Code_ChannelMessage)

		w1 := w0.Message()
		w1.Id(ch.id)
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

	n := len(msg.Unwrap().Raw())
	if st := ch.decrementWriteWindow(ctx, s, n); !st.OK() {
		return st
	}

	s.wmu.Lock()
	defer s.wmu.Unlock()

	if debug && debugChannel {
		debugPrint(ch.client, "channel.write\t", ch.id)
	}

	return ch.conn.write(ctx, msg)
}

func (ch *channel) writeWindow(ctx async.Context, s *channelState, delta int) status.Status {
	s.wmu.Lock()
	defer s.wmu.Unlock()

	var msg ptcp.Message
	{
		s.writeBuf.Reset()
		s.writer.Reset(s.writeBuf)

		w0 := ptcp.NewMessageWriterTo(s.writer.Message())
		w0.Code(ptcp.Code_ChannelWindow)

		w1 := w0.Window()
		w1.Id(ch.id)
		w1.Delta(int32(delta))
		if err := w1.End(); err != nil {
			return tcpError(err)
		}

		var err error
		msg, err = w0.Build()
		if err != nil {
			return tcpError(err)
		}
	}

	return ch.conn.write(ctx, msg)
}

func (ch *channel) writeClose(ctx async.Context, s *channelState, data []byte) status.Status {
	s.wmu.Lock()
	defer s.wmu.Unlock()

	var msg ptcp.Message
	{
		s.writeBuf.Reset()
		s.writer.Reset(s.writeBuf)

		w0 := ptcp.NewMessageWriterTo(s.writer.Message())
		w0.Code(ptcp.Code_CloseChannel)

		w1 := w0.Close()
		w1.Id(ch.id)
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

	return ch.conn.write(ctx, msg)
}

// write window

func (ch *channel) decrementWriteWindow(ctx async.Context, s *channelState, n int) status.Status {
	for {
		// Decrement write window for normal small messages.
		s.wmu.Lock()
		if s.writeWindow >= n {
			s.writeWindow -= n
			s.wmu.Unlock()
			return status.OK
		}

		// Decrement write window for large messages, when the remaining window
		// is greater than the half of the initial window, but the message size
		// still exceeds it.
		if s.writeWindow >= s.window/2 {
			s.writeWindow -= n
			s.wmu.Unlock()
			return status.OK
		}
		s.wmu.Unlock()

		// Wait for write window increment.
		select {
		case <-ctx.Wait():
			return ctx.Status()
		case <-s.writeWait:
		}
	}
}

func (ch *channel) incrementWriteWindow(s *channelState, delta int) {
	s.wmu.Lock()
	defer s.wmu.Unlock()

	s.writeWindow += delta

	for {
		select {
		case s.writeWait <- struct{}{}:
		default:
			return
		}
	}
}

// state

type channelState struct {
	context async.Context
	window  int // Initial window size

	readQueue alloc.MQueue
	readBytes int // Read bytes since last window increment

	wmu         sync.Mutex
	writer      spec.Writer
	writeBuf    *alloc.Buffer
	writeWait   chan struct{} // Wait for window increment
	writeWindow int           // Remaining write window, can become negative on large messages

	mu      sync.RWMutex
	closed  bool
	started bool
	newSent bool
}

var statePool = &sync.Pool{
	New: func() any {
		return newChannelState()
	},
}

func acquireState() *channelState {
	return statePool.Get().(*channelState)
}

func releaseState(s *channelState) {
	s.reset()
	statePool.Put(s)
}

func newChannelState() *channelState {
	buf := alloc.NewBuffer()

	return &channelState{
		context:   async.NewContext(),
		readQueue: alloc.NewMQueue(),
		writeBuf:  buf,
		writer:    spec.NewWriterBuffer(buf),
		writeWait: make(chan struct{}, 1),
	}
}

func (s *channelState) close() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return false
	}

	s.closed = true
	s.context.Cancel()
	s.readQueue.Close()
	return true
}

func (s *channelState) isClosed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.closed
}

func (s *channelState) start() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed || s.started {
		return false
	}

	s.started = true
	return true
}

func (s *channelState) reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.context.Free()
	s.context = async.NewContext()
	s.window = 0

	s.readQueue.Reset()
	s.readBytes = 0

	s.writeBuf.Reset()
	s.writer.Reset(s.writeBuf)
	s.writeWindow = 0
	select {
	case <-s.writeWait:
	default:
	}

	s.closed = false
	s.started = false
	s.newSent = false
}

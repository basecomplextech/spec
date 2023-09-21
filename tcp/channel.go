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

	// Write writes a message to the channel.
	Write(cancel <-chan struct{}, msg []byte) status.Status

	// Close closes the channel and sends the close message.
	Close() status.Status

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

func openChannel(conn *conn, id bin.Bin128) *channel {
	if debug {
		debugPrint(conn.client, "channel.open\t", id)
	}

	s := acquireState()
	s.started = true
	s.window = channelWindow
	s.writeWindow = channelWindow

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

// Read reads a message from the ch, the message is valid until the next iteration.
func (ch *channel) Read(cancel <-chan struct{}) ([]byte, status.Status) {
	s, ok := ch.rlock()
	if !ok {
		return nil, statusChannelClosed
	}
	defer ch.stateMu.RUnlock()

	for {
		b, ok, st := s.readQueue.Read()
		switch {
		case !st.OK():
			// End
			return nil, st

		case !ok:
			// Wait for next messages or end
			select {
			case <-cancel:
				return nil, status.Cancelled
			case <-s.readQueue.ReadWait():
				continue
			}
		}

		if debug && debugChannel {
			debugPrint(ch.client, "channel.read\t", ch.id, ok, st)
		}

		// Parse message
		msg, _, err := ptcp.ParseMessage(b)
		if err != nil {
			ch.close()
			return nil, tcpError(err)
		}

		// Handle message
		code := msg.Code()
		switch code {
		case ptcp.Code_ChannelMessage:
			data := msg.Message().Data()
			if st := ch.incrementReadBytes(cancel, s, len(b)); !st.OK() {
				return nil, st
			}
			return data, status.OK

		case ptcp.Code_ChannelWindow:
			delta := msg.Window().Delta()
			ch.incrementWriteWindow(s, int(delta))

			if debug {
				debugPrint(ch.client, "channel.increment-window\t", ch.id)
			}

		case ptcp.Code_CloseChannel:
			s.close()

			if debug {
				debugPrint(ch.client, "channel.remote-closed\t", ch.id)
			}
			return nil, status.End
		}

		return nil, tcpErrorf("unexpected message code %d", code)
	}
}

// Write writes a message to the channel.
func (ch *channel) Write(cancel <-chan struct{}, msg []byte) status.Status {
	s, ok := ch.rlock()
	if !ok {
		return statusChannelClosed
	}
	defer ch.stateMu.RUnlock()

	if s.closed {
		return statusChannelClosed
	}

	if !s.newSent {
		if st := ch.writeNew(cancel, s); !st.OK() {
			return st
		}
		s.newSent = true
	}
	return ch.writeMessage(cancel, s, msg)
}

// Close closes the channel and sends the close message.
func (ch *channel) Close() status.Status {
	return ch.close()
}

// Internal

// Free closes the channel and releases its resources.
func (ch *channel) Free() {
	ch.close()
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
func (ch *channel) receiveMessage(cancel <-chan struct{}, msg ptcp.Message) {
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
func (ch *channel) receiveWindow(cancel <-chan struct{}, msg ptcp.Message) {
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

func (ch *channel) close() status.Status {
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
	return ch.writeClose(nil /* no cancel */, s)
}

// run

func (ch *channel) run() {
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

// read bytes

func (ch *channel) incrementReadBytes(cancel <-chan struct{}, s *channelState, n int) status.Status {
	s.readBytes += n
	if s.readBytes < s.window/2 {
		return status.OK
	}

	delta := s.readBytes
	s.readBytes = 0
	return ch.writeWindow(cancel, s, delta)
}

// write

func (ch *channel) writeNew(cancel <-chan struct{}, s *channelState) status.Status {
	if s.isClosed() {
		return statusChannelClosed
	}

	s.wmu.Lock()
	defer s.wmu.Unlock()

	var msg ptcp.Message
	{
		s.writeBuf.Reset()
		s.writer.Reset(s.writeBuf)

		w0 := ptcp.NewMessageWriterTo(s.writer.Message())
		w0.Code(ptcp.Code_OpenChannel)

		w1 := w0.Open()
		w1.Id(ch.id)
		w1.Window(int32(s.window))
		if err := w1.End(); err != nil {
			return tcpError(err)
		}

		var err error
		msg, err = w0.Build()
		if err != nil {
			return tcpError(err)
		}
	}

	return ch.conn.write(cancel, msg)
}

func (ch *channel) writeMessage(cancel <-chan struct{}, s *channelState, data []byte) status.Status {
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
	if st := ch.decrementWriteWindow(cancel, s, n); !st.OK() {
		return st
	}

	s.wmu.Lock()
	defer s.wmu.Unlock()

	if debug && debugChannel {
		debugPrint(ch.client, "channel.write\t", ch.id)
	}

	return ch.conn.write(cancel, msg)
}

func (ch *channel) writeWindow(cancel <-chan struct{}, s *channelState, delta int) status.Status {
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

	return ch.conn.write(cancel, msg)
}

func (ch *channel) writeClose(cancel <-chan struct{}, s *channelState) status.Status {
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
		if err := w1.End(); err != nil {
			return tcpError(err)
		}

		var err error
		msg, err = w0.Build()
		if err != nil {
			return tcpError(err)
		}
	}

	return ch.conn.write(cancel, msg)
}

// write window

func (ch *channel) decrementWriteWindow(cancel <-chan struct{}, s *channelState, n int) status.Status {
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
		case <-cancel:
			return status.Cancelled
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

var statePool = &sync.Pool{}

type channelState struct {
	window int // Initial window size

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

func acquireState() *channelState {
	v := statePool.Get()
	if v != nil {
		return v.(*channelState)
	}
	return newChannelState()
}

func releaseState(s *channelState) {
	s.reset()
	statePool.Put(s)
}

func newChannelState() *channelState {
	buf := alloc.NewBuffer()

	return &channelState{
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

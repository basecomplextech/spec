package mpx

import (
	"sync"

	"github.com/basecomplextech/baselibrary/alloc"
	"github.com/basecomplextech/baselibrary/async"
	"github.com/basecomplextech/baselibrary/bin"
	"github.com/basecomplextech/baselibrary/pools"
	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/spec"
	"github.com/basecomplextech/spec/proto/pmpx"
)

type Channel interface {
	// Context returns a channel context.
	Context() async.Context

	// Receive

	// Receive receives and returns a message, or false/end.
	//
	// The message is valid until the next call to Receive.
	// The method does not block if no messages, and returns false instead.
	Receive(ctx async.Context) ([]byte, bool, status.Status)

	// ReceiveSync receives and returns a message, or an end.
	//
	// The message is valid until the next call to Receive.
	// The method blocks until a message is received, or the channel is closed.
	ReceiveSync(ctx async.Context) ([]byte, status.Status)

	// ReceiveWait returns a channel that is notified on a new message, or a channel close.
	ReceiveWait() <-chan struct{}

	// Send

	// Send sends a message to the channel.
	Send(ctx async.Context, msg []byte) status.Status

	// SendAndEnd sends an end message with a payload.
	SendAndEnd(ctx async.Context, msg []byte) status.Status

	// SendAndClose sends a close message with a payload.
	SendAndClose(ctx async.Context, msg []byte) status.Status

	// SendEnd sends an end message.
	//
	// The channel is still open after an end message.
	// No more messages can be sent, but messages can still be received.
	SendEnd(ctx async.Context) status.Status

	// SendClose sends a close message, and closes the channel.
	//
	// The channel is closed after a close message.
	// No more messages can be sent or received.
	SendClose(ctx async.Context) status.Status

	// Internal

	// Free closes the channel and releases its resources.
	Free()
}

// internal

type internalChannel interface {
	// ReceiveFree is called when the channel is freed by the connection.
	ReceiveFree()

	// ReceiveMessage is called when the channel receives a message from the connection.
	ReceiveMessage(msg pmpx.Message) status.Status
}

var (
	_ Channel         = (*channel)(nil)
	_ internalChannel = (*channel)(nil)
)

type channel struct {
	stateMu sync.RWMutex
	state   *channelState
}

type channelState struct {
	id     bin.Bin128
	ctx    async.Context
	conn   internalConn
	client bool // distinguishes client and server channels
	window int  // initial window size

	// send
	sendMu     sync.Mutex
	sendWriter spec.Writer
	sendBuffer *alloc.Buffer

	sendOpen  bool // open message sent
	sendClose bool // close message sent
	sendEnd   bool // end message sent
	sendFree  bool // freed by user

	// receive
	recvMu    sync.Mutex
	recvQueue alloc.MQueue

	recvOpen  bool // open message received
	recvClose bool // close message received
	recvEnd   bool // end message received
	recvFree  bool // freed by connection

	// windows
	windowMu   sync.Mutex
	windowSend int           // remaining send window, can become negative on sending large messages
	windowRecv int           // remaining recv window, can become negative on receiving large messages
	windowWait chan struct{} // wait for send window increment
}

// newChannel returns a new channel.
func newChannel(c internalConn, id bin.Bin128, client bool) *channel {
	s := acquireChannelState()
	s.id = id
	s.conn = c
	s.client = client

	return &channel{state: s}
}

func newChannelState() *channelState {
	sendBuffer := alloc.NewBuffer()

	return &channelState{
		ctx: async.NewContext(),

		sendBuffer: sendBuffer,
		sendWriter: spec.NewWriterBuffer(sendBuffer),

		recvQueue:  alloc.NewMQueue(),
		windowWait: make(chan struct{}, 1),
	}
}

// Context returns a channel context.
func (ch *channel) Context() async.Context {
	s, ok := ch.rlock()
	if !ok {
		return async.CancelledContext()
	}
	defer ch.stateMu.RUnlock()

	return s.ctx
}

// Receive

// Receive receives and returns a message, or false/end.
// The message is valid until the next call to Receive.
// The method does not block if no messages, and returns false instead.
func (ch *channel) Receive(ctx async.Context) ([]byte, bool, status.Status) {
	// RLock state
	s, ok := ch.rlock()
	if !ok {
		return nil, false, statusChannelClosed
	}
	defer ch.stateMu.RUnlock()

	// Lock receive
	s.recvMu.Lock()
	defer s.recvMu.Unlock()

	// Poll queue
	data, ok, st := s.recvQueue.Read()
	switch {
	case !st.OK():
		return nil, false, st
	case ok:
		return data, true, status.OK
	}

	// Maybe ended
	if s.recvEnd {
		return nil, false, status.End
	}
	return nil, false, status.OK
}

// ReceiveSync receives and returns a message, or an end.
// The message is valid until the next call to Receive.
// The method blocks until a message is received, or the channel is closed.
func (ch *channel) ReceiveSync(ctx async.Context) ([]byte, status.Status) {
	for {
		// Poll channel
		data, ok, st := ch.Receive(ctx)
		switch {
		case !st.OK():
			return nil, st
		case ok:
			return data, status.OK
		}

		// Await new message or close
		select {
		case <-ctx.Wait():
			return nil, ctx.Status()
		case <-ch.ReceiveWait():
			return nil, status.OK
		}
	}
}

// ReceiveWait returns a channel that is notified on a new message, or a channel close.
func (ch *channel) ReceiveWait() <-chan struct{} {
	s, ok := ch.rlock()
	if !ok {
		return closedChan
	}
	defer ch.stateMu.Unlock()

	return s.recvQueue.ReadWait()
}

// Send

// Send sends a message to the channel.
func (ch *channel) Send(ctx async.Context, data []byte) status.Status {
	input := pmpx.SendMessageInput{
		Data: data,
	}
	return ch.sendMessage(ctx, input)
}

// SendAndEnd sends an end message with a payload.
func (ch *channel) SendAndEnd(ctx async.Context, msg []byte) status.Status {
	input := pmpx.SendMessageInput{
		Data: msg,
		End:  true,
	}
	return ch.sendMessage(ctx, input)
}

// SendAndClose sends a close message with a payload.
func (ch *channel) SendAndClose(ctx async.Context, data []byte) status.Status {
	input := pmpx.SendMessageInput{
		Data:  data,
		Close: true,
	}
	return ch.sendMessage(ctx, input)
}

// SendEnd sends an end message.
func (ch *channel) SendEnd(ctx async.Context) status.Status {
	input := pmpx.SendMessageInput{
		End: true,
	}
	return ch.sendMessage(ctx, input)
}

// SendClose sends a close message, and closes the channel.
func (ch *channel) SendClose(ctx async.Context) status.Status {
	input := pmpx.SendMessageInput{
		Close: true,
	}
	return ch.sendMessage(ctx, input)
}

// Internal

// Free closes the channel and releases its resources.
func (ch *channel) Free() {
	ctx := async.NoContext()

	ch.sendMessage(ctx, pmpx.SendMessageInput{Close: true}) // ignore error
	ch.sendFree()
	ch.free()
}

// ReceiveFree is called when the channel is freed by the connection.
func (ch *channel) ReceiveFree() {
	ch.receiveFree()
	ch.free()
}

// ReceiveMessage is called when a message is received.
func (ch *channel) ReceiveMessage(msg pmpx.Message) status.Status {
	code := msg.Code()

	switch code {
	case pmpx.Code_ChannelOpen:
		return ch.receiveOpen(msg)
	case pmpx.Code_ChannelClose:
		return ch.receiveClose(msg)
	case pmpx.Code_ChannelEnd:
		return ch.receiveEnd(msg)
	case pmpx.Code_ChannelWindow:
		return ch.receiveWindow(msg)

	case pmpx.Code_ChannelMessage:
		var increment bool
		if st := ch.receiveMessage(msg, &increment); !st.OK() {
			return st
		}
		if !increment {
			return status.OK
		}

		delta, st := ch.incrementRecvWindow()
		if delta == 0 || !st.OK() {
			return st
		}
		return ch.sendWindow(delta)
	}

	return mpxErrorf("received unexpected message, code=%v", code)
}

// private

func (ch *channel) free() {
	ch.stateMu.Lock()
	defer ch.stateMu.Unlock()

	s := ch.state
	if s == nil {
		return
	}

	if !s.sendFree || s.recvFree {
		return
	}

	ch.state = nil
	releaseChannelState(s)
}

func (ch *channel) rlock() (*channelState, bool) {
	ch.stateMu.RLock()
	s := ch.state
	if s == nil {
		ch.stateMu.RUnlock()
		return nil, false
	}
	return s, true
}

// send

func (ch *channel) sendMessage(ctx async.Context, input pmpx.SendMessageInput) status.Status {
	s, ok := ch.rlock()
	if !ok {
		return statusChannelClosed
	}
	defer ch.stateMu.RUnlock()

	// Lock send
	s.sendMu.Lock()
	defer s.sendMu.Unlock()

	// Check closed/ended
	switch {
	case s.sendClose:
		return statusChannelClosed
	case s.sendEnd:
		return statusChannelEnded
	}

	// Make message
	input.ID = s.id
	input.Open = !s.sendOpen

	s.sendBuffer.Reset()
	s.sendWriter.Reset(s.sendBuffer)

	msg, err := pmpx.MakeSendMessage(s.sendWriter.Message(), input)
	if err != nil {
		return mpxError(err)
	}

	// Decrement send window, or wait
	n := len(msg.Unwrap().Raw())
	if st := s.decrementSendWindow(ctx, n, true /* wait */); !st.OK() {
		return st
	}

	// Write message
	return s.conn.SendMessage(ctx, msg)
}

func (ch *channel) sendWindow(delta uint32) status.Status {
	s, ok := ch.rlock()
	if !ok {
		return statusChannelClosed
	}
	defer ch.stateMu.RUnlock()

	// Lock send
	s.sendMu.Lock()
	defer s.sendMu.Unlock()

	// Make message
	s.sendBuffer.Reset()
	s.sendWriter.Reset(s.sendBuffer)

	msg, err := pmpx.MakeSendWindow(s.sendWriter.Message(), s.id, delta)
	if err != nil {
		return mpxError(err)
	}

	// Write message
	ctx := async.NoContext()
	return s.conn.SendMessage(ctx, msg)
}

func (ch *channel) sendFree() {
	s, ok := ch.rlock()
	if !ok {
		return
	}
	defer ch.stateMu.RUnlock()

	// Lock send
	s.sendMu.Lock()
	defer s.sendMu.Unlock()

	// Close and free send
	s.sendClose = true
	s.sendEnd = true
	s.sendFree = true
}

// receive

func (ch *channel) receiveOpen(msg pmpx.Message) status.Status {
	s, ok := ch.rlock()
	if !ok {
		return statusChannelClosed
	}
	defer ch.stateMu.RUnlock()

	// Check server
	if s.client {
		return mpxErrorf("received open message on client channel, channel=%v", s.id)
	}

	// Lock send/receive
	s.sendMu.Lock()
	defer s.sendMu.Unlock()

	s.recvMu.Lock()
	defer s.recvMu.Unlock()

	if s.recvOpen {
		return mpxErrorf("received duplicate open message, channel=%v", s.id)
	}

	// Handle open
	m := msg.Open()
	data := m.Data()
	close := m.Close()
	end := m.End_()
	window := m.Window()

	// Init state
	s.sendOpen = true

	s.recvOpen = true
	s.recvClose = close
	s.recvEnd = end

	s.windowSend = int(window)
	s.windowRecv = int(window)

	// Maybe write data
	if len(data) != 0 {
		_, _ = s.recvQueue.Write(data) // ignore end and false, read queues are unbounded
	}

	// Maybe close queue
	if end {
		s.recvQueue.Close()
	}
	return status.OK
}

func (ch *channel) receiveClose(msg pmpx.Message) status.Status {
	s, ok := ch.rlock()
	if !ok {
		return statusChannelClosed
	}
	defer ch.stateMu.RUnlock()

	// Lock receive
	s.recvMu.Lock()
	defer s.recvMu.Unlock()

	if s.recvClose {
		return mpxErrorf("received duplicate close message, channel=%v", s.id)
	}

	// Handle close
	m := msg.Close()
	data := m.Data()
	size := len(msg.Unwrap().Raw())

	s.recvClose = true
	s.recvEnd = true
	s.decrementRecvWindow(size)

	// Maybe write data
	if len(data) != 0 {
		_, _ = s.recvQueue.Write(data) // ignore end and false, read queues are unbounded
	}

	// Close queue
	s.recvQueue.Close()
	return status.OK
}

func (ch *channel) receiveEnd(msg pmpx.Message) status.Status {
	s, ok := ch.rlock()
	if !ok {
		return statusChannelClosed
	}
	defer ch.stateMu.RUnlock()

	// Lock receive
	s.recvMu.Lock()
	defer s.recvMu.Unlock()

	if s.recvEnd {
		return mpxErrorf("received duplicate end message, channel=%v", s.id)
	}

	// Handle end
	m := msg.End_()
	data := m.Data()
	size := len(msg.Unwrap().Raw())

	s.recvEnd = true
	s.recvQueue.Close()
	s.decrementRecvWindow(size)

	// Maybe write data
	if len(data) != 0 {
		_, _ = s.recvQueue.Write(data) // ignore end and false, read queues are unbounded
	}

	// Close queue
	s.recvQueue.Close()
	return status.OK
}

func (ch *channel) receiveMessage(msg pmpx.Message, incrementWindow *bool) status.Status {
	s, ok := ch.rlock()
	if !ok {
		return statusChannelClosed
	}
	defer ch.stateMu.RUnlock()

	// Lock receive
	s.recvMu.Lock()
	defer s.recvMu.Unlock()

	if s.recvClose || s.recvEnd {
		return mpxErrorf("received message after close/end, channel=%v", s.id)
	}

	// Handle message
	m := msg.Message()
	data := m.Data()
	size := len(msg.Unwrap().Raw())

	*incrementWindow = s.decrementRecvWindow(size)

	// Write data
	_, _ = s.recvQueue.Write(data) // ignore end and false, read queues are unbounded
	return status.OK
}

func (ch *channel) receiveWindow(msg pmpx.Message) status.Status {
	s, ok := ch.rlock()
	if !ok {
		return statusChannelClosed
	}
	defer ch.stateMu.RUnlock()

	// Lock receive
	s.recvMu.Lock()
	defer s.recvMu.Unlock()

	if s.recvClose || s.recvEnd {
		return mpxErrorf("received window after close/end, channel=%v", s.id)
	}

	// Handle message
	m := msg.Window()
	delta := int(m.Delta())
	s.incrementSendWindow(delta)
	return status.OK
}

func (ch *channel) receiveFree() {
	s, ok := ch.rlock()
	if !ok {
		return
	}
	defer ch.stateMu.RUnlock()

	// Lock send/receive
	s.sendMu.Lock()
	defer s.sendMu.Unlock()

	s.recvMu.Lock()
	defer s.recvMu.Unlock()

	// Close send
	s.sendClose = true
	s.sendEnd = true

	// Close and free recv
	s.recvClose = true
	s.recvEnd = true
	s.recvFree = true
	s.recvQueue.Close()
}

// windows

func (s *channelState) incrementSendWindow(delta int) {
	s.windowMu.Lock()
	defer s.windowMu.Unlock()

	s.windowSend += delta

	select {
	case s.windowWait <- struct{}{}:
	default:
	}
}

func (s *channelState) decrementSendWindow(ctx async.Context, n int, wait bool) status.Status {
	for {
		// Decrement send window for normal small messages
		s.windowMu.Lock()
		if !wait || s.windowSend >= n {
			s.windowSend -= n
			s.windowMu.Unlock()
			return status.OK
		}

		// Decrement send window for large messages, when the remaining window
		// is greater than the half of the initial window, but the message size
		// still exceeds it.
		if s.windowSend >= s.window/2 {
			s.windowSend -= n
			s.windowMu.Unlock()
			return status.OK
		}
		s.windowMu.Unlock()

		// Wait for send window increment
		select {
		case <-ctx.Wait():
			return ctx.Status()
		case <-s.ctx.Wait():
			return s.ctx.Status()
		case <-s.windowWait:
		}
	}
}

func (ch *channel) incrementRecvWindow() (uint32, status.Status) {
	s, ok := ch.rlock()
	if !ok {
		return 0, statusChannelClosed
	}
	defer ch.stateMu.RUnlock()

	// Lock window
	s.windowMu.Lock()
	defer s.windowMu.Unlock()

	if s.windowRecv >= s.window {
		return 0, status.OK
	}

	// Increment recv window
	delta := s.window - s.windowRecv
	s.windowRecv += delta
	return uint32(delta), status.OK
}

func (s *channelState) decrementRecvWindow(delta int) bool {
	s.windowMu.Lock()
	defer s.windowMu.Unlock()

	s.windowRecv -= delta
	return s.windowRecv <= s.window/2
}

// state pool

var channelPool = pools.MakePool(newChannelState)

func acquireChannelState() *channelState {
	return channelPool.New()
}

func releaseChannelState(s *channelState) {
	s.reset()
	channelPool.Put(s)
}

func (s *channelState) reset() {
	s.id = bin.Bin128{}
	s.conn = nil
	s.client = false
	s.window = 0

	s.ctx.Free()
	s.ctx = async.NewContext()

	// Reset send
	s.sendMu.Lock()
	defer s.sendMu.Unlock()

	s.sendBuffer.Reset()
	s.sendWriter.Reset(s.sendBuffer)

	s.sendOpen = false
	s.sendClose = false
	s.sendEnd = false
	s.sendFree = false

	// Reset receive
	s.recvMu.Lock()
	defer s.recvMu.Unlock()

	s.recvQueue.Reset()

	s.recvOpen = false
	s.recvClose = false
	s.recvEnd = false
	s.recvFree = false

	// Reset windows
	s.windowMu.Lock()
	defer s.windowMu.Unlock()

	s.windowSend = 0
	s.windowRecv = 0
	select {
	case <-s.windowWait:
	default:
	}
}

// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package mpx

import (
	"net"
	"sync"

	"github.com/basecomplextech/baselibrary/alloc"
	"github.com/basecomplextech/baselibrary/async"
	"github.com/basecomplextech/baselibrary/bin"
	"github.com/basecomplextech/baselibrary/logging"
	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/spec/proto/pmpx"
)

type Conn interface {
	// Close closes the connection and frees its internal resources.
	Close() status.Status

	// Closed returns a flag that is set when the connection is closed.
	Closed() async.Flag

	// Channel opens a new channel.
	Channel(ctx async.Context) (Channel, status.Status)

	// OnClosed adds a disconnect listener, and returns an unsubscribe function.
	//
	// The unsubscribe function does not deadlock, even if the listener is being called right now.
	OnClosed(fn func()) (unsub func())

	// Internal

	// Free closes and frees the connection, allows to wrap the connection into ref.R[Conn].
	Free()
}

// Connect dials an address and returns a connection.
func Connect(address string, logger logging.Logger, opts Options) (Conn, status.Status) {
	delegate := noopConnDelegate{}
	opts = opts.clean()
	return connect(address, delegate, logger, opts)
}

// internal

var _ conn = (*connImpl)(nil)

type conn interface {
	Conn

	// OnClosed adds a disconnection listener, and returns an unsubscribe function.
	OnClosed(fn func()) (unsub func())

	// SendMessage write an outgoing message to the write queue.
	SendMessage(ctx async.Context, msg pmpx.Message) status.Status
}

type connImpl struct {
	conn     net.Conn
	client   bool
	delegate connDelegate
	handler  Handler
	logger   logging.Logger
	options  Options

	// flags
	closed      async.MutFlag
	negotiated  async.MutFlag
	negotiated_ bool

	// reader/writer
	reader *reader
	writer *writer
	writeq alloc.ByteQueue

	// channels
	channelMu       sync.Mutex
	channels        map[bin.Bin128]internalChannel
	channelsClosed  bool
	channelsReached bool // number of channels reached the target number, notify delegate once

	// close listeners
	closedMu        sync.Mutex
	closedSeq       int64
	closedFlag      bool
	closedListeners map[int64]func()
}

// connect connects to an address and returns a client connection.
func connect(address string, delegate connDelegate, logger logging.Logger, opts Options) (
	*connImpl, status.Status) {

	// Dial address
	conn, err := net.DialTimeout("tcp", address, opts.DialTimeout)
	if err != nil {
		return nil, mpxError(err)
	}

	// Incoming handler
	handler := HandleFunc(func(_ Context, ch Channel) status.Status {
		return status.ExternalError("client connection does not support incoming channels")
	})

	// Make connection
	c := newConn(conn, true /* client */, delegate, handler, logger, opts)
	go func() {
		defer func() {
			if e := recover(); e != nil {
				st := status.Recover(e)
				logger.ErrorStatus("Conn panic", st)
			}
		}()

		c.run()
	}()
	return c, status.OK
}

func newConn(
	conn net.Conn,
	client bool,
	delegate connDelegate,
	handler Handler,
	logger logging.Logger,
	opts Options,
) *connImpl {
	return &connImpl{
		conn:     conn,
		client:   client,
		delegate: delegate,
		handler:  handler,
		logger:   logger,
		options:  opts.clean(),

		closed:     async.UnsetFlag(),
		negotiated: async.UnsetFlag(),

		reader: newReader(conn, client, int(opts.ReadBufferSize)),
		writer: newWriter(conn, client, int(opts.WriteBufferSize)),
		writeq: alloc.NewByteQueueCap(int(opts.WriteQueueSize)),

		channels:        make(map[bin.Bin128]internalChannel),
		closedListeners: make(map[int64]func()),
	}
}

// Close closes the connection and frees its internal resources.
func (c *connImpl) Close() status.Status {
	err := c.conn.Close()
	if err != nil {
		return mpxError(err)
	}
	return status.OK
}

// Closed returns a flag that is set when the connection is closed.
func (c *connImpl) Closed() async.Flag {
	return c.closed
}

// Channel opens a new channel.
func (c *connImpl) Channel(ctx async.Context) (Channel, status.Status) {
	for {
		// Create new channel
		ch, ok, st := c.createChannel()
		switch {
		case !st.OK():
			return nil, st
		case ok:
			return ch, status.OK
		}

		// Wait for negotiation or close
		select {
		case <-ctx.Wait():
			return nil, ctx.Status()
		case <-c.closed.Wait():
			return nil, statusConnClosed
		case <-c.negotiated.Wait():
		}
	}
}

// OnClosed adds a disconnection listener, and returns an unsubscribe function.
func (c *connImpl) OnClosed(fn func()) (unsub func()) {
	// Add listener
	id := c.addClosed(fn)
	if id == 0 {
		return func() {}
	}

	// Return unsubscribe
	return func() {
		c.removeClosed(id)
	}
}

// Internal

// Free closes and frees the connection.
func (c *connImpl) Free() {
	c.Close()
}

// SendMessage write an outgoing message to the write queue.
func (c *connImpl) SendMessage(ctx async.Context, msg pmpx.Message) status.Status {
	b := msg.Unwrap().Raw()

	for {
		ok, st := c.writeq.Write(b)
		switch {
		case !st.OK():
			return statusConnClosed
		case ok:
			return status.OK
		}

		// Wait for space
		select {
		case <-ctx.Wait():
			return ctx.Status()
		case <-c.writeq.WriteWait(len(b)):
			continue
		}
	}
}

// private

// run is the main run loop of the connection.
func (c *connImpl) run() {
	defer c.free()
	defer c.close()

	// Negotiate protocol
	st := c.negotiate()
	switch st.Code {
	case status.CodeOK:
	case status.CodeCancelled,
		status.CodeEnd,
		status.CodeClosed:
		return
	default:
		c.logger.ErrorStatus("Conn error", st)
		return
	}

	// Run loops
	recv := async.Go(c.receiveLoop)
	send := async.Go(c.sendLoop)
	defer async.StopWaitAll(recv, send)
	defer c.close()

	// Await exit
	select {
	case <-recv.Wait():
		st = recv.Status()
	case <-send.Wait():
		st = send.Status()
	}

	// Maybe log error
	switch st.Code {
	case status.CodeOK,
		status.CodeCancelled,
		status.CodeEnd,
		status.CodeClosed:
	default:
		c.logger.ErrorStatus("Conn error", st)
	}
}

func (c *connImpl) close() {
	if c.closed.Get() {
		return
	}

	defer c.notifyClosed()
	defer c.delegate.onConnClosed(c)
	defer c.closeChannels()

	c.conn.Close()
	c.closed.Set()
	c.writeq.Close()
}

func (c *connImpl) free() {
	c.reader.free()
	c.writeq.Free()
}

// negotiate

func (c *connImpl) negotiate() status.Status {
	if c.client {
		return c.negotiateClient()
	} else {
		return c.negotiateServer()
	}
}

func (c *connImpl) negotiateClient() status.Status {
	// Write protocol line
	if st := c.writer.writeString(ProtocolLine); !st.OK() {
		return st
	}

	// Write connect request
	{
		w := pmpx.NewConnectRequestWriter()

		vv := w.Versions()
		vv.Add(pmpx.Version_Version10)
		vv.End()

		if c.options.Compress {
			cc := w.Compress()
			cc.Add(pmpx.Compress_Lz4)
			cc.End()
		}

		req, err := w.Build()
		if err != nil {
			return mpxError(err)
		}

		if st := c.writer.writeRequest(req); !st.OK() {
			return st
		}
	}

	// Read/check protocol line
	line, st := c.reader.readLine()
	if !st.OK() {
		return st
	}
	if line != ProtocolLine {
		return mpxErrorf("invalid protocol, expected %q, got %q", ProtocolLine, line)
	}

	// Read connect response
	resp, st := c.reader.readResponse()
	if !st.OK() {
		return st
	}

	// Check status
	ok := resp.Ok()
	if !ok {
		return mpxErrorf("server refused connection: %v", resp.Error())
	}

	// Check version
	v := resp.Version()
	if v != pmpx.Version_Version10 {
		return mpxErrorf("server returned unsupported version %d", v)
	}

	// Init compression
	comp := resp.Compress()
	switch comp {
	case pmpx.Compress_None:
	case pmpx.Compress_Lz4:
		if st := c.reader.initLZ4(); !st.OK() {
			return st
		}
		if st := c.writer.initLZ4(); !st.OK() {
			return st
		}
	default:
		return mpxErrorf("server returned unsupported compression %d", comp)
	}

	c.negotiated.Set()
	c.negotiated_ = true
	return status.OK
}

func (c *connImpl) negotiateServer() status.Status {
	// Write protocol line
	if st := c.writer.writeString(ProtocolLine); !st.OK() {
		return st
	}

	// Read/check protocol line
	line, st := c.reader.readLine()
	if !st.OK() {
		return st
	}
	if line != ProtocolLine {
		return mpxErrorf("invalid protocol, expected %q, got %q", ProtocolLine, line)
	}

	// Read connect request
	req, st := c.reader.readRequest()
	if !st.OK() {
		return st
	}

	// Check version
	ok := false
	versions := req.Versions()
	for i := 0; i < versions.Len(); i++ {
		v := versions.Get(i)
		if v == pmpx.Version_Version10 {
			ok = true
			break
		}
	}
	if !ok {
		w := pmpx.NewConnectResponseWriter()
		w.Ok(false)
		w.Error("unsupported protocol versions")

		resp, err := w.Build()
		if err != nil {
			return mpxError(err)
		}
		return c.writer.writeResponse(resp)
	}

	// Select compression
	comp := pmpx.Compress_None
	comps := req.Compress()
	for i := 0; i < comps.Len(); i++ {
		c := comps.Get(i)
		if c == pmpx.Compress_Lz4 {
			comp = pmpx.Compress_Lz4
			break
		}
	}

	// Return response
	{
		w := pmpx.NewConnectResponseWriter()
		w.Ok(true)
		w.Version(pmpx.Version_Version10)
		w.Compress(comp)

		resp, err := w.Build()
		if err != nil {
			return mpxError(err)
		}
		if st := c.writer.writeResponse(resp); !st.OK() {
			return st
		}
	}

	// Init compression
	switch comp {
	case pmpx.Compress_None:
	case pmpx.Compress_Lz4:
		if st := c.reader.initLZ4(); !st.OK() {
			return st
		}
		if st := c.writer.initLZ4(); !st.OK() {
			return st
		}
	}

	c.negotiated.Set()
	c.negotiated_ = true
	return status.OK
}

// receive

func (c *connImpl) receiveLoop(ctx async.Context) status.Status {
	for {
		// Receive message
		msg, st := c.reader.readMessage()
		if !st.OK() {
			return st
		}

		// Handle message
		code := msg.Code()
		switch code {
		case pmpx.Code_ChannelOpen:
			if st := c.receiveOpen(msg); !st.OK() {
				return st
			}
		case pmpx.Code_ChannelClose:
			if st := c.receiveClose(msg); !st.OK() {
				return st
			}
		case pmpx.Code_ChannelMessage:
			if st := c.receiveMessage(msg); !st.OK() {
				return st
			}
		case pmpx.Code_ChannelWindow:
			if st := c.receiveWindow(msg); !st.OK() {
				return st
			}

		default:
			return mpxErrorf("unexpected mpx message, code=%d", code)
		}
	}
}

func (c *connImpl) receiveOpen(msg pmpx.Message) status.Status {
	m := msg.Open()
	id := m.Id()

	c.channelMu.Lock()
	defer c.channelMu.Unlock()

	// Check not exist, impossible
	if _, ok := c.channels[id]; ok {
		return mpxErrorf("received open message for existing channel, channel=%v", id)
	}

	// Make channel
	ch := openChannel(c, c.client, m)
	c.channels[id] = ch

	// Free on error
	done := false
	defer func() {
		if !done {
			ch.Free1()
			ch.Free()
		}
	}()

	// Handle message
	if st := ch.Receive1(msg); !st.OK() {
		delete(c.channels, id)
		return st
	}
	c.maybeChannelsReached()

	// Start handler
	workerPool.Go(func() {
		c.handleChannel(ch)
	})
	done = true
	return status.OK
}

func (c *connImpl) receiveClose(msg pmpx.Message) status.Status {
	m := msg.Close()
	id := m.Id()

	c.channelMu.Lock()
	defer c.channelMu.Unlock()

	ch, ok := c.channels[id]
	if !ok {
		return status.OK
	}

	defer ch.Free1()
	delete(c.channels, id)

	return ch.Receive1(msg)
}

func (c *connImpl) receiveMessage(msg pmpx.Message) status.Status {
	m := msg.Message()
	id := m.Id()

	c.channelMu.Lock()
	defer c.channelMu.Unlock()

	ch, ok := c.channels[id]
	if !ok {
		return status.OK
	}

	return ch.Receive1(msg)
}

func (c *connImpl) receiveWindow(msg pmpx.Message) status.Status {
	m := msg.Window()
	id := m.Id()

	c.channelMu.Lock()
	defer c.channelMu.Unlock()

	ch, ok := c.channels[id]
	if !ok {
		return status.OK
	}

	return ch.Receive1(msg)
}

// send

func (c *connImpl) sendLoop(ctx async.Context) status.Status {
	for {
		// Write pending messages
		b, ok, st := c.writeq.Read()
		switch {
		case !st.OK():
			return st
		case ok:
			if st := c.sendMessage(b); !st.OK() {
				return st
			}
			continue
		}

		// Flush buffered writes
		if st := c.writer.flush(); !st.OK() {
			return st
		}

		// Wait for more messages
		select {
		case <-ctx.Wait():
			return ctx.Status()
		case <-c.writeq.ReadWait():
		}
	}
}

func (c *connImpl) sendMessage(b []byte) status.Status {
	msg, err := pmpx.NewMessageErr(b)
	if err != nil {
		return mpxError(err)
	}

	// Maybe delete and free channel
	code := msg.Code()
	if code == pmpx.Code_ChannelClose {
		id := msg.Close().Id()
		c.removeChannel(id)
	}

	// Write message
	return c.writer.writeMessage(b)
}

// channels

func (c *connImpl) closeChannels() {
	c.channelMu.Lock()
	defer c.channelMu.Unlock()

	if c.channelsClosed {
		return
	}
	c.channelsClosed = true

	for _, ch := range c.channels {
		ch.Free1()
	}
	c.channels = nil
}

func (c *connImpl) createChannel() (Channel, bool, status.Status) {
	id := bin.Random128()
	window := int(c.options.ChannelWindowSize)
	ch := createChannel(c, c.client, id, window)

	done := false
	defer func() {
		if !done {
			ch.Free()
		}
	}()

	c.channelMu.Lock()
	defer c.channelMu.Unlock()

	switch {
	case c.channelsClosed:
		return nil, false, statusConnClosed
	case !c.negotiated_:
		return nil, false, status.OK
	}

	c.channels[id] = ch
	c.maybeChannelsReached()
	done = true
	return ch, true, status.OK
}

func (c *connImpl) removeChannel(id bin.Bin128) {
	c.channelMu.Lock()
	defer c.channelMu.Unlock()

	if c.channelsClosed {
		return
	}

	ch, ok := c.channels[id]
	if !ok {
		return
	}

	defer ch.Free1()
	delete(c.channels, id)
}

func (c *connImpl) handleChannel(ch Channel) {
	// No need to use async.Go here, because we don't need the result/cancellation,
	// and recover panics manually.
	defer func() {
		if e := recover(); e != nil {
			st := status.Recover(e)
			c.logger.ErrorStatus("Channel panic", st)
		}
	}()
	defer ch.Free()

	// Handle channel
	ctx := ch.Context()
	st := c.handler.HandleChannel(ctx, ch)
	switch st.Code {
	case status.CodeOK,
		status.CodeCancelled,
		status.CodeEnd,
		status.CodeClosed:
		return
	}

	// Log errors
	c.logger.ErrorStatus("Channel error", st)
}

func (c *connImpl) maybeChannelsReached() {
	if c.channelsReached {
		return
	}

	target := c.options.ClientConnChannels
	if target <= 0 || len(c.channels) < target {
		return
	}

	c.channelsReached = true
	c.delegate.onConnChannelsReached(c)
}

// close listeners

func (c *connImpl) addClosed(fn func()) int64 {
	c.closedMu.Lock()
	defer c.closedMu.Unlock()

	if c.closedFlag {
		return 0
	}

	c.closedSeq++
	id := c.closedSeq

	c.closedListeners[id] = fn
	return id
}

func (c *connImpl) removeClosed(id int64) {
	c.closedMu.Lock()
	defer c.closedMu.Unlock()

	if c.closedFlag {
		return
	}

	delete(c.closedListeners, id)
}

func (c *connImpl) notifyClosed() {
	c.closedMu.Lock()
	if c.closedFlag {
		c.closedMu.Unlock()
		return
	}
	c.closedFlag = true
	c.closedMu.Unlock()

	// Notify outside of lock to avoid deadlocks
	for _, l := range c.closedListeners {
		l()
	}
	clear(c.closedListeners)
}

// worker pool

// workerPool allows to reuse goroutines with bigger stacks for handling channels.
// It does not provide any performance benefits in test benchmarks, but it does provide
// performance gains in real-world scenarios with big stacks, especially with chained RPC handlers.
//
//	BenchmarkTable_Get_Parallel-10:
//	goroutines		144731 ops	568 B/op	23 allocs/op
//	goroutine pool	184801 ops	558 B/op	23 allocs/op
var workerPool = async.NewPool()

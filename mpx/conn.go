// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package mpx

import (
	"net"
	"sync"
	"sync/atomic"

	"github.com/basecomplextech/baselibrary/alloc"
	"github.com/basecomplextech/baselibrary/async"
	"github.com/basecomplextech/baselibrary/async/asyncmap"
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
func Connect(ctx async.Context, addr string, logger logging.Logger, opts Options) (
	Conn, status.Status) {

	opts = opts.clean()
	dialer := newDialer(opts)
	delegate := noopConnDelegate{}
	return newConnector(dialer, delegate, logger, opts).connect(ctx, addr)
}

// ConnectDialer dials an address and returns a connection.
func ConnectDialer(ctx async.Context, addr string, dialer *net.Dialer, logger logging.Logger,
	opts Options) (Conn, status.Status) {

	opts = opts.clean()
	delegate := noopConnDelegate{}
	return newConnector(dialer, delegate, logger, opts).connect(ctx, addr)
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

type internalConn interface {
	send(ctx async.Context, msg pmpx.Message) status.Status
}

type connImpl struct {
	conn     net.Conn
	client   bool
	delegate connDelegate
	handler  Handler
	logger   logging.Logger
	options  Options

	// flags
	closed     async.MutFlag
	negotiated async.MutFlag

	// reader/writer
	reader *reader
	writer *writer
	writeq alloc.ByteQueue

	// channels
	channels        asyncmap.AtomicMap[bin.Bin128, internalChannel]
	channelsClosed  atomic.Bool
	channelsReached atomic.Bool // number of channels reached the target number, notify delegate once

	// close listeners
	closedMu        sync.Mutex
	closedSeq       int64
	closedFlag      bool
	closedListeners map[int64]func()
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
		options:  opts,

		closed:     async.UnsetFlag(),
		negotiated: async.UnsetFlag(),

		reader: newReader(conn, client, int(opts.ReadBufferSize)),
		writer: newWriter(conn, client, int(opts.WriteBufferSize)),
		writeq: alloc.NewByteQueueCap(int(opts.WriteQueueSize)),

		channels:        asyncmap.NewAtomicMap[bin.Bin128, internalChannel](),
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
	if c.closed.IsSet() {
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
	if st := c.writer.writeLine(ProtocolLine); !st.OK() {
		return st
	}

	// Write connect request
	req, err := pmpx.NewConnectInput().
		WithCompression(c.options.Compression).
		Build()
	if err != nil {
		return mpxError(err)
	}
	if st := c.writer.writeAndFlush(req); !st.OK() {
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

	// Read connect response
	resp, st := c.reader.readResponse()
	if !st.OK() {
		return st
	}
	if ok := resp.Ok(); !ok {
		return mpxErrorf("server refused connection: %v", resp.Error())
	}

	// Check version
	v := resp.Version()
	if v != pmpx.Version_Version10 {
		return mpxErrorf("server returned unsupported version %d", v)
	}

	// Init compression
	comp := resp.Compression()
	switch comp {
	case pmpx.ConnectCompression_None:
	case pmpx.ConnectCompression_Lz4:
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
	return status.OK
}

func (c *connImpl) negotiateServer() status.Status {
	// Write protocol line
	if st := c.writer.writeLine(ProtocolLine); !st.OK() {
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
		resp, err := pmpx.BuildConnectError("unsupported protocol versions")
		if err != nil {
			return mpxError(err)
		}
		return c.writer.writeAndFlush(resp)
	}

	// Select compression
	comp := pmpx.ConnectCompression_None
	comps := req.Compression()
	for i := 0; i < comps.Len(); i++ {
		c := comps.Get(i)
		if c == pmpx.ConnectCompression_Lz4 {
			comp = pmpx.ConnectCompression_Lz4
			break
		}
	}

	// Write response
	resp, err := pmpx.BuildConnectResponse(pmpx.Version_Version10, comp)
	if err != nil {
		return mpxError(err)
	}
	if st := c.writer.writeAndFlush(resp); !st.OK() {
		return st
	}

	// Init compression
	switch comp {
	case pmpx.ConnectCompression_None:
	case pmpx.ConnectCompression_Lz4:
		if st := c.reader.initLZ4(); !st.OK() {
			return st
		}
		if st := c.writer.initLZ4(); !st.OK() {
			return st
		}
	}

	c.negotiated.Set()
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
		case pmpx.Code_ChannelData:
			if st := c.receiveData(msg); !st.OK() {
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
	m := msg.ChannelOpen()
	id := m.Id()

	// Check not exist, impossible
	ok := c.channels.Contains(id)
	if ok {
		return mpxErrorf("received open message for existing channel, channel=%v", id)
	}

	// Make channel
	ch := openChannel(c, c.client, m)
	c.channels.Set(id, ch)

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
		c.channels.Delete(id)
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
	m := msg.ChannelClose()
	id := m.Id()

	ch, ok := c.channels.Get(id)
	if !ok {
		return status.OK
	}

	defer ch.Free1()
	c.channels.Delete(id)

	return ch.Receive1(msg)
}

func (c *connImpl) receiveData(msg pmpx.Message) status.Status {
	m := msg.ChannelData()
	id := m.Id()

	ch, ok := c.channels.Get(id)
	if !ok {
		return status.OK
	}

	return ch.Receive1(msg)
}

func (c *connImpl) receiveWindow(msg pmpx.Message) status.Status {
	m := msg.ChannelWindow()
	id := m.Id()

	ch, ok := c.channels.Get(id)
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
		id := msg.ChannelClose().Id()
		c.removeChannel(id)
	}

	// Write message
	return c.writer.write(msg)
}

// channels

func (c *connImpl) closeChannels() {
	if c.channelsClosed.Load() {
		return
	}
	c.channelsClosed.Store(true)

	c.channels.Range(func(_ bin.Bin128, ch internalChannel) bool {
		ch.Free1()
		return true
	})
}

func (c *connImpl) createChannel() (Channel, bool, status.Status) {
	id := bin.Random128()
	window := int(c.options.ChannelWindowSize)
	ch := createChannel(c, c.client, id, window)

	done := false
	defer func() {
		if !done {
			ch.free()
		}
	}()

	switch {
	case c.channelsClosed.Load():
		return nil, false, statusConnClosed
	case !c.negotiated.IsSet():
		return nil, false, status.OK
	}

	c.channels.Set(id, ch)
	c.maybeChannelsReached()

	if c.channelsClosed.Load() {
		c.channels.Delete(id)
		return nil, false, statusConnClosed
	}

	done = true
	return ch, true, status.OK
}

func (c *connImpl) removeChannel(id bin.Bin128) {
	if c.channelsClosed.Load() {
		return
	}

	ch, ok := c.channels.Get(id)
	if !ok {
		return
	}

	defer ch.Free1()
	c.channels.Delete(id)
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
	if !c.client || c.channelsReached.Load() {
		return
	}

	target := c.options.ClientConnChannels
	if target <= 0 {
		return
	}

	n := c.channels.Len()
	if n < target {
		return
	}

	set := c.channelsReached.CompareAndSwap(false, true)
	if !set {
		return
	}
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

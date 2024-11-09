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
	// Context returns a connection context.
	Context() ConnContext

	// Close

	// Close closes the connection and frees its internal resources.
	Close() status.Status

	// Closed returns a flag that is set when the connection is closed.
	Closed() async.Flag

	// OnClosed adds a disconnect listener, and returns an unsubscribe function.
	//
	// The unsubscribe function does not deadlock, even if the listener is being called right now.
	OnClosed(fn func()) (unsub func())

	// Channel

	// Channel opens a new channel.
	Channel(ctx async.Context) (Channel, status.Status)

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

type internalConn interface {
	Conn

	// send sends a message to the connection, or returns a connection closed or an end status.
	send(ctx async.Context, msg pmpx.Message) status.Status
}

// implementation

var _ internalConn = (*conn)(nil)

type conn struct {
	ctx      *connContext
	conn     net.Conn
	client   bool
	delegate connDelegate
	handler  Handler
	logger   logging.Logger
	options  Options

	// flags
	closed     async.MutFlag
	handshaked async.MutFlag

	// reader/writer
	reader *connReader
	writer *connWriter
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
	nc net.Conn,
	client bool,
	delegate connDelegate,
	handler Handler,
	logger logging.Logger,
	opts Options,
) *conn {
	c := &conn{
		conn:     nc,
		client:   client,
		delegate: delegate,
		handler:  handler,
		logger:   logger,
		options:  opts,

		closed:     async.UnsetFlag(),
		handshaked: async.UnsetFlag(),

		reader: newConnReader(nc, client, int(opts.ReadBufferSize)),
		writer: newConnWriter(nc, client, int(opts.WriteBufferSize)),
		writeq: alloc.NewByteQueueCap(int(opts.WriteQueueSize)),

		channels:        asyncmap.NewAtomicMap[bin.Bin128, internalChannel](),
		closedListeners: make(map[int64]func()),
	}
	c.ctx = newConnContext(c)
	return c
}

// Context returns a connection context.
func (c *conn) Context() ConnContext {
	return c.ctx
}

// Close

// Close closes the connection and frees its internal resources.
func (c *conn) Close() status.Status {
	err := c.conn.Close()
	if err != nil {
		return mpxError(err)
	}
	return status.OK
}

// Closed returns a flag that is set when the connection is closed.
func (c *conn) Closed() async.Flag {
	return c.closed
}

// OnClosed adds a disconnection listener, and returns an unsubscribe function.
func (c *conn) OnClosed(fn func()) (unsub func()) {
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

// Channel

// Channel opens a new channel.
func (c *conn) Channel(ctx async.Context) (Channel, status.Status) {
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
		case <-c.handshaked.Wait():
		}
	}
}

// Internal

// Free closes and frees the connection.
func (c *conn) Free() {
	c.Close()
}

// internal

// send write an outgoing message to the write queue.
func (c *conn) send(ctx async.Context, msg pmpx.Message) status.Status {
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
func (c *conn) run() {
	defer c.free()
	defer c.close()

	// Negotiate protocol
	st := c.handshake()
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

func (c *conn) close() {
	if c.closed.IsSet() {
		return
	}

	defer c.notifyClosed()
	defer c.delegate.onConnClosed(c)
	defer c.closeChannels()

	c.ctx.Cancel()
	c.conn.Close()
	c.closed.Set()
	c.writeq.Close()
}

func (c *conn) free() {
	c.reader.free()
	c.writeq.Free()
}

// channels

func (c *conn) closeChannels() {
	if c.channelsClosed.Load() {
		return
	}
	c.channelsClosed.Store(true)

	c.channels.Range(func(_ bin.Bin128, ch internalChannel) bool {
		ch.free()
		return true
	})
}

func (c *conn) createChannel() (Channel, bool, status.Status) {
	// Check flags
	switch {
	case c.channelsClosed.Load():
		return nil, false, statusConnClosed
	case !c.handshaked.IsSet():
		return nil, false, status.OK
	}

	// Create channel
	id := bin.Random128()
	window := int32(c.options.ChannelWindowSize)
	ch := newChannel(c, c.client, id, window)

	// Free on error
	done := false
	defer func() {
		if !done {
			ch.Free()
			ch.free()
		}
	}()

	// Add channel
	c.channels.Set(id, ch)
	c.maybeChannelsReached()

	// Check again
	if c.channelsClosed.Load() {
		c.channels.Delete(id)
		return nil, false, statusConnClosed
	}

	done = true
	return ch, true, status.OK
}

func (c *conn) channelHandler(ch Channel) {
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

func (c *conn) maybeChannelsReached() {
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

func (c *conn) addClosed(fn func()) int64 {
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

func (c *conn) removeClosed(id int64) {
	c.closedMu.Lock()
	defer c.closedMu.Unlock()

	if c.closedFlag {
		return
	}

	delete(c.closedListeners, id)
}

func (c *conn) notifyClosed() {
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

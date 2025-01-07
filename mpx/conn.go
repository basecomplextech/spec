// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package mpx

import (
	"net"
	"sync/atomic"

	"github.com/basecomplextech/baselibrary/alloc/bytequeue"
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

	// OnClosed adds a disconnect listener, and returns an unsubscribe function,
	// or false if the connection is already closed.
	OnClosed(fn func()) (unsub func(), _ bool)

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

	dialer := newDialer(opts)
	return ConnectDialer(ctx, addr, dialer, logger, opts)
}

// ConnectDialer dials an address and returns a connection.
func ConnectDialer(ctx async.Context, addr string, dialer *net.Dialer, logger logging.Logger,
	opts Options) (Conn, status.Status) {

	opts = opts.clean()
	delegate := noopConnDelegate{}

	conn, st := newConnector(dialer, delegate, logger, opts).connect(ctx, addr)
	if !st.OK() {
		return nil, st
	}

	go func() {
		defer func() {
			if e := recover(); e != nil {
				st := status.Recover(e)
				logger.ErrorStatus("Connection panic", st)
			}
		}()

		st := conn.run()
		switch st.Code {
		case status.CodeOK,
			status.CodeCancelled,
			status.CodeClosed,
			status.CodeEnd:
		default:
			logger.ErrorStatus("Connection error", st)
		}
	}()
	return conn, status.OK
}

// internal

type internalConn interface {
	Conn

	// run runs the main connection loop.
	run() status.Status

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
	writeq bytequeue.Queue

	// channels
	channels        asyncmap.AtomicMap[bin.Bin128, internalChannel]
	channelsClosed  atomic.Bool
	channelsReached atomic.Bool // number of channels reached the target number, notify delegate once

	// closed listeners
	closedListeners   asyncmap.AtomicMap[int64, func()]
	closedListenerSeq atomic.Int64
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
		writeq: bytequeue.NewCap(int(opts.WriteQueueSize)),

		channels:        asyncmap.NewAtomicMap[bin.Bin128, internalChannel](),
		closedListeners: asyncmap.NewAtomicMap[int64, func()](),
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

// OnClosed adds a disconnect listener, and returns an unsubscribe function,
// or false if the connection is already closed.
func (c *conn) OnClosed(fn func()) (unsub func(), _ bool) {
	// Ensure fn is called only once
	called := &atomic.Bool{}
	fn1 := func() {
		ok := called.CompareAndSwap(false, true)
		if ok {
			fn()
		}
	}

	// Add listener
	id := c.addClosed(fn1)
	if id == 0 {
		return nil, false
	}

	// Return unsubscribe
	unsub = func() {
		c.removeClosed(id)
	}
	return unsub, true
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

// run runs the main connection loop.
func (c *conn) run() status.Status {
	defer c.free()
	defer c.close()

	// Handshake, negotiate
	st := c.handshake()
	if !st.OK() {
		return st
	}

	// Run loops
	recv := async.RunVoid(c.receiveLoop)
	send := async.RunVoid(c.sendLoop)
	defer async.StopWaitAll(recv, send)
	defer c.close()

	// Await exit
	select {
	case <-recv.Wait():
		return recv.Status()
	case <-send.Wait():
		return send.Status()
	}
}

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
	// Check if closed
	if c.closed.IsSet() {
		return 0
	}

	// Add listener
	id := c.closedListenerSeq.Add(1)
	c.closedListeners.Set(id, fn)

	// Check again if closed
	if c.closed.IsSet() {
		c.closedListeners.Delete(id)
		return 0
	}
	return id
}

func (c *conn) removeClosed(id int64) {
	c.closedListeners.Delete(id)
}

func (c *conn) notifyClosed() {
	c.closedListeners.Range(func(_ int64, fn func()) bool {
		fn()
		return true
	})
	c.closedListeners.Clear()
}

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

type Connection interface {
	// Close closes the connection and frees its internal resources.
	Close() status.Status

	// Closed returns a flag that is set when the connection is closed.
	Closed() async.Flag

	// Channel opens a new channel.
	Channel(ctx async.Context) (Channel, status.Status)

	// Internal

	// Free closes and frees the connection, allows to wrap the connection into ref.R[Connection].
	Free()
}

// internal

type internalConn interface {
	SendMessage(ctx async.Context, msg pmpx.Message) status.Status
}

var (
	_ Connection   = (*conn)(nil)
	_ internalConn = (*conn)(nil)
)

type conn struct {
	conn    net.Conn
	client  bool
	logger  logging.Logger
	options Options

	closed     async.MutFlag
	negotiated async.MutFlag

	channelMu     sync.Mutex
	channels      map[bin.Bin128]internalChannel
	channelClosed bool

	reader *reader
	writer *writer
	writeq alloc.MQueue
}

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
		case <-c.negotiated.Wait():
		}
	}
}

// Internal

// Free closes and frees the connection.
func (c *conn) Free() {
	c.Close()
}

// SendMessage write an outgoing message to the write queue.
func (c *conn) SendMessage(ctx async.Context, msg pmpx.Message) status.Status {
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
	st := c.negotiate()
	switch st.Code {
	case status.CodeOK:
	case status.CodeCancelled,
		status.CodeEnd,
		status.CodeClosed:
		return
	default:
		c.logger.ErrorStatus("Connection error", st)
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
		c.logger.ErrorStatus("Connection error", st)
	}
}

func (c *conn) close() {
	c.conn.Close()
	c.closed.Set()

	c.writeq.Close()
	c.closeChannels()
}

func (c *conn) free() {
	c.reader.free()
	c.writeq.Free()
}

// negotiate

func (c *conn) negotiate() status.Status {
	if c.client {
		return c.negotiateClient()
	} else {
		return c.negotiateServer()
	}
}

func (c *conn) negotiateClient() status.Status {
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
	return status.OK
}

func (c *conn) negotiateServer() status.Status {
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
	return status.OK
}

// receive

func (c *conn) receiveLoop(ctx async.Context) status.Status {
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
			c.receiveOpen(msg)
		case pmpx.Code_ChannelClose:
			c.receiveClose(msg)
		case pmpx.Code_ChannelEnd:
			c.receiveEnd(msg)
		case pmpx.Code_ChannelMessage:
			c.receiveMessage(msg)
		case pmpx.Code_ChannelWindow:
			c.receiveWindow(msg)

		default:
			return mpxErrorf("unexpected mpx message, code=%d", code)
		}
	}
}

func (c *conn) receiveOpen(msg pmpx.Message) status.Status {
	m := msg.Open()
	id := m.Id()

	c.channelMu.Lock()
	defer c.channelMu.Unlock()

	// Check not exist, impossible
	if _, ok := c.channels[id]; ok {
		return mpxErrorf("received open message for existing channel, channel=%v", id)
	}

	// Make channel
	ch := newChannel(c, id, c.client)
	c.channels[id] = ch

	// Free on error
	done := false
	defer func() {
		if !done {
			ch.ReceiveFree()
		}
	}()

	// Handle message
	st := ch.ReceiveMessage(msg)
	if !st.OK() {
		delete(c.channels, id)
		return st
	}

	// TODO: Start handler
	done = true
	return st
}

func (c *conn) receiveClose(msg pmpx.Message) status.Status {
	m := msg.Close()
	id := m.Id()

	c.channelMu.Lock()
	defer c.channelMu.Unlock()

	ch, ok := c.channels[id]
	if !ok {
		return status.OK
	}

	defer ch.ReceiveFree()
	delete(c.channels, id)

	return ch.ReceiveMessage(msg)
}

func (c *conn) receiveEnd(msg pmpx.Message) status.Status {
	m := msg.End_()
	id := m.Id()

	c.channelMu.Lock()
	defer c.channelMu.Unlock()

	ch, ok := c.channels[id]
	if !ok {
		return status.OK
	}

	return ch.ReceiveMessage(msg)
}

func (c *conn) receiveMessage(msg pmpx.Message) status.Status {
	m := msg.Message()
	id := m.Id()

	c.channelMu.Lock()
	defer c.channelMu.Unlock()

	ch, ok := c.channels[id]
	if !ok {
		return status.OK
	}

	return ch.ReceiveMessage(msg)
}

func (c *conn) receiveWindow(msg pmpx.Message) status.Status {
	m := msg.Window()
	id := m.Id()

	c.channelMu.Lock()
	defer c.channelMu.Unlock()

	ch, ok := c.channels[id]
	if !ok {
		return status.OK
	}

	return ch.ReceiveMessage(msg)
}

// send

func (c *conn) sendLoop(ctx async.Context) status.Status {
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

func (c *conn) sendMessage(b []byte) status.Status {
	msg, err := pmpx.NewMessageErr(b)
	if err != nil {
		return mpxError(err)
	}

	// Maybe delete and free channel
	code := msg.Code()
	switch code {
	case pmpx.Code_ChannelOpen:
		id := msg.Open().Id()
		close := msg.Open().Close()
		if close {
			c.removeChannel(id)
		}

	case pmpx.Code_ChannelClose:
		id := msg.Open().Id()
		c.removeChannel(id)
	}

	// Write message
	return c.writer.writeMessage(b)
}

// channels

func (c *conn) closeChannels() {
	c.channelMu.Lock()
	defer c.channelMu.Unlock()

	if c.channelClosed {
		return
	}
	c.channelClosed = true

	for _, ch := range c.channels {
		ch.ReceiveFree()
	}
	c.channels = nil
}

func (c *conn) createChannel() (Channel, bool, status.Status) {
	c.channelMu.Lock()
	defer c.channelMu.Unlock()

	switch {
	case c.channelClosed:
		return nil, false, statusConnClosed
	case !c.negotiated.Get():
		return nil, false, status.OK
	}

	id := bin.Random128()
	ch := newChannel(c, id, c.client)
	c.channels[id] = ch
	return ch, true, status.OK
}

func (c *conn) removeChannel(id bin.Bin128) {
	c.channelMu.Lock()
	defer c.channelMu.Unlock()

	if c.channelClosed {
		return
	}

	ch, ok := c.channels[id]
	if !ok {
		return
	}

	defer ch.ReceiveFree()
	delete(c.channels, id)
}
